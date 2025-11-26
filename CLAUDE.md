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
- **Stage:** ✅ Phase 1 Complete - Production backend + minimal frontend
- **What exists:**
  - Fully functional HTML/CSS/JS prototype with mock data (index.html)
  - Production Go backend with Informers, WebSocket streaming, and generic relationship system
  - Minimal working frontend with bidirectional relationship navigation
  - Single 62MB binary (k8v) ready to use
- **Next:** Phase 2 - Build full frontend matching prototype quality

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
- Mermaid.js for topology diagrams
- Google Fonts (Space Grotesk, Inter)
- Self-contained single-file application

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
│   ├── k8s/                      # K8s client, watcher, cache
│   └── browser/                  # Cross-platform browser launcher
├── web/                          # Frontend assets (extracted from prototype)
│   ├── index.html
│   ├── static/css/
│   ├── static/js/
│   └── embed.go                  # Go embed directives
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

## 6. Phase 1 Complete (Production Backend + Minimal Frontend)

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

### What's Next (Phase 2)

1. **Full Frontend Implementation**
   - Extract design from index.html prototype
   - Dashboard view with statistics cards
   - Advanced filtering and search
   - Topology view with Mermaid diagrams
   - Events timeline
   - Dark/light theme toggle

2. **Pod Logs Feature**
   - Stream logs via WebSocket
   - Log viewer in detail panel
   - Follow mode for live logs

3. **Polish & Features**
   - Multi-cluster support (context switching)
   - Export functionality (YAML download)
   - Resource editing (kubectl apply)
   - Namespace filtering

---

## 7. Quick Reference

| Aspect | Details |
|--------|---------|
| **Current Stage** | Implementation phase - POC validated, data model designed |
| **Tech Stack** | Go backend + HTML/CSS/JS frontend |
| **Backend Language** | Go (for K8s ecosystem, single binary) |
| **Communication** | WebSocket (for real-time updates) |
| **Target User** | Developers with kubectl/kubeconfig access |
| **Deployment Model** | Single binary CLI tool (`k8v` command) |
| **Similar Tools** | k9s (TUI), Lens (Electron), kubectl proxy (proxy-only) |
| **Core Resources** | Pods, Deployments, Services, Ingress, ReplicaSets, ConfigMaps, Secrets |
| **MVP Priority** | Resource visualization + relationship graph + live streaming |
| **UI Prototype** | ✅ Complete (index.html has production-quality UI) |
| **POC** | ✅ Complete (k8v-poc validates streaming architecture) |
| **Data Model** | ✅ Complete (DATA_MODEL.md defines relationship model) |
| **Production Backend** | ⏳ In Progress (next step: implement with full data model) |
| **K8s Integration** | ⏳ In Progress (POC has basic integration, need full model) |

---

## 8. Document References

- **README.md** - Basic documentation about the prototype features and usage
- **IDEAS.md** - Detailed feature requirements and user vision (read for MVP priorities)
- **DESIGN.md** - Complete technical design and architecture (read for implementation details)
- **DATA_MODEL.md** - Complete data model with relationships (read for backend implementation)
- **index.html** - Working UI prototype demonstrating full UX vision
- **k8v-poc/** - Minimal POC validating streaming architecture (Pods, Deployments, ReplicaSets)

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

**Last Updated:** 2025-11-26 - Added POC validation results and DATA_MODEL.md summary
