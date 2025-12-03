# k8v - Kubernetes Visualizer

A modern, real-time Kubernetes cluster visualization tool with a beautiful web UI. Like k9s, but with a web interface and superior user experience.

![Status: Phase 3 In Progress](https://img.shields.io/badge/status-Phase%203%20In%20Progress-blue)
![Go Version](https://img.shields.io/badge/go-1.23%2B-blue)
![Binary Size](https://img.shields.io/badge/binary-62MB-orange)

## 🎯 What is k8v?

k8v is a **single-binary CLI tool** that connects to your Kubernetes cluster and provides a **modern web interface** for real-time cluster visualization. It's designed for developers who want the power of kubectl with the convenience of a visual interface.

## 🎥 Demo
![demo](https://github.com/user-attachments/assets/cae0b55a-d208-44c8-b997-da440eb8df46)

### Key Features

- ✅ **Vim-Like Command Mode** - Keyboard-first navigation with `:` command palette and kubectl-style aliases
- ✅ **Real-time Updates** - Live streaming of cluster changes via WebSocket
- ✅ **Resource Visualization** - View Pods, Deployments, Services, Ingress, ReplicaSets, ConfigMaps, Secrets
- ✅ **Pod Logs Viewer** - Stream and view container logs in real-time with configurable modes (1-6 hotkeys)
- ✅ **Search Functionality** - Quick search by resource name with keyboard shortcut (/)
- ✅ **Multi-Context Support** - Switch between Kubernetes contexts with reactive state management
- ✅ **Namespace Filtering** - Server-side filtering with searchable dropdown and keyboard navigation
- ✅ **Relationship Mapping** - Click any resource to see bidirectional relationships
- ✅ **Complete Keyboard Navigation** - `:`, `/`, `d`, `1-6`, `Esc` shortcuts for power users
- ✅ **Scale Tested** - Handles 20,000+ resources smoothly
- ✅ **Zero Dependencies** - Single binary with embedded web UI
- ✅ **Production Ready** - Battle-tested with large production clusters

## 🚀 Quick Start

### Installation

```bash
# Download the binary
git clone https://github.com/user/k8v.git
cd k8v
go build -o k8v cmd/k8v/main.go

# Make it executable
chmod +x k8v
```

### Usage

```bash
# Start k8v (uses current kubectl context)
./k8v

# Specify a different port
./k8v -port 3000
```

The web UI will automatically open in your browser at `http://localhost:8080`.

## 📊 Features

### Dashboard View

- **Resource Statistics** - See counts for all resource types at a glance
- **Filterable Lists** - Filter by Pods, Deployments, Services, Ingress, ReplicaSets, ConfigMaps, Secrets
- **Health Indicators** - Visual status (healthy/warning/error) for every resource
- **Detail Panel** - Click any resource to view:
  - Overview with metadata and status
  - Full YAML configuration
  - Bidirectional relationships (click to navigate)

### Real-Time Updates

- **Live Streaming** - Changes appear in < 500ms
- **Incremental Updates** - Only affected resources refresh (no flickering)
- **Events Timeline** - See recent cluster events with severity indicators
- **Connection Status** - Always know if you're connected to the cluster

### Resource Relationships

The UI automatically discovers and displays relationships:

- **Ownership:** Deployment → ReplicaSet → Pod
- **Dependencies:** Pod → ConfigMap/Secret (via volume mounts, env vars)
- **Network:** Service → Pods (via selector), Ingress → Service (via routes)

**Example:** Click a Service to see:
- `Exposes: Pod api-1, api-2, api-3` (clickable)
- `Routed by: Ingress api` (clickable)

### Keyboard Shortcuts

k8v is designed for keyboard-first workflows with comprehensive shortcuts:

- **`:` (Command Mode)** - Vim-style command palette for instant navigation
  - Type resource names: `pod`, `svc`, `deploy`, etc.
  - Use kubectl aliases: `po`, `rs`, `cm`, `ing`, etc.
  - Special commands: `namespace`, `context`
  - Example: `:svc` → instantly switch to Services view
- **`/`** - Quick search to filter resources by name
- **`d`** - Toggle debug drawer (view cache data)
- **`1-6`** - Switch log viewer modes (when viewing Pod logs)
  - 1: Head (first 500 lines)
  - 2: Tail (last 100 + follow)
  - 3: Last 5 minutes
  - 4: Last 15 minutes
  - 5: Last 500 lines
  - 6: Last 1000 lines
- **`Esc`** - Hierarchical close (command → debug → detail → search)
- **`↑↓`** - Navigate dropdowns and autocomplete
- **`Tab`** - Auto-complete in command mode
- **`Enter`** - Execute/select

See `HOTKEYS.md` for complete documentation.

## 🏗️ Architecture

```
k8v (CLI binary)
  ↓
  ├── Go Backend
  │   ├── Kubernetes Client (client-go with Informers)
  │   ├── WebSocket Hub (real-time streaming)
  │   └── Resource Transformers (7 resource types)
  ↓
  └── Web UI (embedded in binary)
      ├── Modular ES6 architecture (7 modules)
      ├── Incremental DOM updates
      └── WebSocket client
```

### Technical Stack

- **Backend:** Go 1.23+ with `client-go` v0.31.0
- **Communication:** WebSocket (bidirectional real-time updates)
- **Frontend:** Modular ES6 JavaScript (config, state, ws, app, dropdown components)
- **UI Framework:** None - Pure HTML/CSS/JS (no build step required)
- **Authentication:** Uses your local kubeconfig (supports in-cluster mode too)
- **Deployment:** Single binary with embedded assets

## 📈 Performance

Tested with a production cluster of **21,867 resources**:

- **Snapshot Load:** 2-5 seconds with progress logging
- **Update Latency:** < 100ms for incremental updates
- **Memory Usage:** Stable, no leaks observed
- **Binary Size:** 62MB (includes web UI)

## 🎨 UI Design

- **Modern Dark Theme** - Professional glassmorphic interface
- **Smooth Animations** - Polished transitions and hover effects
- **Responsive Layout** - Adapts to content and screen size
- **Color-coded Health** - Green (healthy), Yellow (warning), Red (error)
- **Compact Statistics** - Focus on what matters: the resource list

## 🔧 Configuration

k8v uses your existing kubeconfig:

```bash
# Use current context
./k8v

# Switch context first
kubectl config use-context my-cluster
./k8v

# Or specify port
./k8v -port 3000
```

## 📚 Documentation

- **[CLAUDE.md](./CLAUDE.md)** - Complete project context and architecture
- **[DESIGN.md](./DESIGN.md)** - Technical design decisions
- **[DATA_MODEL.md](./DATA_MODEL.md)** - Data model and relationships
- **[CHANGELOG.md](./CHANGELOG.md)** - Version history and changes
- **[IDEAS.md](./IDEAS.md)** - Feature roadmap and vision

## 🛣️ Roadmap

### ✅ Phase 1 (Complete)
- Production Go backend with Informers
- WebSocket streaming
- Generic relationship system
- Minimal frontend integration

### ✅ Phase 2 (Complete)
- Polished web UI with incremental updates
- WebSocket stability for large clusters
- Compact statistics and refined UX
- Scale testing (21k+ resources)

### 🚧 Phase 3 (In Progress)
- ✅ **Namespace Filtering:** Server-side filtering with searchable dropdown, keyboard navigation, and localStorage persistence (200x network reduction)
- ✅ **Icon Consistency:** Replaced emojis with Feather Icons for cohesive glassmorphic design
- ✅ **Pod Logs Viewer:** Real-time log streaming via WebSocket with container selection and auto-select first container
- ✅ **Search Functionality:** Search resources by name with keyboard shortcut (/) and real-time filtering
- ✅ **Multi-Context Support:** Switch between Kubernetes contexts with reactive state synchronization
- **Enhanced YAML View:** Syntax highlighting and clickable resource references
- **Frontend Performance:** Virtual scrolling and lazy rendering for massive clusters

### 📋 Future
- Additional K8s resources (StatefulSets, DaemonSets, Jobs, PVs, etc.)
- Enhanced search (by labels and annotations)
- Topology graph view (relationship visualization)
- Resource editing (kubectl apply)
- YAML syntax highlighting and clickable references
- YAML export/download
- Custom resource definitions (CRDs)
- Events timeline with filtering

## 🤝 Contributing

This is currently a personal project, but feedback and suggestions are welcome!

## 📄 License

MIT License - see LICENSE file for details.

## 🙏 Acknowledgments

- Inspired by [k9s](https://k9scli.io/) and [Lens](https://k8slens.dev/)
- Built with [client-go](https://github.com/kubernetes/client-go)
- Uses [gorilla/websocket](https://github.com/gorilla/websocket)

## 📞 Support

For questions or issues, please open a GitHub issue.

---

**Made with ❤️ for Kubernetes developers**
