package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"

	"github.com/user/k8v/internal/k8s"
)

// ExecClient represents a WebSocket client for exec streaming
type ExecClient struct {
	conn       *websocket.Conn
	send       chan k8s.ExecMessage
	done       chan struct{} // closed when client is shutting down
	hub        *ExecHub
	podKey     string // "namespace/pod/container"
	logger     *Logger
	cancelFunc context.CancelFunc
	sizeQueue  *k8s.TerminalSizeQueue
	stdinPipe  io.WriteCloser
}

// ExecHub manages all active exec WebSocket connections
type ExecHub struct {
	clients    map[*ExecClient]bool
	register   chan *ExecClient
	unregister chan *ExecClient
	mu         sync.RWMutex
	logger     *Logger
}

// NewExecHub creates a new ExecHub
func NewExecHub(logger *Logger) *ExecHub {
	return &ExecHub{
		clients:    make(map[*ExecClient]bool),
		register:   make(chan *ExecClient),
		unregister: make(chan *ExecClient),
		logger:     logger,
	}
}

// Run starts the exec hub's main loop
func (h *ExecHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			h.logger.Printf("[ExecHub] Client connected: %s (total: %d)", client.podKey, len(h.clients))

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
			h.logger.Printf("[ExecHub] Client disconnected: %s (total: %d)", client.podKey, len(h.clients))
		}
	}
}

// DisconnectAll forcefully disconnects all exec clients
func (h *ExecHub) DisconnectAll() {
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
	h.logger.Printf("[ExecHub] All clients disconnected")
}

// handleExecWebSocket handles WebSocket upgrade and exec streaming
func (s *Server) handleExecWebSocket(w http.ResponseWriter, r *http.Request) {
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
		s.logger.Printf("[ExecStream] WebSocket upgrade failed: %v", err)
		return
	}

	podKey := fmt.Sprintf("%s/%s/%s", namespace, pod, container)
	s.logger.Printf("[ExecStream] New connection: %s", podKey)

	// Create context for this exec session
	ctx, cancel := context.WithCancel(context.Background())

	// Create terminal size queue
	sizeQueue := k8s.NewTerminalSizeQueue()

	// Create pipes for stdin
	stdinReader, stdinWriter := io.Pipe()

	// Create client
	client := &ExecClient{
		conn:       conn,
		send:       make(chan k8s.ExecMessage, 256),
		done:       make(chan struct{}),
		hub:        s.execHub,
		podKey:     podKey,
		logger:     s.logger,
		cancelFunc: cancel,
		sizeQueue:  sizeQueue,
		stdinPipe:  stdinWriter,
	}

	s.execHub.register <- client

	// Detect shell and start exec session
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

		// Detect available shell
		shell, err := k8sClient.DetectShell(ctx, namespace, pod, container)
		if err != nil {
			client.safeSend(k8s.ExecMessage{
				Type: k8s.ExecMessageError,
				Data: fmt.Sprintf("shell detection failed: %v", err),
			})
			return
		}

		// Notify client that we're connected
		if !client.safeSend(k8s.ExecMessage{
			Type: k8s.ExecMessageConnected,
			Data: shell[0],
		}) {
			return // Client disconnected
		}

		// Create stdout writer that sends to WebSocket
		stdoutWriter := &execOutputWriter{
			client:     client,
			outputType: k8s.ExecMessageOutput,
		}

		// Start exec session
		err = k8sClient.ExecPodShell(
			ctx,
			namespace,
			pod,
			container,
			shell,
			stdinReader,
			stdoutWriter,
			stdoutWriter, // stderr goes to same output
			sizeQueue,
		)

		if err != nil {
			s.logger.Printf("[ExecStream] Exec error for %s: %v", podKey, err)
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

// execOutputWriter implements io.Writer and sends output to WebSocket
type execOutputWriter struct {
	client     *ExecClient
	outputType string
}

func (w *execOutputWriter) Write(p []byte) (n int, err error) {
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
func (c *ExecClient) safeSend(msg k8s.ExecMessage) (sent bool) {
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
func (c *ExecClient) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				c.logger.Printf("[ExecStream] Read error for %s: %v", c.podKey, err)
			}
			break
		}

		// Parse the message
		var msg k8s.ExecMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			c.logger.Printf("[ExecStream] Invalid message for %s: %v", c.podKey, err)
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
func (c *ExecClient) writePump() {
	defer c.conn.Close()

	for message := range c.send {
		if err := c.conn.WriteJSON(message); err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				c.logger.Printf("[ExecStream] Write error for %s: %v", c.podKey, err)
			}
			return
		}
	}
}
