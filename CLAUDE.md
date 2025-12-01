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
- **Stage:** ✅ Phase 3 Nearly Complete - Core Features Implemented
- **What exists:**
  - Production Go backend with Informers, WebSocket streaming, and generic relationship system
  - Polished web UI with real-time updates and optimized rendering
  - Bidirectional relationship navigation
  - **Vim-like command mode**: Keyboard-first navigation with `:` command palette (2025-12-02)
  - **Namespace filtering**: Server-side filtering with searchable dropdown UI
  - **Resource type lazy loading**: Instant stats + filtered WebSocket snapshots
  - **Pod logs viewer**: Real-time log streaming with 6 configurable modes (Head, Tail, Last 5m/15m/500/1000) and keyboard shortcuts (1-6)
  - **Search functionality**: Quick search by name with keyboard shortcut (/)
  - **Multi-context support**: Switch between clusters without restarting
  - **Data-driven UI**: Log modes and commands auto-generated from configuration data
  - **Complete keyboard navigation**: `:`, `/`, `d`, `1-6`, `Esc` shortcuts with hierarchical handling
  - **Performance optimized**: 40-100x network reduction for large clusters
  - Single 62MB binary (k8v) ready to use
  - Tested with large production clusters (21,000+ resources)
- **Next:** Enhanced YAML view, additional resource types, and virtual scrolling

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

## 6. Phases Summary

### Phase 2 (Production-Ready - 2025-11-27)
- Incremental DOM updates (no flickering, preserved scroll)
- WebSocket race condition fix (synchronous snapshot, handles 21k+ resources)
- UI polish (compact stats, alphabetical sorting, removed ALL filter)

### Phase 3 Early Progress (2025-11-27/28/30)
**Namespace Filtering**: Server-side filtering (200x network reduction), searchable dropdown with keyboard navigation
**Performance**: Lazy loading by resource type (40-100x reduction), instant stats API
**Context Switching**: Multi-cluster support with reactive state management
**Frontend Architecture**: Modular ES6 modules (config.js, state.js, ws.js, dropdown.js)
**Logging**: Server logging to logs/k8v.log with request/WebSocket tracking

---

## 6.9. Phase 3 Continued (Core MVP Features - 2025-11-30)

### What Was Built

Three major features completed, fulfilling the core MVP requirements from IDEAS.md.

✅ **Pod Logs Viewer**
- Real-time log streaming via WebSocket
- Container selection for multi-container pods
- Auto-select first container (saves user a click)
- Integrated into detail panel as "Logs" tab
- Connection state indicators (loading, error, closed)
- Log hub pattern for managing concurrent streams

✅ **Search Functionality**
- Keyboard shortcut `/` to activate (vim/GitHub style)
- Real-time filtering as user types
- Search by resource name (case-insensitive)
- Clear button and Escape key to exit
- Respects namespace and resource type filters

✅ **Multi-Context Support**
- Context dropdown in header with searchable UI
- Switch between clusters without restarting
- Reactive state management with `SYNC_STATUS` events
- Loading overlay during cache sync
- Automatic namespace reset on context switch

### Technical Implementation

**Pod Logs Viewer**

Backend files:
- `internal/server/logstream.go` - WebSocket handler for log streaming
- `internal/server/loghub.go` - Hub pattern for managing multiple log streams
- `internal/k8s/client.go` - `StreamPodLogs()` method using K8s client

Key features:
- Tail 100 lines on connection
- Follow mode for real-time streaming
- Message types: `LOG_LINE`, `LOG_END`, `LOG_ERROR`
- Clean disconnection when switching pods

Frontend integration (`app.js`):
```javascript
loadLogs() {
  const container = this.containerDropdown.getValue();
  const wsUrl = `${wsProtocol}//${host}/ws/logs?namespace=${ns}&pod=${pod}&container=${container}`;

  this.state.log.socket = new WebSocket(wsUrl);
  this.state.log.socket.onmessage = (event) => {
    const message = JSON.parse(event.data);
    if (message.type === 'LOG_LINE') {
      this.appendLogLine(message.line);
    }
  };
}

// Auto-select first container
if (containers.length > 1) {
  this.containerDropdown.setValue(containers[0].name);
  if (this.state.ui.activeDetailTab === 'logs') {
    this.loadLogs();
  }
}
```

**Search Functionality**

Frontend implementation (`app.js`):
```javascript
// Keyboard shortcut
handleGlobalKeydown(event) {
  if (event.key === '/' && !isInputFocused) {
    event.preventDefault();
    this.activateSearch();
  }
  if (event.key === 'Escape' && this.state.ui.searchActive) {
    this.clearSearch();
  }
}

// Real-time filtering
handleSearchInput(event) {
  this.state.filters.search = event.target.value.toLowerCase().trim();
  this.renderResourceList(); // Filters applied here
}

// Filter integration
renderResourceList() {
  const list = Array.from(this.state.resources.values())
    .filter(r => {
      if (r.type !== this.state.filters.type) return false;
      if (this.state.filters.search && !r.name.toLowerCase().includes(this.state.filters.search)) {
        return false;
      }
      return true;
    });
}
```

**Multi-Context Support**

Backend files:
- `internal/server/app.go` - `SwitchContext()` method, `SYNC_STATUS` broadcasting
- `internal/server/handlers.go` - `/api/contexts`, `/api/context/current`, `/api/context/switch` endpoints

Key architecture:
- `App.context` is the single source of truth (not kubeconfig's "current" field)
- `SwitchContext()` stops old watcher, starts new watcher, waits for cache sync
- Broadcasts `SYNC_STATUS` events during sync lifecycle

Frontend integration (`app.js`):
```javascript
// Get actual backend context (not kubeconfig)
async fetchCurrentContext() {
  const response = await fetch('/api/context/current');
  const data = await response.json();
  this.contextDropdown.setValue(data.context);
}

// Switch context
async switchContext(newContext) {
  await fetch(`/api/context/switch?context=${newContext}`, { method: 'POST' });

  // Reset state
  resetForNewConnection(this.state);
  this.renderResourceList();

  // Reset namespace to "all"
  this.state.filters.namespace = 'all';
  localStorage.setItem('k8v-namespace', 'all');
  this.namespaceDropdown.setValue('all');

  // Reconnect (SYNC_STATUS will trigger data refresh)
  this.wsManager.connect();
}

// Reactive data refresh
handleSyncStatus(syncEvent) {
  if (syncEvent.synced && !syncEvent.syncing) {
    this.fetchNamespaces();     // Get new context's namespaces
    this.fetchAndDisplayStats(); // Get new context's stats
  }
}
```

### UX Improvements

1. **Pod Logs**:
   - No need to leave UI or use `kubectl logs`
   - First container auto-selected for convenience
   - Real-time streaming for debugging

2. **Search**:
   - Keyboard-first navigation (`/` shortcut)
   - Instant feedback as you type
   - Works seamlessly with other filters

3. **Context Switching**:
   - No restart required
   - Loading overlay shows sync progress
   - Fresh data guaranteed by reactive updates

### Architecture Patterns

**Reactive State Management**:
- Backend broadcasts state changes via WebSocket events
- Frontend reacts to events, never polls or guesses
- Single source of truth on backend, UI is always in sync

**Event-Driven Data Flow**:
```
User switches context
  ↓
POST /api/context/switch
  ↓
Backend: Stop old watcher → Start new → WaitForCacheSync
  ↓
Broadcast: SYNC_STATUS syncing=true → Show loading overlay
  ↓
Backend: Cache synced
  ↓
Broadcast: SYNC_STATUS synced=true
  ↓
Frontend: handleSyncStatus() → Fetch namespaces & stats
  ↓
UI updated with fresh data
```

**Separation of Concerns**:
- Backend owns K8s state and cluster connections
- Frontend owns UI state and user interactions
- WebSocket bridges the two with typed messages

### Impact on MVP

These features complete the core MVP requirements from IDEAS.md:
- ✅ Resource Visualization (Phase 1)
- ✅ Live Streaming (Phase 1)
- ✅ Relationship Mapping (Phase 1)
- ✅ Pod Logs Viewing (Phase 3) ← **Completed**
- ✅ Search Functionality (Phase 3) ← **Completed**
- ✅ Multi-Context Support (Phase 3) ← **Completed**
- ✅ Top-Tier UI (Phase 2)
- ✅ Extensible Data Model (Phase 1)

**Remaining for full Phase 3**:
- Enhanced YAML view (syntax highlighting, clickable refs)
- Virtual scrolling for massive resource lists
- Enhanced search (by labels/annotations)

---

## 6.10. Phase 3 Continued (Pod Log Viewer Enhancements - 2025-12-01)

### What Was Built

Major improvements to the pod log viewer with configurable viewing modes and data-driven architecture.

✅ **Log Viewing Modes**
- Six configurable modes with keyboard shortcuts (1-6)
- **Head** (1): Show first 500 lines from beginning
- **Tail** (2): Show last 100 lines then follow (default)
- **Last 5m** (3): Show logs from last 5 minutes with follow
- **Last 15m** (4): Show logs from last 15 minutes with follow
- **Last 500** (5): Show last 500 lines with follow
- **Last 1000** (6): Show last 1000 lines with follow

✅ **Data-Driven Architecture**
- LOG_MODES migrated from object to array with metadata (id, label, hotkey)
- Moved from app.js to config.js (separation of data and logic)
- Buttons auto-generated from data during initialization
- Hotkey handlers generated dynamically from data
- Click handlers attached inline during creation

✅ **HeadLines Support**
- Backend counting mechanism to show first N lines
- K8s API doesn't support this natively (TailLines only gets last N)
- Implemented line counting in StreamPodLogs()
- Stops streaming after reaching head limit

### Technical Implementation

**Backend Changes**:
- `internal/k8s/logs.go`:
  - Added `HeadLines *int64` field to LogOptions struct
  - Added line counting logic with early termination
  - Sends LOG_END message when head limit reached
- `internal/server/logs.go`:
  - Parse `headLines` query parameter
  - Pass to streaming handler

**Frontend Changes**:
- `internal/server/static/config.js`:
  ```javascript
  export const LOG_MODES = [
    { id: 'head', label: 'Head', hotkey: '1', headLines: 500, tailLines: null, sinceSeconds: null, follow: false },
    { id: 'tail', label: 'Tail', hotkey: '2', headLines: null, tailLines: 100, sinceSeconds: null, follow: true },
    // ... 4 more modes
  ];
  ```

- `internal/server/static/app.js`:
  - Added `getLogMode(modeId)` helper to find mode by id
  - Added `renderLogModeButtons()` to generate buttons from data
  - Updated `setLogMode()` to use helper
  - Updated `loadLogs()` to add headLines parameter
  - Updated `handleGlobalKeydown()` to dynamically find mode by hotkey
  - Removed hardcoded hotkey map

- `internal/server/static/index.html`:
  - Removed all hardcoded button elements
  - Added empty container `<div id="logs-mode-buttons">` for dynamic generation

### Architecture Benefits

**Data-Centric Principles Applied**:
1. **Single Source of Truth**: All mode configuration in one place (config.js)
2. **Zero Duplication**: Hotkeys defined once, used everywhere
3. **Easy to Extend**: Adding new mode = one line in config
4. **Self-Documenting**: Data structure shows all available modes
5. **Maintainable**: No need to sync HTML/JS/CSS when adding modes
6. **Pure Data-Centric**: Data drives UI and behavior, not code

**Adding New Mode Example**:
```javascript
// In config.js, add one line:
{ id: 'last-1h', label: '-1h', hotkey: '7', tailLines: null, sinceSeconds: 3600, follow: true }

// Everything auto-generated:
// - Button with label and hotkey indicator
// - Hotkey handler (press 7)
// - Click handler
// - WebSocket URL parameters
```

### Key Code Patterns

**Head Mode Implementation** (internal/k8s/logs.go):
```go
lineCount := int64(0)
for scanner.Scan() {
  broadcast <- LogMessage{Type: "LOG_LINE", Line: scanner.Text() + "\n"}
  lineCount++

  // Stop if we've reached the head limit
  if opts.HeadLines != nil && lineCount >= *opts.HeadLines {
    broadcast <- LogMessage{Type: "LOG_END", Reason: fmt.Sprintf("Head limit reached (%d lines)", *opts.HeadLines)}
    return nil
  }
}
```

**Data-Driven Button Generation** (app.js):
```javascript
renderLogModeButtons() {
  LOG_MODES.forEach(mode => {
    const btn = document.createElement('button');
    btn.className = 'logs-mode-btn';
    btn.dataset.mode = mode.id;
    btn.title = `Hotkey: ${mode.hotkey}`;

    // Label and hotkey spans
    btn.appendChild(createLabel(mode.label));
    btn.appendChild(createHotkey(mode.hotkey));

    // Inline click handler
    btn.addEventListener('click', () => this.setLogMode(mode.id));

    container.appendChild(btn);
  });
}
```

**Dynamic Hotkey Handling** (app.js):
```javascript
// Before: Hardcoded map
const modeMap = { '1': 'head', '2': 'tail', ... };

// After: Generated from data
const mode = LOG_MODES.find(m => m.hotkey === event.key);
if (mode) this.setLogMode(mode.id);
```

### Bug Fixes

**Head mode showing last 500 instead of first 500**:
- Root cause: K8s TailLines parameter always gets last N lines
- Solution: Added HeadLines field and line counting mechanism
- Backend now counts lines and stops after N lines from beginning

### UX Improvements

- Hotkeys 1-6 only active when logs tab is open (contextual)
- Visual mode indicator with active state highlighting
- Compact labels for better space utilization (-5m, -15m, -500, -1000)
- Immediate feedback when switching modes

### Impact

This enhancement demonstrates the data-centric architecture in action:
- Configuration is data, not code
- UI is generated from data
- Adding features requires changing data, not logic
- Maintainability improved significantly
- Extensibility achieved through pure data

---

## 6.11. Phase 3 Continued (Vim-Like Command Mode - 2025-12-02)

### What Was Built

**Major milestone**: Complete keyboard-driven navigation system transforming k8v into a true keyboard-first power tool comparable to k9s.

✅ **Command Mode (`:`)** - Vim-style command palette
- Press `:` to activate full-screen command mode with glassmorphic backdrop
- "COMMAND" mode indicator (like vim's mode indicator)
- Real-time autocomplete with visual type badges
- Arrow key navigation (↑/↓) through suggestions
- Tab completion for faster typing
- Enter to execute commands
- Escape to close command mode

✅ **Resource Type Commands** - Instant navigation
- All 8 resource types: Pod, Deployment, ReplicaSet, Service, Ingress, ConfigMap, Secret, Node
- Kubectl-style aliases: `po`, `svc`, `rs`, `deploy`, `cm`, `ing`, `no`
- Example: `:svc` → instantly switch to Services view
- Prefix matching autocomplete (e.g., "dep" matches "deployment")

✅ **Special Action Commands** - Trigger UI actions
- `namespace` / `ns` - Opens namespace dropdown
- `context` / `ctx` - Opens context dropdown for cluster switching
- `cluster` - Opens context dropdown (alias)
- 100ms delay for smooth transitions

✅ **Data-Centric Architecture**
- All commands defined in `COMMANDS` array in config.js
- Helper functions: `findCommand()` for exact match, `getCommandSuggestions()` for filtering
- Zero hardcoded command lists in logic
- Adding new command = one line in config array
- Everything else (autocomplete, rendering, keyboard handling) auto-generated

✅ **Comprehensive Documentation**
- New `HOTKEYS.md` file documenting all keyboard shortcuts
- Global shortcuts (`:`, `/`, `d`, `Esc`)
- Command mode commands with aliases
- Log viewer hotkeys (1-6)
- Escape key hierarchy explanation
- Examples and tips for power users
- Instructions for adding custom commands

### Technical Implementation

**Data Structure** (config.js):
```javascript
export const COMMANDS = [
  // Resource type commands
  { id: 'pod', type: 'resource', label: 'Pod',
    aliases: ['pods', 'po'], target: 'Pod',
    description: 'Switch to Pods view' },

  // Action commands
  { id: 'namespace', type: 'action', label: 'namespace',
    aliases: ['ns'], action: 'openNamespaceDropdown',
    description: 'Open namespace selector' },
];

export function findCommand(input) {
  const normalized = input.toLowerCase().trim();
  return COMMANDS.find(cmd =>
    cmd.label.toLowerCase() === normalized ||
    cmd.aliases.some(alias => alias === normalized)
  );
}

export function getCommandSuggestions(input) {
  if (!input) return COMMANDS;
  const normalized = input.toLowerCase().trim();
  return COMMANDS.filter(cmd =>
    cmd.label.toLowerCase().startsWith(normalized) ||
    cmd.aliases.some(alias => alias.startsWith(normalized))
  );
}
```

**Command Mode Methods** (app.js - 7 new methods, 238 lines):
1. `activateCommandMode()` - Show overlay, focus input, initialize suggestions
2. `deactivateCommandMode()` - Hide overlay, reset state
3. `handleCommandInput(event)` - Update suggestions as user types
4. `renderCommandSuggestions()` - Generate DOM for autocomplete list with badges
5. `handleCommandKeydown(event)` - Handle keyboard navigation (↑↓, Tab, Enter, Esc)
6. `executeCommand(cmd)` - Execute resource filter or action command
7. `scrollCommandSuggestionIntoView()` - Auto-scroll to highlighted suggestion
8. `setupCommandMode()` - Wire up event listeners (input, keydown, backdrop click)

**Keyboard Integration** (handleGlobalKeydown updates):
```javascript
// Command mode activation - highest priority
if (event.key === ':' && !isInputFocused) {
  event.preventDefault();
  this.activateCommandMode();
  return;
}

// Escape hierarchy - command mode first
if (event.key === 'Escape') {
  if (this.state.command.active) {
    this.deactivateCommandMode();
    return;
  }
  // ... existing escape logic (debug, dropdown, detail, events, search)
}
```

**Command Execution Flow**:
```javascript
executeCommand(cmd) {
  if (cmd.type === 'resource') {
    // Switch resource type filter
    this.setFilter(cmd.target);
    this.deactivateCommandMode();
  } else if (cmd.type === 'action') {
    // Execute special action
    this.deactivateCommandMode();
    setTimeout(() => {
      if (cmd.action === 'openNamespaceDropdown') {
        this.namespaceDropdown.open();
      } else if (cmd.action === 'openContextDropdown') {
        this.contextDropdown.open();
      }
    }, 100); // Small delay for smooth transition
  }
}
```

**Visual Design**:
- Full-screen overlay (z-index: 3000, highest layer)
- Glassmorphic container (600px wide, blur effects)
- Brand color (#C4F561) for active states and borders
- Smooth fade-in animation (0.2s, scale + translate)
- Type-specific badges:
  - Resource: Blue (`rgba(33,150,243,0.2)`, color: `#03A9F4`)
  - Action: Pink (`rgba(156,39,176,0.2)`, color: `#E91E63`)
- Highlighted suggestions with left border accent
- Responsive layout (90vw max-width for mobile)

### Files Modified (6 total)

1. **`config.js`** (+45 lines)
   - COMMANDS array with 11 commands (8 resource types + 3 actions)
   - findCommand() helper function
   - getCommandSuggestions() helper function

2. **`state.js`** (+5 lines)
   - Command state object with 4 properties

3. **`index.html`** (+18 lines)
   - Command overlay HTML structure

4. **`style.css`** (+175 lines)
   - Complete command mode styling

5. **`app.js`** (+243 lines)
   - 7 command mode methods
   - Updated handleGlobalKeydown() for `:` and Escape
   - setupCommandMode() call in init()

6. **`HOTKEYS.md`** (NEW FILE)
   - Complete keyboard shortcut documentation

### Architecture Benefits

**Data-Centric Pattern**:
- Commands are pure configuration data
- Extensible: Add new command = one line in config array
- Zero duplication: Aliases, descriptions defined once
- Self-documenting: Data structure shows all available commands
- Type-safe: Resource vs action commands have different execution paths
- Maintainable: No hardcoded command lists in logic

**Adding Custom Commands Example**:
```javascript
// Resource type command
{
  id: 'daemonset',
  type: 'resource',
  label: 'DaemonSet',
  aliases: ['daemonsets', 'ds'],
  target: 'DaemonSet',
  description: 'Switch to DaemonSets view'
}

// Action command
{
  id: 'help',
  type: 'action',
  label: 'help',
  aliases: ['h'],
  action: 'openHelpModal',
  description: 'Show help documentation'
}
```

Everything else (autocomplete, rendering, keyboard handling) is auto-generated!

### UX Improvements

- **Keyboard-first workflow**: Never need to touch mouse for navigation
- **Muscle memory**: Vim users feel instantly at home with `:` command mode
- **Kubectl familiarity**: Aliases match kubectl conventions (po, svc, rs)
- **Visual feedback**: Clear mode indicator, smooth animations, highlighted selections
- **Discoverable**: Autocomplete shows all available commands with descriptions
- **Fast**: Instant response, no network latency

### Complete Keyboard Navigation

k8v now has comprehensive keyboard shortcuts:
- `:` - Command mode (resource navigation, actions)
- `/` - Search (filter by name)
- `d` - Debug drawer (view cache data)
- `1-6` - Log modes (when viewing Pod logs)
- `Esc` - Hierarchical close (command → debug → dropdown → detail → events → search)
- `↑↓` - Navigate dropdowns and autocomplete
- `Tab` - Auto-complete
- `Enter` - Execute/select

### Impact

This feature transforms k8v into a true keyboard-driven power tool:
- ✅ Vim-like navigation (`:` command mode)
- ✅ Kubectl-style aliases (po, svc, rs, deploy)
- ✅ Complete keyboard accessibility
- ✅ Data-centric extensible architecture
- ✅ Comprehensive documentation (HOTKEYS.md)

**k8v now offers a complete keyboard-first experience comparable to k9s, but with the power of a modern web UI.**

---

## 7. Quick Reference

| Aspect | Details |
|--------|---------|
| **Current Stage** | ✅ Phase 3 Nearly Complete - Core features implemented |
| **Tech Stack** | Go backend + Modular ES6 frontend (no frameworks) |
| **Frontend Architecture** | 7 ES6 modules: app.js, config.js, state.js, ws.js, dropdown.js, style.css, index.html |
| **Backend Language** | Go 1.23+ with client-go v0.31.0 |
| **Communication** | WebSocket (real-time bidirectional updates) |
| **Target User** | Developers with kubectl/kubeconfig access |
| **Deployment Model** | Single 62MB binary (`./k8v` command) |
| **Similar Tools** | k9s (TUI), Lens (Electron), kubectl proxy (proxy-only) |
| **Core Resources** | Pods, Deployments, Services, Ingress, ReplicaSets, ConfigMaps, Secrets |
| **MVP Status** | ✅ Visualization, ✅ Relationships, ✅ Live streaming, ✅ Pod logs, ✅ Search, ✅ Multi-context |
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

17. **Persistent Logging is Essential for Debugging:** Writing all server activity (HTTP requests, WebSocket events, errors) to a persistent log file (`logs/k8v.log`) enables effective debugging across sessions. Claude can read the log file to diagnose issues, identify patterns, and understand the sequence of events leading to errors. Logging middleware must implement `http.Hijacker` interface to support WebSocket upgrades.

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

**Last Updated:** 2025-12-02 - Implemented vim-like command mode with `:` keyboard shortcut, transforming k8v into a true keyboard-first power tool. Added complete command palette with autocomplete, kubectl-style aliases (po, svc, rs), and special action commands (namespace, context). All commands defined as pure configuration data in `COMMANDS` array following the data-centric architecture pattern. Created comprehensive `HOTKEYS.md` documentation. k8v now offers a complete keyboard-first experience comparable to k9s, with the power of a modern web UI. This feature represents a major milestone in making k8v accessible to power users who prefer keyboard-driven workflows.
