export const RESOURCE_TYPES = ['Pod', 'Deployment', 'ReplicaSet', 'Service', 'Ingress', 'ConfigMap', 'Secret'];

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
];

export const API_PATHS = {
  namespaces: '/api/namespaces',
  stats: '/api/stats',
  resourcesWs: '/ws',
  logsWs: '/ws/logs',
};
