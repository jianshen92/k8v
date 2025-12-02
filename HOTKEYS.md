# K8V Keyboard Shortcuts

This document describes all keyboard shortcuts available in the k8v dashboard.

## Global Shortcuts

These shortcuts work anywhere in the dashboard unless an input field is focused.

| Key | Action | Description |
|-----|--------|-------------|
| `:` | Command Mode | Open vim-style command mode for quick navigation |
| `/` | Search | Activate search to filter resources by name |
| `d` | Debug Drawer | Toggle debug drawer (shows frontend cache data) |
| `↑` / `k` | Navigate Up | Move selection to previous row in table |
| `↓` / `j` | Navigate Down | Move selection to next row in table |
| `Enter` | Open Details | Open detail panel for selected row |
| `Esc` | Close/Exit | Hierarchical exit (closes most recent overlay/panel) |

## Command Mode (`:`)

Press `:` to enter command mode, then type a command and press Enter.

### Resource Type Commands

Switch to a specific resource type view:

| Command | Aliases | Action |
|---------|---------|--------|
| `pod` | `pods`, `po` | Switch to Pods view |
| `deployment` | `deployments`, `deploy` | Switch to Deployments view |
| `replicaset` | `replicasets`, `rs` | Switch to ReplicaSets view |
| `service` | `services`, `svc` | Switch to Services view |
| `ingress` | `ingresses`, `ing` | Switch to Ingress view |
| `configmap` | `configmaps`, `cm` | Switch to ConfigMaps view |
| `secret` | `secrets` | Switch to Secrets view |
| `node` | `nodes`, `no` | Switch to Nodes view |

### Special Commands

Trigger actions or open UI elements:

| Command | Aliases | Action |
|---------|---------|--------|
| `namespace` | `ns` | Open namespace dropdown |
| `context` | `ctx` | Open context dropdown |
| `cluster` | - | Open context dropdown |

### Command Mode Navigation

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate suggestions |
| `Tab` | Auto-complete with highlighted suggestion |
| `Enter` | Execute highlighted command |
| `Esc` | Exit command mode |

### Examples

```
:pod           → Switch to Pods view
:svc           → Switch to Services view (using alias)
:namespace     → Open namespace selector
:ctx           → Open context selector (using alias)
```

## Pod Logs Viewer

These shortcuts only work when viewing Pod logs (Logs tab is active).

| Key | Mode | Description |
|-----|------|-------------|
| `1` | Head | Show first 500 lines from beginning |
| `2` | Tail | Show last 100 lines and follow (default) |
| `3` | Last 5m | Show logs from last 5 minutes and follow |
| `4` | Last 15m | Show logs from last 15 minutes and follow |
| `5` | Last 500 | Show last 500 lines and follow |
| `6` | Last 1000 | Show last 1000 lines and follow |

**Note:** Log mode hotkeys only work when:
1. Detail panel is open
2. A Pod is selected
3. Logs tab is active

## Escape Key Hierarchy

The `Esc` key follows a hierarchical pattern, closing UI elements in this order:

1. Command mode (if active)
2. Debug drawer (if open)
3. Namespace/Context dropdowns (if open)
4. Detail panel fullscreen mode (if active)
5. Detail panel (if open)
6. Events drawer (if open)
7. Search (if active)

This means pressing `Esc` multiple times will progressively close overlays from innermost to outermost.

## Tips

- **Command mode aliases**: Use kubectl-style shortcuts like `po`, `svc`, `rs` for faster typing
- **Tab completion**: Start typing a command and press `Tab` to auto-complete
- **Arrow key navigation**: Use `↑`/`↓` to browse command suggestions
- **Prefix matching**: Commands match from the beginning (e.g., "dep" matches "deployment")
- **Quick switching**: Command mode is the fastest way to switch between resource types

## Configuration

All keyboard shortcuts are defined in the codebase as data, making them easy to extend:

- **Command mode**: `/internal/server/static/config.js` - `COMMANDS` array
- **Log modes**: `/internal/server/static/config.js` - `LOG_MODES` array
- **Global shortcuts**: `/internal/server/static/app.js` - `handleGlobalKeydown()` method

To add a new command, simply add an entry to the `COMMANDS` array. No code changes required!

## Adding Custom Commands

To add a new command, edit `/internal/server/static/config.js` and add an entry to the `COMMANDS` array:

```javascript
// Resource type command
{
  id: 'statefulset',
  type: 'resource',
  label: 'StatefulSet',
  aliases: ['statefulsets', 'sts'],
  target: 'StatefulSet',
  description: 'Switch to StatefulSets view'
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

The data-centric architecture means commands are pure configuration - just add to the array and the UI automatically generates buttons, autocomplete, and keyboard handlers.

## Table Navigation

k8v uses a kubectl-style table view for displaying resources. You can navigate through the table using vim-like keyboard shortcuts:

| Key | Action | Details |
|-----|--------|---------|
| `↑` or `k` | Move Up | Navigate to the previous row in the table |
| `↓` or `j` | Move Down | Navigate to the next row in the table |
| `Enter` | Open Details | Open the detail panel for the currently selected row |

**Navigation Behavior:**
- When you first use arrow keys or j/k, the first row is automatically selected
- Navigation wraps at the boundaries (stays on first/last row)
- Selected row is highlighted with a colored left border matching the resource type
- Selected row automatically scrolls into view if needed
- Navigation only works when no input field is focused

**Visual Feedback:**
- Selected row has a subtle background highlight
- Left border color indicates resource type (green for Pods, blue for Deployments, etc.)
- Smooth scrolling keeps the selected row visible

**Tip:** Combine with command mode for fastest navigation:
1. Press `:` to open command mode
2. Type `pod` or `svc` to switch views
3. Use `j`/`k` to navigate rows
4. Press `Enter` to view details
