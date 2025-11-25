# K8V - Kubernetes Cluster Visualizer Design Document

## Executive Summary

Transform the existing beautiful HTML/CSS/JS prototype into a production-ready CLI tool that connects to real Kubernetes clusters and provides real-time visualization. The tool will follow patterns established by successful K8s tools like k9s and kubectl proxy, prioritizing developer experience and single-binary distribution.

---

## 1. Technology Stack

### 1.1 Backend: **Go**

**Rationale:**
- **Native K8s ecosystem**: Official `client-go` library is the gold standard for K8s interactions
- **Single binary compilation**: Cross-compile to any platform with zero runtime dependencies
- **Asset embedding**: Native `embed` package (Go 1.16+) makes bundling HTML/CSS/JS trivial
- **Performance**: Excellent for concurrent watch streams and WebSocket handling
- **Community patterns**: Most K8s CLI tools (kubectl, k9s, helm, kind) use Go
- **Kubeconfig handling**: Built-in support via `client-go/tools/clientcmd`
- **Small footprint**: Produces compact binaries (~15-30MB for full app)

**Alternatives Considered:**
- **Node.js**: Good JavaScript ecosystem, but requires runtime; harder single-binary distribution
- **Python**: Excellent for prototyping, but packaging complexity and slower performance for watch streams

### 1.2 Communication Layer: **WebSocket**

**Rationale:**
- **Bidirectional**: Enables server-push for K8s watch events AND client commands
- **Low latency**: Critical for real-time cluster updates
- **Browser support**: Excellent across all modern browsers
- **Libraries**: Mature Go WebSocket libraries (gorilla/websocket, nhooyr.io/websocket)
- **K8s watch compatibility**: Natural mapping from K8s watch API to WebSocket events

**Alternative: Server-Sent Events (SSE)**
- Simpler for server→client only
- Less ideal for potential future features (exec into pods, port-forward UI, etc.)

---

## 2. System Architecture

### 2.1 High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         User Machine                         │
│                                                              │
│  ┌──────────────┐      ┌───────────────────────────────┐   │
│  │  CLI Binary  │─────▶│   Embedded HTTP/WS Server     │   │
│  │   (k8v)      │      │   (localhost:8080)            │   │
│  └──────────────┘      │                               │   │
│                        │  ┌─────────────────────────┐  │   │
│                        │  │  Static File Handler    │  │   │
│                        │  │  (index.html, CSS, JS)  │  │   │
│                        │  └─────────────────────────┘  │   │
│                        │                               │   │
│                        │  ┌─────────────────────────┐  │   │
│                        │  │  WebSocket Handler      │  │   │
│                        │  │  - Push K8s events      │  │   │
│                        │  │  - Stream watch updates │  │   │
│                        │  └─────────────────────────┘  │   │
│                        │                               │   │
│                        │  ┌─────────────────────────┐  │   │
│                        │  │  REST API Handler       │  │   │
│                        │  │  - /api/resources       │  │   │
│                        │  │  - /api/namespaces      │  │   │
│                        │  └─────────────────────────┘  │   │
│                        └──────────┬────────────────────┘   │
│                                   │                         │
│                        ┌──────────▼────────────────────┐   │
│                        │   K8s Client Manager          │   │
│                        │   (client-go)                 │   │
│                        │                               │   │
│                        │  - Kubeconfig loader          │   │
│                        │  - Watch API manager          │   │
│                        │  - Resource cache             │   │
│                        │  - Multi-namespace support    │   │
│                        └──────────┬────────────────────┘   │
│                                   │                         │
└───────────────────────────────────┼─────────────────────────┘
                                    │
                                    │ HTTPS/K8s API
                                    │
                        ┌───────────▼───────────┐
                        │  Kubernetes Cluster   │
                        │  API Server           │
                        └───────────────────────┘
```

### 2.2 Component Breakdown

#### **CLI Entry Point**
- Parse flags: `--port`, `--context`, `--namespace`, `--kubeconfig`
- Initialize K8s client with context/namespace
- Start HTTP server with embedded static assets
- Open browser automatically to `http://localhost:8080`
- Graceful shutdown on SIGTERM/SIGINT

#### **HTTP Server**
- Serve embedded static files (index.html, no external CDN dependencies)
- REST API endpoints for initial data load
- WebSocket upgrade endpoint for real-time streams
- CORS handled (localhost only)

#### **K8s Client Manager**
- Load kubeconfig (default `~/.kube/config` or `--kubeconfig`)
- Support context switching
- Watch multiple resource types concurrently:
  - Pods, Deployments, ReplicaSets, Services, Ingress
  - ConfigMaps, Secrets (limited data for security)
- Maintain in-memory cache with sync on watch events
- Handle reconnection on watch failures

#### **WebSocket Event Stream**
```go
// Message types sent to frontend
type WSMessage struct {
    Type     string      `json:"type"`    // "init", "update", "delete"
    Resource string      `json:"resource"` // "pod", "service", etc.
    Data     interface{} `json:"data"`
}
```

### 2.3 Data Flow

**Initial Load:**
```
Browser → GET /
  ↓
Server → Return index.html (embedded)
  ↓
Browser → Execute JS → WebSocket connect to ws://localhost:8080/ws
  ↓
Server → Send full resource snapshot via WebSocket
  ↓
Browser → Render dashboard/topology
```

**Real-Time Updates:**
```
K8s API → Watch event (Pod created)
  ↓
K8s Client → Process event → Update cache
  ↓
WebSocket Handler → Broadcast to all connected clients
  ↓
Browser → Receive event → Update UI (no full refresh)
```

---

## 3. Project Structure

```
k8v/
├── cmd/
│   └── k8v/
│       └── main.go                 # CLI entry point, cobra commands
│
├── internal/
│   ├── server/
│   │   ├── server.go               # HTTP server setup
│   │   ├── handlers.go             # REST API handlers
│   │   ├── websocket.go            # WebSocket handler
│   │   └── middleware.go           # Logging, CORS
│   │
│   ├── k8s/
│   │   ├── client.go               # K8s client initialization
│   │   ├── watcher.go              # Watch API manager
│   │   ├── cache.go                # In-memory resource cache
│   │   └── transformer.go          # K8s objects → JSON for frontend
│   │
│   └── browser/
│       └── open.go                 # Cross-platform browser launcher
│
├── web/                            # Frontend assets
│   ├── index.html                  # Modified from prototype
│   ├── static/
│   │   ├── css/
│   │   │   └── styles.css          # Extracted from inline
│   │   ├── js/
│   │   │   ├── app.js              # Main app logic
│   │   │   ├── websocket.js        # WebSocket client
│   │   │   └── renderer.js         # UI rendering
│   │   └── vendor/
│   │       └── mermaid.min.js      # Downloaded, no CDN dependency
│   │
│   └── embed.go                    # Go embed directives
│
├── pkg/
│   └── types/
│       └── resources.go            # Shared types between server/frontend
│
├── scripts/
│   ├── build.sh                    # Cross-compilation script
│   └── release.sh                  # GitHub release builder
│
├── go.mod
├── go.sum
├── README.md
├── IDEAS.md
├── DESIGN.md                       # This document
└── LICENSE
```

### 3.1 Frontend Organization Strategy

**Approach: Extract and Modularize**
1. Keep existing prototype HTML structure
2. Extract inline CSS to `styles.css`
3. Extract JavaScript to modular files:
   - `app.js`: Main app logic, data management
   - `websocket.js`: WebSocket connection, reconnection logic
   - `renderer.js`: DOM updates, topology rendering
4. Replace CDN dependencies with embedded versions:
   - Download Mermaid.js and bundle it
   - Download Google Fonts and serve locally (optional, or keep CDN for fonts)

### 3.2 Asset Embedding Strategy

**Go 1.16+ embed package:**

```go
// web/embed.go
package web

import "embed"

//go:embed index.html static/*
var Assets embed.FS
```

**Usage in server:**
```go
// Serve embedded files
http.Handle("/", http.FileServer(http.FS(web.Assets)))
```

**Benefits:**
- Single binary distribution
- No runtime file dependencies
- Simplifies deployment

---

## 4. Key Technical Decisions

### 4.1 Kubernetes Client Library

**Decision: Use `client-go` official library**

**Implementation:**
```go
import (
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
    "k8s.io/client-go/informers"
)

// Use SharedInformerFactory for efficient watching
informerFactory := informers.NewSharedInformerFactory(clientset, time.Minute)
```

**Why:**
- Official, well-maintained
- Informer pattern for efficient caching
- Built-in retry/reconnection
- Same library as kubectl

### 4.2 Authentication & Kubeconfig Handling

**Decision: Leverage `client-go/tools/clientcmd`**

**Features:**
- Auto-detect kubeconfig from `$KUBECONFIG` or `~/.kube/config`
- Support for all auth methods (certs, tokens, exec plugins, OIDC)
- Context switching via CLI flags

**CLI Flags:**
```bash
k8v                                    # Use default context
k8v --context=prod                    # Specific context
k8v --namespace=kube-system           # Specific namespace (default: all)
k8v --kubeconfig=~/custom-config      # Custom kubeconfig path
k8v --port=9090                       # Custom port
```

### 4.3 WebSocket Message Protocol

**Decision: JSON-based event protocol**

**Message Format:**
```json
{
  "type": "RESOURCE_ADDED",
  "resourceType": "pod",
  "namespace": "default",
  "data": { /* full pod object */ }
}

{
  "type": "RESOURCE_UPDATED",
  "resourceType": "deployment",
  "namespace": "production",
  "data": { /* updated deployment */ }
}

{
  "type": "RESOURCE_DELETED",
  "resourceType": "service",
  "namespace": "default",
  "name": "frontend-service"
}
```

### 4.4 State Management in Frontend

**Decision: Immutable state updates with event sourcing pattern**

**Approach:**
```javascript
// Central state object
const state = {
  resources: {
    pods: {},
    services: {},
    deployments: {},
    // ...
  },
  filters: {
    namespace: 'all',
    type: 'all',
    health: 'all'
  }
};

// WebSocket message handler
ws.onmessage = (event) => {
  const msg = JSON.parse(event.data);

  switch (msg.type) {
    case 'RESOURCE_ADDED':
    case 'RESOURCE_UPDATED':
      state.resources[msg.resourceType][msg.data.name] = msg.data;
      break;
    case 'RESOURCE_DELETED':
      delete state.resources[msg.resourceType][msg.name];
      break;
  }

  // Re-render affected UI components
  updateDashboard();
  updateTopology();
};
```

**Benefits:**
- No full page reload
- Incremental UI updates
- Easy debugging (all state in one place)

### 4.5 Resource Caching Strategy

**Decision: Two-tier caching**

**Tier 1: Server-side (Go)**
- Use K8s Informer pattern (SharedInformerFactory)
- Watches maintain local cache automatically
- Serves initial snapshot to new WebSocket connections
- Memory-efficient (only active resources)

**Tier 2: Client-side (JavaScript)**
- Browser maintains full resource map
- Updates incrementally via WebSocket
- No need to re-fetch on view changes

**Benefits:**
- Fast initial load (from server cache)
- Real-time updates (from watch)
- No polling required
- Reduced API server load

---

## 5. Implementation Phases

### Phase 1: Basic CLI + Static File Serving (MVP)
**Goal:** Get the prototype running as a CLI tool with embedded assets

**Tasks:**
1. Initialize Go module, add dependencies (client-go, gorilla/websocket, cobra)
2. Create basic project structure (cmd/, internal/, web/)
3. Extract CSS/JS from index.html into separate files
4. Download Mermaid.js vendor file (no CDN dependency)
5. Implement embed.go to bundle static assets
6. Create basic HTTP server that serves embedded index.html
7. Add CLI entry point with `--port` flag
8. Test: `k8v` opens browser showing prototype (still with mock data)

**Success Criteria:**
- Single command `k8v` opens browser
- Prototype UI loads from embedded assets
- No external dependencies at runtime

**Estimated Effort:** 2-3 days

### Phase 2: K8s API Integration (Read-Only)
**Goal:** Replace mock data with real cluster resources

**Tasks:**
1. Implement kubeconfig loading with context support
2. Create K8s client initialization
3. Build resource fetcher for initial snapshot:
   - Pods, Services, Deployments, ReplicaSets, Ingress
   - ConfigMaps, Secrets (metadata only)
4. Create REST endpoint `/api/resources` returning JSON
5. Transform K8s objects to frontend format (match mock data schema)
6. Compute relationships (Service→Pods, Ingress→Service, etc.)
7. Update frontend to fetch from `/api/resources` instead of mock data
8. Handle errors (cluster unreachable, auth failures)

**Success Criteria:**
- `k8v --context=minikube` shows real cluster data
- All resource types display correctly
- Relationships render properly
- Graceful error handling

**Estimated Effort:** 4-5 days

### Phase 3: Real-Time Watch Mode
**Goal:** Add live updates using K8s watch API and WebSockets

**Tasks:**
1. Implement WebSocket upgrade handler `/ws`
2. Create K8s watch manager using Informers:
   - Start watches for all resource types
   - Handle Add/Update/Delete events
   - Maintain server-side cache
3. Build WebSocket broadcaster:
   - Send initial snapshot on connect
   - Stream watch events to all clients
   - Handle client disconnects
4. Update frontend WebSocket client:
   - Connect on page load
   - Process event messages
   - Update state incrementally
5. Implement UI partial updates:
   - Add resource to list/topology
   - Update resource in place
   - Remove deleted resources
   - Animate changes
6. Add reconnection logic (both K8s watch and WebSocket)

**Success Criteria:**
- Create/update/delete pod in cluster → UI updates within 1 second
- No full page refresh required
- Handles temporary disconnections gracefully
- Multiple browser tabs receive same updates

**Estimated Effort:** 5-6 days

### Phase 4: Polish & Additional Features (Post-MVP)
**Optional enhancements:**
- Namespace filtering via UI
- Health computation from pod status/events
- Event timeline feed
- Performance metrics (resource usage graphs)
- Export functionality (YAML download, screenshots)
- Dark/light theme toggle
- Search/filter improvements

---

## 6. Deployment & Distribution

### 6.1 Installation Methods

**Primary: Binary Releases (GitHub Releases)**
```bash
# Download latest release
curl -LO https://github.com/user/k8v/releases/latest/download/k8v_darwin_amd64
chmod +x k8v_darwin_amd64
mv k8v_darwin_amd64 /usr/local/bin/k8v
```

**Secondary: Homebrew (macOS/Linux)**
```bash
brew tap user/k8v
brew install k8v
```

**Tertiary: Go Install (requires Go toolchain)**
```bash
go install github.com/user/k8v/cmd/k8v@latest
```

### 6.2 Cross-Platform Considerations

**Supported Platforms:**
- macOS (darwin/amd64, darwin/arm64 - Apple Silicon)
- Linux (linux/amd64, linux/arm64)
- Windows (windows/amd64)

**Build Script:**
```bash
#!/bin/bash
# scripts/build.sh

platforms=(
  "darwin/amd64"
  "darwin/arm64"
  "linux/amd64"
  "linux/arm64"
  "windows/amd64"
)

for platform in "${platforms[@]}"; do
  GOOS=${platform%/*}
  GOARCH=${platform#*/}
  output="k8v_${GOOS}_${GOARCH}"
  [ "$GOOS" = "windows" ] && output="${output}.exe"

  CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH \
    go build -ldflags="-s -w" -o "dist/$output" ./cmd/k8v
done
```

**Platform-Specific Notes:**
- **Browser opening**: Use `pkg/browser` package (different commands per OS)
- **Path handling**: Use `filepath.Join()` for cross-platform paths
- **Signals**: Handle SIGTERM (Unix) and SIGINT (Windows) for graceful shutdown

### 6.3 Single Binary Embedding

**Strategy:**
1. **Static assets**: Embedded via `//go:embed` (HTML, CSS, JS, fonts)
2. **Vendor JavaScript**: Download and commit Mermaid.js (avoid CDN at runtime)
3. **Build flags**: Use `-ldflags="-s -w"` to strip debug info (reduce size)
4. **Compression**: Consider UPX for additional compression (optional)

**Expected Binary Sizes:**
- Without compression: ~20-30MB
- With UPX: ~8-12MB
- Trade-off: Startup time vs. download size

---

## 7. Technical Considerations

### 7.1 Known Limitations (MVP)

1. **No persistence**: All state in-memory, restarts lose history
2. **Single-user**: No multi-user collaboration features
3. **Read-only**: No kubectl-like exec/logs/port-forward
4. **No RBAC awareness**: Shows all resources user can access (may be noisy)

### 7.2 Security Considerations

1. **Localhost binding**: Default to `127.0.0.1` only
2. **No auth**: Assumes local trusted environment (like kubectl proxy)
3. **Secrets handling**: Never send full secret values to frontend
4. **CORS**: Strict origin checking if remote access added
5. **Kubeconfig permissions**: Inherit from filesystem (no additional auth layer)

### 7.3 Success Metrics

**Technical Metrics:**
- **Startup time**: < 2 seconds from command to browser open
- **Initial load**: < 1 second to render 100 resources
- **Update latency**: < 500ms from K8s event to UI update
- **Memory usage**: < 100MB for typical cluster (1000 resources)
- **Binary size**: < 30MB uncompressed

**User Experience Metrics:**
- **Installation**: Single binary download, no dependencies
- **First run**: Zero configuration for default kubeconfig
- **Visual clarity**: All existing prototype features preserved
- **Stability**: No crashes on watch reconnection or malformed data

### 7.4 Future Enhancements

1. **Historical data**: Store events in SQLite for trend analysis
2. **Metrics integration**: Fetch Prometheus metrics for resource usage
3. **Multi-cluster**: Switch between clusters via UI
4. **Plugins**: Allow custom visualizations via plugin system
5. **Remote access**: Optional flag to bind to 0.0.0.0 (with auth)
6. **CRD support**: Auto-discover and visualize custom resources

---

## 8. References & Inspiration

### 8.1 Similar Tools Analysis

| Tool | Language | Approach | Lessons |
|------|----------|----------|---------|
| **k9s** | Go | TUI (terminal) | Excellent K8s client patterns, watch handling |
| **kubectl proxy** | Go | Reverse proxy to K8s API | Simple server model, browser-based |
| **Lens** | Electron/Node | Desktop app | Feature-rich but heavy; avoid this path |
| **Octant** | Go | Web UI | Similar architecture, good reference for server structure |
| **Headlamp** | Go/React | Web UI | Modern approach, check out their watch implementation |

### 8.2 Key Libraries

- **Kubernetes**: `k8s.io/client-go` (official K8s client)
- **CLI**: `github.com/spf13/cobra` (standard for Go CLIs)
- **WebSocket**: `github.com/gorilla/websocket` or `nhooyr.io/websocket`
- **HTTP Router**: `github.com/gorilla/mux` or stdlib `net/http` (sufficient)
- **Browser Opening**: `github.com/pkg/browser` or custom implementation

---

## 9. Conclusion

This design balances pragmatism with the goal of creating a production-ready tool. Go is the clear choice for K8s ecosystem integration and single-binary distribution. The phased approach allows for incremental value delivery, with Phase 1-3 forming a complete MVP. The architecture is simple enough to implement quickly but extensible for future enhancements.

**Recommended First Steps:**
1. Set up Go project structure
2. Extract and modularize existing prototype assets
3. Implement basic embedded server (Phase 1)
4. Validate approach before proceeding to K8s integration

**Key Success Factors:**
- Preserve the beautiful existing UI design
- Leverage Go's strengths for K8s integration
- Prioritize real-time updates (user's #1 feature request)
- Single binary for easy distribution
- Phased implementation for incremental delivery
