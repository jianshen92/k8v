# K8V Implementation Phases

This document outlines the phased approach to building k8v, from POC to production.

---

## ✅ Phase 0: POC (COMPLETED)

**Goal:** Validate the streaming architecture

**What We Built:**
- Location: `k8v-poc/`
- Simple table UI (unstyled)
- WebSocket streaming from Go backend
- K8s Watch API integration (direct, no Informers)
- Three resource types: Pods, Deployments, ReplicaSets
- Real-time ADD/MODIFY/DELETE events

**What We Validated:**
- ✅ K8s Watch API works correctly
- ✅ WebSocket streaming to browser works
- ✅ Real-time UI updates work (< 1 second latency)
- ✅ Basic table UI successfully displays resources
- ✅ Event handling (ADD/MODIFY/DELETE) works

**Key Learnings:**
- Direct Watch API is simple and works
- Need mutex for concurrent WebSocket writes
- k8s.io/client-go v0.31.0 works with Go 1.23
- Browser WebSocket API is straightforward

**Duration:** 1 day

---

## ✅ Phase 1: Production Backend + Minimal Frontend (COMPLETED)

**Goal:** Build production-quality backend with full data model and validate relationship navigation

**Status:** Complete - Production backend with generic relationship system and minimal frontend implemented

### Backend (Production Quality)

**1. Project Structure**
```
k8v/
├── cmd/k8v/
│   └── main.go              # CLI entry point
├── internal/
│   ├── server/
│   │   ├── server.go        # HTTP/WebSocket server
│   │   ├── handlers.go      # API handlers
│   │   └── websocket.go     # WebSocket streaming
│   ├── k8s/
│   │   ├── client.go        # K8s client setup
│   │   ├── watcher.go       # Watch manager with Informers
│   │   ├── cache.go         # Resource cache
│   │   ├── transformers.go  # K8s objects → Resource model
│   │   └── relationships.go # Compute relationships
│   └── types/
│       └── resource.go      # Data model structs
├── web/
│   ├── index.html           # Minimal frontend
│   └── embed.go             # Go embed directives
└── go.mod
```

**2. Core Implementation**

**Data Model (internal/types/resource.go):**
- `Resource` struct with full relationship model
- `Relationships` with 8 types (OwnedBy, Owns, DependsOn, UsedBy, etc.)
- `ResourceRef`, `ResourceStatus`, `HealthState` types
- Resource ID format: `type:namespace:name`

**K8s Integration (internal/k8s/):**
- Use **Informers** (not direct Watch) for efficiency
- `SharedInformerFactory` for all resource types
- Maintain **ResourceCache** (map[string]*Resource)
- Compute relationships on every update:
  - Ownership: Parse OwnerReferences
  - Dependencies: Extract ConfigMap/Secret refs from Pod specs
  - Network: Match Service selectors to Pod labels, Ingress rules to Services
  - **Bidirectional**: Update both sides (Service→Pods AND Pods→Service)

**Resource Types (7 total):**
1. Pod
2. Deployment
3. ReplicaSet
4. Service
5. Ingress
6. ConfigMap
7. Secret

**Transformers (internal/k8s/transformers.go):**
- `TransformPod(pod *v1.Pod, cache *ResourceCache) *Resource`
- `TransformDeployment(dep *appsv1.Deployment, cache *ResourceCache) *Resource`
- `TransformService(svc *v1.Service, cache *ResourceCache) *Resource`
- ... one per resource type

**Relationship Computation (internal/k8s/relationships.go):**
```go
// Extract ownership from OwnerReferences
func computeOwnership(obj metav1.Object) []ResourceRef

// Find ConfigMaps/Secrets used by Pod
func computeDependencies(pod *v1.Pod) []ResourceRef

// Find Pods exposed by Service
func computeExposedPods(service *v1.Service, cache *ResourceCache) []ResourceRef

// Find Services routed by Ingress
func computeRoutedServices(ingress *netv1.Ingress) []ResourceRef
```

**WebSocket Protocol:**
```json
{
  "type": "ADDED",
  "resource": {
    "id": "pod:default:nginx-abc",
    "type": "Pod",
    "name": "nginx-abc",
    "namespace": "default",
    "status": {"phase": "Running", "ready": "1/1"},
    "health": "healthy",
    "relationships": {
      "ownedBy": [{"id": "replicaset:default:nginx-rs", ...}],
      "exposedBy": [{"id": "service:default:nginx-svc", ...}],
      "dependsOn": [{"id": "configmap:default:app-config", ...}]
    },
    "labels": {"app": "nginx"},
    "yaml": "apiVersion: v1\nkind: Pod\n..."
  }
}
```

**3. Server Implementation**

**HTTP Routes:**
- `GET /` → Serve index.html
- `GET /ws` → WebSocket upgrade
- `GET /health` → Health check endpoint

**WebSocket Behavior:**
1. Client connects → Send full snapshot of all resources (ADDED events)
2. K8s watch fires → Send incremental updates (ADDED/MODIFIED/DELETED)
3. Relationship changes → Send updated resource with new relationships

### Frontend (Minimal but Functional)

**UI Components:**

**1. Resource Table**
```html
<table>
  <thead>
    <tr>
      <th>Type</th>
      <th>Namespace</th>
      <th>Name</th>
      <th>Status</th>
      <th>Health</th>
    </tr>
  </thead>
  <tbody id="resources">
    <!-- Rows dynamically populated -->
  </tbody>
</table>
```

**2. Detail Panel (Slide-in from right)**
```html
<div id="detail-panel" class="hidden">
  <div class="panel-header">
    <h2 id="detail-name">Pod: nginx-abc</h2>
    <button onclick="closeDetail()">×</button>
  </div>

  <div class="panel-content">
    <!-- Tab 1: Overview -->
    <div id="overview-tab">
      <div class="section">
        <h3>Status</h3>
        <p>Phase: Running</p>
        <p>Ready: 1/1</p>
      </div>

      <div class="section">
        <h3>Relationships</h3>

        <h4>Owned By</h4>
        <ul id="ownedBy-list">
          <!-- Clickable links to related resources -->
          <li><a href="#" onclick="showResource('replicaset:default:nginx-rs')">
            ReplicaSet: nginx-rs
          </a></li>
        </ul>

        <h4>Exposed By</h4>
        <ul id="exposedBy-list">
          <li><a href="#" onclick="showResource('service:default:nginx-svc')">
            Service: nginx-svc
          </a></li>
        </ul>

        <h4>Depends On</h4>
        <ul id="dependsOn-list">
          <li><a href="#" onclick="showResource('configmap:default:app-config')">
            ConfigMap: app-config
          </a></li>
        </ul>
      </div>
    </div>

    <!-- Tab 2: YAML -->
    <div id="yaml-tab" class="hidden">
      <pre><code id="yaml-content"></code></pre>
    </div>
  </div>
</div>
```

**3. Styling (Minimal)**
- Basic CSS for table borders, padding
- Health colors: Green (healthy), Yellow (warning), Red (error)
- Detail panel: Simple slide-in animation
- Clickable links: Blue underline
- NO glassmorphism, NO fancy animations, NO Mermaid topology

**4. JavaScript Logic**

```javascript
// State management
const resources = {}; // resourceId → Resource object

// WebSocket handling
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onmessage = (event) => {
  const msg = JSON.parse(event.data);

  if (msg.type === 'DELETED') {
    delete resources[msg.resource.id];
    removeTableRow(msg.resource.id);
  } else {
    resources[msg.resource.id] = msg.resource;
    updateTableRow(msg.resource);
  }
};

// Show resource details
function showResource(resourceId) {
  const resource = resources[resourceId];

  // Populate detail panel
  document.getElementById('detail-name').textContent =
    `${resource.type}: ${resource.name}`;

  // Populate relationships (clickable links)
  renderRelationships(resource.relationships);

  // Show YAML
  document.getElementById('yaml-content').textContent = resource.yaml;

  // Open panel
  document.getElementById('detail-panel').classList.remove('hidden');
}

// Render relationship links
function renderRelationships(relationships) {
  for (const [relType, refs] of Object.entries(relationships)) {
    const list = document.getElementById(`${relType}-list`);
    list.innerHTML = refs.map(ref =>
      `<li><a href="#" onclick="showResource('${ref.id}')">
        ${ref.type}: ${ref.name}
      </a></li>`
    ).join('');
  }
}
```

**Key Features:**
✅ Click row → Shows detail panel
✅ Detail panel shows all relationships as clickable links
✅ Click relationship → Navigates to that resource's detail
✅ YAML tab shows full resource definition
✅ Real-time updates (table updates automatically)
✅ Health indicators (colored status)

**What's NOT Included:**
❌ No topology graph
❌ No advanced filtering/search
❌ No fancy UI design (that's Phase 2)
❌ No pod logs yet
❌ No Mermaid diagrams

### Success Criteria

**Backend:**
- ✅ All 7 resource types stream correctly
- ✅ Relationships are computed accurately
- ✅ Bidirectional relationships work (both directions)
- ✅ Click Service → shows all Pods it exposes
- ✅ Click Pod → shows Service that exposes it
- ✅ Real-time updates within 500ms
- ✅ Memory usage < 100MB for 1000 resources

**Frontend:**
- ✅ Table shows all resources
- ✅ Click resource → detail panel opens
- ✅ Relationships shown as clickable links
- ✅ Click relationship → navigates to that resource
- ✅ YAML tab displays correctly
- ✅ Real-time updates work without refresh

**Integration:**
- ✅ `k8v` command starts server, opens browser
- ✅ Connects to local K8s cluster (kubeconfig)
- ✅ Can explore entire cluster by clicking through relationships

### What Was Built

**Backend:**
- ✅ Complete data model with 8 relationship types (OwnedBy, Owns, DependsOn, UsedBy, Exposes, ExposedBy, RoutesTo, RoutedBy)
- ✅ Thread-safe resource cache
- ✅ **Generic relationship computation system** (refactored from 4 specific functions to 1 generic function)
- ✅ Transformers for all 7 resource types (Pod, Deployment, ReplicaSet, Service, Ingress, ConfigMap, Secret)
- ✅ K8s client with SharedInformerFactory pattern
- ✅ Event-driven watcher with handlers for all resource types
- ✅ HTTP/WebSocket server with hub pattern
- ✅ Proper error handling and panic recovery for large clusters (2000+ resources)

**Frontend:**
- ✅ Enhanced table with Type, Namespace, Name, Status, Health columns
- ✅ Detail panel with Overview and YAML tabs
- ✅ Clickable relationship links for bidirectional navigation
- ✅ Real-time WebSocket updates with console logging
- ✅ Health indicators (green/yellow/red)
- ✅ Dark theme with clean styling
- ✅ Connection status monitoring

**Binary:**
- ✅ Single 62MB binary with embedded web UI
- ✅ CLI with port configuration
- ✅ Handles large clusters (tested with 2296 resources)

### Key Improvements

**Generic Relationship System:**
- Replaced 4 specific functions with 1 generic `FindReverseRelationships()`
- Added `RelationshipType` enum for type safety
- Bidirectional relationships computed on-demand from cache
- Scalable: Add new relationship types without new functions
- Reduced code duplication significantly

**Console Logging for Development:**
- Color-coded event logging (green=ADDED, yellow=MODIFIED, red=DELETED)
- Relationship summaries in console
- Connection status tracking
- Snapshot completion detection

### Actual Duration

- **Backend implementation**: 1 day
- **Minimal frontend**: 1 day
- **Bug fixes & improvements**: 1 day (WebSocket panic fix, generic relationships refactor)
- **Total: 3 days**

---

## 📋 Phase 2: Full Frontend Implementation (FUTURE)

**Goal:** Replace minimal UI with production-quality design from index.html prototype

### What to Build

**1. Extract from index.html Prototype:**
- Dark-themed glassmorphic design
- Smooth animations and transitions
- Dashboard view with statistics cards
- Advanced filtering (by type, namespace, health)
- Search functionality
- Events timeline
- Responsive grid layout

**2. Topology View (Optional):**
- Mermaid diagram showing resource relationships
- Interactive graph with pan/zoom
- Click nodes to show details
- Visual flow: Ingress → Service → Deployment → Pods

**3. Additional Features:**
- Pod logs viewer (stream logs via WebSocket)
- Resource editing (kubectl apply via backend)
- Namespace filtering
- Multi-cluster support (switch contexts)
- Export functionality (YAML download, screenshots)
- Dark/light theme toggle

### Success Criteria

- ✅ Matches index.html prototype quality
- ✅ All prototype features implemented
- ✅ Topology view works smoothly
- ✅ Pod logs streaming works
- ✅ Advanced filtering and search work
- ✅ Production-ready UI/UX

### Estimated Duration

- **Dashboard view**: 2-3 days
- **Topology view**: 2-3 days
- **Pod logs**: 1-2 days
- **Polish & testing**: 2-3 days
- **Total: 7-11 days**

---

## Summary Timeline

| Phase | Goal | Duration | Status |
|-------|------|----------|--------|
| **Phase 0** | Validate architecture | 1 day | ✅ Done |
| **Phase 1** | Production backend + minimal UI | 5-7 days | 🔄 Current |
| **Phase 2** | Full frontend like prototype | 7-11 days | 📋 Future |

**Total estimated time: 13-19 days** for complete production system.

---

## Next Actions (Phase 1)

1. Set up production Go project structure
2. Implement data model types (Resource, Relationships)
3. Implement K8s client with Informers
4. Build transformers for 7 resource types
5. Implement relationship computation
6. Build WebSocket streaming with full snapshots
7. Create minimal frontend (enhanced table + detail panel)
8. Test relationship navigation end-to-end
9. Validate with real cluster (create/delete resources, verify updates)

**Let's start with #1: Project structure setup!** 🚀
