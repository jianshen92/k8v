package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/user/k8v/internal/k8s"
)

// NodeExecClient represents a WebSocket client for node exec streaming
type NodeExecClient struct {
	conn              *websocket.Conn
	send              chan k8s.ExecMessage
	done              chan struct{} // closed when client is shutting down
	hub               *NodeExecHub
	nodeName          string // target node name
	debugPodName      string // created debug pod name
	debugPodNamespace string // namespace where debug pod is created
	logger            *Logger
	cancelFunc        context.CancelFunc
	sizeQueue         *k8s.TerminalSizeQueue
	stdinPipe         io.WriteCloser
}

// NodeExecHub manages all active node exec WebSocket connections
type NodeExecHub struct {
	clients    map[*NodeExecClient]bool
	register   chan *NodeExecClient
	unregister chan *NodeExecClient
	mu         sync.RWMutex
	logger     *Logger
}

// NewNodeExecHub creates a new NodeExecHub
func NewNodeExecHub(logger *Logger) *NodeExecHub {
	return &NodeExecHub{
		clients:    make(map[*NodeExecClient]bool),
		register:   make(chan *NodeExecClient),
		unregister: make(chan *NodeExecClient),
		logger:     logger,
	}
}

// Run starts the node exec hub's main loop
func (h *NodeExecHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			h.logger.Printf("[NodeExecHub] Client connected: %s (total: %d)", client.nodeName, len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				// Close done first to signal shutdown to other goroutines
				close(client.done)
				if client.cancelFunc != nil {
					client.cancelFunc()
				}
				if client.sizeQueue != nil {
					client.sizeQueue.Close()
				}
				if client.stdinPipe != nil {
					client.stdinPipe.Close()
				}
				close(client.send)
			}
			h.mu.Unlock()
			h.logger.Printf("[NodeExecHub] Client disconnected: %s (total: %d)", client.nodeName, len(h.clients))
		}
	}
}

// DisconnectAll forcefully disconnects all node exec clients
func (h *NodeExecHub) DisconnectAll() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for client := range h.clients {
		// Close done first to signal shutdown
		close(client.done)
		if client.cancelFunc != nil {
			client.cancelFunc()
		}
		if client.sizeQueue != nil {
			client.sizeQueue.Close()
		}
		if client.stdinPipe != nil {
			client.stdinPipe.Close()
		}
		close(client.send)
		client.conn.Close()
		delete(h.clients, client)
	}
	h.logger.Printf("[NodeExecHub] All clients disconnected")
}

// handleNodeExecWebSocket handles WebSocket upgrade and node exec streaming
func (s *Server) handleNodeExecWebSocket(w http.ResponseWriter, r *http.Request) {
	// Parse required query parameters
	nodeName := r.URL.Query().Get("node")

	if nodeName == "" {
		http.Error(w, "missing required parameter: node", http.StatusBadRequest)
		return
	}

	// Upgrade connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Printf("[NodeExecStream] WebSocket upgrade failed: %v", err)
		return
	}

	s.logger.Printf("[NodeExecStream] New connection for node: %s", nodeName)

	// Create context for this exec session
	ctx, cancel := context.WithCancel(context.Background())

	// Create terminal size queue
	sizeQueue := k8s.NewTerminalSizeQueue()

	// Create pipes for stdin
	stdinReader, stdinWriter := io.Pipe()

	// Get debug pod options
	opts := k8s.DefaultNodeDebugPodOptions()

	// Create client
	client := &NodeExecClient{
		conn:              conn,
		send:              make(chan k8s.ExecMessage, 256),
		done:              make(chan struct{}),
		hub:               s.nodeExecHub,
		nodeName:          nodeName,
		debugPodNamespace: opts.Namespace,
		logger:            s.logger,
		cancelFunc:        cancel,
		sizeQueue:         sizeQueue,
		stdinPipe:         stdinWriter,
	}

	s.nodeExecHub.register <- client

	// Start debug pod lifecycle management
	go func() {
		defer cancel() // Always cancel context when this goroutine exits

		watcher := s.watcherProvider.GetWatcher()
		if watcher == nil {
			client.safeSend(k8s.ExecMessage{
				Type: k8s.ExecMessageError,
				Data: "watcher not available",
			})
			return
		}

		k8sClient := watcher.GetClient()

		// Send CREATING status
		if !client.safeSend(k8s.ExecMessage{
			Type: k8s.ExecMessageCreating,
			Data: fmt.Sprintf("Creating debug pod on node %s...", nodeName),
		}) {
			return // Client disconnected
		}

		// Create debug pod
		podName, err := k8sClient.CreateNodeDebugPod(ctx, nodeName, opts)
		if err != nil {
			client.safeSend(k8s.ExecMessage{
				Type: k8s.ExecMessageError,
				Data: fmt.Sprintf("failed to create debug pod: %v", err),
			})
			return
		}

		// Store pod name for cleanup
		client.debugPodName = podName

		// Ensure cleanup on exit
		defer func() {
			s.cleanupDebugPod(k8sClient, opts.Namespace, podName)
		}()

		// Send WAITING status
		if !client.safeSend(k8s.ExecMessage{
			Type: k8s.ExecMessageWaiting,
			Data: fmt.Sprintf("Waiting for debug pod %s to be ready...", podName),
		}) {
			return // Client disconnected
		}

		// Wait for pod to be ready
		err = k8sClient.WaitForPodReady(ctx, opts.Namespace, podName, opts.TimeoutSeconds)
		if err != nil {
			client.safeSend(k8s.ExecMessage{
				Type: k8s.ExecMessageError,
				Data: fmt.Sprintf("debug pod failed to start: %v", err),
			})
			return
		}

		// Notify client that we're connected
		if !client.safeSend(k8s.ExecMessage{
			Type: k8s.ExecMessageConnected,
			Data: "bash (node)",
		}) {
			return // Client disconnected
		}

		// Create stdout writer that sends to WebSocket
		stdoutWriter := &nodeExecOutputWriter{
			client:     client,
			outputType: k8s.ExecMessageOutput,
		}

		// Start exec session with chroot
		err = k8sClient.ExecNodeDebugShell(
			ctx,
			opts.Namespace,
			podName,
			stdinReader,
			stdoutWriter,
			stdoutWriter, // stderr goes to same output
			sizeQueue,
		)

		if err != nil {
			s.logger.Printf("[NodeExecStream] Exec error for node %s: %v", nodeName, err)
			client.safeSend(k8s.ExecMessage{
				Type: k8s.ExecMessageError,
				Data: err.Error(),
			})
		}

		// Send close message
		client.safeSend(k8s.ExecMessage{
			Type: k8s.ExecMessageClose,
			Data: "session ended",
		})
	}()

	// Start pumps
	go client.writePump()
	go client.readPump()
}

// cleanupDebugPod deletes the debug pod with a timeout
func (s *Server) cleanupDebugPod(k8sClient *k8s.Client, namespace, podName string) {
	// Use a separate context with timeout for cleanup
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := k8sClient.DeleteNodeDebugPod(ctx, namespace, podName)
	if err != nil {
		s.logger.Printf("[NodeExecStream] Failed to cleanup debug pod %s/%s: %v", namespace, podName, err)
	} else {
		s.logger.Printf("[NodeExecStream] Cleaned up debug pod %s/%s", namespace, podName)
	}
}

// nodeExecOutputWriter implements io.Writer and sends output to WebSocket
type nodeExecOutputWriter struct {
	client     *NodeExecClient
	outputType string
}

func (w *nodeExecOutputWriter) Write(p []byte) (n int, err error) {
	defer func() {
		if r := recover(); r != nil {
			// Channel was closed, that's okay
		}
	}()

	select {
	case <-w.client.done:
		// Client is shutting down
		return len(p), nil
	case w.client.send <- k8s.ExecMessage{
		Type: w.outputType,
		Data: string(p),
	}:
		return len(p), nil
	default:
		// Channel full, drop message
		return len(p), nil
	}
}

// safeSend sends a message to the client, returns false if client is shutting down
func (c *NodeExecClient) safeSend(msg k8s.ExecMessage) (sent bool) {
	defer func() {
		if r := recover(); r != nil {
			// Channel was closed, that's okay
			sent = false
		}
	}()

	select {
	case <-c.done:
		return false
	case c.send <- msg:
		return true
	}
}

// readPump pumps messages from the WebSocket connection
func (c *NodeExecClient) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				c.logger.Printf("[NodeExecStream] Read error for node %s: %v", c.nodeName, err)
			}
			break
		}

		// Parse the message
		var msg k8s.ExecMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			c.logger.Printf("[NodeExecStream] Invalid message for node %s: %v", c.nodeName, err)
			continue
		}

		switch msg.Type {
		case k8s.ExecMessageInput:
			// Write to stdin pipe
			if c.stdinPipe != nil {
				c.stdinPipe.Write([]byte(msg.Data))
			}

		case k8s.ExecMessageResize:
			// Send resize to terminal size queue
			if c.sizeQueue != nil {
				c.sizeQueue.Send(msg.Cols, msg.Rows)
			}
		}
	}
}

// writePump pumps messages to the WebSocket connection
func (c *NodeExecClient) writePump() {
	defer c.conn.Close()

	for message := range c.send {
		if err := c.conn.WriteJSON(message); err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				c.logger.Printf("[NodeExecStream] Write error for node %s: %v", c.nodeName, err)
			}
			return
		}
	}
}
