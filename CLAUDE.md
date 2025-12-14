# K8V Project Context

> Kubernetes cluster visualization tool - like k9s but with a modern web UI

---

## What is k8v?

**Vision:** CLI tool (`k8v`) → starts local server → opens browser → displays live Kubernetes cluster with beautiful UI

**Core Features:**
- Real-time cluster visualization via WebSocket streaming
- Bidirectional resource relationship navigation
- Pod logs viewer with 6 modes (Head, Tail, Last 5m/15m/500/1000)
- **Pod shell/exec** - Interactive terminal access to pods via xterm.js
- **Node shell** - Interactive shell on nodes via debug pod (`kubectl debug node` approach)
- Vim-like command mode (`:` palette with kubectl aliases)
- Namespace/context switching without restart
- Complete keyboard navigation (`:`, `/`, `d`, `1-6`, `Esc`)
- Search by name (`/`)
- Server-side filtering (40-100x network reduction)

**Current State:** ✅ Phase 3 Complete
- Production Go backend (Informers, WebSocket, relationships)
- Polished ES6 frontend (7 modules: app, config, state, ws, dropdown, style, index.html)
- Single 66MB binary ready to use (includes embedded xterm.js)
- Tested with 21,000+ resources in production clusters

**Next:** Enhanced YAML view, virtual scrolling, additional resource types

---

## Tech Stack

| Component | Technology | Why |
|-----------|------------|-----|
| **Backend** | Go 1.23+ with client-go v0.31.0 | Native K8s ecosystem, single binary, small footprint |
| **Frontend** | ES6 modules (no framework) | Zero dependencies, fast, maintainable |
| **Communication** | WebSocket | Real-time bidirectional updates |
| **UI Design** | Dark glassmorphic | Professional, modern aesthetic |
| **Deployment** | Single binary | Zero runtime dependencies |

---

## Architecture

```
CLI Binary (k8v)
  ↓
Embedded HTTP/WebSocket Server (localhost:8080)
  ↓
K8s Client Manager (client-go + Informers)
  ↓
Kubernetes API Server
```

**Project Structure:**
```
k8v/
├── cmd/k8v/main.go              # CLI entry
├── internal/
│   ├── server/                   # HTTP/WebSocket server
│   │   └── static/               # Frontend (embedded)
│   │       ├── index.html
│   │       ├── style.css
│   │       ├── app.js           # Main app logic
│   │       ├── config.js        # Data-driven config
│   │       ├── state.js         # State management
│   │       ├── ws.js            # WebSocket client
│   │       └── dropdown.js      # Reusable component
│   ├── k8s/                      # K8s client, watchers
│   └── browser/                  # Browser launcher
└── pkg/types/                    # Shared types
```

**Data Model:**
```go
type Resource struct {
    ID            string         // "type:namespace:name"
    Type          string         // "Pod", "Deployment", etc.
    Name          string
    Namespace     string
    Status        ResourceStatus
    Health        HealthState    // "healthy", "warning", "error"
    Relationships Relationships  // Bidirectional links
    Labels        map[string]string
    YAML          string
}

type Relationships struct {
    OwnedBy   []ResourceRef  // Deployment → ReplicaSet → Pod
    Owns      []ResourceRef
    DependsOn []ResourceRef  // Pod → ConfigMap/Secret
    UsedBy    []ResourceRef
    Exposes   []ResourceRef  // Service → Pods
    ExposedBy []ResourceRef
    RoutesTo  []ResourceRef  // Ingress → Service
    RoutedBy  []ResourceRef
}
```

**WebSocket Protocol:**
```json
{
  "type": "RESOURCE_ADDED|RESOURCE_MODIFIED|RESOURCE_DELETED|SYNC_STATUS|LOG_LINE",
  "resourceType": "pod",
  "namespace": "default",
  "data": { /* resource object */ }
}
```

---

## MVP Status

| Feature | Status |
|---------|--------|
| Resource Visualization | ✅ Complete |
| Live Streaming | ✅ Complete |
| Relationship Mapping | ✅ Complete |
| Pod Logs Viewing | ✅ Complete |
| Pod Shell/Exec | ✅ Complete |
| Node Shell | ✅ Complete |
| Search Functionality | ✅ Complete |
| Multi-Context Support | ✅ Complete |
| Namespace Filtering | ✅ Complete |
| Keyboard Navigation | ✅ Complete |
| Top-Tier UI | ✅ Complete |
| Extensible Data Model | ✅ Complete |

**Remaining:**
- Enhanced YAML view (syntax highlighting, clickable refs)
- Virtual scrolling for massive lists
- Additional resource types (StatefulSets, DaemonSets, etc.)

---

## Key Insights

1. **Data-Centric Architecture**: Configuration is data (LOG_MODES, COMMANDS arrays), not code. UI/behavior auto-generated from data. Add feature = add one line to config array.

2. **Incremental DOM Updates**: With large clusters, full rerenders cause flickering. Modify individual DOM elements on RESOURCE_ADDED/MODIFIED/DELETED events.

3. **WebSocket Race Conditions**: Async snapshot sending causes panics. Send snapshot synchronously before starting pumps.

4. **Server-Side Filtering**: Client-side filtering wastes bandwidth. Filter on backend (namespace, resource type) = 40-100x network reduction.

5. **Reactive State Management**: Backend broadcasts state changes (SYNC_STATUS), frontend reacts. Never poll or guess. Single source of truth on server.

6. **Relationship-First Model**: Resource connections (Deployment → ConfigMap, Service → Pods) are as important as resources themselves. Bidirectional navigation is essential.

7. **Keyboard-First UX**: Vim-like command mode (`:`) + kubectl aliases (po, svc, rs) make k8v a power tool comparable to k9s.

8. **Modular Frontend**: 7 ES6 modules with single responsibilities. Easy to locate/modify functionality. Zero framework dependencies.

9. **Persistent Logging**: `logs/k8v.log` enables debugging across sessions. Essential for diagnosing issues.

10. **Informers > Direct Watch**: SharedInformerFactory provides caching, automatic reconnection, efficient updates. Essential for production.

---

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `:` | Command mode (`:po`, `:svc`, `:ns`, `:ctx`) |
| `/` | Search by name |
| `d` | Debug drawer (cache data) |
| `1-6` | Log modes (when viewing logs) |
| `Esc` | Hierarchical close (command → debug → dropdown → detail → events → search) |
| `↑↓` | Navigate dropdowns/autocomplete |
| `Tab` | Auto-complete |
| `Enter` | Execute/select |

See `HOTKEYS.md` for full documentation.

---

## Document References

- **README.md** - Public documentation, quickstart
- **IDEAS.md** - Original vision, MVP priorities
- **DESIGN.md** - Technical architecture
- **DATA_MODEL.md** - Resource relationships system
- **CHANGELOG.md** - Version history
- **HOTKEYS.md** - Keyboard shortcuts guide

---

## Quick Reference

| Aspect | Details |
|--------|---------|
| **Current Stage** | ✅ Phase 3 Complete |
| **Tech Stack** | Go + ES6 modules + xterm.js |
| **Communication** | WebSocket |
| **Deployment** | Single 66MB binary |
| **Binary** | `./k8v` |
| **Resources** | Pod, Deployment, ReplicaSet, Service, Ingress, ConfigMap, Secret, Node |
| **Scale Tested** | 21,867 resources |
| **Performance** | 40-100x network reduction |

---

**Last Updated:** 2025-12-13 - Node shell implemented. Interactive shell access to nodes via debug pod (`kubectl debug node` approach) with chroot to host filesystem. Also fixed terminal reconnection bugs (double keystrokes, Tab key, artifacts on reconnect).
