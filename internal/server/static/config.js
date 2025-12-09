export const RESOURCE_TYPES = ['Pod', 'Deployment', 'ReplicaSet', 'Service', 'Ingress', 'ConfigMap', 'Secret', 'Node'];

export const LOCAL_STORAGE_KEYS = {
  namespace: 'k8v-namespace',
};

export const EVENTS_LIMIT = 100;

export const RELATIONSHIP_TYPES = [
  { key: 'ownedBy', label: 'Owned By' },
  { key: 'owns', label: 'Owns' },
  { key: 'dependsOn', label: 'Depends On' },
  { key: 'usedBy', label: 'Used By' },
  { key: 'exposes', label: 'Exposes' },
  { key: 'exposedBy', label: 'Exposed By' },
  { key: 'routesTo', label: 'Routes To' },
  { key: 'routedBy', label: 'Routed By' },
  { key: 'scheduledOn', label: 'Scheduled On' },
  { key: 'schedules', label: 'Schedules' },
];

export const API_PATHS = {
  namespaces: '/api/namespaces',
  stats: '/api/stats',
  resource: '/api/resource',
  resourcesWs: '/ws',
  logsWs: '/ws/logs',
  execWs: '/ws/exec',
};

// Single source of truth for all keyboard shortcuts
// Other data structures reference hotkey IDs to get the actual key bindings
export const HOTKEYS = {
  // General
  help:        { key: '?', description: 'Show keyboard shortcuts', category: 'General' },
  command:     { key: ':', description: 'Open command mode', category: 'General' },
  search:      { key: '/', description: 'Search by name', category: 'General' },
  escape:      { key: 'Escape', displayKey: 'Esc', description: 'Close modal / panel / search', category: 'General' },
  debug:       { key: 'd', description: 'Toggle debug drawer', category: 'General' },
  // Navigation
  navDown:     { key: 'j', altKey: 'ArrowDown', displayKey: 'j / ↓', description: 'Navigate down', category: 'Navigation' },
  navUp:       { key: 'k', altKey: 'ArrowUp', displayKey: 'k / ↑', description: 'Navigate up', category: 'Navigation' },
  select:      { key: 'Enter', description: 'Open selected resource', category: 'Navigation' },
  // Detail Panel
  fullscreen:   { key: '`', description: 'Toggle fullscreen detail panel', category: 'Detail Panel' },
  tabOverview:  { key: 'o', description: 'Switch to Overview tab', category: 'Detail Panel' },
  tabYaml:      { key: 'y', description: 'Switch to YAML tab', category: 'Detail Panel' },
  tabLogs:      { key: 'l', description: 'Switch to Logs tab', category: 'Detail Panel' },
  tabShell:     { key: 's', description: 'Switch to Shell tab', category: 'Detail Panel' },
  // Logs
  logHead:     { key: '1', description: 'Head (first 500 lines)', category: 'Logs' },
  logTail:     { key: '2', description: 'Tail (follow last 100)', category: 'Logs' },
  logLast5m:   { key: '3', description: 'Last 5 minutes', category: 'Logs' },
  logLast15m:  { key: '4', description: 'Last 15 minutes', category: 'Logs' },
  logLast500:  { key: '5', description: 'Last 500 lines', category: 'Logs' },
  logLast1000: { key: '6', description: 'Last 1000 lines', category: 'Logs' },
};

// Helper to check if a key event matches a hotkey ID
export function matchesHotkey(event, hotkeyId) {
  const hotkey = HOTKEYS[hotkeyId];
  if (!hotkey) return false;
  return event.key === hotkey.key || event.key === hotkey.altKey;
}

// Get display key for UI (e.g., button labels)
export function getHotkeyDisplay(hotkeyId) {
  const hotkey = HOTKEYS[hotkeyId];
  return hotkey?.displayKey || hotkey?.key || '';
}

export const LOG_MODES = [
  { id: 'head', label: 'Head', hotkeyId: 'logHead', headLines: 500, tailLines: null, sinceSeconds: null, follow: false },
  { id: 'tail', label: 'Tail', hotkeyId: 'logTail', headLines: null, tailLines: 100, sinceSeconds: null, follow: true },
  { id: 'last-5m', label: '-5m', hotkeyId: 'logLast5m', headLines: null, tailLines: null, sinceSeconds: 300, follow: true },
  { id: 'last-15m', label: '-15m', hotkeyId: 'logLast15m', headLines: null, tailLines: null, sinceSeconds: 900, follow: true },
  { id: 'last-500', label: '-500', hotkeyId: 'logLast500', headLines: null, tailLines: 500, sinceSeconds: null, follow: true },
  { id: 'last-1000', label: '-1000', hotkeyId: 'logLast1000', headLines: null, tailLines: 1000, sinceSeconds: null, follow: true },
];

export const COMMANDS = [
  // Resource type commands
  { id: 'pod', type: 'resource', label: 'Pod', aliases: ['pods', 'po'], target: 'Pod', description: 'Switch to Pods view' },
  { id: 'deployment', type: 'resource', label: 'Deployment', aliases: ['deployments', 'deploy'], target: 'Deployment', description: 'Switch to Deployments view' },
  { id: 'replicaset', type: 'resource', label: 'ReplicaSet', aliases: ['replicasets', 'rs'], target: 'ReplicaSet', description: 'Switch to ReplicaSets view' },
  { id: 'service', type: 'resource', label: 'Service', aliases: ['services', 'svc'], target: 'Service', description: 'Switch to Services view' },
  { id: 'ingress', type: 'resource', label: 'Ingress', aliases: ['ingresses', 'ing'], target: 'Ingress', description: 'Switch to Ingress view' },
  { id: 'configmap', type: 'resource', label: 'ConfigMap', aliases: ['configmaps', 'cm'], target: 'ConfigMap', description: 'Switch to ConfigMaps view' },
  { id: 'secret', type: 'resource', label: 'Secret', aliases: ['secrets'], target: 'Secret', description: 'Switch to Secrets view' },
  { id: 'node', type: 'resource', label: 'Node', aliases: ['nodes', 'no'], target: 'Node', description: 'Switch to Nodes view' },

  // Special commands
  { id: 'namespace', type: 'action', label: 'namespace', aliases: ['ns'], action: 'openNamespaceDropdown', description: 'Open namespace selector' },
  { id: 'context', type: 'action', label: 'context', aliases: ['ctx'], action: 'openContextDropdown', description: 'Open context selector' },
  { id: 'cluster', type: 'action', label: 'cluster', aliases: [], action: 'openContextDropdown', description: 'Open context selector' },
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

// Table view column configuration (kubectl-style)
export const TABLE_COLUMNS = {
  Pod: [
    { id: 'name', label: 'NAME', width: '200px', align: 'left', sortable: true },
    { id: 'ready', label: 'READY', width: '80px', align: 'center', sortable: false },
    { id: 'status', label: 'STATUS', width: '120px', align: 'left', sortable: false },
    { id: 'restarts', label: 'RESTARTS', width: '90px', align: 'center', sortable: false },
    { id: 'age', label: 'AGE', width: '80px', align: 'right', sortable: false },
    { id: 'namespace', label: 'NAMESPACE', width: '150px', align: 'left', sortable: false },
  ],
  Deployment: [
    { id: 'name', label: 'NAME', width: '200px', align: 'left', sortable: true },
    { id: 'ready', label: 'READY', width: '80px', align: 'center', sortable: false },
    { id: 'upToDate', label: 'UP-TO-DATE', width: '100px', align: 'center', sortable: false },
    { id: 'available', label: 'AVAILABLE', width: '100px', align: 'center', sortable: false },
    { id: 'age', label: 'AGE', width: '80px', align: 'right', sortable: false },
    { id: 'namespace', label: 'NAMESPACE', width: '150px', align: 'left', sortable: false },
  ],
  ReplicaSet: [
    { id: 'name', label: 'NAME', width: '250px', align: 'left', sortable: true },
    { id: 'desired', label: 'DESIRED', width: '80px', align: 'center', sortable: false },
    { id: 'current', label: 'CURRENT', width: '80px', align: 'center', sortable: false },
    { id: 'ready', label: 'READY', width: '80px', align: 'center', sortable: false },
    { id: 'age', label: 'AGE', width: '80px', align: 'right', sortable: false },
    { id: 'namespace', label: 'NAMESPACE', width: '150px', align: 'left', sortable: false },
  ],
  Service: [
    { id: 'name', label: 'NAME', width: '200px', align: 'left', sortable: true },
    { id: 'type', label: 'TYPE', width: '120px', align: 'left', sortable: false },
    { id: 'clusterIp', label: 'CLUSTER-IP', width: '120px', align: 'left', sortable: false },
    { id: 'externalIp', label: 'EXTERNAL-IP', width: '120px', align: 'left', sortable: false },
    { id: 'ports', label: 'PORT(S)', width: '150px', align: 'left', sortable: false },
    { id: 'age', label: 'AGE', width: '80px', align: 'right', sortable: false },
    { id: 'namespace', label: 'NAMESPACE', width: '150px', align: 'left', sortable: false },
  ],
  Ingress: [
    { id: 'name', label: 'NAME', width: '200px', align: 'left', sortable: true },
    { id: 'class', label: 'CLASS', width: '120px', align: 'left', sortable: false },
    { id: 'hosts', label: 'HOSTS', width: '250px', align: 'left', sortable: false },
    { id: 'address', label: 'ADDRESS', width: '150px', align: 'left', sortable: false },
    { id: 'ports', label: 'PORTS', width: '100px', align: 'left', sortable: false },
    { id: 'age', label: 'AGE', width: '80px', align: 'right', sortable: false },
    { id: 'namespace', label: 'NAMESPACE', width: '150px', align: 'left', sortable: false },
  ],
  ConfigMap: [
    { id: 'name', label: 'NAME', width: '250px', align: 'left', sortable: true },
    { id: 'data', label: 'DATA', width: '80px', align: 'center', sortable: false },
    { id: 'age', label: 'AGE', width: '80px', align: 'right', sortable: false },
    { id: 'namespace', label: 'NAMESPACE', width: '150px', align: 'left', sortable: false },
  ],
  Secret: [
    { id: 'name', label: 'NAME', width: '250px', align: 'left', sortable: true },
    { id: 'type', label: 'TYPE', width: '200px', align: 'left', sortable: false },
    { id: 'data', label: 'DATA', width: '80px', align: 'center', sortable: false },
    { id: 'age', label: 'AGE', width: '80px', align: 'right', sortable: false },
    { id: 'namespace', label: 'NAMESPACE', width: '150px', align: 'left', sortable: false },
  ],
  Node: [
    { id: 'name', label: 'NAME', width: '200px', align: 'left', sortable: true },
    { id: 'status', label: 'STATUS', width: '100px', align: 'left', sortable: false },
    { id: 'roles', label: 'ROLES', width: '150px', align: 'left', sortable: false },
    { id: 'age', label: 'AGE', width: '80px', align: 'right', sortable: false },
    { id: 'version', label: 'VERSION', width: '120px', align: 'left', sortable: false },
    { id: 'internalIp', label: 'INTERNAL-IP', width: '120px', align: 'left', sortable: false },
    { id: 'externalIp', label: 'EXTERNAL-IP', width: '120px', align: 'left', sortable: false },
  ],
};

export function getColumnsForType(resourceType) {
  return TABLE_COLUMNS[resourceType] || [];
}
