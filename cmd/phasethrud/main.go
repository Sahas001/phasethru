package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Sahas001/phasethru/pkg/tunnel"
)

func main() {
	controlAddr := flag.String("control-addr", ":9000", "Control plane address to listen on")
	proxyAddr := flag.String("proxy-addr", ":8080", "Public HTTP proxy address to listen on")
	domain := flag.String("domain", "localhost", "Domain name suffix for subdomains (e.g. phasethru.dev)")
	tlsCert := flag.String("tls-cert", "", "Path to TLS certificate for control plane")
	tlsKey := flag.String("tls-key", "", "Path to TLS key for control plane")
	flag.Parse()

	log.Println("[Daemon] Initializing phasethrud...")

	server := tunnel.NewServer(*controlAddr, *proxyAddr, *domain, *tlsCert, *tlsKey)

	// Listen for system signals to shutdown gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("[Daemon] Received shutdown signal. Starting graceful shutdown...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("[Daemon] Graceful shutdown error: %v", err)
		}
		os.Exit(0)
	}()

	if err := server.Start(); err != nil {
		log.Fatalf("[Daemon] Server runtime error: %v", err)
	}
}
