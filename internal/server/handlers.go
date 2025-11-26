package server

import (
	"encoding/json"
	"io/fs"
	"net/http"
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
		"resources": s.watcher.GetResourceCount(),
	})
}

// handleNamespaces returns list of namespaces in the cluster
func (s *Server) handleNamespaces(w http.ResponseWriter, r *http.Request) {
	namespaces := s.watcher.GetNamespaces()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"namespaces": namespaces,
	})
}
