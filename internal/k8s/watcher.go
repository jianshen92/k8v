package k8s

import (
	"fmt"
	"log"
	"sort"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/user/k8v/internal/types"
)

// EventType represents the type of Kubernetes event
type EventType string

const (
	EventAdded    EventType = "ADDED"
	EventModified EventType = "MODIFIED"
	EventDeleted  EventType = "DELETED"
)

// ResourceEvent represents a resource change event
type ResourceEvent struct {
	Type     EventType       `json:"type"`
	Resource *types.Resource `json:"resource"`
}

// EventHandler is a callback function for resource events
type EventHandler func(event ResourceEvent)

// Watcher manages all Kubernetes resource watchers using Informers
type Watcher struct {
	client  *Client
	cache   *ResourceCache
	handler EventHandler
}

// NewWatcher creates a new watcher with the given client and cache
func NewWatcher(client *Client, resourceCache *ResourceCache, handler EventHandler) *Watcher {
	return &Watcher{
		client:  client,
		cache:   resourceCache,
		handler: handler,
	}
}

// Start registers all informer event handlers and starts watching
func (w *Watcher) Start() error {
	// Register Pod handlers
	podInformer := w.client.InformerFactory.Core().V1().Pods().Informer()
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    w.handlePodAdd,
		UpdateFunc: w.handlePodUpdate,
		DeleteFunc: w.handlePodDelete,
	})

	// Register Deployment handlers
	deploymentInformer := w.client.InformerFactory.Apps().V1().Deployments().Informer()
	deploymentInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    w.handleDeploymentAdd,
		UpdateFunc: w.handleDeploymentUpdate,
		DeleteFunc: w.handleDeploymentDelete,
	})

	// Register ReplicaSet handlers
	replicaSetInformer := w.client.InformerFactory.Apps().V1().ReplicaSets().Informer()
	replicaSetInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    w.handleReplicaSetAdd,
		UpdateFunc: w.handleReplicaSetUpdate,
		DeleteFunc: w.handleReplicaSetDelete,
	})

	// Register Service handlers
	serviceInformer := w.client.InformerFactory.Core().V1().Services().Informer()
	serviceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    w.handleServiceAdd,
		UpdateFunc: w.handleServiceUpdate,
		DeleteFunc: w.handleServiceDelete,
	})

	// Register Ingress handlers
	ingressInformer := w.client.InformerFactory.Networking().V1().Ingresses().Informer()
	ingressInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    w.handleIngressAdd,
		UpdateFunc: w.handleIngressUpdate,
		DeleteFunc: w.handleIngressDelete,
	})

	// Register ConfigMap handlers
	configMapInformer := w.client.InformerFactory.Core().V1().ConfigMaps().Informer()
	configMapInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    w.handleConfigMapAdd,
		UpdateFunc: w.handleConfigMapUpdate,
		DeleteFunc: w.handleConfigMapDelete,
	})

	// Register Secret handlers
	secretInformer := w.client.InformerFactory.Core().V1().Secrets().Informer()
	secretInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    w.handleSecretAdd,
		UpdateFunc: w.handleSecretUpdate,
		DeleteFunc: w.handleSecretDelete,
	})

	log.Println("All informer handlers registered")
	return nil
}

// Pod event handlers

func (w *Watcher) handlePodAdd(obj interface{}) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		return
	}

	resource := TransformPod(pod, w.cache)
	w.cache.Set(resource)
	UpdateBidirectionalRelationships(w.cache, resource)

	if w.handler != nil {
		w.handler(ResourceEvent{Type: EventAdded, Resource: resource})
	}
}

func (w *Watcher) handlePodUpdate(oldObj, newObj interface{}) {
	pod, ok := newObj.(*v1.Pod)
	if !ok {
		return
	}

	resource := TransformPod(pod, w.cache)
	w.cache.Set(resource)
	UpdateBidirectionalRelationships(w.cache, resource)

	if w.handler != nil {
		w.handler(ResourceEvent{Type: EventModified, Resource: resource})
	}
}

func (w *Watcher) handlePodDelete(obj interface{}) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		return
	}

	id := types.BuildID("Pod", pod.Namespace, pod.Name)
	resource, _ := w.cache.Get(id)
	w.cache.Delete(id)

	if w.handler != nil && resource != nil {
		w.handler(ResourceEvent{Type: EventDeleted, Resource: resource})
	}
}

// Deployment event handlers

func (w *Watcher) handleDeploymentAdd(obj interface{}) {
	deployment, ok := obj.(*appsv1.Deployment)
	if !ok {
		return
	}

	resource := TransformDeployment(deployment, w.cache)
	w.cache.Set(resource)
	UpdateBidirectionalRelationships(w.cache, resource)

	if w.handler != nil {
		w.handler(ResourceEvent{Type: EventAdded, Resource: resource})
	}
}

func (w *Watcher) handleDeploymentUpdate(oldObj, newObj interface{}) {
	deployment, ok := newObj.(*appsv1.Deployment)
	if !ok {
		return
	}

	resource := TransformDeployment(deployment, w.cache)
	w.cache.Set(resource)
	UpdateBidirectionalRelationships(w.cache, resource)

	if w.handler != nil {
		w.handler(ResourceEvent{Type: EventModified, Resource: resource})
	}
}

func (w *Watcher) handleDeploymentDelete(obj interface{}) {
	deployment, ok := obj.(*appsv1.Deployment)
	if !ok {
		return
	}

	id := types.BuildID("Deployment", deployment.Namespace, deployment.Name)
	resource, _ := w.cache.Get(id)
	w.cache.Delete(id)

	if w.handler != nil && resource != nil {
		w.handler(ResourceEvent{Type: EventDeleted, Resource: resource})
	}
}

// ReplicaSet event handlers

func (w *Watcher) handleReplicaSetAdd(obj interface{}) {
	rs, ok := obj.(*appsv1.ReplicaSet)
	if !ok {
		return
	}

	resource := TransformReplicaSet(rs, w.cache)
	w.cache.Set(resource)
	UpdateBidirectionalRelationships(w.cache, resource)

	if w.handler != nil {
		w.handler(ResourceEvent{Type: EventAdded, Resource: resource})
	}
}

func (w *Watcher) handleReplicaSetUpdate(oldObj, newObj interface{}) {
	rs, ok := newObj.(*appsv1.ReplicaSet)
	if !ok {
		return
	}

	resource := TransformReplicaSet(rs, w.cache)
	w.cache.Set(resource)
	UpdateBidirectionalRelationships(w.cache, resource)

	if w.handler != nil {
		w.handler(ResourceEvent{Type: EventModified, Resource: resource})
	}
}

func (w *Watcher) handleReplicaSetDelete(obj interface{}) {
	rs, ok := obj.(*appsv1.ReplicaSet)
	if !ok {
		return
	}

	id := types.BuildID("ReplicaSet", rs.Namespace, rs.Name)
	resource, _ := w.cache.Get(id)
	w.cache.Delete(id)

	if w.handler != nil && resource != nil {
		w.handler(ResourceEvent{Type: EventDeleted, Resource: resource})
	}
}

// Service event handlers

func (w *Watcher) handleServiceAdd(obj interface{}) {
	service, ok := obj.(*v1.Service)
	if !ok {
		return
	}

	resource := TransformService(service, w.cache)
	w.cache.Set(resource)
	UpdateBidirectionalRelationships(w.cache, resource)

	if w.handler != nil {
		w.handler(ResourceEvent{Type: EventAdded, Resource: resource})
	}
}

func (w *Watcher) handleServiceUpdate(oldObj, newObj interface{}) {
	service, ok := newObj.(*v1.Service)
	if !ok {
		return
	}

	resource := TransformService(service, w.cache)
	w.cache.Set(resource)
	UpdateBidirectionalRelationships(w.cache, resource)

	if w.handler != nil {
		w.handler(ResourceEvent{Type: EventModified, Resource: resource})
	}
}

func (w *Watcher) handleServiceDelete(obj interface{}) {
	service, ok := obj.(*v1.Service)
	if !ok {
		return
	}

	id := types.BuildID("Service", service.Namespace, service.Name)
	resource, _ := w.cache.Get(id)
	w.cache.Delete(id)

	if w.handler != nil && resource != nil {
		w.handler(ResourceEvent{Type: EventDeleted, Resource: resource})
	}
}

// Ingress event handlers

func (w *Watcher) handleIngressAdd(obj interface{}) {
	ingress, ok := obj.(*netv1.Ingress)
	if !ok {
		return
	}

	resource := TransformIngress(ingress, w.cache)
	w.cache.Set(resource)
	UpdateBidirectionalRelationships(w.cache, resource)

	if w.handler != nil {
		w.handler(ResourceEvent{Type: EventAdded, Resource: resource})
	}
}

func (w *Watcher) handleIngressUpdate(oldObj, newObj interface{}) {
	ingress, ok := newObj.(*netv1.Ingress)
	if !ok {
		return
	}

	resource := TransformIngress(ingress, w.cache)
	w.cache.Set(resource)
	UpdateBidirectionalRelationships(w.cache, resource)

	if w.handler != nil {
		w.handler(ResourceEvent{Type: EventModified, Resource: resource})
	}
}

func (w *Watcher) handleIngressDelete(obj interface{}) {
	ingress, ok := obj.(*netv1.Ingress)
	if !ok {
		return
	}

	id := types.BuildID("Ingress", ingress.Namespace, ingress.Name)
	resource, _ := w.cache.Get(id)
	w.cache.Delete(id)

	if w.handler != nil && resource != nil {
		w.handler(ResourceEvent{Type: EventDeleted, Resource: resource})
	}
}

// ConfigMap event handlers

func (w *Watcher) handleConfigMapAdd(obj interface{}) {
	cm, ok := obj.(*v1.ConfigMap)
	if !ok {
		return
	}

	resource := TransformConfigMap(cm, w.cache)
	w.cache.Set(resource)
	UpdateBidirectionalRelationships(w.cache, resource)

	if w.handler != nil {
		w.handler(ResourceEvent{Type: EventAdded, Resource: resource})
	}
}

func (w *Watcher) handleConfigMapUpdate(oldObj, newObj interface{}) {
	cm, ok := newObj.(*v1.ConfigMap)
	if !ok {
		return
	}

	resource := TransformConfigMap(cm, w.cache)
	w.cache.Set(resource)
	UpdateBidirectionalRelationships(w.cache, resource)

	if w.handler != nil {
		w.handler(ResourceEvent{Type: EventModified, Resource: resource})
	}
}

func (w *Watcher) handleConfigMapDelete(obj interface{}) {
	cm, ok := obj.(*v1.ConfigMap)
	if !ok {
		return
	}

	id := types.BuildID("ConfigMap", cm.Namespace, cm.Name)
	resource, _ := w.cache.Get(id)
	w.cache.Delete(id)

	if w.handler != nil && resource != nil {
		w.handler(ResourceEvent{Type: EventDeleted, Resource: resource})
	}
}

// Secret event handlers

func (w *Watcher) handleSecretAdd(obj interface{}) {
	secret, ok := obj.(*v1.Secret)
	if !ok {
		return
	}

	resource := TransformSecret(secret, w.cache)
	w.cache.Set(resource)
	UpdateBidirectionalRelationships(w.cache, resource)

	if w.handler != nil {
		w.handler(ResourceEvent{Type: EventAdded, Resource: resource})
	}
}

func (w *Watcher) handleSecretUpdate(oldObj, newObj interface{}) {
	secret, ok := newObj.(*v1.Secret)
	if !ok {
		return
	}

	resource := TransformSecret(secret, w.cache)
	w.cache.Set(resource)
	UpdateBidirectionalRelationships(w.cache, resource)

	if w.handler != nil {
		w.handler(ResourceEvent{Type: EventModified, Resource: resource})
	}
}

func (w *Watcher) handleSecretDelete(obj interface{}) {
	secret, ok := obj.(*v1.Secret)
	if !ok {
		return
	}

	id := types.BuildID("Secret", secret.Namespace, secret.Name)
	resource, _ := w.cache.Get(id)
	w.cache.Delete(id)

	if w.handler != nil && resource != nil {
		w.handler(ResourceEvent{Type: EventDeleted, Resource: resource})
	}
}

// GetSnapshot returns all current resources in the cache
func (w *Watcher) GetSnapshot() []ResourceEvent {
	resources := w.cache.List()
	events := make([]ResourceEvent, len(resources))

	for i, resource := range resources {
		events[i] = ResourceEvent{
			Type:     EventAdded,
			Resource: resource,
		}
	}

	fmt.Printf("Snapshot contains %d resources\n", len(events))
	return events
}

// GetNamespaces returns all unique namespaces from cached resources
func (w *Watcher) GetNamespaces() []string {
	nsMap := make(map[string]bool)
	resources := w.cache.List()
	for _, r := range resources {
		if r.Namespace != "" {
			nsMap[r.Namespace] = true
		}
	}

	namespaces := make([]string, 0, len(nsMap))
	for ns := range nsMap {
		namespaces = append(namespaces, ns)
	}
	sort.Strings(namespaces)
	return namespaces
}

// GetSnapshotFiltered returns resources filtered by namespace
func (w *Watcher) GetSnapshotFiltered(namespace string) []ResourceEvent {
	var resources []*types.Resource
	if namespace == "" {
		resources = w.cache.List()
	} else {
		resources = w.cache.ListByNamespace(namespace)
	}

	events := make([]ResourceEvent, len(resources))
	for i, resource := range resources {
		events[i] = ResourceEvent{
			Type:     EventAdded,
			Resource: resource,
		}
	}

	fmt.Printf("Filtered snapshot contains %d resources (namespace=%s)\n",
		len(events), namespace)
	return events
}

// GetResourceCount returns the number of resources in the cache
func (w *Watcher) GetResourceCount() int {
	return w.cache.Count()
}

// GetResourceCounts returns counts by resource type
func (w *Watcher) GetResourceCounts(namespace string) map[string]int {
	var resources []*types.Resource
	if namespace == "" || namespace == "all" {
		resources = w.cache.List()
	} else {
		resources = w.cache.ListByNamespace(namespace)
	}

	counts := make(map[string]int)
	for _, r := range resources {
		counts[r.Type]++
	}
	counts["total"] = len(resources)

	return counts
}

// GetSnapshotFilteredByType returns resources filtered by namespace and type
func (w *Watcher) GetSnapshotFilteredByType(namespace string, resourceType string) []ResourceEvent {
	var resources []*types.Resource
	if namespace == "" || namespace == "all" {
		resources = w.cache.List()
	} else {
		resources = w.cache.ListByNamespace(namespace)
	}

	// Filter by resource type
	filtered := []*types.Resource{}
	for _, r := range resources {
		if resourceType == "" || resourceType == "all" || r.Type == resourceType {
			filtered = append(filtered, r)
		}
	}

	events := make([]ResourceEvent, len(filtered))
	for i, resource := range filtered {
		events[i] = ResourceEvent{
			Type:     EventAdded,
			Resource: resource,
		}
	}

	fmt.Printf("Filtered snapshot by type contains %d resources (namespace=%s, type=%s)\n",
		len(events), namespace, resourceType)
	return events
}
