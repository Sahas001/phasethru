package tunnel

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/yamux"
)

// Server handles accepting client control connections and proxying public HTTP requests.
type Server struct {
	ControlAddr string
	ProxyAddr   string
	Domain      string
	TLSCert     string
	TLSKey      string
	registry    *TunnelRegistry

	mu         sync.Mutex
	controlLn  net.Listener
	httpServer *http.Server
	closed     bool
}

// NewServer initializes and returns a new Server instance.
func NewServer(controlAddr, proxyAddr, domain, tlsCert, tlsKey string) *Server {
	return &Server{
		ControlAddr: controlAddr,
		ProxyAddr:   proxyAddr,
		Domain:      domain,
		TLSCert:     tlsCert,
		TLSKey:      tlsKey,
		registry:    NewTunnelRegistry(),
	}
}

// Start runs the Control listener and blocks on the HTTP proxy listener.
func (s *Server) Start() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return fmt.Errorf("server is already closed")
	}

	// Start Control Plane Listener
	var controlLn net.Listener
	var err error
	if s.TLSCert != "" && s.TLSKey != "" {
		cert, err := tls.LoadX509KeyPair(s.TLSCert, s.TLSKey)
		if err != nil {
			s.mu.Unlock()
			return fmt.Errorf("failed to load TLS key pair: %w", err)
		}
		tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}
		controlLn, err = tls.Listen("tcp", s.ControlAddr, tlsConfig)
	} else {
		controlLn, err = net.Listen("tcp", s.ControlAddr)
	}
	if err != nil {
		s.mu.Unlock()
		return fmt.Errorf("failed to start control listener: %w", err)
	}
	s.controlLn = controlLn
	s.mu.Unlock()

	log.Printf("[Control Server] Listening on %s (TLS: %t)", s.ControlAddr, s.TLSCert != "")

	// Accept client control connections in a background loop
	go func() {
		for {
			conn, err := controlLn.Accept()
			if err != nil {
				s.mu.Lock()
				closed := s.closed
				s.mu.Unlock()
				if closed {
					return // Clean exit on shutdown
				}
				log.Printf("[Control Server] Connection accept error: %v", err)
				return
			}
			go s.handleControlConnection(conn)
		}
	}()

	// Start Proxy HTTP Server (blocks)
	s.mu.Lock()
	s.httpServer = &http.Server{
		Addr:    s.ProxyAddr,
		Handler: s,
	}
	s.mu.Unlock()

	log.Printf("[Proxy Server] Listening on %s", s.ProxyAddr)
	err = s.httpServer.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

// Shutdown gracefully stops listeners and active client sessions.
func (s *Server) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	s.mu.Unlock()

	log.Println("[Server] Initiating graceful shutdown...")

	var err error
	if s.controlLn != nil {
		err = s.controlLn.Close()
	}

	// Gracefully close all Yamux sessions (notifies clients and clears registry)
	log.Println("[Server] Disconnecting all active client tunnels...")
	s.registry.CloseAll()

	// Shutdown HTTP proxy server
	if s.httpServer != nil {
		if httpErr := s.httpServer.Shutdown(ctx); httpErr != nil {
			err = httpErr
		}
	}

	log.Println("[Server] Shutdown complete.")
	return err
}

// handleControlConnection processes a client CLI connection: handshakes and starts the Yamux session.
func (s *Server) handleControlConnection(conn net.Conn) {
	defer conn.Close()

	// Enforce 10-second read deadline for incoming client handshake
	_ = conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	// Read Client Handshake Request
	var req HandshakeRequest
	if err := ReadHandshake(conn, &req); err != nil {
		log.Printf("[Control Server] Handshake read error from %s: %v", conn.RemoteAddr(), err)
		return
	}

	subdomain := req.RequestedSubdomain
	var err error
	if subdomain == "" {
		subdomain, err = s.generateRandomSubdomain()
		if err != nil {
			log.Printf("[Control Server] Subdomain generation failed: %v", err)
			_ = WriteHandshake(conn, HandshakeResponse{Success: false, Error: "failed to generate subdomain"})
			return
		}
	} else {
		subdomain = strings.ToLower(subdomain)
		if !s.isValidSubdomain(subdomain) {
			log.Printf("[Control Server] Invalid subdomain requested from %s: %s", conn.RemoteAddr(), subdomain)
			_ = WriteHandshake(conn, HandshakeResponse{Success: false, Error: "invalid subdomain: must be 3-63 chars, lowercase alphanumeric or dashes"})
			return
		}
	}

	// Attempt to reserve the subdomain in the registry
	if err := s.registry.Register(subdomain, nil); err != nil {
		log.Printf("[Control Server] Registration conflict for subdomain %s: %v", subdomain, err)
		_ = WriteHandshake(conn, HandshakeResponse{Success: false, Error: "subdomain is already taken"})
		return
	}
	defer s.registry.Unregister(subdomain)

	// Enforce 10-second write deadline for server response
	_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	// Send success handshake response to the client CLI
	err = WriteHandshake(conn, HandshakeResponse{
		Subdomain: subdomain,
		Success:   true,
	})
	if err != nil {
		log.Printf("[Control Server] Handshake write error to %s: %v", conn.RemoteAddr(), err)
		return
	}

	// Reset deadlines before transferring control of connection to Yamux multiplexer
	_ = conn.SetDeadline(time.Time{})

	// Initialize Yamux session (Server mode) with production keep-alive settings
	cfg := yamux.DefaultConfig()
	cfg.KeepAliveInterval = 30 * time.Second
	cfg.ConnectionWriteTimeout = 15 * time.Second

	session, err := yamux.Server(conn, cfg)
	if err != nil {
		log.Printf("[Control Server] Yamux server init failed for %s: %v", subdomain, err)
		return
	}
	defer session.Close()

	// Update the registry to point to the active session
	s.registry.Put(subdomain, session)
	log.Printf("[Control Server] Tunnel established: %s -> control connection from %s", subdomain, conn.RemoteAddr())

	// Wait until the session closes (client disconnects or connection drops)
	<-session.CloseChan()
	log.Printf("[Control Server] Tunnel terminated: %s", subdomain)
}

// ServeHTTP implements http.Handler to route requests to the appropriate Yamux session.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	subdomain := s.extractSubdomain(r.Host)
	if subdomain == "" {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head>
    <title>phasethru</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif; background-color: #0f172a; color: #f1f5f9; display: flex; justify-content: center; align-items: center; height: 100vh; margin: 0; }
        .card { background: #1e293b; padding: 2.5rem; border-radius: 12px; box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.3); text-align: center; max-width: 450px; border: 1px solid #334155; }
        h1 { color: #38bdf8; margin-top: 0; }
        p { color: #94a3b8; line-height: 1.6; }
        code { background: #0f172a; padding: 0.2rem 0.5rem; border-radius: 4px; color: #f43f5e; font-family: monospace; }
    </style>
</head>
<body>
    <div class="card">
        <h1>phasethru</h1>
        <p>Your lightweight, open-source reverse tunnel server is up and running.</p>
        <p>To expose a local port, run the client:</p>
        <p><code>phasethru 3000</code></p>
    </div>
</body>
</html>`)
		return
	}

	session, exists := s.registry.Get(subdomain)
	if !exists || session == nil {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>Tunnel Not Found</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif; background-color: #0f172a; color: #f1f5f9; display: flex; justify-content: center; align-items: center; height: 100vh; margin: 0; }
        .card { background: #1e293b; padding: 2.5rem; border-radius: 12px; box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.3); text-align: center; max-width: 450px; border: 1px solid #334155; }
        h1 { color: #f43f5e; margin-top: 0; }
        p { color: #94a3b8; line-height: 1.6; }
    </style>
</head>
<body>
    <div class="card">
        <h1>Tunnel Not Found or Offline</h1>
        <p>The tunnel for subdomain <strong>%s</strong> is not active or has disconnected.</p>
    </div>
</body>
</html>`, subdomain)
		return
	}

	// Reverse proxy through the Yamux session
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			// Set standard forwarding headers
			req.Header.Set("X-Forwarded-Host", req.Host)
			req.URL.Scheme = "http"
			req.URL.Host = req.Host
			req.Host = req.URL.Host // Adjust Host header to match upstream profile
			req.RequestURI = ""
		},
		Transport: &YamuxTransport{Session: session},
		ErrorLog:  log.New(io.Discard, "", 0),
		ErrorHandler: func(rw http.ResponseWriter, req *http.Request, err error) {
			log.Printf("[Proxy] Proxy error for %s: %v", subdomain, err)
			rw.WriteHeader(http.StatusBadGateway)
			fmt.Fprintf(rw, "Bad Gateway: failed to proxy request: %v", err)
		},
	}
	proxy.ServeHTTP(w, r)
}

// extractSubdomain parses hostnames and identifies the active subdomain prefix.
func (s *Server) extractSubdomain(host string) string {
	h := host
	if strings.Contains(h, ":") {
		var err error
		h, _, err = net.SplitHostPort(host)
		if err != nil {
			h = host
		}
	}

	domain := strings.ToLower(s.Domain)
	h = strings.ToLower(h)

	// Local development support (e.g. user.localhost -> user)
	if domain == "localhost" {
		if h == "localhost" || h == "127.0.0.1" {
			return ""
		}
		if strings.HasSuffix(h, ".localhost") {
			return strings.TrimSuffix(h, ".localhost")
		}
	}

	if h == domain {
		return ""
	}
	suffix := "." + domain
	if strings.HasSuffix(h, suffix) {
		return strings.TrimSuffix(h, suffix)
	}

	// Fallback/wildcard matching for arbitrary hosts
	parts := strings.Split(h, ".")
	if len(parts) > 1 {
		return parts[0]
	}

	return ""
}

// generateRandomSubdomain creates a cryptographically secure random string of 8 alphanumeric characters.
func (s *Server) generateRandomSubdomain() (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b), nil
}

// isValidSubdomain verifies the requested subdomain string format.
func (s *Server) isValidSubdomain(subdomain string) bool {
	if len(subdomain) < 3 || len(subdomain) > 63 {
		return false
	}
	for _, ch := range subdomain {
		if !((ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-') {
			return false
		}
	}
	return true
}

// YamuxTransport implements http.RoundTripper by routing requests into new Yamux streams.
type YamuxTransport struct {
	Session *yamux.Session
}

// RoundTrip opens a new stream on the Yamux session, writes the request, and parses the response.
func (t *YamuxTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	stream, err := t.Session.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open multiplexed stream: %w", err)
	}

	// Send HTTP/1.1 request wire format down the stream
	if err := req.Write(stream); err != nil {
		_ = stream.Close()
		return nil, fmt.Errorf("failed to write HTTP request to stream: %w", err)
	}

	// Read HTTP/1.1 response wire format from the stream
	resp, err := http.ReadResponse(bufio.NewReader(stream), req)
	if err != nil {
		_ = stream.Close()
		return nil, fmt.Errorf("failed to read HTTP response from stream: %w", err)
	}

	// Wrap response body to ensure stream is closed when response body is closed
	resp.Body = &streamCloser{
		ReadCloser: resp.Body,
		stream:     stream,
	}

	return resp, nil
}

// streamCloser wraps the HTTP response body and guarantees closure of both response body and Yamux stream.
type streamCloser struct {
	io.ReadCloser
	stream io.Closer
}

// Close closes the read closer, then closes the underlying multiplexed stream.
func (sc *streamCloser) Close() error {
	err1 := sc.ReadCloser.Close()
	err2 := sc.stream.Close()
	if err1 != nil {
		return err1
	}
	return err2
}
