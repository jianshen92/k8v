package server

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"

	"github.com/user/k8v/internal/k8s"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

// Client represents a WebSocket client connection
type Client struct {
	conn         *websocket.Conn
	send         chan k8s.ResourceEvent
	hub          *Hub
	namespace    string // namespace filter ("" = all namespaces)
	resourceType string // resource type filter ("" = all types)
}

// Hub manages all active WebSocket connections
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan k8s.ResourceEvent
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan k8s.ResourceEvent, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("Client connected (total: %d)\n", len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("Client disconnected (total: %d)\n", len(h.clients))

		case event := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				// Skip if client has namespace filter and resource doesn't match
				if client.namespace != "" && event.Resource.Namespace != client.namespace {
					continue
				}

				// Skip if client has resource type filter and resource doesn't match
				if client.resourceType != "" && event.Resource.Type != client.resourceType {
					continue
				}

				select {
				case client.send <- event:
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

// Broadcast sends an event to all connected clients
func (h *Hub) Broadcast(event k8s.ResourceEvent) {
	h.broadcast <- event
}

// handleWebSocket handles WebSocket upgrade and connection
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// Parse namespace filter from query params
	namespace := r.URL.Query().Get("namespace")
	if namespace == "" || namespace == "all" {
		namespace = "" // Empty string = all namespaces
	}

	// Parse resource type filter from query params
	resourceType := r.URL.Query().Get("type")
	if resourceType == "" || resourceType == "all" {
		resourceType = "" // Empty string = all types
	}

	log.Printf("New WebSocket connection with filters - namespace: '%s', type: '%s'", namespace, resourceType)

	client := &Client{
		conn:         conn,
		send:         make(chan k8s.ResourceEvent, 10000), // Large buffer for initial snapshot
		hub:          s.hub,
		namespace:    namespace,
		resourceType: resourceType,
	}

	s.hub.register <- client

	// Send initial snapshot of resources (filtered by namespace and type) synchronously before starting pumps
	snapshot := s.watcher.GetSnapshotFilteredByType(namespace, resourceType)
	log.Printf("Sending filtered snapshot of %d resources (namespace=%s, type=%s) to new client", len(snapshot), namespace, resourceType)

	// Log first few resources in snapshot for debugging
	if len(snapshot) > 0 && len(snapshot) <= 10 {
		log.Printf("Snapshot resources:")
		for _, event := range snapshot {
			log.Printf("  - %s/%s (type:%s)", event.Resource.Namespace, event.Resource.Name, event.Resource.Type)
		}
	}

	// Send snapshot directly without using the channel to avoid race condition
	batchSize := 1000
	for i, event := range snapshot {
		err := conn.WriteJSON(event)
		if err != nil {
			log.Printf("Failed to send snapshot event %d/%d: %v", i+1, len(snapshot), err)
			conn.Close()
			s.hub.unregister <- client
			return
		}
		// Log progress every batch
		if (i+1)%batchSize == 0 {
			log.Printf("Snapshot progress: %d/%d resources sent", i+1, len(snapshot))
		}
	}
	log.Printf("Snapshot sent successfully: %d resources", len(snapshot))

	// Start goroutines for read/write
	go client.writePump()
	go client.readPump()
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		// We don't expect messages from clients yet
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	defer c.conn.Close()

	for event := range c.send {
		err := c.conn.WriteJSON(event)
		if err != nil {
			// Don't log error if connection is closed, it's expected
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				return
			}
			log.Printf("Write error: %v", err)
			return
		}
	}
}
