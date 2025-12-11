import { LOCAL_STORAGE_KEYS, RESOURCE_TYPES } from './config.js';

export function createInitialState() {
  return {
    resources: new Map(),
    events: [],
    snapshotComplete: false,
    snapshotCount: 0,
    namespaces: [],
    highlightedNamespaceIndex: -1,
    resourceTypes: RESOURCE_TYPES.slice(),
    filters: {
      type: 'Pod',
      namespace: localStorage.getItem(LOCAL_STORAGE_KEYS.namespace) || 'all',
      search: '',
    },
    ui: {
      eventsOpen: false,
      debugOpen: false,
      unreadEvents: 0,
      searchActive: false,
      activeMainTab: 'dashboard',
      activeDetailTab: 'overview',
      detailResourceId: null,
      detailFullscreen: false,
      selectedRowIndex: -1,
    },
    ws: {
      connectionId: 0,
      reconnectTimeout: null,
      manual: false,
    },
    log: {
      socket: null,
      currentKey: null,
      mode: 'tail',
    },
    exec: {
      socket: null,
      connected: false,
      container: '',
      terminalInstance: null,
      fitAddon: null,
    },
    sync: {
      syncing: false,
      synced: false,
      error: null,
      context: '',
    },
    command: {
      active: false,
      input: '',
      highlightedIndex: -1,
      suggestions: [],
    },
  };
}

export function resetForNewConnection(state) {
  state.resources.clear();
  state.events.length = 0;
  state.snapshotComplete = false;
  state.snapshotCount = 0;
  state.ui.unreadEvents = 0;

  // Reset sync state
  state.sync.syncing = true;
  state.sync.synced = false;
  state.sync.error = null;
}
