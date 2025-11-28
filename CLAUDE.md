# K8V Project Context

> This document provides a comprehensive overview of the k8v (Kubernetes Visualizer) project for onboarding and context understanding.

---

## 1. Project Overview

**What is k8v?**
K8v is a Kubernetes cluster visualization tool designed to be like k9s but with a modern web UI and superior user experience.

**Vision:**
- CLI tool that connects to any Kubernetes cluster
- Beautiful, intuitive web interface for cluster visualization
- Real-time updates streaming from the cluster
- Resource relationship graphs showing how components connect
- Pod logs viewing and resource inspection

**Current State:**
- **Stage:** ✅ Phase 3 Ongoing - Advanced Filtering & Performance Optimization
- **What exists:**
  - Production Go backend with Informers, WebSocket streaming, and generic relationship system
  - Polished web UI with real-time updates and optimized rendering
  - Bidirectional relationship navigation
  - **Namespace filtering**: Server-side filtering with searchable dropdown UI
  - **Resource type lazy loading**: Instant stats + filtered WebSocket snapshots
  - **Performance optimized**: 40-100x network reduction for large clusters
  - Single 62MB binary (k8v) ready to use
  - Tested with large production clusters (21,000+ resources)
- **Next:** Pod logs viewing, search functionality, and additional resource types

---

## 2. Index.html POC - Key Findings

### What Exists

The prototype (`index.html`) is a **complete, production-ready frontend** that demonstrates the full vision:

**UI Design:**
- Dark-themed glassmorphic interface with smooth animations
- Professional, modern aesthetic comparable to commercial products
- Responsive grid layout that adapts to content
- Color-coded health indicators (green=healthy, yellow=warning, red=error)

**Two Main Views:**

1. **Dashboard View** (Primary)
   - Resource statistics cards showing counts and health status
   - Filterable resource lists by type (Pods, Deployments, Services, etc.)
   - Recent events timeline with severity indicators
   - Click any resource to see detailed information

2. **Topology View** (Secondary)
   - Interactive Mermaid diagram showing resource relationships
   - Visual representation of traffic flow: Ingress → Services → Deployments → ReplicaSets → Pods
   - Pan and zoom capabilities for exploring complex clusters

**Interactive Features:**
- **Resource filtering:** Filter by type using buttons (All, Pods, Deployments, Services, Ingress, ReplicaSets, ConfigMaps, Secrets)
- **Detail panels:** Click any resource to view:
  - Overview with metadata and status
  - Full YAML configuration with syntax highlighting
  - Relationships showing connected resources
- **YAML navigation:** Clickable resource references in YAML that navigate to related resources
- **Copy to clipboard:** One-click YAML copying functionality
- **Health indicators:** Visual status for every resource (healthy, warning, error with pulsing animation)

**Data Model:**
- Well-structured mock data simulating a medium-sized cluster
- Comprehensive resource type coverage:
  - Ingress (routing entry points)
  - Services (network abstractions)
  - Deployments (application deployments)
  - ReplicaSets (pod replica management)
  - Pods (running containers)
  - ConfigMaps (configuration data)
  - Secrets (sensitive data)
- Relationship mapping: `Ingress → Service → Deployment → ReplicaSet → Pod`
- Includes three namespaces: default, monitoring, logging

**Technology Stack:**
- Pure HTML/CSS/JavaScript (no framework dependencies)
- ES6 modules for code organization
- Feather Icons for consistent iconography
- Google Fonts (Space Grotesk, Inter)
- Modular architecture with separation of concerns

### Strengths

1. **UI/UX Excellence:** Production-quality design with attention to detail
2. **Comprehensive Coverage:** All major K8s resource types represented
3. **Relationship Mapping:** Clear visualization of how resources connect
4. **Intuitive Interactions:** Natural click-to-explore navigation pattern
5. **Professional Polish:** Smooth animations, hover effects, visual feedback
6. **Data Structure:** Well-organized schema ready for real K8s data

### What's Missing

1. **No Kubernetes API connection** - Uses hardcoded mock data
2. **No real-time updates** - Static data, no live streaming
3. **No backend server** - Pure client-side application
4. **No CLI tool wrapper** - Just an HTML file, not a command
5. **No pod logs viewing** - Only resource metadata/YAML shown

---

## 3. IDEAS.MD Summary

**Purpose:** Documents the user's vision, feature requirements, and MVP priorities.

### Core Vision
Build a tool that combines:
- The power of k9s (Kubernetes TUI)
- The accessibility of a web interface
- Superior user experience and visual design

### User Workflow
1. User has kubeconfig setup locally
2. User runs `k8v` command
3. Backend server starts and opens browser
4. UI displays live cluster state with streaming updates

### Core MVP Priorities

**Must-Have Features:**
1. **Resource Visualization**
   - Display all core K8s resources (Pods, Deployments, Services, Ingress, ConfigMaps, Secrets, ReplicaSets)
   - Construct relationship graphs showing connections (e.g., Deployment → ConfigMap)
   - Click any resource to see what's linked to it

2. **Live Streaming**
   - Sync to Kubernetes cluster
   - Stream live updates to dashboard
   - Real-time reflection of cluster changes

3. **Pod Logs Viewing**
   - View logs for any pod
   - Essential for debugging and monitoring

4. **Top-Tier UI**
   - Preserve the existing prototype's quality
   - Smooth, responsive, professional

5. **Extensible Data Model**
   - Easy to add new resource types
   - Support for future K8s resources

**Secondary Features:**
- Topology graph view (acknowledged as a hard problem, not critical for MVP)
- Search functionality (important but can come in v2)

### Feature Categories

**Core Visualization:**
- Interactive dashboard with real-time metrics
- Multi-namespace support
- Resource type filtering
- Health status indicators

**Resource Details:**
- Detailed resource view with specifications
- YAML configuration display
- Cross-reference navigation between related resources
- Relationship visualization

**Monitoring & Events:**
- Recent events timeline
- Event severity levels
- Event source tracking

**User Experience:**
- Modern UI design
- Search functionality
- Responsive layout
- Quick actions

---

## 4. DESIGN.MD Summary

**Purpose:** Technical architecture and implementation plan for transforming the prototype into a production tool.

### Technology Choices

**Backend: Go**
- Native Kubernetes ecosystem (official `client-go` library)
- Single binary compilation with zero runtime dependencies
- Built-in `embed` package for bundling HTML/CSS/JS assets
- Excellent performance for concurrent watch streams
- Small footprint (~15-30MB binaries)
- Cross-platform support (macOS, Linux, Windows)

**Communication: WebSocket**
- Bidirectional communication for real-time updates
- Low latency for cluster event streaming
- Natural mapping from K8s watch API to browser updates
- Future-proof for interactive features (exec, port-forward)

### System Architecture

```
CLI Binary (k8v)
  ↓
Embedded HTTP/WebSocket Server (localhost:8080)
  ↓
K8s Client Manager (client-go + Informers)
  ↓
Kubernetes API Server
```

**Components:**
1. **CLI Entry Point** - Parse flags, initialize K8s client, start server, open browser
2. **HTTP Server** - Serve embedded static files (the prototype)
3. **WebSocket Handler** - Stream K8s events to browser in real-time
4. **K8s Client Manager** - Watch cluster resources, maintain cache, handle reconnections

### Project Structure

```
k8v/
├── cmd/k8v/main.go              # CLI entry point
├── internal/
│   ├── server/                   # HTTP/WebSocket server
│   │   ├── static/               # Frontend assets (embedded)
│   │   │   ├── index.html        # HTML structure
│   │   │   ├── style.css         # Styles
│   │   │   ├── app.js            # Main application logic
│   │   │   ├── config.js         # Configuration constants
│   │   │   ├── state.js          # State management
│   │   │   ├── ws.js             # WebSocket connection management
│   │   │   └── dropdown.js       # Reusable dropdown component
│   ├── k8s/                      # K8s client, watcher, cache
│   └── browser/                  # Cross-platform browser launcher
├── pkg/types/                    # Shared types
└── scripts/                      # Build scripts
```

### Implementation Phases

**Phase 1: Basic CLI + Static File Serving** (2-3 days)
- Extract CSS/JS from prototype into separate files
- Implement Go embed for bundling assets
- Create basic HTTP server serving embedded files
- Add CLI with `--port` flag
- Result: `k8v` command opens browser showing prototype

**Phase 2: K8s API Integration** (4-5 days)
- Implement kubeconfig loading with context support
- Create K8s client initialization
- Build resource fetcher for initial snapshot
- Create REST endpoint `/api/resources`
- Transform K8s objects to frontend format
- Result: `k8v --context=minikube` shows real cluster data

**Phase 3: Real-Time Watch Mode** (5-6 days)
- Implement WebSocket upgrade handler `/ws`
- Create K8s watch manager using Informers
- Build WebSocket broadcaster for events
- Update frontend WebSocket client
- Implement UI partial updates
- Add reconnection logic
- Result: Live updates within 1 second of cluster changes

**Phase 4: Polish & Pod Logs** (Post-MVP)
- Add pod logs viewing endpoint
- Implement logs streaming via WebSocket
- Additional UI enhancements
- Performance optimizations

### Key Technical Decisions

1. **K8s Client:** Use official `client-go` with SharedInformerFactory pattern
2. **Authentication:** Leverage `client-go/tools/clientcmd` for kubeconfig handling
3. **Caching:** Two-tier strategy (server-side Informers + client-side state)
4. **State Management:** Immutable state updates with event sourcing in frontend
5. **Security:** Localhost-only binding by default (like kubectl proxy)

### WebSocket Message Protocol

```json
{
  "type": "RESOURCE_ADDED",
  "resourceType": "pod",
  "namespace": "default",
  "data": { /* full pod object */ }
}
```

### Success Metrics

- **Startup time:** < 2 seconds from command to browser open
- **Initial load:** < 1 second to render 100 resources
- **Update latency:** < 500ms from K8s event to UI update
- **Memory usage:** < 100MB for typical cluster (1000 resources)
- **Binary size:** < 30MB uncompressed

### Distribution

- **Primary:** Binary releases via GitHub Releases
- **Secondary:** Homebrew tap
- **Tertiary:** `go install` for Go developers

**Supported Platforms:**
- macOS (Intel + Apple Silicon)
- Linux (amd64 + arm64)
- Windows (amd64)

---

## 4.5. DATA_MODEL.MD Summary

**Purpose:** Defines the complete data model for resources and relationships

**Key Content:**

**Core Structure:**
```go
type Resource struct {
    ID            string         // "type:namespace:name"
    Type          string         // "Pod", "Deployment", etc.
    Name          string
    Namespace     string
    Status        ResourceStatus
    Health        HealthState    // "healthy", "warning", "error"
    Relationships Relationships  // The key part!
    Labels        map[string]string
    YAML          string
}

type Relationships struct {
    OwnedBy   []ResourceRef  // Ownership hierarchy
    Owns      []ResourceRef
    DependsOn []ResourceRef  // Dependencies (ConfigMap, Secret)
    UsedBy    []ResourceRef
    Exposes   []ResourceRef  // Network relationships
    ExposedBy []ResourceRef
    RoutesTo  []ResourceRef  // Traffic routing
    RoutedBy  []ResourceRef
}
```

**Key Design Decisions:**
- **Relationship-first approach**: Resource connections are core, not an afterthought
- **Bidirectional references**: Navigate both ways (Service → Pods AND Pods → Service)
- **Resource ID format**: `type:namespace:name` for unique identification
- **Health computation**: Derived from status (Running = healthy, CrashLoop = error, etc.)
- **Extensible pattern**: Add new resource types with ~50 lines of code

**Relationship Types:**
- **Ownership**: Deployment → ReplicaSet → Pod (via OwnerReferences)
- **Dependencies**: Pod → ConfigMap/Secret (via volume mounts, env vars)
- **Network**: Service → Pods (via selector), Ingress → Service (via routes)

**Example Relationship Chain:**
```
Ingress "api"
  ├─ routesTo → Service "api-svc"
  │              └─ exposes → Pod "api-1", "api-2", "api-3"
  │
Deployment "api-deploy"
  ├─ owns → ReplicaSet "api-rs"
  │          └─ owns → Pod "api-1", "api-2", "api-3"
  └─ dependsOn → ConfigMap "api-config"
                  Secret "api-secrets"
```

**Click-to-explore flow:**
1. User clicks "Service: api-svc"
2. UI shows "Exposes: Pod api-1, api-2, api-3" (clickable)
3. UI shows "Routed by: Ingress api" (clickable)
4. User clicks Pod → sees ownedBy ReplicaSet, dependsOn ConfigMap, etc.

**Extensibility:** Adding StatefulSet requires:
1. Write `TransformStatefulSet()` function
2. Add `watchStatefulSets()` goroutine
3. Register in main → Done! No frontend changes needed.

---

## 5. POC Validation (Completed)

**Status:** ✅ Minimal streaming POC built and validated

**Location:** `/Users/jianshenyap/code/k8v/k8v-poc/`

**What was validated:**
- ✅ K8s watch API works correctly
- ✅ WebSocket streaming to browser works
- ✅ Real-time UI updates work (< 1 second latency)
- ✅ Simple table UI successfully displays Pods, Deployments, ReplicaSets
- ✅ ADD/MODIFY/DELETE events handled correctly

**Key learnings:**
- Direct Watch API (not Informers) is simple and works for POC
- gorilla/websocket handles concurrent writes (need mutex)
- Browser WebSocket API is straightforward
- k8s.io/client-go requires Go 1.23 (use v0.31.0, not latest)

**Next:** Build production system with full data model

---

## 6. Phase 2 Complete (Production-Ready Application)

### What Was Built in Phase 2

✅ **UI Refinements and Optimizations**
- Removed "ALL" filter tab - users now view specific resource types
- Default view set to "Pods" for immediate usefulness
- Compact statistics cards (reduced from 220px to 140px minwidth)
- Alphabetical sorting by resource name (not grouped by type)
- Fixed resource pill height issues for long names

✅ **Performance Optimizations**
- **Incremental DOM updates**: Only affected resources are updated, not full list rerenders
  - ADDED events: Insert new pill in correct sorted position
  - MODIFIED events: Replace only the changed pill in place
  - DELETED events: Remove only the deleted pill
  - No more flickering or scroll position loss
- **Filter-aware updates**: Respects current filter, only shows/hides matching resources
- **Initial snapshot optimization**: Full render during snapshot, incremental after

✅ **WebSocket Stability for Large Clusters**
- **Fixed race condition**: Snapshot now sent synchronously before starting read/write pumps
- **Direct WriteJSON**: Snapshot bypasses buffered channel to avoid "send on closed channel" panics
- **Progress logging**: Shows progress every 1000 resources during snapshot transmission
- **Graceful error handling**: Proper cleanup if client disconnects during snapshot
- **Tested at scale**: Successfully handles 21,867 resources in production cluster

✅ **UI Polish**
- Consistent pill heights with `min-height: 56px`
- Proper alignment with `align-items: flex-start`
- Smaller, more compact statistics section
- Better visual hierarchy focusing on resource list

### Technical Implementation Details

**Incremental Updates (`internal/server/static/index.html`)**
```javascript
// Before: Full rerender on every event (slow, flickering)
function handleResourceEvent(event) {
  updateStatCards();
  renderResourceList(); // ← Cleared and rebuilt entire list
}

// After: Targeted DOM updates (fast, smooth)
function handleResourceEvent(event) {
  updateStatCards();
  if (snapshotComplete) {
    updateResourceInList(resourceId, event.type); // ← Only update one resource
  } else {
    renderResourceList(); // During snapshot, use full render
  }
}
```

**WebSocket Race Condition Fix (`internal/server/websocket.go`)**
```go
// Before: Async snapshot with race condition
go func() {
  for _, event := range snapshot {
    client.send <- event // Could panic if channel closed
  }
}()
go client.writePump()
go client.readPump()

// After: Synchronous snapshot, no race
snapshot := s.watcher.GetSnapshot()
for i, event := range snapshot {
  err := conn.WriteJSON(event) // Direct write, no channel
  if err != nil {
    conn.Close()
    return
  }
  // Progress logging every 1000 resources
}
// Now start pumps after snapshot complete
go client.writePump()
go client.readPump()
```

### Performance Characteristics

**Large Cluster Performance (21,867 resources)**:
- Snapshot transmission: ~2-5 seconds (with progress logging)
- Memory usage: Stable, no leaks observed
- Update latency: < 100ms for incremental updates
- No flickering or visual artifacts
- No WebSocket panics or crashes

**UI Rendering Performance**:
- Initial render: Full list render during snapshot (expected)
- Live updates: Single DOM element add/update/remove (optimized)
- Filter changes: Full list render (expected, infrequent)
- Scroll position: Preserved during updates

### Current Binary Specs

- **Size:** 62MB
- **Platform:** macOS (tested), Linux/Windows (should work)
- **Dependencies:** None (embedded web UI)
- **Command:** `./k8v` or `./k8v -port 8080`

### Known Limitations

1. ~~**Frontend lag with large clusters**~~ - ✅ **PARTIALLY FIXED** Lazy loading by resource type reduces load significantly (virtualization still future work)
2. ~~**No namespace filtering**~~ - ✅ **COMPLETED** Server-side namespace filtering with searchable dropdown
3. **No pod logs viewing** - Cannot view container logs yet (High priority)
4. **Basic YAML view** - No syntax highlighting or clickable references (Future)
5. **No search functionality** - Cannot search by name or labels yet (Future)
6. **Limited resource types** - Only 7 core types (StatefulSets, DaemonSets, etc. in Future)
7. **No multi-cluster support** - Single context only (Future)
8. **Topology view not implemented** - Placeholder shown (Future)

---

## 6.5. Phase 3 Progress (Namespace Filtering & UI Polish)

### What Was Built in Phase 3

✅ **Server-Side Namespace Filtering**
- **Backend filtering at WebSocket level**: Resources filtered before transmission
- **Filter broadcasts**: Hub only sends events matching client's namespace
- **Query parameter support**: WebSocket accepts `?namespace=xxx` parameter
- **New API endpoint**: `/api/namespaces` lists available namespaces
- **200x performance improvement**: 20k resources → 100 resources for typical namespace

✅ **Advanced Namespace Selector UI**
- **Searchable dropdown**: Type to filter namespaces in real-time
- **Keyboard navigation**: Arrow keys (↓/↑), Enter to select, Escape to close
- **Auto-scroll**: Highlighted option automatically scrolls into view
- **Visual feedback**: Yellow highlight for keyboard focus, distinct from active state
- **localStorage persistence**: Remembers last selected namespace across sessions
- **Smart reconnection**: Clears state and reconnects WebSocket when namespace changes
- **Empty state detection**: Auto-switches to "All Namespaces" if selected namespace is deleted

✅ **Icon Consistency Improvements**
- **Feather Icons integration**: Replaced all emojis with consistent line-based icons
- **Events button**: `📋` → `<i data-feather="activity">`
- **Topology placeholder**: `🚧` → `<i data-feather="git-branch">`
- **Empty state**: `📭` → `<i data-feather="inbox">`
- **Detail panel tabs**: Added icons to Overview (info) and YAML (code) tabs
- **Professional appearance**: Unified visual language matching glassmorphic theme

### Technical Implementation Details

**Backend: Namespace Filtering (`internal/server/websocket.go`)**
```go
// Parse namespace from query parameter
namespace := r.URL.Query().Get("namespace")
if namespace == "" || namespace == "all" {
    namespace = "" // Empty string = all namespaces
}

client := &Client{
    conn:      conn,
    send:      make(chan k8s.ResourceEvent, 10000),
    hub:       s.hub,
    namespace: namespace,
}

// Send filtered snapshot
snapshot := s.watcher.GetSnapshotFiltered(namespace)

// Filter broadcasts per client
if client.namespace != "" && event.Resource.Namespace != client.namespace {
    continue // Skip this client
}
```

**Frontend: Searchable Dropdown with Keyboard Navigation (`index.html`)**
```javascript
function handleNamespaceKeyboard(e) {
  if (e.key === 'ArrowDown') {
    highlightedNamespaceIndex = Math.min(highlightedNamespaceIndex + 1, filtered.length - 1);
    scrollToHighlighted();
  } else if (e.key === 'Enter') {
    setNamespace(filtered[highlightedNamespaceIndex]);
    closeNamespaceDropdown();
  } else if (e.key === 'Escape') {
    closeNamespaceDropdown();
  }
}
```

**Icon Consistency: Feather Icons**
```html
<!-- Feather Icons CDN -->
<script src="https://unpkg.com/feather-icons"></script>

<!-- Usage in HTML -->
<i data-feather="activity"></i>  <!-- Events -->
<i data-feather="inbox"></i>     <!-- Empty state -->
<i data-feather="info"></i>      <!-- Overview tab -->
<i data-feather="code"></i>      <!-- YAML tab -->

<!-- Render icons -->
<script>feather.replace();</script>
```

### Performance Characteristics

**Namespace Filtering Impact (21,867 resource cluster)**:
- Full snapshot: 21,867 resources → 50MB transfer → 3-5s load
- Filtered (default namespace): 100 resources → 250KB transfer → <1s load
- **Network reduction**: 200x smaller payload
- **Memory reduction**: Client holds only filtered resources
- **Update efficiency**: Only receives events for selected namespace

**Keyboard Navigation**:
- Instant highlight updates (no lag)
- Smooth auto-scroll with `scrollIntoView({ block: 'nearest', behavior: 'smooth' })`
- Works seamlessly with real-time search filtering

---

## 6.6. Phase 3 Continued (Performance Optimization - 2025-11-28)

### What Was Built

✅ **Lazy Loading by Resource Type**
- **REST API for instant stats**: `/api/stats` endpoint returns counts without streaming
- **Resource type filtering**: WebSocket filters by type before transmission (`?type=Pod`)
- **Server-side filtering**: `GetSnapshotFilteredByType(namespace, resourceType)` method
- **Dual filtering**: Combines namespace + type filtering (e.g., `?namespace=yatai&type=Pod`)
- **40-100x network reduction**: Only send filtered resources (e.g., 171 Deployments vs 3,037 total)

✅ **Bug Fixes**
- **Stats overwriting**: Removed client-side counting, now always fetch from `/api/stats`
- **Automatic namespace switching**: Removed unwanted behavior that violated user expectations

### Technical Implementation

**Backend Changes**:
```go
// GET /api/stats?namespace=xxx
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

// WebSocket with dual filtering
func (w *Watcher) GetSnapshotFilteredByType(namespace string, resourceType string) []ResourceEvent {
  // Filter by namespace first
  var resources []*types.Resource
  if namespace == "" {
    resources = w.cache.List()
  } else {
    resources = w.cache.ListByNamespace(namespace)
  }

  // Then filter by type
  filtered := []*types.Resource{}
  for _, r := range resources {
    if resourceType == "" || r.Type == resourceType {
      filtered = append(filtered, r)
    }
  }
  return events
}
```

**Frontend Changes**:
```javascript
// Fetch stats instantly (no streaming)
async function fetchAndDisplayStats() {
  const nsParam = currentNamespace === 'all' ? '' : `?namespace=${currentNamespace}`;
  const response = await fetch(`/api/stats${nsParam}`);
  const counts = await response.json();

  // Update stat cards (instant <100ms)
  document.getElementById('stat-total').textContent = counts.total || 0;
  // ... other stats
}

// Reconnect with new filter (lazy load)
function reconnectWithFilter() {
  clearResources();
  fetchAndDisplayStats().then(() => {
    connect(); // WebSocket with ?type= parameter
  });
}

// Build WebSocket URL with dual filtering
const params = [];
if (currentNamespace !== 'all') params.push(`namespace=${currentNamespace}`);
if (currentFilter !== 'all') params.push(`type=${currentFilter}`);
const wsUrl = `/ws${params.length > 0 ? '?' + params.join('&') : ''}`;
```

### Performance Impact

**Network Transfer Reduction**:
- Pods: 1,218 resources (60% reduction vs all types)
- Deployments: 171 resources (94% reduction)
- Services: 575 resources (81% reduction)
- Ingress: 72 resources (98% reduction)

**Load Time Improvements**:
- Stats loading: <100ms (vs 2-5s for full snapshot)
- Filter switching: <1s (reconnect + filtered snapshot)
- 20k cluster becomes manageable with instant stats + lazy lists

---

## 6.7. Frontend Architecture (Modular Data-Centric Design)

### What Was Refactored

The frontend evolved from a single-file prototype into a **modular, data-centric architecture** following ES6 module patterns:

✅ **File Structure** (`internal/server/static/`)
```
├── index.html        # HTML structure only (8.7KB)
├── style.css         # All styles (16.3KB)
├── app.js            # Main application logic (26.9KB)
├── config.js         # Configuration constants (717 bytes)
├── state.js          # State management (962 bytes)
├── ws.js             # WebSocket connection management (1.9KB)
└── dropdown.js       # Reusable dropdown component (4.9KB)
```

### Module Responsibilities

**config.js** - Central configuration
- Resource types list (`RESOURCE_TYPES`)
- API endpoint paths (`API_PATHS`)
- Relationship type definitions (`RELATIONSHIP_TYPES`)
- localStorage keys (`LOCAL_STORAGE_KEYS`)
- Constants (events limit, etc.)

**state.js** - State management
- `createInitialState()` - Initialize application state
- `resetForNewConnection()` - Clear state for reconnections
- State structure: resources, filters, UI state, WebSocket state, log state

**ws.js** - WebSocket lifecycle
- `createResourceSocket()` - Factory for WebSocket manager
- Connection management with auto-reconnect
- Connection ID tracking to prevent race conditions
- Snapshot completion detection
- Configurable handlers (onOpen, onMessage, onClose, onError)

**dropdown.js** - Reusable component
- Custom web component (`<k8v-dropdown>`)
- Searchable dropdown with keyboard navigation
- Used for namespace and container selection
- Emits standard change events

**app.js** - Application orchestration
- Main `App` class coordinating all functionality
- UI event handling and rendering
- Resource list management with incremental updates
- Detail panel and logs viewer
- Search and filtering logic
- Namespace and resource type switching

**index.html** - Markup only
- Semantic HTML structure
- No inline JavaScript (all in modules)
- Minimal inline styles (button styling only)

**style.css** - Complete styling
- Glassmorphic dark theme
- Responsive grid layouts
- Component styles (cards, dropdowns, panels)
- Animation and transition definitions

### Architecture Benefits

1. **Separation of Concerns**: Each file has a single, clear responsibility
2. **Testability**: Modules can be tested independently
3. **Maintainability**: Easy to locate and modify specific functionality
4. **Reusability**: Components like dropdown can be reused
5. **Code Organization**: Logical grouping by function, not file size
6. **Performance**: Browser can cache individual modules
7. **Developer Experience**: Easier to navigate and understand codebase

### Data Flow

```
User Action
    ↓
app.js (Event Handler)
    ↓
state.js (Update State)
    ↓
ws.js (Send to Server) OR app.js (Update UI)
    ↓
app.js (Render Changes)
    ↓
DOM Update
```

### Key Design Patterns

- **Module Pattern**: ES6 imports/exports for clean dependencies
- **Factory Pattern**: `createResourceSocket()`, `createInitialState()`
- **Observer Pattern**: WebSocket handlers, event listeners
- **Component Pattern**: Custom web component (dropdown)
- **Singleton Pattern**: Single App instance manages global state

---

## 7. Phase 1 Complete (Production Backend + Minimal Frontend)

### What Was Built

✅ **Go Project Setup**
- Production directory structure (cmd/, internal/types, internal/k8s, internal/server)
- Dependencies: client-go v0.31.0, gorilla/websocket, sigs.k8s.io/yaml
- Single binary build with embedded web assets

✅ **Data Model Implementation**
- Complete Resource struct with 8 relationship types
- **Generic relationship system** with RelationshipType enum
- Thread-safe ResourceCache
- Bidirectional relationship computation

✅ **Kubernetes Integration**
- Kubeconfig loading (supports both local and in-cluster)
- SharedInformerFactory pattern for efficiency
- Watcher with event handlers for 7 resource types
- Transformers: Pod, Deployment, ReplicaSet, Service, Ingress, ConfigMap, Secret

✅ **WebSocket Streaming**
- HTTP/WebSocket server with hub pattern
- Initial snapshot + incremental updates
- Panic recovery for large clusters (2000+ resources tested)
- Channel buffer optimization (10,000 events)

✅ **Minimal Frontend**
- Enhanced table view with all resource types
- Detail panel with Overview and YAML tabs
- **Clickable bidirectional relationship navigation**
- Real-time updates via WebSocket
- Health indicators (green/yellow/red)
- Console logging for development (color-coded events)

✅ **Key Improvements**
- **Generic Relationship Computation:**
  - Reduced from 4 specific functions to 1 generic `FindReverseRelationships()`
  - Type-safe with RelationshipType enum
  - Scalable: add new relationships without new functions
  - Cleaner, more maintainable code

### Binary Details

- **Size:** 62MB
- **Command:** `./k8v` or `./k8v -port 8080`
- **Handles:** 2000+ resources tested in production
- **Performance:** Real-time updates < 500ms latency

### What's Next (Phase 3 and Beyond)

**Phase 3 Priorities:**

1. **Frontend Performance Optimization**
   - Virtual scrolling/pagination for large resource lists (1000+ resources)
   - Lazy rendering to prevent lag with many resources
   - Debounced updates during rapid event streams
   - Memory optimization for long-running sessions

2. **Namespace Filtering**
   - Namespace selector dropdown
   - Filter resources by namespace(s)
   - Reduce frontend load by hiding unwanted namespaces
   - Persist namespace filter preference

3. **Pod Logs Viewer**
   - Stream logs via WebSocket
   - Log viewer in detail panel
   - Follow mode for live logs
   - Log search and filtering
   - Container selection for multi-container pods

4. **Enhanced YAML View**
   - Syntax highlighting for YAML
   - Clickable resource references in YAML (e.g., click ConfigMap name → navigate to ConfigMap)
   - Highlight relationship fields (ownerReferences, selectors, etc.)
   - Copy specific YAML sections

5. **Search Functionality**
   - Search by resource name
   - Filter by labels and annotations
   - Quick navigation to specific resources

**Future Enhancements:**

6. **Additional Kubernetes Resources**
   - StatefulSets, DaemonSets, Jobs, CronJobs
   - PersistentVolumes, PersistentVolumeClaims
   - NetworkPolicies, ServiceAccounts, Roles, RoleBindings
   - Custom Resource Definitions (CRDs)

7. **Multi-Cluster Support**
   - Context switching UI
   - Multiple cluster views
   - Cross-cluster comparison

8. **Advanced Features**
   - Topology graph view (relationship visualization)
   - Resource editing (kubectl apply)
   - YAML export/download
   - Events timeline with filtering
   - Dark/light theme toggle

---

## 7. Quick Reference

| Aspect | Details |
|--------|---------|
| **Current Stage** | ✅ Phase 3 Complete - Advanced filtering, performance optimization, modular frontend |
| **Tech Stack** | Go backend + Modular ES6 frontend (no frameworks) |
| **Frontend Architecture** | 7 ES6 modules: app.js, config.js, state.js, ws.js, dropdown.js, style.css, index.html |
| **Backend Language** | Go 1.23+ with client-go v0.31.0 |
| **Communication** | WebSocket (real-time bidirectional updates) |
| **Target User** | Developers with kubectl/kubeconfig access |
| **Deployment Model** | Single 62MB binary (`./k8v` command) |
| **Similar Tools** | k9s (TUI), Lens (Electron), kubectl proxy (proxy-only) |
| **Core Resources** | Pods, Deployments, Services, Ingress, ReplicaSets, ConfigMaps, Secrets |
| **MVP Status** | ✅ Resource visualization, ✅ Relationships, ✅ Live streaming, ✅ Pod logs |
| **UI** | ✅ Complete - Modular architecture with incremental updates |
| **POC** | ✅ Complete (k8v-poc validates streaming architecture) |
| **Data Model** | ✅ Complete - Generic relationship system implemented |
| **Production Backend** | ✅ Complete - Informers, WebSocket hub, transformers |
| **K8s Integration** | ✅ Complete - Watches 7 resource types, handles 20k+ resources |
| **Scale Tested** | ✅ 21,867 resources in production cluster |

---

## 8. Document References

- **README.md** - Public-facing documentation with quickstart and roadmap
- **CLAUDE.md** - This file - comprehensive project context and architecture overview
- **IDEAS.md** - Original feature requirements and user vision
- **DESIGN.md** - Technical design and architecture decisions
- **DATA_MODEL.md** - Complete data model with relationship system
- **CHANGELOG.md** - Version history and changes across all phases
- **PHASE2_SUMMARY.md** - Phase 2 technical achievements and lessons learned
- **PHASE3_PLAN.md** - Phase 3 implementation plan with detailed task breakdowns
- **index.html** - Original UI prototype demonstrating UX vision
- **k8v-poc/** - Minimal POC validating streaming architecture

---

## 9. Key Insights

1. **UI is Already Done:** The prototype is production-ready. No need to redesign or rebuild the frontend - just extract and modularize it.

2. **Focus on Backend:** The main implementation work is building the Go backend to connect to real Kubernetes clusters.

3. **Real-Time is Priority #1:** The user specifically requested real-time updates via K8s watch API. This should be core to the architecture.

4. **Start Simple:** Phase 1 (embedded server) validates the approach before tackling complex K8s integration.

5. **Pod Logs are MVP:** Unlike typical dashboards, logs viewing is a must-have for the initial release.

6. **Topology is Secondary:** While impressive in the prototype, graph topology is acknowledged as a hard problem and not critical for MVP.

7. **Extensibility Matters:** The data model should make it easy to add new K8s resource types in the future.

8. **Single Binary FTW:** Following the Go/K8s ecosystem pattern of single binary distribution simplifies everything.

9. **Relationships are Core:** The data model is relationship-first. Resource connections (Deployment → ConfigMap, Service → Pods) are as important as the resources themselves.

10. **POC Validated Approach:** The minimal POC proved that Watch API + WebSocket + simple UI works. No need to guess - the architecture is validated.

11. **Incremental DOM Updates are Critical:** With large clusters, full list rerenders cause flickering and poor UX. Incremental updates (add/modify/delete single elements) are essential for smooth real-time updates.

12. **WebSocket Race Conditions at Scale:** Async snapshot sending creates race conditions where channels close before snapshot completes. Synchronous snapshot transmission before starting pumps prevents panics.

13. **Progress Feedback for Large Clusters:** When handling 20k+ resources, progress logging is essential for understanding what's happening during initial load. Silent waits create uncertainty.

14. **Simplicity Wins:** Removing the "ALL" filter simplified the UX and code. Users naturally want to focus on specific resource types, not see everything mixed together.

15. **Alphabetical > Grouped Sorting:** Within a filtered view (e.g., Pods only), alphabetical sorting by name is more useful than grouping by namespace then name. Users know what they're looking for.

16. **Modular Frontend Architecture:** Splitting the frontend into ES6 modules (config, state, ws, app) dramatically improves maintainability and developer experience. Each module has a single responsibility, making it easy to locate and modify specific functionality. The data-centric approach (separating state and configuration from logic) makes the codebase more testable and extensible.

---

## 10. Maintaining This Document

**⚠️ IMPORTANT: Keep CLAUDE.md Updated**

This document serves as the primary context file for understanding the project. As development progresses and decisions change, **update this document** to reflect the current state.

### When to Update CLAUDE.md

Update this file whenever:
- **Direction changes:** New architectural decisions or approach pivots
- **Priorities shift:** MVP scope changes or feature priorities reordered
- **Key decisions made:** Important technical choices that differ from DESIGN.md
- **Significant progress:** Major milestones completed (e.g., Phase 1 done)
- **New insights:** Discoveries that change understanding of the problem
- **Scope changes:** Features added or removed from MVP

### What to Update

When making changes, update the relevant sections:
- **Quick Reference table** - Change status indicators (❌ → ✅)
- **Next Steps** - Mark completed tasks, add new ones
- **Key Insights** - Add new learnings or remove outdated assumptions
- **Index.html POC Findings** - If prototype is modified
- **IDEAS.md/DESIGN.md summaries** - If those documents change significantly

### How to Update

1. **Keep it concise** - This is a summary document, not a detailed spec
2. **Update inline** - Modify existing sections rather than appending
3. **Remove outdated info** - Delete decisions that were reversed
4. **Add date markers** - For significant changes, note "Updated: YYYY-MM-DD"
5. **Preserve context** - Keep enough history to understand why decisions were made

### Example Updates

```markdown
# Before
**Backend Status** | ❌ Not started (need to build Go server)

# After
**Backend Status** | ✅ Phase 1 complete (embedded server working)
```

```markdown
# Adding a new insight
9. **Mermaid Too Heavy:** Discovered Mermaid.js is 1MB minified.
   Considering D3.js or custom SVG for topology view instead. (Updated: 2025-01-15)
```

**Goal:** Anyone (including future Claude sessions) should be able to read this document and understand the current project state accurately. If CLAUDE.md conflicts with reality, update CLAUDE.md.

---

**Last Updated:** 2025-11-29 - Frontend Refactored: Split into modular ES6 architecture (config.js, state.js, ws.js, app.js, dropdown.js). Added fullscreen mode for detail panel. Improved UI consistency with Feather icons for all buttons.
