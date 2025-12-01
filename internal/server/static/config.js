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
};

export const LOG_MODES = [
  { id: 'head', label: 'Head', hotkey: '1', headLines: 500, tailLines: null, sinceSeconds: null, follow: false },
  { id: 'tail', label: 'Tail', hotkey: '2', headLines: null, tailLines: 100, sinceSeconds: null, follow: true },
  { id: 'last-5m', label: '-5m', hotkey: '3', headLines: null, tailLines: null, sinceSeconds: 300, follow: true },
  { id: 'last-15m', label: '-15m', hotkey: '4', headLines: null, tailLines: null, sinceSeconds: 900, follow: true },
  { id: 'last-500', label: '-500', hotkey: '5', headLines: null, tailLines: 500, sinceSeconds: null, follow: true },
  { id: 'last-1000', label: '-1000', hotkey: '6', headLines: null, tailLines: 1000, sinceSeconds: null, follow: true },
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
