# K8V Data Model

This document defines the data model for the k8v Kubernetes visualizer, focusing on resource relationships and extensibility.

## Design Principles

1. **Relationship-First**: Resource connections are core to the model, not an afterthought
2. **Extensible**: Easy to add new Kubernetes resource types without breaking existing code
3. **Bidirectional**: Relationships work both ways (Service → Pods AND Pods → Service)
4. **Type-Safe**: Strong typing in Go backend, clear JSON contracts for frontend
5. **Graph-Ready**: Supports both list views and topology graph visualization

---

## Core Data Structures

### Resource

The `Resource` struct represents any Kubernetes object with computed relationships.

```go
// Resource represents any Kubernetes resource
type Resource struct {
    // Identity
    ID        string `json:"id"`        // Unique: "type:namespace:name"
    Type      string `json:"type"`      // "Pod", "Deployment", "Service", etc.
    Name      string `json:"name"`
    Namespace string `json:"namespace"`

    // Status & Health
    Status ResourceStatus `json:"status"`
    Health HealthState    `json:"health"` // "healthy", "warning", "error"

    // Relationships (the key part!)
    Relationships Relationships `json:"relationships"`

    // Metadata
    Labels      map[string]string `json:"labels"`
    Annotations map[string]string `json:"annotations"`
    CreatedAt   time.Time         `json:"createdAt"`

    // Raw data for detail views
    Spec interface{} `json:"spec,omitempty"` // Type-specific data
    YAML string      `json:"yaml"`           // Full YAML for viewing
}
```

**Field Explanations:**

- **ID**: Globally unique identifier following pattern `type:namespace:name`
  - Examples: `pod:default:nginx-abc123`, `service:kube-system:kube-dns`
  - Used for quick lookups and as graph node IDs

- **Type**: Kubernetes resource kind (Pod, Deployment, Service, etc.)

- **Status**: Computed status information (phase, ready count, messages)

- **Health**: High-level health state for visual indicators

- **Relationships**: All connections to other resources (see below)

- **Spec**: Type-specific data (e.g., for Pods: container specs, for Services: ports)

- **YAML**: Full YAML representation for detail view

### Relationships

Captures all connections between resources.

```go
type Relationships struct {
    // Ownership hierarchy
    OwnedBy  []ResourceRef `json:"ownedBy"`  // e.g., ReplicaSet owned by Deployment
    Owns     []ResourceRef `json:"owns"`     // e.g., Deployment owns ReplicaSets

    // Dependencies
    DependsOn []ResourceRef `json:"dependsOn"` // e.g., Pod depends on ConfigMap/Secret
    UsedBy    []ResourceRef `json:"usedBy"`    // e.g., ConfigMap used by Pods

    // Network relationships
    Exposes   []ResourceRef `json:"exposes"`   // e.g., Service exposes Pods
    ExposedBy []ResourceRef `json:"exposedBy"` // e.g., Pod exposed by Service
    RoutesTo  []ResourceRef `json:"routesTo"`  // e.g., Ingress routes to Service
    RoutedBy  []ResourceRef `json:"routedBy"`  // e.g., Service routed by Ingress
}
```

**Relationship Types Explained:**

| Relationship | Description | Example |
|-------------|-------------|---------|
| **OwnedBy** | Kubernetes ownership (OwnerReferences) | ReplicaSet ← Deployment |
| **Owns** | Kubernetes ownership (reverse) | Deployment → ReplicaSet |
| **DependsOn** | Resource needs this to function | Pod → ConfigMap, Pod → Secret |
| **UsedBy** | Other resources depend on this | ConfigMap ← Pod |
| **Exposes** | Network exposure | Service → Pod (endpoints) |
| **ExposedBy** | Exposed by network resource | Pod ← Service |
| **RoutesTo** | Traffic routing | Ingress → Service |
| **RoutedBy** | Receives routed traffic | Service ← Ingress |

### ResourceRef

Lightweight reference to another resource (avoids circular dependencies).

```go
type ResourceRef struct {
    ID        string `json:"id"`        // "type:namespace:name"
    Type      string `json:"type"`      // "Pod", "Service", etc.
    Name      string `json:"name"`
    Namespace string `json:"namespace"`
}
```

### ResourceStatus

Type-specific status information.

```go
type ResourceStatus struct {
    Phase   string `json:"phase"`   // Type-specific: "Running", "Pending", "Active", etc.
    Ready   string `json:"ready"`   // e.g., "3/3" for Deployment replicas
    Message string `json:"message"` // Human-readable status explanation
}
```

### HealthState

High-level health indicator for visual representation.

```go
type HealthState string

const (
    HealthHealthy HealthState = "healthy" // Green: All good
    HealthWarning HealthState = "warning" // Yellow: Degraded or attention needed
    HealthError   HealthState = "error"   // Red: Failed or critical issue
    HealthUnknown HealthState = "unknown" // Gray: Cannot determine health
)
```

**Health Computation Rules:**

- **Pods**: Healthy if Running, Error if CrashLoopBackOff/Failed, Warning if Pending
- **Deployments**: Healthy if ReadyReplicas == Replicas, Warning if partial, Error if 0
- **Services**: Healthy if has endpoints, Warning if partial endpoints, Error if none
- **ConfigMaps/Secrets**: Always Healthy (no runtime failures)

---

## Relationship Examples

### Example 1: Full Application Stack

```
Ingress "api-ingress"
  └─ routesTo → Service "api-service"
                  ├─ exposes → Pod "api-pod-1" (Running)
                  ├─ exposes → Pod "api-pod-2" (Running)
                  └─ exposes → Pod "api-pod-3" (CrashLoopBackOff)

Deployment "api-deployment"
  ├─ owns → ReplicaSet "api-rs-abc123"
  │           └─ owns → Pod "api-pod-1"
  │                     Pod "api-pod-2"
  │                     Pod "api-pod-3"
  └─ dependsOn → ConfigMap "api-config"
                  Secret "api-secrets"
```

**JSON Representation:**

```json
{
  "id": "deployment:default:api-deployment",
  "type": "Deployment",
  "name": "api-deployment",
  "namespace": "default",
  "status": {
    "phase": "Progressing",
    "ready": "2/3",
    "message": "1 pod failing"
  },
  "health": "warning",
  "relationships": {
    "owns": [
      {"id": "replicaset:default:api-rs-abc123", "type": "ReplicaSet", ...}
    ],
    "dependsOn": [
      {"id": "configmap:default:api-config", "type": "ConfigMap", ...},
      {"id": "secret:default:api-secrets", "type": "Secret", ...}
    ]
  }
}
```

```json
{
  "id": "service:default:api-service",
  "type": "Service",
  "name": "api-service",
  "namespace": "default",
  "status": {
    "phase": "Active",
    "ready": "3 endpoints",
    "message": ""
  },
  "health": "healthy",
  "relationships": {
    "exposes": [
      {"id": "pod:default:api-pod-1", "type": "Pod", ...},
      {"id": "pod:default:api-pod-2", "type": "Pod", ...},
      {"id": "pod:default:api-pod-3", "type": "Pod", ...}
    ],
    "routedBy": [
      {"id": "ingress:default:api-ingress", "type": "Ingress", ...}
    ]
  }
}
```

```json
{
  "id": "configmap:default:api-config",
  "type": "ConfigMap",
  "name": "api-config",
  "namespace": "default",
  "health": "healthy",
  "relationships": {
    "usedBy": [
      {"id": "deployment:default:api-deployment", "type": "Deployment", ...},
      {"id": "pod:default:api-pod-1", "type": "Pod", ...},
      {"id": "pod:default:api-pod-2", "type": "Pod", ...},
      {"id": "pod:default:api-pod-3", "type": "Pod", ...}
    ]
  }
}
```

### Example 2: Click to Explore Flow

**User clicks on "api-service" in the UI:**

```javascript
// Frontend has the service resource
const service = {
  id: "service:default:api-service",
  relationships: {
    exposes: [/* 3 pod refs */],
    routedBy: [/* 1 ingress ref */]
  }
}

// UI shows:
// "This service exposes:"
//   - Pod: api-pod-1 (click to navigate)
//   - Pod: api-pod-2 (click to navigate)
//   - Pod: api-pod-3 (click to navigate)
//
// "This service is routed by:"
//   - Ingress: api-ingress (click to navigate)
```

**User clicks on "api-pod-1":**

```javascript
// Navigate to pod, show its relationships
const pod = {
  id: "pod:default:api-pod-1",
  relationships: {
    ownedBy: [/* ReplicaSet ref */],
    exposedBy: [/* Service ref */],
    dependsOn: [/* ConfigMap, Secret refs */]
  }
}

// UI shows:
// "This pod is owned by:"
//   - ReplicaSet: api-rs-abc123 (click to navigate)
//
// "This pod is exposed by:"
//   - Service: api-service (click to navigate)
//
// "This pod depends on:"
//   - ConfigMap: api-config (click to navigate)
//   - Secret: api-secrets (click to navigate)
```

---

## Extensibility Pattern

To add a new resource type (e.g., **StatefulSet**):

### Step 1: Add Transformer Function

```go
// internal/k8s/transformers.go

func TransformStatefulSet(sts *appsv1.StatefulSet) *Resource {
    return &Resource{
        ID:        buildID("StatefulSet", sts.Namespace, sts.Name),
        Type:      "StatefulSet",
        Name:      sts.Name,
        Namespace: sts.Namespace,

        Status: ResourceStatus{
            Phase:   "Running",
            Ready:   fmt.Sprintf("%d/%d", sts.Status.ReadyReplicas, *sts.Spec.Replicas),
            Message: getStatefulSetMessage(sts),
        },

        Health: computeStatefulSetHealth(sts),

        Relationships: Relationships{
            Owns:      extractOwnedPods(sts),        // Pods owned by StatefulSet
            DependsOn: extractVolumeClaims(sts),     // PersistentVolumeClaims
            ExposedBy: findServicesForStatefulSet(sts), // Headless service
        },

        Labels:      sts.Labels,
        Annotations: sts.Annotations,
        CreatedAt:   sts.CreationTimestamp.Time,
        Spec:        sts.Spec,
        YAML:        marshalToYAML(sts),
    }
}

func computeStatefulSetHealth(sts *appsv1.StatefulSet) HealthState {
    if sts.Status.ReadyReplicas == 0 {
        return HealthError
    }
    if sts.Status.ReadyReplicas < *sts.Spec.Replicas {
        return HealthWarning
    }
    return HealthHealthy
}
```

### Step 2: Add Watcher

```go
// internal/k8s/watcher.go

func watchStatefulSets(conn *websocket.Conn, clientset *kubernetes.Clientset, mu *sync.Mutex) {
    watcher, err := clientset.AppsV1().StatefulSets("").Watch(ctx, metav1.ListOptions{})
    if err != nil {
        log.Printf("Failed to watch StatefulSets: %v", err)
        return
    }
    defer watcher.Stop()

    for event := range watcher.ResultChan() {
        sts := event.Object.(*appsv1.StatefulSet)
        resource := TransformStatefulSet(sts)

        mu.Lock()
        conn.WriteJSON(ResourceEvent{
            Type:     string(event.Type),
            Resource: resource,
        })
        mu.Unlock()
    }
}
```

### Step 3: Register in Main

```go
// main.go

func handleWebSocket(w http.ResponseWriter, r *http.Request, clientset *kubernetes.Clientset) {
    // ... existing code ...

    wg.Add(1)
    go watchStatefulSets(conn, clientset, &mu)  // Add this line
}
```

**That's it!** No changes to frontend data structures needed.

---

## Frontend TypeScript Interface

```typescript
interface Resource {
    id: string;
    type: string;
    name: string;
    namespace: string;

    status: ResourceStatus;
    health: 'healthy' | 'warning' | 'error' | 'unknown';

    relationships: Relationships;

    labels: Record<string, string>;
    annotations: Record<string, string>;
    createdAt: string;

    spec?: any;
    yaml: string;
}

interface Relationships {
    ownedBy: ResourceRef[];
    owns: ResourceRef[];
    dependsOn: ResourceRef[];
    usedBy: ResourceRef[];
    exposes: ResourceRef[];
    exposedBy: ResourceRef[];
    routesTo: ResourceRef[];
    routedBy: ResourceRef[];
}

interface ResourceRef {
    id: string;
    type: string;
    name: string;
    namespace: string;
}

interface ResourceStatus {
    phase: string;
    ready: string;
    message: string;
}

interface ResourceEvent {
    type: 'ADDED' | 'MODIFIED' | 'DELETED';
    resource: Resource;
}
```

---

## Computing Relationships

### Ownership (OwnedBy/Owns)

Use Kubernetes `OwnerReferences`:

```go
func extractOwners(obj metav1.Object) []ResourceRef {
    refs := []ResourceRef{}
    for _, owner := range obj.GetOwnerReferences() {
        refs = append(refs, ResourceRef{
            ID:        buildID(owner.Kind, obj.GetNamespace(), owner.Name),
            Type:      owner.Kind,
            Name:      owner.Name,
            Namespace: obj.GetNamespace(),
        })
    }
    return refs
}
```

### Dependencies (DependsOn)

Parse Pod spec for ConfigMaps/Secrets:

```go
func extractConfigMapDeps(pod *v1.Pod) []ResourceRef {
    refs := []ResourceRef{}

    // Volume mounts
    for _, volume := range pod.Spec.Volumes {
        if volume.ConfigMap != nil {
            refs = append(refs, ResourceRef{
                ID:        buildID("ConfigMap", pod.Namespace, volume.ConfigMap.Name),
                Type:      "ConfigMap",
                Name:      volume.ConfigMap.Name,
                Namespace: pod.Namespace,
            })
        }
    }

    // Env from
    for _, container := range pod.Spec.Containers {
        for _, envFrom := range container.EnvFrom {
            if envFrom.ConfigMapRef != nil {
                refs = append(refs, ResourceRef{
                    ID:        buildID("ConfigMap", pod.Namespace, envFrom.ConfigMapRef.Name),
                    Type:      "ConfigMap",
                    Name:      envFrom.ConfigMapRef.Name,
                    Namespace: pod.Namespace,
                })
            }
        }
    }

    return refs
}
```

### Network (Exposes/RoutesTo)

Match Service selectors to Pod labels:

```go
func findExposedPods(service *v1.Service, allPods []*v1.Pod) []ResourceRef {
    refs := []ResourceRef{}

    for _, pod := range allPods {
        if pod.Namespace != service.Namespace {
            continue
        }

        // Check if pod labels match service selector
        if labelsMatch(pod.Labels, service.Spec.Selector) {
            refs = append(refs, ResourceRef{
                ID:        buildID("Pod", pod.Namespace, pod.Name),
                Type:      "Pod",
                Name:      pod.Name,
                Namespace: pod.Namespace,
            })
        }
    }

    return refs
}
```

---

## Implementation Notes

### ID Format

Always use: `type:namespace:name`

Examples:
- `pod:default:nginx-abc123`
- `service:kube-system:kube-dns`
- `configmap:monitoring:prometheus-config`

For cluster-scoped resources (no namespace):
- `namespace::kube-system`
- `clusterrole::admin`

### Bidirectional Relationships

Maintain both directions for efficient lookups:

```go
// When processing a Deployment that owns a ReplicaSet:
deployment.Relationships.Owns = []ResourceRef{rsRef}
replicaSet.Relationships.OwnedBy = []ResourceRef{deployRef}

// When processing a Service that exposes Pods:
service.Relationships.Exposes = []ResourceRef{pod1, pod2, pod3}
pod1.Relationships.ExposedBy = []ResourceRef{serviceRef}
pod2.Relationships.ExposedBy = []ResourceRef{serviceRef}
pod3.Relationships.ExposedBy = []ResourceRef{serviceRef}
```

### Caching Strategy

Use a central cache keyed by resource ID:

```go
type ResourceCache struct {
    mu        sync.RWMutex
    resources map[string]*Resource // ID -> Resource
}

func (c *ResourceCache) Get(id string) (*Resource, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    r, ok := c.resources[id]
    return r, ok
}

func (c *ResourceCache) Set(r *Resource) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.resources[r.ID] = r
}
```

### Relationship Updates

When a resource changes, update both sides:

```go
func updateRelationships(cache *ResourceCache, resource *Resource) {
    // Update forward relationships
    for _, ref := range resource.Relationships.DependsOn {
        if dep, ok := cache.Get(ref.ID); ok {
            addToUsedBy(dep, resource)
        }
    }

    // Update reverse relationships
    for _, ref := range resource.Relationships.UsedBy {
        if user, ok := cache.Get(ref.ID); ok {
            addToDependsOn(user, resource)
        }
    }
}
```

---

## Resource Type Support

### Phase 1 (MVP)

| Resource | Relationships | Priority |
|----------|---------------|----------|
| **Pod** | OwnedBy, ExposedBy, DependsOn | High |
| **Deployment** | Owns, DependsOn | High |
| **ReplicaSet** | OwnedBy, Owns | High |
| **Service** | Exposes, RoutedBy | High |
| **Ingress** | RoutesTo | High |
| **ConfigMap** | UsedBy | High |
| **Secret** | UsedBy | High |

### Phase 2 (Future)

- StatefulSet
- DaemonSet
- Job / CronJob
- PersistentVolumeClaim
- Namespace
- ServiceAccount

---

## Summary

This data model provides:

✅ **Rich relationships** - Capture all important connections between resources
✅ **Extensibility** - Add new types with ~50 lines of code
✅ **Type safety** - Strong contracts between backend and frontend
✅ **Graph support** - Ready for topology visualization
✅ **Performance** - Efficient lookups and caching
✅ **Bidirectional** - Navigate relationships in both directions

The model is designed to support the core feature: **clicking on a resource shows what it's linked to**, making cluster exploration intuitive and powerful.
