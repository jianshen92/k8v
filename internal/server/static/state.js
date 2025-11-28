import { LOCAL_STORAGE_KEYS } from './config.js';

export function createInitialState() {
  return {
    resources: new Map(),
    events: [],
    snapshotComplete: false,
    snapshotCount: 0,
    namespaces: [],
    highlightedNamespaceIndex: -1,
    filters: {
      type: 'Pod',
      namespace: localStorage.getItem(LOCAL_STORAGE_KEYS.namespace) || 'all',
      search: '',
    },
    ui: {
      eventsOpen: false,
      unreadEvents: 0,
      searchActive: false,
      activeMainTab: 'dashboard',
      activeDetailTab: 'overview',
      detailResourceId: null,
    },
    ws: {
      connectionId: 0,
      reconnectTimeout: null,
      manual: false,
    },
    log: {
      socket: null,
      currentKey: null,
    },
  };
}

export function resetForNewConnection(state) {
  state.resources.clear();
  state.events.length = 0;
  state.snapshotComplete = false;
  state.snapshotCount = 0;
  state.ui.unreadEvents = 0;
}
