package k8s

import (
	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/user/k8v/internal/types"
)

// ExtractOwners extracts ownership relationships from OwnerReferences
func ExtractOwners(obj metav1.Object) []types.ResourceRef {
	refs := []types.ResourceRef{}
	for _, owner := range obj.GetOwnerReferences() {
		refs = append(refs, types.NewResourceRef(
			owner.Kind,
			obj.GetNamespace(),
			owner.Name,
		))
	}
	return refs
}

// FindReverseRelationships finds all resources that have a relationship pointing TO the target
// This is a generic function that works for all relationship types
func FindReverseRelationships(
	targetID string,
	forwardRelType types.RelationshipType,
	cache *ResourceCache,
) []types.ResourceRef {
	refs := []types.ResourceRef{}

	// Search all resources in cache
	for _, resource := range cache.List() {
		// Get the forward relationship field (e.g., OwnedBy, DependsOn)
		forwardRefs := resource.GetRelationship(forwardRelType)

		// Check if this resource has our target in its forward relationship
		for _, ref := range forwardRefs {
			if ref.ID == targetID {
				refs = append(refs, types.NewResourceRef(
					resource.Type,
					resource.Namespace,
					resource.Name,
				))
				break
			}
		}
	}

	return refs
}

// ExtractConfigMapDeps extracts ConfigMap dependencies from a Pod spec
func ExtractConfigMapDeps(pod *v1.Pod) []types.ResourceRef {
	refs := []types.ResourceRef{}
	seen := make(map[string]bool)

	// Volume mounts
	for _, volume := range pod.Spec.Volumes {
		if volume.ConfigMap != nil {
			id := types.BuildID("ConfigMap", pod.Namespace, volume.ConfigMap.Name)
			if !seen[id] {
				refs = append(refs, types.NewResourceRef("ConfigMap", pod.Namespace, volume.ConfigMap.Name))
				seen[id] = true
			}
		}
	}

	// Env from
	for _, container := range pod.Spec.Containers {
		for _, envFrom := range container.EnvFrom {
			if envFrom.ConfigMapRef != nil {
				id := types.BuildID("ConfigMap", pod.Namespace, envFrom.ConfigMapRef.Name)
				if !seen[id] {
					refs = append(refs, types.NewResourceRef("ConfigMap", pod.Namespace, envFrom.ConfigMapRef.Name))
					seen[id] = true
				}
			}
		}

		// Individual env vars
		for _, env := range container.Env {
			if env.ValueFrom != nil && env.ValueFrom.ConfigMapKeyRef != nil {
				id := types.BuildID("ConfigMap", pod.Namespace, env.ValueFrom.ConfigMapKeyRef.Name)
				if !seen[id] {
					refs = append(refs, types.NewResourceRef("ConfigMap", pod.Namespace, env.ValueFrom.ConfigMapKeyRef.Name))
					seen[id] = true
				}
			}
		}
	}

	return refs
}

// ExtractSecretDeps extracts Secret dependencies from a Pod spec
func ExtractSecretDeps(pod *v1.Pod) []types.ResourceRef {
	refs := []types.ResourceRef{}
	seen := make(map[string]bool)

	// Volume mounts
	for _, volume := range pod.Spec.Volumes {
		if volume.Secret != nil {
			id := types.BuildID("Secret", pod.Namespace, volume.Secret.SecretName)
			if !seen[id] {
				refs = append(refs, types.NewResourceRef("Secret", pod.Namespace, volume.Secret.SecretName))
				seen[id] = true
			}
		}
	}

	// Env from
	for _, container := range pod.Spec.Containers {
		for _, envFrom := range container.EnvFrom {
			if envFrom.SecretRef != nil {
				id := types.BuildID("Secret", pod.Namespace, envFrom.SecretRef.Name)
				if !seen[id] {
					refs = append(refs, types.NewResourceRef("Secret", pod.Namespace, envFrom.SecretRef.Name))
					seen[id] = true
				}
			}
		}

		// Individual env vars
		for _, env := range container.Env {
			if env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil {
				id := types.BuildID("Secret", pod.Namespace, env.ValueFrom.SecretKeyRef.Name)
				if !seen[id] {
					refs = append(refs, types.NewResourceRef("Secret", pod.Namespace, env.ValueFrom.SecretKeyRef.Name))
					seen[id] = true
				}
			}
		}
	}

	return refs
}

// FindExposedPods finds all Pods that match a Service's selector
func FindExposedPods(service *v1.Service, cache *ResourceCache) []types.ResourceRef {
	refs := []types.ResourceRef{}

	// Get all pods from cache
	pods := cache.ListByType("Pod")

	for _, resource := range pods {
		// Skip if different namespace
		if resource.Namespace != service.Namespace {
			continue
		}

		// Check if pod labels match service selector
		if LabelsMatch(resource.Labels, service.Spec.Selector) {
			refs = append(refs, types.NewResourceRef("Pod", resource.Namespace, resource.Name))
		}
	}

	return refs
}

// FindRoutedServices finds all Services that an Ingress routes to
func FindRoutedServices(ingress *netv1.Ingress) []types.ResourceRef {
	refs := []types.ResourceRef{}
	seen := make(map[string]bool)

	// Default backend
	if ingress.Spec.DefaultBackend != nil && ingress.Spec.DefaultBackend.Service != nil {
		id := types.BuildID("Service", ingress.Namespace, ingress.Spec.DefaultBackend.Service.Name)
		if !seen[id] {
			refs = append(refs, types.NewResourceRef("Service", ingress.Namespace, ingress.Spec.DefaultBackend.Service.Name))
			seen[id] = true
		}
	}

	// Rules
	for _, rule := range ingress.Spec.Rules {
		if rule.HTTP == nil {
			continue
		}
		for _, path := range rule.HTTP.Paths {
			if path.Backend.Service != nil {
				id := types.BuildID("Service", ingress.Namespace, path.Backend.Service.Name)
				if !seen[id] {
					refs = append(refs, types.NewResourceRef("Service", ingress.Namespace, path.Backend.Service.Name))
					seen[id] = true
				}
			}
		}
	}

	return refs
}

// LabelsMatch checks if a set of labels matches a selector
func LabelsMatch(labels map[string]string, selector map[string]string) bool {
	if len(selector) == 0 {
		return false
	}

	for key, value := range selector {
		if labels[key] != value {
			return false
		}
	}

	return true
}

// UpdateBidirectionalRelationships updates both sides of a relationship
// For example, when a Service exposes Pods, update both:
// - Service.Relationships.Exposes -> Pods
// - Pod.Relationships.ExposedBy -> Service
func UpdateBidirectionalRelationships(cache *ResourceCache, resource *types.Resource) {
	// Update reverse ownership relationships
	for _, ownerRef := range resource.Relationships.OwnedBy {
		if owner, ok := cache.Get(ownerRef.ID); ok {
			addToOwns(owner, resource)
			cache.Set(owner)
		}
	}

	// Update reverse dependency relationships
	for _, depRef := range resource.Relationships.DependsOn {
		if dep, ok := cache.Get(depRef.ID); ok {
			addToUsedBy(dep, resource)
			cache.Set(dep)
		}
	}

	// Update reverse network relationships
	for _, exposedRef := range resource.Relationships.Exposes {
		if exposed, ok := cache.Get(exposedRef.ID); ok {
			addToExposedBy(exposed, resource)
			cache.Set(exposed)
		}
	}

	// Update reverse routing relationships
	for _, routeRef := range resource.Relationships.RoutesTo {
		if routed, ok := cache.Get(routeRef.ID); ok {
			addToRoutedBy(routed, resource)
			cache.Set(routed)
		}
	}
}

// Helper functions to add relationships without duplicates

func addToOwns(resource *types.Resource, owned *types.Resource) {
	ref := types.NewResourceRef(owned.Type, owned.Namespace, owned.Name)
	if !containsRef(resource.Relationships.Owns, ref) {
		resource.Relationships.Owns = append(resource.Relationships.Owns, ref)
	}
}

func addToUsedBy(resource *types.Resource, user *types.Resource) {
	ref := types.NewResourceRef(user.Type, user.Namespace, user.Name)
	if !containsRef(resource.Relationships.UsedBy, ref) {
		resource.Relationships.UsedBy = append(resource.Relationships.UsedBy, ref)
	}
}

func addToExposedBy(resource *types.Resource, exposer *types.Resource) {
	ref := types.NewResourceRef(exposer.Type, exposer.Namespace, exposer.Name)
	if !containsRef(resource.Relationships.ExposedBy, ref) {
		resource.Relationships.ExposedBy = append(resource.Relationships.ExposedBy, ref)
	}
}

func addToRoutedBy(resource *types.Resource, router *types.Resource) {
	ref := types.NewResourceRef(router.Type, router.Namespace, router.Name)
	if !containsRef(resource.Relationships.RoutedBy, ref) {
		resource.Relationships.RoutedBy = append(resource.Relationships.RoutedBy, ref)
	}
}

func containsRef(refs []types.ResourceRef, ref types.ResourceRef) bool {
	for _, r := range refs {
		if r.ID == ref.ID {
			return true
		}
	}
	return false
}

// ExtractPodNodeScheduling extracts the Node a Pod is scheduled on
func ExtractPodNodeScheduling(pod *v1.Pod) []types.ResourceRef {
	if pod.Spec.NodeName == "" {
		return []types.ResourceRef{} // Pod not yet scheduled
	}
	return []types.ResourceRef{
		types.NewResourceRef("Node", "", pod.Spec.NodeName), // Nodes are cluster-scoped
	}
}
