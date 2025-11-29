package server

import (
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
	sendSync     chan k8s.SyncStatusEvent
	hub          *Hub
	namespace    string // namespace filter ("" = all namespaces)
	resourceType string // resource type filter ("" = all types)
	logger       *Logger
}

// Hub manages all active WebSocket connections
type Hub struct {
	clients           map[*Client]bool
	broadcast         chan k8s.ResourceEvent
	broadcastSync     chan k8s.SyncStatusEvent
	register          chan *Client
	unregister        chan *Client
	mu                sync.RWMutex
	logger            *Logger
	currentSyncStatus *k8s.SyncStatusEvent
	syncMu            sync.RWMutex
}

// NewHub creates a new Hub
func NewHub(logger *Logger) *Hub {
	return &Hub{
		clients:           make(map[*Client]bool),
		broadcast:         make(chan k8s.ResourceEvent, 256),
		broadcastSync:     make(chan k8s.SyncStatusEvent, 10),
		register:          make(chan *Client),
		unregister:        make(chan *Client),
		logger:            logger,
		currentSyncStatus: nil,
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
			h.logger.Printf("[WebSocket] Client connected (total: %d)", len(h.clients))

			// Send cached sync status to new client immediately
			h.syncMu.RLock()
			if h.currentSyncStatus != nil {
				select {
				case client.sendSync <- *h.currentSyncStatus:
				default:
					h.logger.Printf("[WebSocket] Failed to send sync status to new client")
				}
			}
			h.syncMu.RUnlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				close(client.sendSync)
			}
			h.mu.Unlock()
			h.logger.Printf("[WebSocket] Client disconnected (total: %d)", len(h.clients))

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

		case syncEvent := <-h.broadcastSync:
			// Cache the latest sync status
			h.syncMu.Lock()
			h.currentSyncStatus = &syncEvent
			h.syncMu.Unlock()

			// Broadcast to all clients
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.sendSync <- syncEvent:
				default:
					// Client is slow, close it
					h.logger.Printf("[WebSocket] Client slow during sync broadcast, closing")
					close(client.send)
					close(client.sendSync)
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

// BroadcastSyncStatus sends sync status update to all clients
func (h *Hub) BroadcastSyncStatus(event k8s.SyncStatusEvent) {
	h.broadcastSync <- event
}

// DisconnectAll forcefully disconnects all clients
func (h *Hub) DisconnectAll() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for client := range h.clients {
		close(client.send)
		client.conn.Close()
		delete(h.clients, client)
	}
	h.logger.Printf("[WebSocket] All clients disconnected")
}

// handleWebSocket handles WebSocket upgrade and connection
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Printf("[WebSocket] Upgrade failed: %v", err)
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

	s.logger.Printf("[WebSocket] New connection with filters - namespace: '%s', type: '%s'", namespace, resourceType)

	client := &Client{
		conn:         conn,
		send:         make(chan k8s.ResourceEvent, 10000), // Large buffer for initial snapshot
		sendSync:     make(chan k8s.SyncStatusEvent, 10),
		hub:          s.hub,
		namespace:    namespace,
		resourceType: resourceType,
		logger:       s.logger,
	}

	s.hub.register <- client

	// Send initial snapshot of resources (filtered by namespace and type) synchronously before starting pumps
	snapshot := s.watcherProvider.GetWatcher().GetSnapshotFilteredByType(namespace, resourceType)
	s.logger.Printf("[WebSocket] Sending filtered snapshot of %d resources (namespace=%s, type=%s) to new client", len(snapshot), namespace, resourceType)

	// Log first few resources in snapshot for debugging
	if len(snapshot) > 0 && len(snapshot) <= 10 {
		s.logger.Printf("[WebSocket] Snapshot resources:")
		for _, event := range snapshot {
			s.logger.Printf("[WebSocket]   - %s/%s (type:%s)", event.Resource.Namespace, event.Resource.Name, event.Resource.Type)
		}
	}

	// Send snapshot directly without using the channel to avoid race condition
	batchSize := 1000
	for i, event := range snapshot {
		err := conn.WriteJSON(event)
		if err != nil {
			s.logger.Printf("[WebSocket] Failed to send snapshot event %d/%d: %v", i+1, len(snapshot), err)
			conn.Close()
			s.hub.unregister <- client
			return
		}
		// Log progress every batch
		if (i+1)%batchSize == 0 {
			s.logger.Printf("[WebSocket] Snapshot progress: %d/%d resources sent", i+1, len(snapshot))
		}
	}
	s.logger.Printf("[WebSocket] Snapshot sent successfully: %d resources", len(snapshot))

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

	for {
		select {
		case event, ok := <-c.send:
			if !ok {
				return
			}
			err := c.conn.WriteJSON(event)
			if err != nil {
				// Don't log error if connection is closed, it's expected
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					return
				}
				c.logger.Printf("[WebSocket] Write error: %v", err)
				return
			}

		case syncEvent, ok := <-c.sendSync:
			if !ok {
				return
			}
			err := c.conn.WriteJSON(syncEvent)
			if err != nil {
				// Don't log error if connection is closed, it's expected
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					return
				}
				c.logger.Printf("[WebSocket] Write sync error: %v", err)
				return
			}
		}
	}
}
