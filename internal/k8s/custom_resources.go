package k8s

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/yaml"

	"github.com/user/k8v/internal/types"
)

type customResourceInfo struct {
	GVR        schema.GroupVersionResource
	Kind       string
	Namespaced bool
	TypeName   string
}

// discoverCustomResources lists CRDs and returns the served GVRs we should watch.
func (w *Watcher) discoverCustomResources(ctx context.Context) ([]customResourceInfo, error) {
	if w.client.DynamicClient == nil {
		return nil, fmt.Errorf("dynamic client not initialized")
	}

	crdGVR := schema.GroupVersionResource{
		Group:    "apiextensions.k8s.io",
		Version:  "v1",
		Resource: "customresourcedefinitions",
	}

	list, err := w.client.DynamicClient.Resource(crdGVR).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list CRDs: %w", err)
	}

	infos := make([]customResourceInfo, 0, len(list.Items))

	for _, item := range list.Items {
		group, found, _ := unstructured.NestedString(item.Object, "spec", "group")
		if !found || group == "" {
			continue
		}

		kind, found, _ := unstructured.NestedString(item.Object, "spec", "names", "kind")
		if !found || kind == "" {
			continue
		}

		plural, found, _ := unstructured.NestedString(item.Object, "spec", "names", "plural")
		if !found || plural == "" {
			continue
		}

		scope, _, _ := unstructured.NestedString(item.Object, "spec", "scope")
		namespaced := strings.EqualFold(scope, "Namespaced")

		versions, found, _ := unstructured.NestedSlice(item.Object, "spec", "versions")
		if !found || len(versions) == 0 {
			continue
		}

		var version string
		for _, v := range versions {
			if mv, ok := v.(map[string]interface{}); ok {
				served, _, _ := unstructured.NestedBool(mv, "served")
				if served {
					vName, _, _ := unstructured.NestedString(mv, "name")
					if vName != "" {
						version = vName
						break
					}
				}
			}
		}
		if version == "" {
			continue
		}

		typeName := kind
		if group != "" {
			typeName = fmt.Sprintf("%s.%s", kind, group)
		}

		infos = append(infos, customResourceInfo{
			GVR: schema.GroupVersionResource{
				Group:    group,
				Version:  version,
				Resource: plural,
			},
			Kind:       kind,
			Namespaced: namespaced,
			TypeName:   typeName,
		})
	}

	return infos, nil
}

func (w *Watcher) registerCustomResourceInformers(ctx context.Context) {
	crInfos, err := w.discoverCustomResources(ctx)
	if err != nil {
		w.client.logf("Failed to discover CRDs: %v", err)
		return
	}

	if len(crInfos) == 0 {
		w.client.logf("No custom resources discovered")
		return
	}

	for _, info := range crInfos {
		informer := w.client.DynamicInformerFactory.ForResource(info.GVR).Informer()

		infoCopy := info
		informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				u, ok := obj.(*unstructured.Unstructured)
				if !ok {
					return
				}
				resource := TransformCustomResource(u, infoCopy)
				w.cache.Set(resource)
				UpdateBidirectionalRelationships(w.cache, resource)
				if w.handler != nil {
					w.handler(ResourceEvent{Type: EventAdded, Resource: resource})
				}
			},
			UpdateFunc: func(_, newObj interface{}) {
				u, ok := newObj.(*unstructured.Unstructured)
				if !ok {
					return
				}
				resource := TransformCustomResource(u, infoCopy)
				w.cache.Set(resource)
				UpdateBidirectionalRelationships(w.cache, resource)
				if w.handler != nil {
					w.handler(ResourceEvent{Type: EventModified, Resource: resource})
				}
			},
			DeleteFunc: func(obj interface{}) {
				u, ok := obj.(*unstructured.Unstructured)
				if !ok {
					tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
					if !ok {
						return
					}
					u, _ = tombstone.Obj.(*unstructured.Unstructured)
					if u == nil {
						return
					}
				}

				id := types.BuildID(infoCopy.TypeName, u.GetNamespace(), u.GetName())
				resource, _ := w.cache.Get(id)
				w.cache.Delete(id)

				if w.handler != nil && resource != nil {
					w.handler(ResourceEvent{Type: EventDeleted, Resource: resource})
				}
			},
		})
	}

	w.client.logf("Registered custom resource informers for %d CRDs", len(crInfos))
}

// TransformCustomResource converts an unstructured object into our Resource model.
func TransformCustomResource(obj *unstructured.Unstructured, info customResourceInfo) *types.Resource {
	namespace := obj.GetNamespace()
	name := obj.GetName()

	phase, _, _ := unstructured.NestedString(obj.Object, "status", "phase")
	ready := extractReadyCondition(obj)

	yamlData, err := yaml.Marshal(obj.Object)
	if err != nil {
		yamlData = []byte{}
	}

	return &types.Resource{
		ID:        types.BuildID(info.TypeName, namespace, name),
		Type:      info.TypeName,
		Name:      name,
		Namespace: namespace,
		Status: types.ResourceStatus{
			Phase:   phase,
			Ready:   ready,
			Message: "",
		},
		Health: types.HealthUnknown,
		Relationships: types.Relationships{
			OwnedBy:     ExtractOwners(obj),
			Owns:        nil,
			DependsOn:   nil,
			UsedBy:      nil,
			Exposes:     nil,
			ExposedBy:   nil,
			RoutesTo:    nil,
			RoutedBy:    nil,
			ScheduledOn: nil,
			Schedules:   nil,
		},
		Labels:      obj.GetLabels(),
		Annotations: obj.GetAnnotations(),
		CreatedAt:   obj.GetCreationTimestamp().Time,
		Spec:        obj.Object,
		YAML:        string(yamlData),
	}
}

func extractReadyCondition(obj *unstructured.Unstructured) string {
	conditions, found, _ := unstructured.NestedSlice(obj.Object, "status", "conditions")
	if !found {
		return ""
	}

	for _, cond := range conditions {
		m, ok := cond.(map[string]interface{})
		if !ok {
			continue
		}
		t, _ := m["type"].(string)
		s, _ := m["status"].(string)
		if strings.EqualFold(t, "Ready") {
			return s
		}
	}

	return ""
}
