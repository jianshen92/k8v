package k8s

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client wraps the Kubernetes clientset and informer factory
type Client struct {
	Clientset       *kubernetes.Clientset
	InformerFactory informers.SharedInformerFactory
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

// Start starts all informers
func (c *Client) Start(stopCh <-chan struct{}) {
	c.InformerFactory.Start(stopCh)
}

// WaitForCacheSync waits for all informer caches to sync
func (c *Client) WaitForCacheSync(stopCh <-chan struct{}) bool {
	synced := c.InformerFactory.WaitForCacheSync(stopCh)
	for informerType, isSynced := range synced {
		if !isSynced {
			fmt.Printf("Warning: informer %s failed to sync\n", informerType)
			return false
		}
	}
	return true
}
