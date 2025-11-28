package server

import (
	"embed"
	"fmt"
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
	logHub  *LogHub
	logger  *Logger
}

// NewServerWithHub creates a new HTTP server with an existing hub
func NewServerWithHub(port int, watcher *k8s.Watcher, hub *Hub, logHub *LogHub) (*Server, error) {
	logger, err := NewLogger()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	return &Server{
		port:    port,
		watcher: watcher,
		hub:     hub,
		logHub:  logHub,
		logger:  logger,
	}, nil
}

// Close gracefully shuts down the server
func (s *Server) Close() error {
	if s.logger != nil {
		return s.logger.Close()
	}
	return nil
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Set up HTTP routes with logging middleware
	http.HandleFunc("/", s.logger.LoggingMiddleware(s.handleIndex))
	http.HandleFunc("/health", s.logger.LoggingMiddleware(s.handleHealth))
	http.HandleFunc("/api/namespaces", s.logger.LoggingMiddleware(s.handleNamespaces))
	http.HandleFunc("/api/stats", s.logger.LoggingMiddleware(s.handleStats))
	http.HandleFunc("/ws", s.logger.LoggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		s.handleWebSocket(w, r)
	}))
	http.HandleFunc("/ws/logs", s.logger.LoggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		s.handleLogsWebSocket(w, r)
	}))

	addr := fmt.Sprintf(":%d", s.port)
	s.logger.Printf("Starting server on http://localhost%s", addr)

	return http.ListenAndServe(addr, nil)
}
