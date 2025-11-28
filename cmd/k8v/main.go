package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/user/k8v/internal/k8s"
	"github.com/user/k8v/internal/server"
)

func main() {
	// Parse flags
	port := flag.Int("port", 8080, "HTTP server port")
	flag.Parse()

	log.Println("Starting k8v - Kubernetes Visualizer")

	// Create logger for server
	logger, err := server.NewLogger()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Create Kubernetes client
	logger.Printf("Connecting to Kubernetes cluster...")
	client, err := k8s.NewClient()
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}
	client.SetLogger(logger)
	logger.Printf("✓ Connected to Kubernetes cluster")

	// Create resource cache
	cache := k8s.NewResourceCache()
	logger.Printf("✓ Resource cache initialized")

	// Create hub for WebSocket broadcasting
	hub := server.NewHub(logger)
	go hub.Run()

	// Create log hub for log streaming
	logHub := server.NewLogHub(logger)
	go logHub.Run()

	// Create watcher with event handler that broadcasts to hub
	watcher := k8s.NewWatcher(client, cache, hub.Broadcast)
	err = watcher.Start()
	if err != nil {
		log.Fatalf("Failed to start watcher: %v", err)
	}
	logger.Printf("✓ Watcher initialized")

	// Start informers
	stopCh := make(chan struct{})
	client.Start(stopCh)
	logger.Printf("✓ Informers started")

	// Wait for informer caches to sync (logging is done inside WaitForCacheSync)
	if !client.WaitForCacheSync(stopCh) {
		log.Fatal("Failed to sync informer caches")
	}

	// Create and start HTTP server
	srv, err := server.NewServerWithHub(*port, watcher, hub, logHub)
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
		close(stopCh)
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
