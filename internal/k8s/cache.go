package k8s

import (
	"sync"

	"github.com/user/k8v/internal/types"
)

// ResourceCache maintains an in-memory cache of all Kubernetes resources
// with thread-safe access for concurrent read/write operations
type ResourceCache struct {
	mu        sync.RWMutex
	resources map[string]*types.Resource // ID -> Resource
}

// NewResourceCache creates a new empty resource cache
func NewResourceCache() *ResourceCache {
	return &ResourceCache{
		resources: make(map[string]*types.Resource),
	}
}

// Get retrieves a resource by ID
func (c *ResourceCache) Get(id string) (*types.Resource, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	r, ok := c.resources[id]
	return r, ok
}

// Set stores or updates a resource in the cache
func (c *ResourceCache) Set(r *types.Resource) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.resources[r.ID] = r
}

// Delete removes a resource from the cache by ID
func (c *ResourceCache) Delete(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.resources, id)
}

// List returns all resources in the cache
func (c *ResourceCache) List() []*types.Resource {
	c.mu.RLock()
	defer c.mu.RUnlock()

	resources := make([]*types.Resource, 0, len(c.resources))
	for _, r := range c.resources {
		resources = append(resources, r)
	}
	return resources
}

// ListByType returns all resources of a specific type
func (c *ResourceCache) ListByType(resourceType string) []*types.Resource {
	c.mu.RLock()
	defer c.mu.RUnlock()

	resources := []*types.Resource{}
	for _, r := range c.resources {
		if r.Type == resourceType {
			resources = append(resources, r)
		}
	}
	return resources
}

// ListByNamespace returns all resources in a specific namespace
func (c *ResourceCache) ListByNamespace(namespace string) []*types.Resource {
	c.mu.RLock()
	defer c.mu.RUnlock()

	resources := []*types.Resource{}
	for _, r := range c.resources {
		if r.Namespace == namespace {
			resources = append(resources, r)
		}
	}
	return resources
}

// Count returns the total number of resources in the cache
func (c *ResourceCache) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.resources)
}
