# Kubernetes Cluster Visualizer

A modern, interactive web-based visualization tool for Kubernetes cluster resources. This single-page application provides both a dashboard view and a topology view to help you understand and navigate your Kubernetes infrastructure.

## Features

### Dual View Modes

**Dashboard View**
- Grid-based layout displaying comprehensive statistics
- Resource cards with health indicators
- Filterable resource lists organized by type
- Real-time statistics showing resource counts and health status

**Topology View**
- Interactive 3D node visualization with animated cubes
- Visual connections showing relationships between resources
- Drag-and-drop node positioning
- Zoom and pan capabilities for exploring complex clusters

### Resource Management

Supports visualization of all major Kubernetes resource types:
- **Ingress** - Entry points to your cluster
- **Services** - Network abstractions for pod groups
- **Deployments** - Declarative application updates
- **ReplicaSets** - Pod replica management
- **Pods** - Running container instances
- **ConfigMaps** - Configuration data storage
- **Secrets** - Sensitive data storage

### Interactive Features

- **Resource Filtering** - Filter by resource type (All, Ingress, Services, Deployments, Pods, Config)
- **Health Status Indicators** - Visual indicators for healthy, warning, and error states
- **Detailed Resource View** - Click any resource to view:
  - Overview with metadata and status
  - Full YAML configuration with syntax highlighting
  - Relationships showing connected resources
- **YAML Navigation** - Clickable resource references in YAML that navigate to related resources
- **Copy to Clipboard** - One-click YAML copying functionality
- **Resource Search** - Quick navigation through resource lists

### Statistics Dashboard

- Total resource counts by type
- Health status aggregation
- Color-coded warnings and errors
- At-a-glance cluster health overview

## Design

- **Modern Dark Theme** - Easy on the eyes with gradient accents
- **Glassmorphism UI** - Frosted glass effects with backdrop blur
- **Smooth Animations** - Polished transitions and hover effects
- **3D Visualizations** - Animated cube representations for each resource
- **Custom Typography** - Space Grotesk and Inter fonts for optimal readability
- **Responsive Layout** - Adapts to different screen sizes

## Technology Stack

- **Pure HTML/CSS/JavaScript** - No external dependencies or frameworks
- **Self-Contained** - Single file application, no build process required
- **Zero Configuration** - Open in any modern browser and start visualizing

## Usage

1. Open `index.html` in any modern web browser
2. Use the navigation tabs to switch between Dashboard and Topology views
3. Click on resource filter buttons to show specific resource types
4. Click any resource card or node to view detailed information
5. In the detail panel:
   - View resource overview and metadata
   - Inspect full YAML configuration
   - Explore relationships with other resources
   - Copy YAML to clipboard

## Keyboard & Mouse Interactions

- **Click resource node/card** - Show detailed information
- **Click filter buttons** - Filter by resource type
- **Click YAML references** - Navigate to related resources
- **Hover over elements** - See interactive effects and highlights

## Resource Relationships

The visualizer automatically detects and displays relationships:
- **Ingress → Services** - Traffic routing paths
- **Services → Pods** - Service endpoint mappings
- **Deployments → ReplicaSets** - Deployment ownership
- **ReplicaSets → Pods** - Pod replica management
- **Pods → ConfigMaps/Secrets** - Configuration and secret usage

## Sample Data

The application includes sample data demonstrating a typical microservices architecture:
- Frontend application with multiple pods
- API service backend
- Admin interface
- Database service
- Worker processes
- Monitoring stack (Prometheus, Grafana)
- Supporting infrastructure (ingress, configs, secrets)

## Browser Compatibility

Requires a modern browser with support for:
- ES6+ JavaScript
- CSS Grid and Flexbox
- CSS backdrop-filter
- CSS transforms and animations
- Clipboard API

Tested on:
- Chrome/Edge 90+
- Firefox 88+
- Safari 14+

## Customization

To adapt this visualizer for your own cluster:

1. Locate the `k8sData` object in the `<script>` section (around line 1502)
2. Replace the sample data with your actual Kubernetes resources
3. Update the `nodePositions` object for custom topology layouts
4. Modify the color scheme via CSS variables if desired

## File Structure

```
index.html          # Single-page application containing all HTML, CSS, and JavaScript
```

## License

This is a standalone visualization tool. Modify and use as needed for your infrastructure visualization needs.

## Notes

- This is a static visualization tool with hardcoded sample data
- For production use, consider connecting to a real Kubernetes API
- No actual Kubernetes cluster connection is made
- All data is client-side only, no server required
