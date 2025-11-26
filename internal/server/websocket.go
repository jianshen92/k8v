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
	conn *websocket.Conn
	send chan k8s.ResourceEvent
	hub  *Hub
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

	client := &Client{
		conn: conn,
		send: make(chan k8s.ResourceEvent, 10000), // Large buffer for initial snapshot
		hub:  s.hub,
	}

	s.hub.register <- client

	// Send initial snapshot of all resources
	go func() {
		snapshot := s.watcher.GetSnapshot()
		for _, event := range snapshot {
			// Try to send with recovery from panic if channel is closed
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("Recovered from panic while sending snapshot: %v", r)
					}
				}()
				client.send <- event
			}()
		}
	}()

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
			log.Printf("Write error: %v", err)
			return
		}
	}
}
