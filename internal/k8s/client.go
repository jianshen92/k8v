package k8s

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

// Logger interface for logging (to avoid circular dependency)
type Logger interface {
	Printf(format string, v ...interface{})
}

// Client wraps the Kubernetes clientset and informer factory
type Client struct {
	Clientset       *kubernetes.Clientset
	InformerFactory informers.SharedInformerFactory
	logger          Logger
}

// NewClient creates a new Kubernetes client with informers
func NewClient() (*Client, error) {
	config, err := getKubeConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	// Create SharedInformerFactory with 30 second resync period
	informerFactory := informers.NewSharedInformerFactory(clientset, 30*time.Second)

	return &Client{
		Clientset:       clientset,
		InformerFactory: informerFactory,
	}, nil
}

// getKubeConfig returns a Kubernetes client config
// It tries in-cluster config first, then falls back to kubeconfig file
func getKubeConfig() (*rest.Config, error) {
	// Try in-cluster config first
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, nil
	}

	// Fall back to kubeconfig file
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build config from kubeconfig: %w", err)
	}

	return config, nil
}

// SetLogger sets the logger for the client
func (c *Client) SetLogger(logger Logger) {
	c.logger = logger
}

// Start starts all informers
func (c *Client) Start(stopCh <-chan struct{}) {
	c.InformerFactory.Start(stopCh)
}

// logf logs using the logger if available, otherwise falls back to fmt.Printf
func (c *Client) logf(format string, v ...interface{}) {
	if c.logger != nil {
		c.logger.Printf(format, v...)
	} else {
		fmt.Printf(format+"\n", v...)
	}
}

// WaitForCacheSync waits for all informer caches to sync
func (c *Client) WaitForCacheSync(stopCh <-chan struct{}) bool {
	syncStart := time.Now()
	syncTimes := make(map[string]time.Time)
	syncedInformers := make(map[string]bool)

	c.logf("Waiting for informer caches to sync...")

	// Get all registered informers
	informers := map[string]cache.InformerSynced{
		"Pods":        c.InformerFactory.Core().V1().Pods().Informer().HasSynced,
		"Deployments": c.InformerFactory.Apps().V1().Deployments().Informer().HasSynced,
		"ReplicaSets": c.InformerFactory.Apps().V1().ReplicaSets().Informer().HasSynced,
		"Services":    c.InformerFactory.Core().V1().Services().Informer().HasSynced,
		"Ingresses":   c.InformerFactory.Networking().V1().Ingresses().Informer().HasSynced,
		"ConfigMaps":  c.InformerFactory.Core().V1().ConfigMaps().Informer().HasSynced,
		"Secrets":     c.InformerFactory.Core().V1().Secrets().Informer().HasSynced,
	}

	// Poll each informer until all are synced
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	progressTicker := time.NewTicker(5 * time.Second)
	defer progressTicker.Stop()

	lastCheckTime := make(map[string]time.Time)

	for {
		select {
		case <-stopCh:
			c.logf("  ✗ Sync cancelled")
			return false

		case <-progressTicker.C:
			elapsed := time.Since(syncStart)
			synced := len(syncedInformers)
			total := len(informers)
			pending := []string{}
			for name := range informers {
				if !syncedInformers[name] {
					pending = append(pending, name)
				}
			}
			c.logf("  Progress: %d/%d informers synced (%v elapsed) - Pending: %v", synced, total, elapsed.Round(time.Second), pending)

		case <-ticker.C:
			allSynced := true
			for name, hasSynced := range informers {
				if !syncedInformers[name] {
					lastCheckTime[name] = time.Now()
					if hasSynced() {
						// This informer just synced!
						elapsedFromStart := time.Since(syncStart)
						syncTimes[name] = time.Now()
						syncedInformers[name] = true
						c.logf("  ✓ %s synced after %v", name, elapsedFromStart.Round(time.Millisecond))
					} else {
						allSynced = false
					}
				}
			}

			if allSynced {
				totalTime := time.Since(syncStart)
				c.logf("All informers synced successfully in %v", totalTime.Round(time.Millisecond))
				return true
			}
		}
	}
}
