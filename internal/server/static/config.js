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
