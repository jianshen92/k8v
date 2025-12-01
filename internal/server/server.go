package server

import (
	"embed"
	"fmt"
	"net/http"

	"github.com/user/k8v/internal/k8s"
)

//go:embed static/*
var staticFiles embed.FS

// WatcherProvider provides access to the current watcher
type WatcherProvider interface {
	GetWatcher() *k8s.Watcher
	GetCurrentContext() string
	SwitchContext(context string) error
	GetSyncStatus() interface{} // Returns app.SyncStatus or compatible struct
}

// Server represents the HTTP server
type Server struct {
	port            int
	watcherProvider WatcherProvider
	hub             *Hub
	logHub          *LogHub
	logger          *Logger
}

// For backward compatibility - direct watcher wrapper
type directWatcherProvider struct {
	watcher *k8s.Watcher
}

func (d *directWatcherProvider) GetWatcher() *k8s.Watcher {
	return d.watcher
}

func (d *directWatcherProvider) GetCurrentContext() string {
	return "unknown"
}

func (d *directWatcherProvider) SwitchContext(context string) error {
	return fmt.Errorf("context switching not supported with direct watcher")
}

func (d *directWatcherProvider) GetSyncStatus() interface{} {
	// Direct watcher is always synced
	return map[string]interface{}{
		"syncing": false,
		"synced":  true,
		"context": "unknown",
	}
}

// NewServerWithHub creates a new HTTP server with an existing hub (backward compatibility)
func NewServerWithHub(port int, watcher *k8s.Watcher, hub *Hub, logHub *LogHub) (*Server, error) {
	return NewServerWithProvider(port, &directWatcherProvider{watcher: watcher}, hub, logHub)
}

// NewServerWithProvider creates a new HTTP server with a watcher provider
func NewServerWithProvider(port int, provider WatcherProvider, hub *Hub, logHub *LogHub) (*Server, error) {
	logger, err := NewLogger()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	return &Server{
		port:            port,
		watcherProvider: provider,
		hub:             hub,
		logHub:          logHub,
		logger:          logger,
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
	http.HandleFunc("/api/contexts", s.logger.LoggingMiddleware(s.handleContexts))
	http.HandleFunc("/api/context/current", s.logger.LoggingMiddleware(s.handleCurrentContext))
	http.HandleFunc("/api/context/switch", s.logger.LoggingMiddleware(s.handleSwitchContext))
	http.HandleFunc("/api/sync/status", s.logger.LoggingMiddleware(s.handleSyncStatus))
	http.HandleFunc("/api/resource", s.logger.LoggingMiddleware(s.handleGetResource))
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
