package server

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/user/k8v/internal/k8s"
)

// handleIndex serves the main HTML page
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	// Try to serve from embedded static files
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		http.Error(w, "Failed to load static files", http.StatusInternalServerError)
		return
	}

	// Serve index.html for root path
	if r.URL.Path == "/" {
		http.ServeFileFS(w, r, staticFS, "index.html")
		return
	}

	// Serve other static files
	http.FileServerFS(staticFS).ServeHTTP(w, r)
}

// handleHealth returns the health status of the server
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"clients":   len(s.hub.clients),
		"resources": s.watcherProvider.GetWatcher().GetResourceCount(),
		"context":   s.watcherProvider.GetCurrentContext(),
	})
}

// handleNamespaces returns list of namespaces in the cluster
func (s *Server) handleNamespaces(w http.ResponseWriter, r *http.Request) {
	namespaces := s.watcherProvider.GetWatcher().GetNamespaces()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"namespaces": namespaces,
	})
}

// handleStats returns resource counts by type
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")

	counts := s.watcherProvider.GetWatcher().GetResourceCounts(namespace)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(counts)
}

// handleContexts returns list of available Kubernetes contexts
func (s *Server) handleContexts(w http.ResponseWriter, r *http.Request) {
	contexts, err := k8s.ListContexts()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to list contexts: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"contexts": contexts,
	})
}

// handleCurrentContext returns the current Kubernetes context
func (s *Server) handleCurrentContext(w http.ResponseWriter, r *http.Request) {
	context := s.watcherProvider.GetCurrentContext()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"context": context,
	})
}

// handleSwitchContext switches to a different Kubernetes context
func (s *Server) handleSwitchContext(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	context := r.URL.Query().Get("context")
	if context == "" {
		http.Error(w, "context parameter is required", http.StatusBadRequest)
		return
	}

	s.logger.Printf("[API] Switching to context: %s", context)

	err := s.watcherProvider.SwitchContext(context)
	if err != nil {
		s.logger.Printf("[API] Context switch failed: %v", err)
		http.Error(w, fmt.Sprintf("failed to switch context: %v", err), http.StatusInternalServerError)
		return
	}

	s.logger.Printf("[API] Context switched successfully to: %s", context)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"context": context,
	})
}

// handleSyncStatus returns the current sync status
func (s *Server) handleSyncStatus(w http.ResponseWriter, r *http.Request) {
	status := s.watcherProvider.GetSyncStatus()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
