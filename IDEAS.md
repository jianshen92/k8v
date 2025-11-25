# What i want to build
A client that i can just hook into any kubernetes cluster, and have nice overview of the cluster state.
This is like k9s, but with a web UI that has a better user experience.

# Features that i would like to have.

## Core Visualization
- **Interactive Dashboard**: Real-time overview of cluster health with key metrics (healthy nodes, warnings, errors, total resources)
- **Multi-namespace Support**: View resources across multiple namespaces (default, monitoring, logging, etc.)
- **Resource Type Filtering**: Filter and browse by resource type (Pods, Deployments, Services, Ingress, ReplicaSets, ConfigMaps, Secrets)
- **Health Status Indicators**: Visual color-coded status for each resource (healthy, warning, error)

## Topology View
- **Network Topology Visualization**: Interactive diagram showing relationships between resources using Mermaid
- **Resource Flow Mapping**: Visualize traffic flow: Ingress → Services → Deployments → ReplicaSets → Pods
- **Pan and Zoom**: Navigate large cluster architectures with smooth pan/zoom controls
- **Interactive Nodes**: Hover and click on topology nodes to view details

## Resource Details
- **Detailed Resource View**: Click any resource to see full specifications
- **YAML Configuration Display**: View complete YAML manifests for each resource
- **Cross-reference Navigation**: Clickable resource references in YAML to jump between related resources
- **Resource Relationships**: See which resources are connected (e.g., which pods a service exposes, which ingress routes to a service)

## Monitoring & Events
- **Recent Events Timeline**: Real-time feed of cluster events (pod crashes, scaling events, deployments, etc.)
- **Event Severity Levels**: Color-coded events by type (error, warning, normal)
- **Event Source Tracking**: See which resource and namespace generated each event

## User Experience
- **Modern UI Design**: Dark-themed glassmorphic design with smooth animations
- **Search Functionality**: Quick search across all resources
- **Responsive Layout**: Grid-based layout that adapts to content
- **Quick Actions**: One-click access to topology view and report downloads

# Core MVP
- all the core visualization:
- constructing a graph that links the relationship between resources, for examples deployments -> configmap
- clicking on a resources and it will show what is linked to it.
- topology (graph view) is secondary, as graph displaying a hard problem.
- sync to a kubernetes cluster, then start streaming the live updates to the dashboard.
- search (important but can come in v2)
- top tier UI
- logs viewing of pods
- start with core resources (pod, deployments, services, ingress, configmap, secret, replicaset), but data model must be extensible and easy to include more resources.
- 

# Ideas
## Initial Idea
1. user have kubeconfig setup locally.
2. user type `k8v` 
3. it spins up a backend server that streams the cluster state and live changes to a Web UI
