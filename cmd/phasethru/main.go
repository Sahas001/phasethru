package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Sahas001/phasethru/pkg/tunnel"
	"github.com/charmbracelet/lipgloss"
)

func main() {
	serverAddr := flag.String("server-addr", "localhost:9000", "Remote server control address")
	subdomain := flag.String("requested-subdomain", "", "Request a specific subdomain (optional)")
	useTLS := flag.Bool("tls", false, "Use TLS to secure control plane connection")
	insecure := flag.Bool("insecure", false, "Skip TLS verification (for development/self-signed certs)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <local-port>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s 3000\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	localPortOrAddr := args[0]
	localAddr := localPortOrAddr
	if !strings.Contains(localAddr, ":") {
		localAddr = "127.0.0.1:" + localAddr
	}

	client := tunnel.NewClient(*serverAddr, localAddr, *subdomain, *useTLS, *insecure)

	// Configure callbacks
	client.OnStart = func(sub string) {
		// Render beautiful dashboard using lipgloss
		statusStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#4ade80"))
		labelStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#94a3b8")).Width(14)
		valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#f1f5f9"))
		titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#38bdf8")).MarginBottom(1)

		serverHost, _, err := net.SplitHostPort(*serverAddr)
		if err != nil {
			serverHost = *serverAddr
		}
		publicURL := fmt.Sprintf("http://%s.%s", sub, serverHost)

		content := lipgloss.JoinVertical(
			lipgloss.Left,
			titleStyle.Render("⚡ phasethru tunnel established"),
			fmt.Sprintf("%s%s", labelStyle.Render("Status:"), statusStyle.Render("ONLINE")),
			fmt.Sprintf("%s%s", labelStyle.Render("Forwarding:"), valueStyle.Render(localAddr)),
			fmt.Sprintf("%s%s", labelStyle.Render("Public URL:"), lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#38bdf8")).Render(publicURL)),
		)

		box := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#334155")).
			Padding(1, 2).
			MarginBottom(1).
			Render(content)

		fmt.Println(box)
		fmt.Println("Real-time proxy traffic logs:")
	}

	client.OnRequestLog = func(method, path, status string, latency time.Duration) {
		ts := time.Now().Format("15:04:05")
		tsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#64748b")).Render(fmt.Sprintf("[%s]", ts))

		// Color code method
		methodColor := "#38bdf8"
		switch method {
		case "GET":
			methodColor = "#06b6d4"
		case "POST":
			methodColor = "#10b981"
		case "PUT", "PATCH":
			methodColor = "#f59e0b"
		case "DELETE":
			methodColor = "#ef4444"
		}
		methodStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(methodColor)).Width(6).Render(method)

		pathStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#e2e8f0")).Render(path)

		// Color code status
		statusColor := "#10b981"
		if strings.HasPrefix(status, "4") || strings.HasPrefix(status, "5") {
			statusColor = "#ef4444"
		} else if strings.HasPrefix(status, "3") {
			statusColor = "#eab308"
		}
		statusStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(statusColor)).Render(status)

		// Color code latency
		latencyColor := "#8b5cf6"
		if latency > 500*time.Millisecond {
			latencyColor = "#ef4444"
		} else if latency > 200*time.Millisecond {
			latencyColor = "#f59e0b"
		}
		latencyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(latencyColor)).Render(fmt.Sprintf("%dms", latency.Milliseconds()))

		fmt.Printf("%s %s %s | STATUS: %s | LATENCY: %s\n", tsStyle, methodStyle, pathStyle, statusStyle, latencyStyle)
	}

	// Graceful shutdown channel
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n[CLI] Disconnecting and shutting down tunnel gracefully...")
		client.Close()
		os.Exit(0)
	}()

	log.Printf("[CLI] Dialing phasethrud at %s...", *serverAddr)
	if err := client.Start(); err != nil {
		log.Fatalf("[CLI] Tunnel execution error: %v", err)
	}
}
