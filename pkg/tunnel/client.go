package tunnel

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/yamux"
)

// Client represents the tunnel client configuration and connection state.
type Client struct {
	ServerAddr         string
	LocalAddr          string
	RequestedSubdomain string
	InsecureSkipVerify bool
	UseTLS             bool

	// OnRequestLog callback is triggered for each HTTP request processed, allowing UI styling.
	OnRequestLog func(method, path, status string, latency time.Duration)

	// OnStart callback is triggered when the tunnel has been successfully authorized.
	OnStart func(subdomain string)

	mu      sync.Mutex
	session *yamux.Session
	wg      sync.WaitGroup
	closed  bool
}

// NewClient initializes and returns a new Client.
func NewClient(serverAddr, localAddr, requestedSubdomain string, useTLS, insecure bool) *Client {
	return &Client{
		ServerAddr:         serverAddr,
		LocalAddr:          localAddr,
		RequestedSubdomain: requestedSubdomain,
		UseTLS:             useTLS,
		InsecureSkipVerify: insecure,
	}
}

// Start dials the server daemon, runs the handshake, and handles multiplexed connections.
func (c *Client) Start() error {
	var conn net.Conn
	var err error

	if c.UseTLS {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: c.InsecureSkipVerify,
		}
		conn, err = tls.Dial("tcp", c.ServerAddr, tlsConfig)
	} else {
		conn, err = net.Dial("tcp", c.ServerAddr)
	}
	if err != nil {
		return fmt.Errorf("failed to connect to server %s: %w", c.ServerAddr, err)
	}

	// Enforce 10-second deadline for the initial control handshake
	_ = conn.SetDeadline(time.Now().Add(10 * time.Second))

	// 1. Send Handshake Request
	req := HandshakeRequest{
		RequestedSubdomain: c.RequestedSubdomain,
	}
	if err := WriteHandshake(conn, req); err != nil {
		_ = conn.Close()
		return fmt.Errorf("handshake request failed: %w", err)
	}

	// 2. Read Handshake Response
	var resp HandshakeResponse
	if err := ReadHandshake(conn, &resp); err != nil {
		_ = conn.Close()
		return fmt.Errorf("handshake response failed: %w", err)
	}

	if !resp.Success {
		_ = conn.Close()
		return fmt.Errorf("server rejected tunnel: %s", resp.Error)
	}

	// Reset deadlines before transferring connection control to Yamux
	_ = conn.SetDeadline(time.Time{})

	// 3. Initialize Yamux Client session with production timeout parameters
	cfg := yamux.DefaultConfig()
	cfg.KeepAliveInterval = 30 * time.Second
	cfg.ConnectionWriteTimeout = 15 * time.Second

	session, err := yamux.Client(conn, cfg)
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("failed to start Yamux client: %w", err)
	}

	c.mu.Lock()
	c.session = session
	c.mu.Unlock()

	if c.OnStart != nil {
		c.OnStart(resp.Subdomain)
	} else {
		log.Printf("[Client] Tunnel approved! Subdomain: %s", resp.Subdomain)
		log.Printf("[Client] Public URL: http://%s.%s", resp.Subdomain, c.getServerDomain())
		log.Printf("[Client] Forwarding to: %s", c.LocalAddr)
	}

	// 4. Accept virtual stream connection loop
	for {
		stream, err := session.Accept()
		if err != nil {
			c.mu.Lock()
			closed := c.closed
			c.mu.Unlock()
			if !closed {
				log.Printf("[Client] Tunnel session closed: %v", err)
			}
			break
		}

		c.wg.Add(1)
		go func(s net.Conn) {
			defer c.wg.Done()
			c.handleStream(s)
		}(stream)
	}

	return nil
}

// Close gracefully closes the Yamux session and waits for active connections to drain.
func (c *Client) Close() {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return
	}
	c.closed = true
	c.mu.Unlock()

	log.Println("[Client] Closing active tunnel session...")
	if c.session != nil {
		_ = c.session.Close()
	}

	log.Println("[Client] Draining active proxy streams...")
	c.wg.Wait()
	log.Println("[Client] Client shutdown complete.")
}

// getServerDomain returns the hostname/domain portion of the ServerAddr.
func (c *Client) getServerDomain() string {
	host, _, err := net.SplitHostPort(c.ServerAddr)
	if err != nil {
		host = c.ServerAddr
	}
	return host
}

// handleStream handles incoming streams: peeks at header to separate HTTP and raw TCP traffic.
func (c *Client) handleStream(stream net.Conn) {
	// Wrap stream in bufio.Reader to peek without consuming bytes
	br := bufio.NewReader(stream)
	peekBytes, err := br.Peek(8)
	if err != nil {
		c.rawCopy(stream, br)
		return
	}

	// Identify common HTTP methods
	isHTTP := false
	methods := []string{"GET ", "POST", "PUT ", "HEAD", "DELE", "OPTI", "PATC"}
	for _, m := range methods {
		if strings.HasPrefix(string(peekBytes), m) {
			isHTTP = true
			break
		}
	}

	if isHTTP {
		c.httpProxy(stream, br)
	} else {
		c.rawCopy(stream, br)
	}
}

// httpProxy parses incoming HTTP/1.1 requests/responses to support logging.
func (c *Client) httpProxy(stream net.Conn, br *bufio.Reader) {
	start := time.Now()

	req, err := http.ReadRequest(br)
	if err != nil {
		c.rawCopy(stream, br)
		return
	}
	defer req.Body.Close()

	localConn, err := net.Dial("tcp", c.LocalAddr)
	if err != nil {
		log.Printf("[Client] Failed to dial local service: %v", err)
		_ = stream.Close()
		return
	}
	defer localConn.Close()

	// Write HTTP request to local port
	if err := req.Write(localConn); err != nil {
		log.Printf("[Client] Failed to write HTTP request: %v", err)
		_ = stream.Close()
		return
	}

	// Read HTTP response from local port
	resp, err := http.ReadResponse(bufio.NewReader(localConn), req)
	if err != nil {
		log.Printf("[Client] Failed to read HTTP response: %v", err)
		_ = stream.Close()
		return
	}
	defer resp.Body.Close()

	// Write HTTP response back to the Yamux stream
	if err := resp.Write(stream); err != nil {
		log.Printf("[Client] Failed to write HTTP response: %v", err)
		_ = stream.Close()
		return
	}

	// Log HTTP traversal details
	latency := time.Since(start)
	if c.OnRequestLog != nil {
		c.OnRequestLog(req.Method, req.URL.Path, resp.Status, latency)
	} else {
		log.Printf("[Proxy] %s %s | STATUS: %s | LATENCY: %v", req.Method, req.URL.Path, resp.Status, latency)
	}
}

// rawCopy fallbacks to bidirectionally copying raw TCP traffic.
func (c *Client) rawCopy(stream net.Conn, r io.Reader) {
	defer stream.Close()
	localConn, err := net.Dial("tcp", c.LocalAddr)
	if err != nil {
		log.Printf("[Client] Failed to dial local service: %v", err)
		return
	}
	defer localConn.Close()

	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(localConn, r)
		close(done)
	}()

	_, _ = io.Copy(stream, localConn)
	<-done
}
