package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/user/k8v/internal/app"
	"github.com/user/k8v/internal/k8s"
	"github.com/user/k8v/internal/server"
)

// Version is set at build time via -ldflags.
var Version = "dev"

func main() {
	// Parse flags
	port := flag.Int("port", 8080, "HTTP server port")
	versionFlag := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Println(Version)
		return
	}

	log.Println("Starting k8v - Kubernetes Visualizer")

	// Create logger for server
	logger, err := server.NewLogger()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Create hubs for WebSocket broadcasting
	hub := server.NewHub(logger)
	go hub.Run()

	logHub := server.NewLogHub(logger)
	go logHub.Run()

	execHub := server.NewExecHub(logger)
	go execHub.Run()

	// Create and start app with current context
	currentContext, err := k8s.GetCurrentContext()
	if err != nil {
		log.Fatalf("Failed to get current context: %v", err)
	}

	k8vApp := app.NewApp(logger, hub, logHub)
	if err := k8vApp.Start(currentContext); err != nil {
		log.Fatalf("Failed to start app: %v", err)
	}

	// Create and start HTTP server
	srv, err := server.NewServerWithProvider(*port, k8vApp, hub, logHub, execHub)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	defer srv.Close()

	// Handle shutdown gracefully
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh

		logger.Printf("\nShutting down...")
		k8vApp.Stop()
		srv.Close()
		os.Exit(0)
	}()

	// Start server (blocking)
	logger.Printf("✓ Server starting on http://localhost:%d", *port)
	fmt.Printf("\n🚀 K8V is running! Open http://localhost:%d in your browser\n\n", *port)

	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
