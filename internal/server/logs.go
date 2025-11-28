package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"

	"github.com/user/k8v/internal/k8s"
)

// LogClient represents a WebSocket client for log streaming
type LogClient struct {
	conn   *websocket.Conn
	send   chan k8s.LogMessage
	hub    *LogHub
	podKey string // "namespace/pod/container"
}

// LogHub manages all active log streaming WebSocket connections
type LogHub struct {
	clients    map[*LogClient]bool
	broadcast  chan k8s.LogMessage
	register   chan *LogClient
	unregister chan *LogClient
	mu         sync.RWMutex
}

// NewLogHub creates a new LogHub
func NewLogHub() *LogHub {
	return &LogHub{
		clients:    make(map[*LogClient]bool),
		broadcast:  make(chan k8s.LogMessage, 256),
		register:   make(chan *LogClient),
		unregister: make(chan *LogClient),
	}
}

// Run starts the log hub's main loop
func (h *LogHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("[LogHub] Client connected: %s (total: %d)\n", client.podKey, len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("[LogHub] Client disconnected: %s (total: %d)\n", client.podKey, len(h.clients))

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
					// Sent successfully
				default:
					// Client is slow, close it
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// handleLogsWebSocket handles WebSocket upgrade and log streaming
func (s *Server) handleLogsWebSocket(w http.ResponseWriter, r *http.Request) {
	// Parse required query parameters
	namespace := r.URL.Query().Get("namespace")
	pod := r.URL.Query().Get("pod")
	container := r.URL.Query().Get("container")

	if namespace == "" || pod == "" || container == "" {
		http.Error(w, "missing required parameters: namespace, pod, container", http.StatusBadRequest)
		return
	}

	// Upgrade connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	podKey := fmt.Sprintf("%s/%s/%s", namespace, pod, container)
	log.Printf("[LogStream] New connection: %s", podKey)

	// Create client
	client := &LogClient{
		conn:   conn,
		send:   make(chan k8s.LogMessage, 1000),
		hub:    s.logHub,
		podKey: podKey,
	}

	s.logHub.register <- client

	// Start log streaming in background
	// Use background context instead of r.Context() to avoid cancellation after WebSocket upgrade
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		err := s.watcher.StreamPodLogs(ctx, namespace, pod, container, s.logHub.broadcast)
		if err != nil {
			log.Printf("[LogStream] Streaming error for %s: %v", podKey, err)
			// Send error message to client
			s.logHub.broadcast <- k8s.LogMessage{
				Type:  "LOG_ERROR",
				Error: err.Error(),
			}
		}
		cancel()
	}()

	// Start pumps
	go client.writePump()
	go client.readPump(cancel) // Pass cancel to stop streaming on disconnect
}

// readPump pumps messages from the WebSocket connection
func (c *LogClient) readPump(cancel context.CancelFunc) {
	defer func() {
		cancel() // Stop log streaming
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break
		}
		// We don't expect messages from clients
	}
}

// writePump pumps messages to the WebSocket connection
func (c *LogClient) writePump() {
	defer c.conn.Close()

	for message := range c.send {
		if err := c.conn.WriteJSON(message); err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				log.Printf("[LogStream] Write error for %s: %v", c.podKey, err)
			}
			return
		}
	}
}
