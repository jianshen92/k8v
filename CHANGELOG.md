# Changelog

All notable changes to the k8v project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

---

## [Phase 3] - 2025-11-27

### 🚧 Phase 3 In Progress - Namespace Filtering & UI Polish

This phase focuses on advanced filtering capabilities and UI consistency improvements.

### Added
- **Server-side namespace filtering**: Filter resources by namespace before sending to client
  - `/api/namespaces` endpoint returns list of available namespaces
  - WebSocket query parameter support (`?namespace=xxx`)
  - Broadcast-level filtering in Hub (only sends matching resources to clients)
  - `GetNamespaces()` method in watcher.go (extracts unique namespaces, sorted alphabetically)
  - `GetSnapshotFiltered(namespace)` method in watcher.go
  - `Client.namespace` field for per-client filtering
- **Searchable namespace dropdown**: Advanced UI component for namespace selection
  - Dropdown with input field for real-time search/filtering
  - Live filtering as user types
  - Active state indicator showing current selection
  - Click outside to close
- **Keyboard navigation**: Full keyboard accessibility for namespace dropdown
  - Arrow Down/Up to navigate options
  - Enter to select highlighted option
  - Escape to close dropdown
  - Auto-scroll to keep highlighted option visible
  - Visual highlight state for keyboard navigation
- **Feather Icons integration**: Replaced all emojis with consistent icon library
  - Events button: `📋` → activity icon
  - Topology placeholder: `🚧` → git-branch icon
  - Empty state: `📭` → inbox icon
  - Detail panel tabs: info and code icons
  - Cohesive design language matching glassmorphic theme
- **localStorage persistence**: Remember last selected namespace across sessions
  - Automatic restore on page load
  - Graceful handling of deleted namespaces (fallback to "All Namespaces")

### Changed
- **Namespace selection**: From buttons to searchable dropdown (better UX for clusters with many namespaces)
- **Icon design**: Consistent Feather Icons throughout (no more emoji inconsistency)

### Performance
- **200x network reduction**: Filtering single namespace (~100 resources, 250KB) vs all namespaces (~20k resources, 50MB)
- **Sub-second load times**: Filtered namespace loads in <1s vs 3-5s for full cluster
- **Instant namespace switching**: Reconnect with new filter in <1s

### Technical Details

**Files Modified**:
- `internal/k8s/watcher.go`: Added `GetNamespaces()`, `GetSnapshotFiltered()`, `sort` import
- `internal/server/handlers.go`: Added `handleNamespaces()` endpoint
- `internal/server/server.go`: Registered `/api/namespaces` route
- `internal/server/websocket.go`: Added namespace query param parsing, per-client filtering
- `internal/server/static/index.html`: Searchable dropdown UI, keyboard navigation, Feather Icons, localStorage persistence

**Key Code Changes**:
```go
// Namespace filtering in watcher
func (w *Watcher) GetSnapshotFiltered(namespace string) []ResourceEvent {
  var resources []*types.Resource
  if namespace == "" {
    resources = w.cache.List()
  } else {
    resources = w.cache.ListByNamespace(namespace)
  }
  // ... transform to events
}

// Broadcast-level filtering in Hub
case event := <-h.broadcast:
  for client := range h.clients {
    if client.namespace != "" && event.Resource.Namespace != client.namespace {
      continue // Skip non-matching resources
    }
    client.send <- event
  }
```

```javascript
// Searchable dropdown with keyboard navigation
function handleNamespaceKeyboard(e) {
  if (e.key === 'ArrowDown') {
    highlightedNamespaceIndex = Math.min(highlightedNamespaceIndex + 1, filtered.length - 1);
    scrollToHighlighted();
  } else if (e.key === 'ArrowUp') {
    highlightedNamespaceIndex = Math.max(highlightedNamespaceIndex - 1, 0);
    scrollToHighlighted();
  } else if (e.key === 'Enter') {
    setNamespace(filtered[highlightedNamespaceIndex]);
  }
}

// localStorage persistence
let currentNamespace = localStorage.getItem('k8v-namespace') || 'all';
function setNamespace(ns) {
  localStorage.setItem('k8v-namespace', ns);
  currentNamespace = ns;
  reconnectWithNamespace();
}
```

---

## [Phase 2] - 2025-11-27

### ✅ Phase 2 Complete - Production-Ready Application

This phase focused on UI refinement, performance optimization, and stability improvements for production use.

### Added
- **Incremental DOM updates**: Only affected resources are updated in the UI (no more full list rerenders)
  - `ADDED` events insert pills in correct sorted position
  - `MODIFIED` events replace only the changed pill
  - `DELETED` events remove only the deleted pill
- **Progress logging**: Shows transmission progress every 1000 resources during snapshot
- **Alphabetical sorting**: Resources now sorted by name (A-Z) instead of grouped by type
- **Compact statistics cards**: Reduced from 220px to 140px minwidth for better space utilization

### Changed
- **Removed "ALL" filter tab**: Users now view specific resource types only
- **Default view**: Now starts with "Pods" view instead of "All"
- **Resource pill styling**: Fixed height issues with `min-height: 56px` and `align-items: flex-start`
- **Statistics design**: Smaller, more compact cards with reduced padding and font sizes

### Fixed
- **WebSocket race condition**: Snapshot now sent synchronously before starting read/write pumps
  - Eliminates "send on closed channel" panics
  - Direct `WriteJSON` calls bypass buffered channel
  - Proper cleanup if client disconnects during snapshot
- **UI flickering**: Incremental updates prevent visual artifacts and preserve scroll position
- **Filter awareness**: Incremental updates respect current filter state

### Performance
- Successfully tested with **21,867 resources** in production cluster
- Snapshot transmission: ~2-5 seconds (with progress logging)
- Update latency: < 100ms for incremental updates
- Memory usage: Stable, no leaks observed
- No WebSocket panics or crashes

### Technical Details

**Files Modified**:
- `internal/server/static/index.html`: Incremental DOM updates, filter changes, sorting
- `internal/server/websocket.go`: Synchronous snapshot, race condition fix, progress logging

**Key Code Changes**:
```javascript
// Incremental DOM updates
function updateResourceInList(resourceId, eventType) {
  // Only update single resource element, not entire list
}
```

```go
// Synchronous snapshot transmission
snapshot := s.watcher.GetSnapshot()
for i, event := range snapshot {
  err := conn.WriteJSON(event) // Direct write, no race condition
  if err != nil {
    return // Clean exit on error
  }
}
// Start pumps AFTER snapshot complete
go client.writePump()
go client.readPump()
```

---

## [Phase 1] - 2025-11-26

### ✅ Phase 1 Complete - Production Backend + Minimal Frontend

This phase established the core backend architecture and basic frontend integration.

### Added
- **Go project structure**: Production-ready directory layout
  - `cmd/k8v/main.go`: CLI entry point
  - `internal/types/`: Shared type definitions
  - `internal/k8s/`: Kubernetes client, watcher, transformers
  - `internal/server/`: HTTP/WebSocket server
- **Kubernetes integration**:
  - Kubeconfig loading (local and in-cluster)
  - SharedInformerFactory pattern for efficiency
  - Watchers for 7 resource types: Pod, Deployment, ReplicaSet, Service, Ingress, ConfigMap, Secret
- **Generic relationship system**:
  - 8 relationship types (OwnedBy, Owns, DependsOn, UsedBy, Exposes, ExposedBy, RoutesTo, RoutedBy)
  - Bidirectional relationship computation
  - Type-safe RelationshipType enum
  - Single `FindReverseRelationships()` function replaces 4 specific functions
- **WebSocket streaming**:
  - Hub pattern for managing multiple clients
  - Initial snapshot + incremental updates
  - Panic recovery for large clusters
  - 10,000 event channel buffer
- **Resource transformers**:
  - `TransformPod()`, `TransformDeployment()`, `TransformReplicaSet()`
  - `TransformService()`, `TransformIngress()`, `TransformConfigMap()`, `TransformSecret()`
  - Health computation from K8s status
  - YAML serialization
- **Minimal frontend**:
  - Table view with all resource types
  - Detail panel with Overview and YAML tabs
  - Clickable bidirectional relationship navigation
  - Real-time updates via WebSocket
  - Health indicators (green/yellow/red)
  - Console logging for development

### Technical Details
- **Binary size**: 62MB
- **Dependencies**: client-go v0.31.0, gorilla/websocket, sigs.k8s.io/yaml
- **Go version**: 1.23+
- **Architecture**: Single binary with embedded web assets

---

## [POC] - 2025-11-25

### ✅ POC Validation Complete

Minimal proof-of-concept to validate the technical approach.

### Added
- Basic K8s watch API integration
- WebSocket streaming to browser
- Simple table UI with Pods, Deployments, ReplicaSets
- ADD/MODIFY/DELETE event handling

### Validated
- ✅ K8s watch API works correctly
- ✅ WebSocket streaming to browser works
- ✅ Real-time UI updates work (< 1 second latency)
- ✅ Simple table UI successfully displays resources

### Learnings
- Direct Watch API (not Informers) is simple and works for POC
- gorilla/websocket handles concurrent writes (need mutex)
- Browser WebSocket API is straightforward
- k8s.io/client-go requires Go 1.23 (use v0.31.0)

---

## [Prototype] - 2025-11-24

### Initial Prototype

Complete HTML/CSS/JS prototype demonstrating the vision.

### Features
- Dark-themed glassmorphic UI
- Dashboard view with statistics cards
- Filterable resource lists (All, Pods, Deployments, Services, Ingress, ReplicaSets, ConfigMaps, Secrets)
- Detail panel with Overview, YAML, and Relationships tabs
- Topology view placeholder (Mermaid.js)
- Events timeline
- Mock data for 3 namespaces

### Limitations
- No Kubernetes API connection (mock data only)
- No real-time updates
- No backend server
- No CLI tool wrapper
