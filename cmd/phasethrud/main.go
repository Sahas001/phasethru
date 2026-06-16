package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"log"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Sahas001/phasethru/pkg/tunnel"
)

func generateInsecureTLSConfig() (*tls.Config, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{Organization: []string{"PhaseThru Local Dev"}},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, err
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}
	return &tls.Config{Certificates: []tls.Certificate{cert}}, nil
}

func main() {
	controlAddr := flag.String("control-addr", ":9000", "Control plane address to listen on")
	proxyAddr := flag.String("proxy-addr", ":8080", "Public HTTP proxy address to listen on")
	domain := flag.String("domain", "localhost", "Domain name suffix for subdomains (e.g. phasethru.dev)")
	tlsCert := flag.String("tls-cert", "", "Path to TLS certificate for control plane")
	tlsKey := flag.String("tls-key", "", "Path to TLS key for control plane")
	flag.Parse()

	log.Println("[Daemon] Initializing phasethrud...")

	server := tunnel.NewServer(*controlAddr, *proxyAddr, *domain, *tlsCert, *tlsKey)

	// Generate in-memory self-signed certificate for HTTPS proxy on port 8443
	tlsConfig, err := generateInsecureTLSConfig()
	if err != nil {
		log.Fatalf("[Daemon] Failed to generate in-memory TLS config: %v", err)
	}

	httpsServer := &http.Server{
		Addr:      *proxyAddr,
		Handler:   server,
		TLSConfig: tlsConfig,
	}

	// Listen for system signals to shutdown gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("[Daemon] Received shutdown signal. Starting graceful shutdown...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Shutdown HTTPS server
		if err := httpsServer.Shutdown(ctx); err != nil {
			log.Printf("[Daemon] HTTPS proxy graceful shutdown error: %v", err)
		}

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("[Daemon] Graceful shutdown error: %v", err)
		}
		os.Exit(0)
	}()

	// Start HTTPS server in a goroutine
	go func() {
		log.Println("[Daemon] HTTPS proxy listening on :8443 (in-memory self-signed cert)")
		if err := httpsServer.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			log.Printf("[Daemon] HTTPS proxy server runtime error: %v", err)
		}
	}()

	if err := server.Start(); err != nil {
		log.Fatalf("[Daemon] Server runtime error: %v", err)
	}
}
