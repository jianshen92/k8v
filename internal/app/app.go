package app

import (
	"fmt"
	"sync"

	"github.com/user/k8v/internal/k8s"
	"github.com/user/k8v/internal/server"
)

// Logger interface for logging
type Logger interface {
	Printf(format string, v ...interface{})
}

// SyncStatus represents the current sync state
type SyncStatus struct {
	Syncing bool   `json:"syncing"`
	Synced  bool   `json:"synced"`
	Error   string `json:"error,omitempty"`
	Context string `json:"context"`
}

// App manages the Kubernetes client, watcher, and server lifecycle
type App struct {
	logger  Logger
	hub     *server.Hub
	logHub  *server.LogHub
	context string

	mu         sync.RWMutex
	client     *k8s.Client
	cache      *k8s.ResourceCache
	watcher    *k8s.Watcher
	stopCh     chan struct{}
	isRunning  bool
	syncStatus SyncStatus
}

// NewApp creates a new app instance
func NewApp(logger Logger, hub *server.Hub, logHub *server.LogHub) *App {
	return &App{
		logger: logger,
		hub:    hub,
		logHub: logHub,
	}
}

// Start initializes and starts the Kubernetes client and watcher
// It returns immediately and syncs informers in the background
func (a *App) Start(context string) error {
	a.mu.Lock()

	if a.isRunning {
		a.mu.Unlock()
		return fmt.Errorf("app is already running")
	}

	a.logger.Printf("Connecting to Kubernetes cluster (context: %s)...", context)

	// Create Kubernetes client
	client, err := k8s.NewClientWithContext(context)
	if err != nil {
		a.mu.Unlock()
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}
	client.SetLogger(a.logger)
	a.logger.Printf("✓ Connected to Kubernetes cluster")

	// Create resource cache
	cache := k8s.NewResourceCache()
	a.logger.Printf("✓ Resource cache initialized")

	// Create watcher with event handler that broadcasts to hub
	watcher := k8s.NewWatcher(client, cache, a.hub.Broadcast)
	err = watcher.Start()
	if err != nil {
		a.mu.Unlock()
		return fmt.Errorf("failed to start watcher: %w", err)
	}
	a.logger.Printf("✓ Watcher initialized")

	// Start informers
	stopCh := make(chan struct{})
	client.Start(stopCh)
	a.logger.Printf("✓ Informers started")

	// Update app state
	a.client = client
	a.cache = cache
	a.watcher = watcher
	a.stopCh = stopCh
	a.context = context
	a.isRunning = true
	a.syncStatus = SyncStatus{
		Syncing: true,
		Synced:  false,
		Context: context,
	}

	a.mu.Unlock()

	// Broadcast syncing state immediately
	a.hub.BroadcastSyncStatus(k8s.SyncStatusEvent{
		Type:    k8s.EventSyncStatus,
		Syncing: true,
		Synced:  false,
		Context: context,
	})

	// Wait for informer caches to sync in background
	go func() {
		a.logger.Printf("Starting background sync for informer caches...")
		synced := client.WaitForCacheSync(stopCh)

		a.mu.Lock()
		defer a.mu.Unlock()

		if synced {
			a.syncStatus = SyncStatus{
				Syncing: false,
				Synced:  true,
				Context: context,
			}
			a.logger.Printf("✓ App synced successfully with context: %s", context)

			// Broadcast synced state
			a.hub.BroadcastSyncStatus(k8s.SyncStatusEvent{
				Type:    k8s.EventSyncStatus,
				Syncing: false,
				Synced:  true,
				Context: context,
			})
		} else {
			a.syncStatus = SyncStatus{
				Syncing: false,
				Synced:  false,
				Error:   "Failed to sync informer caches",
				Context: context,
			}
			a.logger.Printf("✗ App sync failed for context: %s", context)

			// Broadcast error state
			a.hub.BroadcastSyncStatus(k8s.SyncStatusEvent{
				Type:    k8s.EventSyncStatus,
				Syncing: false,
				Synced:  false,
				Error:   "Failed to sync informer caches",
				Context: context,
			})
		}
	}()

	a.logger.Printf("✓ App started with context: %s (syncing in background)", context)
	return nil
}

// Stop gracefully stops the app
func (a *App) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.isRunning {
		return
	}

	a.logger.Printf("Stopping app...")
	close(a.stopCh)
	a.isRunning = false
	a.logger.Printf("✓ App stopped")
}

// SwitchContext switches to a different Kubernetes context
func (a *App) SwitchContext(newContext string) error {
	a.logger.Printf("Switching context from '%s' to '%s'...", a.context, newContext)

	// Broadcast syncing state immediately (clients stay connected)
	a.hub.BroadcastSyncStatus(k8s.SyncStatusEvent{
		Type:    k8s.EventSyncStatus,
		Syncing: true,
		Synced:  false,
		Context: newContext,
	})

	// Disconnect all log clients (log connections are specific to pods)
	a.logHub.DisconnectAll()
	a.logger.Printf("✓ Log clients disconnected")

	// Stop current app
	a.Stop()
	a.logger.Printf("✓ Previous context stopped")

	// Start with new context (will broadcast sync updates automatically)
	if err := a.Start(newContext); err != nil {
		// Broadcast error state
		a.hub.BroadcastSyncStatus(k8s.SyncStatusEvent{
			Type:    k8s.EventSyncStatus,
			Syncing: false,
			Synced:  false,
			Error:   err.Error(),
			Context: newContext,
		})
		return fmt.Errorf("failed to start with new context: %w", err)
	}

	a.logger.Printf("✓ Context switched successfully to '%s'", newContext)
	return nil
}

// GetWatcher returns the current watcher
func (a *App) GetWatcher() *k8s.Watcher {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.watcher
}

// GetCurrentContext returns the current context name
func (a *App) GetCurrentContext() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.context
}

// GetSyncStatus returns the current sync status
func (a *App) GetSyncStatus() interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.syncStatus
}
