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

	// Create Kubernetes client
	log.Println("Connecting to Kubernetes cluster...")
	client, err := k8s.NewClient()
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}
	log.Println("✓ Connected to Kubernetes cluster")

	// Create resource cache
	cache := k8s.NewResourceCache()
	log.Println("✓ Resource cache initialized")

	// Create hub for WebSocket broadcasting
	hub := server.NewHub()
	go hub.Run()

	// Create log hub for log streaming
	logHub := server.NewLogHub()
	go logHub.Run()

	// Create watcher with event handler that broadcasts to hub
	watcher := k8s.NewWatcher(client, cache, hub.Broadcast)
	err = watcher.Start()
	if err != nil {
		log.Fatalf("Failed to start watcher: %v", err)
	}
	log.Println("✓ Watcher initialized")

	// Start informers
	stopCh := make(chan struct{})
	client.Start(stopCh)
	log.Println("✓ Informers started")

	// Wait for informer caches to sync
	log.Println("Waiting for informer caches to sync...")
	if !client.WaitForCacheSync(stopCh) {
		log.Fatal("Failed to sync informer caches")
	}
	log.Println("✓ Informer caches synced")

	// Create and start HTTP server
	srv := server.NewServerWithHub(*port, watcher, hub, logHub)

	// Handle shutdown gracefully
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh

		log.Println("\nShutting down...")
		close(stopCh)
		os.Exit(0)
	}()

	// Start server (blocking)
	log.Printf("✓ Server starting on http://localhost:%d\n", *port)
	fmt.Printf("\n🚀 K8V is running! Open http://localhost:%d in your browser\n\n", *port)

	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
