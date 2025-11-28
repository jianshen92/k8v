package server

import (
	"embed"
	"fmt"
	"log"
	"net/http"

	"github.com/user/k8v/internal/k8s"
)

//go:embed static/*
var staticFiles embed.FS

// Server represents the HTTP server
type Server struct {
	port    int
	watcher *k8s.Watcher
	hub     *Hub
}

// NewServerWithHub creates a new HTTP server with an existing hub
func NewServerWithHub(port int, watcher *k8s.Watcher, hub *Hub) *Server {
	return &Server{
		port:    port,
		watcher: watcher,
		hub:     hub,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Set up HTTP routes
	http.HandleFunc("/", s.handleIndex)
	http.HandleFunc("/health", s.handleHealth)
	http.HandleFunc("/api/namespaces", s.handleNamespaces)
	http.HandleFunc("/api/stats", s.handleStats)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		s.handleWebSocket(w, r)
	})

	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("Starting server on http://localhost%s\n", addr)

	return http.ListenAndServe(addr, nil)
}
