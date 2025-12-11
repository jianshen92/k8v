package k8s

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
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
	Clientset              *kubernetes.Clientset
	InformerFactory        informers.SharedInformerFactory
	DynamicClient          dynamic.Interface
	DynamicInformerFactory dynamicinformer.DynamicSharedInformerFactory
	config                 *rest.Config
	logger                 Logger
}

// NewClient creates a new Kubernetes client with informers using the current context
func NewClient() (*Client, error) {
	return NewClientWithContext("")
}

// NewClientWithContext creates a new Kubernetes client with informers using a specific context
// If context is empty, uses the current context from kubeconfig
func NewClientWithContext(context string) (*Client, error) {
	config, err := getKubeConfigWithContext(context)
	if err != nil {
		return nil, fmt.Errorf("failed to get kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	// Create SharedInformerFactory with 30 second resync period
	informerFactory := informers.NewSharedInformerFactory(clientset, 30*time.Second)
	dynamicFactory := dynamicinformer.NewDynamicSharedInformerFactory(dynamicClient, 30*time.Second)

	return &Client{
		Clientset:              clientset,
		InformerFactory:        informerFactory,
		DynamicClient:          dynamicClient,
		DynamicInformerFactory: dynamicFactory,
		config:                 config,
	}, nil
}

// getKubeConfig returns a Kubernetes client config using the current context
// It tries in-cluster config first, then falls back to kubeconfig file
func getKubeConfig() (*rest.Config, error) {
	return getKubeConfigWithContext("")
}

// getKubeConfigWithContext returns a Kubernetes client config using a specific context
// If context is empty, uses the current context from kubeconfig
func getKubeConfigWithContext(context string) (*rest.Config, error) {
	// Try in-cluster config first (ignore context in this case)
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, nil
	}

	// Fall back to kubeconfig file
	kubeconfigPath := getKubeconfigPath()

	// Load the kubeconfig file
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.ExplicitPath = kubeconfigPath

	configOverrides := &clientcmd.ConfigOverrides{}
	if context != "" {
		configOverrides.CurrentContext = context
	}

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err = kubeConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to build config from kubeconfig: %w", err)
	}

	return config, nil
}

// getKubeconfigPath returns the path to the kubeconfig file
func getKubeconfigPath() string {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		home, _ := os.UserHomeDir()
		kubeconfig = filepath.Join(home, ".kube", "config")
	}
	return kubeconfig
}

// Context represents a Kubernetes context
type Context struct {
	Name      string `json:"name"`
	Cluster   string `json:"cluster"`
	Namespace string `json:"namespace"`
	Current   bool   `json:"current"`
}

// ListContexts returns all available contexts from kubeconfig
func ListContexts() ([]Context, error) {
	kubeconfigPath := getKubeconfigPath()

	// Load the kubeconfig file
	config, err := clientcmd.LoadFromFile(kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	contexts := make([]Context, 0, len(config.Contexts))
	for name, ctxInfo := range config.Contexts {
		contexts = append(contexts, Context{
			Name:      name,
			Cluster:   ctxInfo.Cluster,
			Namespace: ctxInfo.Namespace,
			Current:   name == config.CurrentContext,
		})
	}

	return contexts, nil
}

// GetCurrentContext returns the current context name
func GetCurrentContext() (string, error) {
	kubeconfigPath := getKubeconfigPath()

	// Load the kubeconfig file
	config, err := clientcmd.LoadFromFile(kubeconfigPath)
	if err != nil {
		return "", fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	return config.CurrentContext, nil
}

// SetLogger sets the logger for the client
func (c *Client) SetLogger(logger Logger) {
	c.logger = logger
}

// Start starts all informers
func (c *Client) Start(stopCh <-chan struct{}) {
	c.InformerFactory.Start(stopCh)
	if c.DynamicInformerFactory != nil {
		c.DynamicInformerFactory.Start(stopCh)
	}
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
		"Nodes":       c.InformerFactory.Core().V1().Nodes().Informer().HasSynced,
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
				goto dynamicSync
			}
		}
	}

dynamicSync:
	if c.DynamicInformerFactory != nil {
		c.logf("Waiting for dynamic informer caches to sync...")
		c.DynamicInformerFactory.WaitForCacheSync(stopCh)
	}

	return true
}
