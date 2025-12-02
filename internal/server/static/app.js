import { API_PATHS, COMMANDS, EVENTS_LIMIT, LOCAL_STORAGE_KEYS, LOG_MODES, RELATIONSHIP_TYPES, RESOURCE_TYPES, findCommand, getCommandSuggestions, getColumnsForType } from './config.js';
import { createInitialState, resetForNewConnection } from './state.js';
import { createResourceSocket } from './ws.js';
import './dropdown.js';

// ========== Data Extraction Helper Functions ==========
// These functions extract cell values from resource objects for table columns

function extractCellValue(resource, columnId) {
  switch (columnId) {
    case 'name':
      return resource.name;
    case 'namespace':
      return resource.namespace || '-';
    case 'age':
      return formatAge(resource.createdAt);
    case 'status':
      return resource.status?.phase || '-';
    case 'ready':
      return resource.status?.ready || '-';

    // Pod-specific
    case 'restarts':
      return getPodRestartCount(resource);

    // Deployment-specific
    case 'upToDate':
      return resource.spec?.status?.updatedReplicas ?? '-';
    case 'available':
      return resource.spec?.status?.availableReplicas ?? '-';

    // ReplicaSet-specific
    case 'desired':
      return resource.spec?.spec?.replicas ?? '-';
    case 'current':
      return resource.spec?.status?.replicas ?? '-';

    // Service-specific
    case 'type':
      return getServiceType(resource);
    case 'clusterIp':
      return getServiceClusterIp(resource);
    case 'externalIp':
      return getServiceExternalIp(resource);
    case 'ports':
      return getServicePorts(resource);

    // Ingress-specific
    case 'class':
      return getIngressClass(resource);
    case 'hosts':
      return getIngressHosts(resource);
    case 'address':
      return getIngressAddress(resource);

    // ConfigMap/Secret-specific
    case 'data':
      return getDataCount(resource);

    // Node-specific
    case 'roles':
      return getNodeRoles(resource);
    case 'version':
      return getNodeVersion(resource);
    case 'internalIp':
      return getNodeInternalIp(resource);
    case 'externalIp':
      return getNodeExternalIp(resource);

    default:
      return '-';
  }
}

function formatAge(createdAt) {
  if (!createdAt) return '-';
  const now = new Date();
  const created = new Date(createdAt);
  const diffMs = now - created;
  const diffSec = Math.floor(diffMs / 1000);
  const diffMin = Math.floor(diffSec / 60);
  const diffHour = Math.floor(diffMin / 60);
  const diffDay = Math.floor(diffHour / 24);

  if (diffDay > 0) return `${diffDay}d`;
  if (diffHour > 0) return `${diffHour}h`;
  if (diffMin > 0) return `${diffMin}m`;
  return `${diffSec}s`;
}

function getPodRestartCount(resource) {
  if (!resource.spec?.status?.containerStatuses) return '0';
  const totalRestarts = resource.spec.status.containerStatuses.reduce((sum, container) => {
    return sum + (container.restartCount || 0);
  }, 0);
  return totalRestarts.toString();
}

function getServiceType(resource) {
  return resource.spec?.spec?.type || 'ClusterIP';
}

function getServiceClusterIp(resource) {
  return resource.spec?.spec?.clusterIP || '-';
}

function getServiceExternalIp(resource) {
  const spec = resource.spec?.spec;
  if (!spec) return '<none>';

  // Check for LoadBalancer IPs
  if (resource.spec?.status?.loadBalancer?.ingress) {
    const ips = resource.spec.status.loadBalancer.ingress
      .map(ing => ing.ip || ing.hostname)
      .filter(Boolean);
    if (ips.length > 0) return ips.join(',');
  }

  // Check for ExternalIPs
  if (spec.externalIPs && spec.externalIPs.length > 0) {
    return spec.externalIPs.join(',');
  }

  return '<none>';
}

function getServicePorts(resource) {
  const ports = resource.spec?.spec?.ports;
  if (!ports || ports.length === 0) return '-';

  return ports.map(p => {
    const port = p.port;
    const nodePort = p.nodePort ? `:${p.nodePort}` : '';
    const protocol = p.protocol && p.protocol !== 'TCP' ? `/${p.protocol}` : '';
    return `${port}${nodePort}${protocol}`;
  }).join(',');
}

function getIngressClass(resource) {
  return resource.spec?.spec?.ingressClassName ||
         resource.annotations?.['kubernetes.io/ingress.class'] ||
         '-';
}

function getIngressHosts(resource) {
  const rules = resource.spec?.spec?.rules;
  if (!rules || rules.length === 0) return '*';

  const hosts = rules.map(r => r.host || '*').filter(Boolean);
  return hosts.join(',') || '*';
}

function getIngressAddress(resource) {
  const ingress = resource.spec?.status?.loadBalancer?.ingress;
  if (!ingress || ingress.length === 0) return '-';

  const addresses = ingress.map(ing => ing.ip || ing.hostname).filter(Boolean);
  return addresses.join(',') || '-';
}

function getDataCount(resource) {
  const data = resource.spec?.data;
  if (!data) return '0';
  return Object.keys(data).length.toString();
}

function getNodeRoles(resource) {
  const labels = resource.labels || {};
  const roles = [];

  for (const [key, value] of Object.entries(labels)) {
    if (key === 'node-role.kubernetes.io/master' || key === 'node-role.kubernetes.io/control-plane') {
      roles.push('control-plane');
    } else if (key.startsWith('node-role.kubernetes.io/')) {
      roles.push(key.replace('node-role.kubernetes.io/', ''));
    }
  }

  return roles.length > 0 ? roles.join(',') : '<none>';
}

function getNodeVersion(resource) {
  return resource.spec?.status?.nodeInfo?.kubeletVersion || '-';
}

function getNodeInternalIp(resource) {
  const addresses = resource.spec?.status?.addresses;
  if (!addresses) return '-';

  const internal = addresses.find(addr => addr.type === 'InternalIP');
  return internal?.address || '-';
}

function getNodeExternalIp(resource) {
  const addresses = resource.spec?.status?.addresses;
  if (!addresses) return '<none>';

  const external = addresses.find(addr => addr.type === 'ExternalIP');
  return external?.address || '<none>';
}

class App {
  constructor() {
    this.state = createInitialState();
    this.filteredNamespaces = [];
    this.currentContainerCount = 0;
    this.currentSingleContainerValue = '';

    this.wsManager = createResourceSocket(this.state, {
      buildUrl: this.buildWsUrl.bind(this),
      onOpen: this.onSocketOpen.bind(this),
      onMessage: this.handleResourceEvent.bind(this),
      onSyncStatus: this.handleSyncStatus.bind(this),
      onClose: this.onSocketClose.bind(this),
      onError: this.onSocketError.bind(this),
      onSnapshotComplete: this.onSnapshotComplete.bind(this),
    });
  }

  // ---------- Init ----------
  async init() {
    this.attachUIListeners();
    this.renderFilters();
    this.renderResourceList();
    this.renderLogModeButtons();
    this.setupContextDropdown();
    this.setupNamespaceDropdown();
    this.setupSearchFilter();
    this.setupCommandMode();

    // Get actual backend context first
    await this.fetchCurrentContext();

    // Then fetch available contexts (for dropdown options)
    this.fetchAndDisplayContexts();
    this.fetchNamespaces();

    await this.fetchAndDisplayStats();

    this.wsManager.connect();
    feather.replace();
  }

  // ---------- Sync Status (WebSocket Push) ----------
  handleSyncStatus(syncEvent) {
    console.log('[App] Sync status update:', syncEvent);

    // Update state
    this.state.sync.syncing = syncEvent.syncing;
    this.state.sync.synced = syncEvent.synced;
    this.state.sync.error = syncEvent.error || null;
    this.state.sync.context = syncEvent.context;

    // Update UI
    this.updateSyncUI();

    // Reactive data fetching when synced
    if (syncEvent.synced && !syncEvent.syncing) {
      console.log('[App] Cache synced, refreshing stats and namespaces');
      this.fetchNamespaces();
      this.fetchAndDisplayStats();
    }
  }

  updateSyncUI() {
    const loadingState = document.getElementById('loading-state');
    const resourceTableWrapper = document.querySelector('.resource-table-wrapper');
    const loadingText = loadingState?.querySelector('.loading-text');
    const loadingSubtext = document.getElementById('loading-subtext');

    if (this.state.sync.syncing) {
      if (loadingState) loadingState.style.display = 'flex';
      if (resourceTableWrapper) resourceTableWrapper.style.display = 'none';
      if (loadingText) loadingText.textContent = 'Syncing informer caches...';
      if (loadingSubtext) loadingSubtext.textContent = 'This may take a while for large clusters';
    } else if (this.state.sync.synced) {
      if (loadingState) loadingState.style.display = 'none';
      if (resourceTableWrapper) resourceTableWrapper.style.display = 'block';
    } else if (this.state.sync.error) {
      if (loadingState) loadingState.style.display = 'flex';
      if (loadingText) loadingText.textContent = 'Sync failed';
      if (loadingSubtext) loadingSubtext.textContent = this.state.sync.error;
    }
  }

  // ---------- WebSocket helpers ----------
  buildWsUrl() {
    const params = [];
    const { namespace, type } = this.state.filters;
    if (namespace !== 'all') params.push(`namespace=${namespace}`);
    if (type !== 'all') params.push(`type=${type}`);
    const queryString = params.length ? `?${params.join('&')}` : '';
    return `${API_PATHS.resourcesWs}${queryString}`;
  }

  onSocketOpen() {
    const nsLabel = this.state.filters.namespace === 'all' ? 'All Namespaces' : this.state.filters.namespace;
    document.getElementById('connection-status').textContent = `Connected (${nsLabel})`;
    this.state.ws.manual = false;
  }

  onSocketError() {
    document.getElementById('connection-status').textContent = 'Error';
  }

  onSocketClose() {
    document.getElementById('connection-status').textContent = 'Disconnected';
  }

  onSnapshotComplete() {
    console.log(`[App] Snapshot complete, rendering ${this.state.resources.size} resources`);
    this.renderResourceList();
    this.renderEvents();
  }

  // ---------- Namespace Filters ----------
  async fetchNamespaces() {
    try {
      const response = await fetch(API_PATHS.namespaces);
      const data = await response.json();
      this.state.namespaces = ['all', ...(data.namespaces || [])];
    } catch (err) {
      console.error('Failed to fetch namespaces:', err);
      this.state.namespaces = ['all'];
    }
    const options = this.state.namespaces.map(ns => ({
      value: ns,
      label: ns === 'all' ? 'All Namespaces' : ns,
    }));
    if (this.namespaceDropdown) {
      this.namespaceDropdown.setOptions(options, this.state.filters.namespace);
    }
  }

  setNamespace(namespace) {
    this.state.filters.namespace = namespace;
    localStorage.setItem(LOCAL_STORAGE_KEYS.namespace, namespace);
    this.reconnectWithNamespace();
  }

  reconnectWithNamespace() {
    this.wsManager.disconnect(true);
    resetForNewConnection(this.state);
    this.renderResourceList();
    if (this.namespaceDropdown) {
      this.namespaceDropdown.setValue(this.state.filters.namespace);
    }
    this.fetchAndDisplayStats().then(() => {
      this.wsManager.connect();
    });
  }

  // ---------- Filters ----------
  renderFilters() {
    const container = document.getElementById('filters');
    container.innerHTML = '';
    for (const t of RESOURCE_TYPES) {
      const btn = document.createElement('button');
      btn.className = 'filter-btn' + (this.state.filters.type === t ? ' active' : '');
      btn.textContent = t + (t === 'Ingress' ? '' : (t.endsWith('s') ? '' : 's'));
      btn.dataset.filterType = t;
      btn.addEventListener('click', () => this.setFilter(t));
      container.appendChild(btn);
    }
  }

  setFilter(type) {
    this.state.filters.type = type;
    this.reconnectWithFilter();
  }

  reconnectWithFilter() {
    this.wsManager.disconnect(true);
    resetForNewConnection(this.state);
    this.renderResourceList();
    this.renderFilters();
    this.fetchAndDisplayStats().then(() => {
      this.wsManager.connect();
    });
  }

  // ---------- Search ----------
  activateSearch() {
    const namespaceMenu = document.getElementById('namespace-menu');
    if (namespaceMenu && namespaceMenu.style.display !== 'none') return;

    const activeElement = document.activeElement;
    if (activeElement && (activeElement.tagName === 'TEXTAREA' || activeElement.tagName === 'INPUT' || activeElement.isContentEditable)) {
      return;
    }

    this.state.ui.searchActive = true;
    document.getElementById('search-trigger').style.display = 'none';
    document.getElementById('search-active').style.display = 'flex';
    const input = document.getElementById('search-input');
    input.focus();
    feather.replace();
  }

  deactivateSearch() {
    this.state.ui.searchActive = false;
    this.state.filters.search = '';
    document.getElementById('search-trigger').style.display = 'flex';
    document.getElementById('search-active').style.display = 'none';
    document.getElementById('search-input').value = '';
    this.renderResourceList();
  }

  clearSearch() {
    this.deactivateSearch();
  }

  handleSearchInput(event) {
    this.state.filters.search = event.target.value.toLowerCase().trim();
    this.renderResourceList();
  }

  // ---------- Command Mode ----------
  activateCommandMode() {
    // Don't activate if input is focused or dropdowns are open
    const activeElement = document.activeElement;
    const isInputFocused = activeElement && (
      activeElement.tagName === 'INPUT' ||
      activeElement.tagName === 'TEXTAREA' ||
      activeElement.isContentEditable
    );
    if (isInputFocused) return;
    if (this.namespaceDropdown && this.namespaceDropdown.isOpen()) return;
    if (this.contextDropdown && this.contextDropdown.isOpen()) return;

    this.state.command.active = true;
    this.state.command.input = '';
    this.state.command.highlightedIndex = -1;
    this.state.command.suggestions = COMMANDS.slice(); // All commands initially

    const overlay = document.getElementById('command-mode-overlay');
    overlay.style.display = 'flex';

    const input = document.getElementById('command-input');
    input.value = '';
    input.focus();

    this.renderCommandSuggestions();
  }

  deactivateCommandMode() {
    this.state.command.active = false;
    this.state.command.input = '';
    this.state.command.highlightedIndex = -1;
    this.state.command.suggestions = [];

    const overlay = document.getElementById('command-mode-overlay');
    overlay.style.display = 'none';
  }

  handleCommandInput(event) {
    this.state.command.input = event.target.value;
    this.state.command.highlightedIndex = -1;
    this.state.command.suggestions = getCommandSuggestions(this.state.command.input);
    this.renderCommandSuggestions();
  }

  renderCommandSuggestions() {
    const container = document.getElementById('command-suggestions');
    container.innerHTML = '';

    if (this.state.command.suggestions.length === 0) {
      const empty = document.createElement('div');
      empty.className = 'command-suggestion';
      empty.style.color = '#666';
      empty.style.cursor = 'default';
      empty.textContent = 'No matching commands';
      container.appendChild(empty);
      return;
    }

    this.state.command.suggestions.forEach((cmd, index) => {
      const item = document.createElement('div');
      item.className = 'command-suggestion';
      if (index === this.state.command.highlightedIndex) {
        item.classList.add('highlighted');
      }
      item.dataset.index = index;

      // Main content
      const main = document.createElement('div');
      main.className = 'command-suggestion-main';

      // Type badge
      const typeBadge = document.createElement('span');
      typeBadge.className = `command-suggestion-type ${cmd.type}`;
      typeBadge.textContent = cmd.type;
      main.appendChild(typeBadge);

      // Label
      const label = document.createElement('span');
      label.className = 'command-suggestion-label';
      label.textContent = cmd.label;
      main.appendChild(label);

      // Aliases (if any)
      if (cmd.aliases.length > 0) {
        const aliases = document.createElement('span');
        aliases.className = 'command-suggestion-aliases';
        aliases.textContent = `(${cmd.aliases.join(', ')})`;
        main.appendChild(aliases);
      }

      item.appendChild(main);

      // Description
      const desc = document.createElement('div');
      desc.className = 'command-suggestion-description';
      desc.textContent = cmd.description;
      item.appendChild(desc);

      // Click handler
      item.addEventListener('click', () => this.executeCommand(cmd));

      container.appendChild(item);
    });
  }

  executeCommand(cmd) {
    if (!cmd) return;

    if (cmd.type === 'resource') {
      // Switch resource type filter
      this.setFilter(cmd.target);
      this.deactivateCommandMode();
    } else if (cmd.type === 'action') {
      // Execute special action
      this.deactivateCommandMode();
      if (cmd.action === 'openNamespaceDropdown') {
        setTimeout(() => {
          if (this.namespaceDropdown) {
            this.namespaceDropdown.open();
          }
        }, 100); // Small delay for smooth transition
      } else if (cmd.action === 'openContextDropdown') {
        setTimeout(() => {
          if (this.contextDropdown) {
            this.contextDropdown.open();
          }
        }, 100);
      }
    }
  }

  handleCommandKeydown(event) {
    if (!this.state.command.active) return;

    const suggestions = this.state.command.suggestions;

    if (event.key === 'Escape') {
      event.preventDefault();
      this.deactivateCommandMode();
      return;
    }

    if (event.key === 'ArrowDown') {
      event.preventDefault();
      if (suggestions.length === 0) return;
      this.state.command.highlightedIndex = Math.min(
        this.state.command.highlightedIndex + 1,
        suggestions.length - 1
      );
      this.renderCommandSuggestions();
      this.scrollCommandSuggestionIntoView();
      return;
    }

    if (event.key === 'ArrowUp') {
      event.preventDefault();
      if (suggestions.length === 0) return;
      this.state.command.highlightedIndex = Math.max(
        this.state.command.highlightedIndex - 1,
        0
      );
      this.renderCommandSuggestions();
      this.scrollCommandSuggestionIntoView();
      return;
    }

    if (event.key === 'Tab') {
      event.preventDefault();
      // Auto-complete with highlighted suggestion
      if (this.state.command.highlightedIndex >= 0 &&
          this.state.command.highlightedIndex < suggestions.length) {
        const cmd = suggestions[this.state.command.highlightedIndex];
        const input = document.getElementById('command-input');
        input.value = cmd.label;
        this.state.command.input = cmd.label;
        this.state.command.suggestions = getCommandSuggestions(cmd.label);
        this.renderCommandSuggestions();
      } else if (suggestions.length === 1) {
        // Auto-complete with only suggestion
        const cmd = suggestions[0];
        const input = document.getElementById('command-input');
        input.value = cmd.label;
        this.state.command.input = cmd.label;
      }
      return;
    }

    if (event.key === 'Enter') {
      event.preventDefault();

      // Execute highlighted suggestion
      if (this.state.command.highlightedIndex >= 0 &&
          this.state.command.highlightedIndex < suggestions.length) {
        this.executeCommand(suggestions[this.state.command.highlightedIndex]);
        return;
      }

      // Execute first matching command if input exactly matches
      const exactMatch = findCommand(this.state.command.input);
      if (exactMatch) {
        this.executeCommand(exactMatch);
        return;
      }

      // Execute first suggestion if only one exists
      if (suggestions.length === 1) {
        this.executeCommand(suggestions[0]);
        return;
      }

      // If no match, do nothing (could add visual feedback here)
      return;
    }
  }

  scrollCommandSuggestionIntoView() {
    const container = document.getElementById('command-suggestions');
    const highlighted = container.querySelector('.command-suggestion.highlighted');
    if (highlighted) {
      highlighted.scrollIntoView({ block: 'nearest', behavior: 'smooth' });
    }
  }

  setupCommandMode() {
    const commandInput = document.getElementById('command-input');
    const commandBackdrop = document.querySelector('.command-backdrop');

    if (commandInput) {
      commandInput.addEventListener('input', (e) => this.handleCommandInput(e));
      commandInput.addEventListener('keydown', (e) => this.handleCommandKeydown(e));
    }

    if (commandBackdrop) {
      commandBackdrop.addEventListener('click', () => this.deactivateCommandMode());
    }
  }

  handleGlobalKeydown(event) {
    const activeElement = document.activeElement;
    const isInputFocused = activeElement && (
      activeElement.tagName === 'INPUT' ||
      activeElement.tagName === 'TEXTAREA' ||
      activeElement.isContentEditable
    );

    // Command mode activation - highest priority
    if (event.key === ':' && !isInputFocused) {
      event.preventDefault();
      this.activateCommandMode();
      return;
    }

    if (event.key === '/' && !isInputFocused) {
      event.preventDefault();
      this.activateSearch();
      return;
    }

    // Table row navigation
    if (!isInputFocused && (event.key === 'ArrowUp' || event.key === 'k' || event.key === 'ArrowDown' || event.key === 'j')) {
      event.preventDefault();
      this.handleTableNavigation(event.key);
      return;
    }

    // Enter to open selected row's detail panel
    if (event.key === 'Enter' && !isInputFocused && this.state.ui.selectedRowIndex >= 0) {
      event.preventDefault();
      this.openSelectedRowDetail();
      return;
    }

    if (event.key === 'd' && !isInputFocused) {
      event.preventDefault();
      this.toggleDebugDrawer();
      return;
    }

    // Log mode hotkeys - dynamically generated from LOG_MODES data
    if (!isInputFocused && this.state.ui.activeDetailTab === 'logs' && this.state.ui.detailResourceId) {
      const resource = this.getResourceById(this.state.ui.detailResourceId);
      if (resource && resource.type === 'Pod') {
        // Find mode by hotkey
        const mode = LOG_MODES.find(m => m.hotkey === event.key);
        if (mode) {
          event.preventDefault();
          this.setLogMode(mode.id);
          return;
        }
      }
    }

    if (event.key === 'Escape') {
      // Command mode - highest priority
      if (this.state.command.active) {
        this.deactivateCommandMode();
        return;
      }

      // Close debug drawer first if open
      const debugDrawer = document.getElementById('debug-drawer');
      if (debugDrawer && debugDrawer.classList.contains('visible')) {
        this.toggleDebugDrawer();
        return;
      }
      if (this.namespaceDropdown && this.namespaceDropdown.isOpen()) {
        this.namespaceDropdown.close();
        return;
      }

      const detailPanel = document.getElementById('detail-panel');
      if (detailPanel && detailPanel.classList.contains('visible')) {
        // If in fullscreen mode, exit fullscreen first
        if (this.state.ui.detailFullscreen) {
          this.toggleFullscreen();
          return;
        }
        // Otherwise close the panel
        this.closeDetail();
        return;
      }

      if (this.state.ui.eventsOpen) {
        this.toggleEventsDrawer();
        return;
      }

      if (this.state.ui.searchActive) {
        this.clearSearch();
      }
    }
  }

  setupSearchFilter() {
    document.getElementById('search-trigger').addEventListener('click', () => this.activateSearch());
    document.getElementById('search-input').addEventListener('input', (e) => this.handleSearchInput(e));
    document.getElementById('search-clear').addEventListener('click', () => this.clearSearch());
    document.addEventListener('keydown', (e) => this.handleGlobalKeydown(e));
  }

  // ---------- Stats ----------
  async fetchAndDisplayStats() {
    try {
      const nsParam = this.state.filters.namespace === 'all' ? '' : `?namespace=${this.state.filters.namespace}`;
      const response = await fetch(`${API_PATHS.stats}${nsParam}`);
      const counts = await response.json();

      document.getElementById('stat-total').textContent = counts.total || 0;
      document.getElementById('stat-pod').textContent = counts['Pod'] || 0;
      document.getElementById('stat-deployment').textContent = counts['Deployment'] || 0;
      document.getElementById('stat-replicaset').textContent = counts['ReplicaSet'] || 0;
      document.getElementById('stat-service').textContent = counts['Service'] || 0;
      document.getElementById('stat-ingress').textContent = counts['Ingress'] || 0;
      document.getElementById('stat-configmap').textContent = counts['ConfigMap'] || 0;
      document.getElementById('stat-secret').textContent = counts['Secret'] || 0;
      document.getElementById('stat-node').textContent = counts['Node'] || 0;
      document.getElementById('resource-count').textContent = `${counts.total || 0} resources`;

      console.log('[Stats] Loaded counts:', counts);
    } catch (error) {
      console.error('[Stats] Failed to fetch stats:', error);
    }
  }

  // Fetch a single resource by ID from the backend
  async fetchResource(resourceId) {
    try {
      const response = await fetch(`${API_PATHS.resource}?id=${encodeURIComponent(resourceId)}`);

      if (!response.ok) {
        if (response.status === 404) {
          throw new Error(`Resource not found: ${resourceId}`);
        }
        throw new Error(`Failed to fetch resource: ${response.statusText}`);
      }

      const resource = await response.json();
      this.state.resources.set(resource.id, resource);
      return resource;
    } catch (error) {
      console.error('Error fetching resource:', error);
      throw error;
    }
  }

  // ---------- Resources rendering (Table View) ----------
  typeToClass(t) { return t.toLowerCase(); }

  renderTableHeader() {
    const headerRow = document.getElementById('table-header-row');
    headerRow.innerHTML = '';

    const columns = getColumnsForType(this.state.filters.type);

    columns.forEach(column => {
      const th = document.createElement('th');
      th.textContent = column.label;
      th.style.width = column.width;
      th.classList.add(`align-${column.align}`);

      if (column.sortable) {
        th.classList.add('sortable');
        th.setAttribute('title', 'Click to sort by ' + column.label);
      }

      headerRow.appendChild(th);
    });
  }

  createTableRow(resource, rowIndex) {
    const row = document.createElement('tr');
    row.className = this.typeToClass(resource.type);
    row.dataset.resourceId = resource.id;
    row.dataset.rowIndex = rowIndex;

    // Mark as selected if this is the selected row
    if (rowIndex === this.state.ui.selectedRowIndex) {
      row.classList.add('selected');
    }

    // Click to open detail panel
    row.addEventListener('click', () => {
      this.selectRow(rowIndex);
      this.showResource(resource.id);
    });

    const columns = getColumnsForType(resource.type);

    columns.forEach(column => {
      const td = document.createElement('td');
      td.classList.add(`align-${column.align}`);

      // Special handling for name column (add health indicator)
      if (column.id === 'name') {
        td.classList.add('cell-name', 'cell-with-health');

        const healthDot = document.createElement('div');
        healthDot.className = `cell-health-dot ${resource.health || 'unknown'}`;
        td.appendChild(healthDot);

        const nameText = document.createTextNode(extractCellValue(resource, column.id));
        td.appendChild(nameText);
      } else if (column.id === 'status') {
        // Special handling for status column (color coding)
        td.classList.add('cell-status');
        const statusValue = extractCellValue(resource, column.id);
        td.textContent = statusValue;
        td.classList.add(statusValue.replace(/\s/g, ''));
      } else {
        // Regular cell
        td.textContent = extractCellValue(resource, column.id);
      }

      row.appendChild(td);
    });

    return row;
  }

  renderResourceList() {
    const tbody = document.getElementById('resource-table-body');
    tbody.innerHTML = '';

    const list = Array.from(this.state.resources.values())
      .filter(r => {
        if (r.type !== this.state.filters.type) return false;
        if (this.state.filters.search && !r.name.toLowerCase().includes(this.state.filters.search)) return false;
        return true;
      })
      .sort((a, b) => a.name.localeCompare(b.name));

    // Render table header
    this.renderTableHeader();

    // Handle empty state
    if (!list.length) {
      const row = document.createElement('tr');
      row.className = 'empty-state-row';
      const td = document.createElement('td');
      td.colSpan = getColumnsForType(this.state.filters.type).length;

      const emptyMessage = this.state.filters.search
        ? `No ${this.state.filters.type}s matching "${this.state.filters.search}"`
        : `No ${this.state.filters.type}s found`;
      const emptyDetail = this.state.filters.namespace !== 'all'
        ? `in namespace "${this.state.filters.namespace}"`
        : 'in cluster';

      const emptyContent = document.createElement('div');
      emptyContent.className = 'empty-state-content';
      emptyContent.innerHTML = `
        <div><i data-feather="inbox" style="width: 48px; height: 48px; stroke-width: 1.5;"></i></div>
        <div style="font-size: 16px; margin: 12px 0 8px;">${emptyMessage}</div>
        <div style="font-size: 13px; color: #666;">${emptyDetail}</div>
      `;

      td.appendChild(emptyContent);
      row.appendChild(td);
      tbody.appendChild(row);
      feather.replace();
      return;
    }

    // Render rows
    list.forEach((resource, index) => {
      tbody.appendChild(this.createTableRow(resource, index));
    });

    // Reset selection if out of bounds
    if (this.state.ui.selectedRowIndex >= list.length) {
      this.state.ui.selectedRowIndex = -1;
    }
  }

  selectRow(rowIndex) {
    // Remove previous selection
    const previousRow = document.querySelector('.resource-table tbody tr.selected');
    if (previousRow) {
      previousRow.classList.remove('selected');
    }

    // Update state
    this.state.ui.selectedRowIndex = rowIndex;

    // Add new selection
    const newRow = document.querySelector(`.resource-table tbody tr[data-row-index="${rowIndex}"]`);
    if (newRow) {
      newRow.classList.add('selected');
      // Scroll into view if needed
      newRow.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
    }
  }

  handleTableNavigation(key) {
    const tbody = document.getElementById('resource-table-body');
    const rows = tbody.querySelectorAll('tr:not(.empty-state-row)');

    if (rows.length === 0) return;

    let newIndex = this.state.ui.selectedRowIndex;

    // Navigate up (ArrowUp or k)
    if (key === 'ArrowUp' || key === 'k') {
      if (newIndex <= 0) {
        newIndex = 0; // Stay at first row
      } else {
        newIndex--;
      }
    }

    // Navigate down (ArrowDown or j)
    if (key === 'ArrowDown' || key === 'j') {
      if (newIndex < 0) {
        newIndex = 0; // Start at first row if nothing selected
      } else if (newIndex >= rows.length - 1) {
        newIndex = rows.length - 1; // Stay at last row
      } else {
        newIndex++;
      }
    }

    this.selectRow(newIndex);
  }

  openSelectedRowDetail() {
    if (this.state.ui.selectedRowIndex < 0) return;

    const selectedRow = document.querySelector(`.resource-table tbody tr[data-row-index="${this.state.ui.selectedRowIndex}"]`);
    if (!selectedRow) return;

    const resourceId = selectedRow.dataset.resourceId;
    if (resourceId) {
      this.showResource(resourceId);
    }
  }

  updateResourceInList(resourceId, eventType) {
    const tbody = document.getElementById('resource-table-body');
    const existingRow = tbody.querySelector(`[data-resource-id="${resourceId}"]`);

    if (eventType === 'DELETED') {
      if (existingRow) {
        const deletedIndex = parseInt(existingRow.dataset.rowIndex, 10);
        existingRow.remove();

        // Adjust row indices after deletion
        this.reindexTableRows();

        // Adjust selected index if needed
        if (this.state.ui.selectedRowIndex === deletedIndex) {
          this.state.ui.selectedRowIndex = -1;
        } else if (this.state.ui.selectedRowIndex > deletedIndex) {
          this.state.ui.selectedRowIndex--;
        }
      }
      return;
    }

    const resource = this.state.resources.get(resourceId);
    if (!resource) return;

    const matchesTypeFilter = resource.type === this.state.filters.type;
    const matchesSearchFilter = !this.state.filters.search || resource.name.toLowerCase().includes(this.state.filters.search);
    const matchesFilter = matchesTypeFilter && matchesSearchFilter;

    if (!matchesFilter) {
      if (existingRow) {
        const deletedIndex = parseInt(existingRow.dataset.rowIndex, 10);
        existingRow.remove();
        this.reindexTableRows();

        if (this.state.ui.selectedRowIndex === deletedIndex) {
          this.state.ui.selectedRowIndex = -1;
        } else if (this.state.ui.selectedRowIndex > deletedIndex) {
          this.state.ui.selectedRowIndex--;
        }
      }
      return;
    }

    if (eventType === 'MODIFIED' && existingRow) {
      const rowIndex = parseInt(existingRow.dataset.rowIndex, 10);
      const newRow = this.createTableRow(resource, rowIndex);
      existingRow.replaceWith(newRow);
    } else if (eventType === 'ADDED') {
      const allVisible = Array.from(this.state.resources.values())
        .filter(r => r.type === this.state.filters.type)
        .filter(r => !this.state.filters.search || r.name.toLowerCase().includes(this.state.filters.search))
        .sort((a, b) => a.name.localeCompare(b.name));

      const index = allVisible.findIndex(r => r.id === resourceId);
      const newRow = this.createTableRow(resource, index);

      if (index === allVisible.length - 1 || tbody.children.length === 0) {
        tbody.appendChild(newRow);
      } else {
        const nextResource = allVisible[index + 1];
        if (nextResource) {
          const nextRow = tbody.querySelector(`[data-resource-id="${nextResource.id}"]`);
          if (nextRow) {
            tbody.insertBefore(newRow, nextRow);
          } else {
            tbody.appendChild(newRow);
          }
        } else {
          tbody.appendChild(newRow);
        }
      }

      // Reindex rows after addition
      this.reindexTableRows();

      // Adjust selected index if needed
      if (this.state.ui.selectedRowIndex >= index) {
        this.state.ui.selectedRowIndex++;
      }
    }
  }

  reindexTableRows() {
    const tbody = document.getElementById('resource-table-body');
    const rows = tbody.querySelectorAll('tr:not(.empty-state-row)');
    rows.forEach((row, index) => {
      row.dataset.rowIndex = index;
    });
  }

  // ---------- Events ----------
  toggleEventsDrawer() {
    this.state.ui.eventsOpen = !this.state.ui.eventsOpen;
    const drawer = document.getElementById('events-drawer');
    if (this.state.ui.eventsOpen) {
      drawer.classList.add('visible');
      this.state.ui.unreadEvents = 0;
      this.updateEventsBadge();
      this.renderEvents();
    } else {
      drawer.classList.remove('visible');
    }
  }

  toggleDebugDrawer() {
    this.state.ui.debugOpen = !this.state.ui.debugOpen;
    const drawer = document.getElementById('debug-drawer');
    if (this.state.ui.debugOpen) {
      drawer.classList.add('visible');
      this.renderDebugData();
    } else {
      drawer.classList.remove('visible');
    }
  }

  renderDebugData() {
    const debugJson = document.getElementById('debug-json');

    // Convert Map to object for JSON serialization
    const resourcesArray = Array.from(this.state.resources.values());

    const debugData = {
      timestamp: new Date().toISOString(),
      resourceCount: this.state.resources.size,
      filters: this.state.filters,
      ui: {
        activeDetailTab: this.state.ui.activeDetailTab,
        detailResourceId: this.state.ui.detailResourceId,
        searchActive: this.state.ui.searchActive,
        eventsOpen: this.state.ui.eventsOpen,
      },
      websocket: {
        connected: this.state.ws.connected,
        snapshotComplete: this.state.ws.snapshotComplete,
      },
      resources: resourcesArray
    };

    debugJson.textContent = JSON.stringify(debugData, null, 2);
  }

  updateEventsBadge() {
    const badge = document.getElementById('events-badge');
    if (this.state.ui.unreadEvents > 0) {
      badge.textContent = this.state.ui.unreadEvents > 99 ? '99+' : this.state.ui.unreadEvents;
      badge.style.display = 'block';
    } else {
      badge.style.display = 'none';
    }
  }

  renderEvents() {
    if (!this.state.ui.eventsOpen) return;
    const el = document.getElementById('events');
    el.innerHTML = '';
    for (const e of this.state.events) {
      const item = document.createElement('div');
      item.className = 'event-item ' + (e.type === 'MODIFIED' ? 'warning' : e.type === 'DELETED' ? 'error' : '');

      const header = document.createElement('div');
      header.className = 'event-header';

      const typeSpan = document.createElement('span');
      typeSpan.className = 'event-type ' + e.type;
      typeSpan.textContent = e.type;

      const time = document.createElement('div');
      time.className = 'event-time';
      time.textContent = new Date(e.time).toLocaleTimeString();

      header.appendChild(typeSpan);
      header.appendChild(time);

      const msg = document.createElement('div');
      msg.className = 'event-message';
      msg.textContent = `${e.resource.type} › ${e.resource.namespace || 'default'} › ${e.resource.name}`;

      item.appendChild(header);
      item.appendChild(msg);
      el.appendChild(item);
    }
  }

  handleResourceEvent(event) {
    const resourceId = event.resource.id;

    if (event.type === 'DELETED') {
      this.state.resources.delete(resourceId);
    } else {
      this.state.resources.set(resourceId, event.resource);
    }

    this.state.events.unshift({ type: event.type, resource: event.resource, time: Date.now() });
    if (this.state.events.length > EVENTS_LIMIT) this.state.events.pop();

    if (!this.state.ui.eventsOpen && this.state.snapshotComplete) {
      this.state.ui.unreadEvents++;
      this.updateEventsBadge();
    }

    // Only render if snapshot is complete (incremental updates)
    // During snapshot, we buffer resources without rendering for speed
    if (this.state.snapshotComplete) {
      this.updateResourceInList(resourceId, event.type);
      this.renderEvents();
    }
    // During snapshot: do nothing, just buffer in state.resources
  }

  // ---------- Detail & Logs ----------
  getResourceById(id) {
    return this.state.resources.get(id);
  }

  // Get log mode config by id
  getLogMode(modeId) {
    return LOG_MODES.find(m => m.id === modeId);
  }

  // Generate log mode buttons from data
  renderLogModeButtons() {
    const container = document.getElementById('logs-mode-buttons');
    if (!container) return;

    container.innerHTML = '';

    LOG_MODES.forEach(mode => {
      const btn = document.createElement('button');
      btn.className = 'logs-mode-btn';
      btn.dataset.mode = mode.id;
      btn.title = `Hotkey: ${mode.hotkey}`;

      // Set active if this is the current mode
      if (mode.id === this.state.log.mode) {
        btn.classList.add('active');
      }

      // Mode label
      const labelSpan = document.createElement('span');
      labelSpan.className = 'mode-label';
      labelSpan.textContent = mode.label;
      btn.appendChild(labelSpan);

      // Hotkey indicator
      const hotkeySpan = document.createElement('span');
      hotkeySpan.className = 'mode-hotkey';
      hotkeySpan.textContent = mode.hotkey;
      btn.appendChild(hotkeySpan);

      // Click handler
      btn.addEventListener('click', () => this.setLogMode(mode.id));

      container.appendChild(btn);
    });
  }

  setLogMode(modeId) {
    const mode = this.getLogMode(modeId);
    if (!mode) {
      console.error('Invalid log mode:', modeId);
      return;
    }

    // Update state
    this.state.log.mode = modeId;

    // Update UI
    document.querySelectorAll('.logs-mode-btn').forEach(btn => {
      btn.classList.toggle('active', btn.dataset.mode === modeId);
    });

    // Reload logs if logs tab is active and we have a pod selected
    if (this.state.ui.activeDetailTab === 'logs' && this.state.ui.detailResourceId) {
      const resource = this.getResourceById(this.state.ui.detailResourceId);
      if (resource && resource.type === 'Pod') {
        this.loadLogs();
      }
    }
  }

  loadLogs() {
    const resource = this.getResourceById(this.state.ui.detailResourceId);
    if (!resource || resource.type !== 'Pod') return;

    const container = this.containerDropdown ? this.containerDropdown.getValue() : '';
    if (!container) {
      document.getElementById('logs-content').textContent = 'Please select a container.';
      return;
    }

    if (this.state.log.socket) {
      this.state.log.socket.close();
      this.state.log.socket = null;
    }

    document.getElementById('logs-content').textContent = '';
    document.getElementById('logs-loading').style.display = 'block';
    document.getElementById('logs-error').style.display = 'none';

    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';

    // Get mode options
    const modeOpts = this.getLogMode(this.state.log.mode) || this.getLogMode('tail');

    // Build WebSocket URL with mode parameters
    const params = new URLSearchParams({
      namespace: resource.namespace,
      pod: resource.name,
      container: container,
      follow: modeOpts.follow.toString()
    });

    if (modeOpts.headLines !== null) {
      params.append('headLines', modeOpts.headLines.toString());
    }

    if (modeOpts.tailLines !== null) {
      params.append('tailLines', modeOpts.tailLines.toString());
    }

    if (modeOpts.sinceSeconds !== null) {
      params.append('sinceSeconds', modeOpts.sinceSeconds.toString());
    }

    const wsUrl = `${wsProtocol}//${window.location.host}${API_PATHS.logsWs}?${params.toString()}`;

    this.state.log.socket = new WebSocket(wsUrl);
    this.state.log.currentKey = `${resource.namespace}/${resource.name}/${container}`;

    this.state.log.socket.onopen = () => {
      console.log('[LogStream] Connected:', this.state.log.currentKey);
      document.getElementById('logs-loading').style.display = 'none';
    };

    this.state.log.socket.onmessage = (event) => {
      const message = JSON.parse(event.data);
      if (message.type === 'LOG_LINE') {
        this.appendLogLine(message.line);
      } else if (message.type === 'LOG_END') {
        this.appendLogLine('\n[End of logs]');
        console.log('[LogStream] Ended:', message.reason);
      } else if (message.type === 'LOG_ERROR') {
        this.showLogError(message.error);
        console.error('[LogStream] Error:', message.error);
      }
    };

    this.state.log.socket.onerror = (error) => {
      console.error('[LogStream] WebSocket error:', error);
      this.showLogError('WebSocket connection error');
    };

    this.state.log.socket.onclose = () => {
      console.log('[LogStream] Disconnected');
      if (this.state.log.currentKey) {
        this.appendLogLine('\n[Connection closed]');
      }
    };
  }

  appendLogLine(line) {
    const logsContent = document.getElementById('logs-content');
    logsContent.textContent += line;
    logsContent.scrollTop = logsContent.scrollHeight;
  }

  showLogError(errorMessage) {
    document.getElementById('logs-loading').style.display = 'none';
    document.getElementById('logs-error').style.display = 'block';
    document.getElementById('logs-error-message').textContent = errorMessage;
  }

  async showResource(resourceId) {
    let resource = this.getResourceById(resourceId);

    // If not in cache, fetch from backend
    if (!resource) {
      try {
        resource = await this.fetchResource(resourceId);
      } catch (error) {
        alert(`Unable to load resource: ${error.message}`);
        return;
      }
    }

    if (this.state.log.socket) {
      this.state.log.socket.close();
      this.state.log.socket = null;
    }
    this.state.log.currentKey = null;
    this.state.ui.detailResourceId = resourceId;

    // Reset tabs to Overview when opening a resource
    document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
    const overviewTabButton = document.querySelector('.tab[data-tab="overview"]');
    if (overviewTabButton) {
      overviewTabButton.classList.add('active');
    }
    document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));
    const overviewTabContent = document.getElementById('overview-tab');
    if (overviewTabContent) {
      overviewTabContent.classList.add('active');
    }
    this.state.ui.activeDetailTab = 'overview';
    document.getElementById('container-selector').style.display = 'none';
    this.currentContainerCount = 0;
    this.currentSingleContainerValue = '';

    document.getElementById('logs-content').textContent = '';
    document.getElementById('logs-loading').style.display = 'none';
    document.getElementById('logs-error').style.display = 'none';

    document.getElementById('detail-title').textContent = `${resource.type}: ${resource.name}`;
    const statusInfo = document.getElementById('status-info');
    statusInfo.innerHTML = `
      <div class="info-label">Namespace:</div><div>${resource.namespace || '-'}</div>
      <div class="info-label">Phase:</div><div>${resource.status.phase || '-'}</div>
      <div class="info-label">Ready:</div><div>${resource.status.ready || '-'}</div>
      <div class="info-label">Health:</div><div><span class="resource-status ${resource.health}" style="display:inline-block"></span> ${resource.health}</div>
      ${resource.status.message ? `<div class="info-label">Message:</div><div>${resource.status.message}</div>` : ''}`;

    const relationshipsSection = document.getElementById('relationships-section');
    let relationshipsHTML = '<h3>Relationships</h3>';
    RELATIONSHIP_TYPES.forEach(({ key, label }) => {
      const refs = resource.relationships[key];
      if (refs && refs.length > 0) {
        relationshipsHTML += `<h4>${label}</h4><ul class="relationship-list">${refs.map(ref => `<li><a class="relationship-link" data-relationship-id="${ref.id}">${ref.type}: ${ref.name}</a></li>`).join('')}</ul>`;
      }
    });
    relationshipsSection.innerHTML = relationshipsHTML;
    relationshipsSection.querySelectorAll('.relationship-link').forEach(link => {
      link.addEventListener('click', (e) => {
        e.preventDefault();
        const targetId = link.dataset.relationshipId;
        if (targetId) this.showResource(targetId);
      });
    });

    document.getElementById('yaml-content').textContent = resource.yaml || 'No YAML available';

    const logsTabButton = document.getElementById('logs-tab-button');
    if (resource.type === 'Pod') {
      logsTabButton.style.display = 'flex';
      const containers = resource.spec?.containers || [];
      const containerSelector = document.getElementById('container-selector');
      if (this.containerDropdown) {
        const options = containers.map(c => ({ value: c.name, label: c.name }));
        this.containerDropdown.setOptions(options, '');
      }
      if (containers.length === 1) {
        this.currentContainerCount = 1;
        this.currentSingleContainerValue = containers[0].name;
        if (this.containerDropdown) this.containerDropdown.setValue(containers[0].name);
        containerSelector.style.display = 'none';
      } else if (containers.length > 1) {
        this.currentContainerCount = containers.length;
        this.currentSingleContainerValue = '';
        containerSelector.style.display = 'flex';
        // Auto-select the first container to save user a click
        if (this.containerDropdown) this.containerDropdown.setValue(containers[0].name);
        // Auto-load logs if currently on logs tab
        if (this.state.ui.activeDetailTab === 'logs') {
          this.loadLogs();
        }
      } else {
        logsTabButton.style.display = 'none';
        this.currentContainerCount = 0;
        this.currentSingleContainerValue = '';
      }
    } else {
      logsTabButton.style.display = 'none';
      document.getElementById('container-selector').style.display = 'none';
      this.currentContainerCount = 0;
      this.currentSingleContainerValue = '';
      if (this.state.ui.activeDetailTab === 'logs') {
        const overviewTab = document.querySelector('.tab[data-tab="overview"]');
        if (overviewTab) this.switchTab('overview', overviewTab);
      }
    }

    document.getElementById('detail-panel').classList.add('visible');
    feather.replace();
  }

  closeDetail() {
    if (this.state.log.socket) {
      this.state.log.socket.close();
      this.state.log.socket = null;
    }
    this.state.log.currentKey = null;
    this.state.ui.detailResourceId = null;
    this.state.ui.detailFullscreen = false;
    document.getElementById('detail-panel').classList.remove('visible');
    document.getElementById('detail-panel').classList.remove('fullscreen');
  }

  toggleFullscreen() {
    this.state.ui.detailFullscreen = !this.state.ui.detailFullscreen;
    const panel = document.getElementById('detail-panel');
    const icon = document.getElementById('fullscreen-icon');

    if (this.state.ui.detailFullscreen) {
      panel.classList.add('fullscreen');
      icon.setAttribute('data-feather', 'minimize');
    } else {
      panel.classList.remove('fullscreen');
      icon.setAttribute('data-feather', 'maximize');
    }

    feather.replace();
  }

  switchTab(tabName, target) {
    document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
    target.classList.add('active');
    document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));
    document.getElementById(`${tabName}-tab`).classList.add('active');

    this.state.ui.activeDetailTab = tabName;

    if (tabName === 'logs') {
      const containerSelector = document.getElementById('container-selector');
      const shouldShowSelector = this.currentContainerCount > 1;
      containerSelector.style.display = shouldShowSelector ? 'flex' : 'none';

      if (this.currentContainerCount === 1 && this.currentSingleContainerValue && this.containerDropdown) {
        this.containerDropdown.setValue(this.currentSingleContainerValue);
      }

      if (this.containerDropdown && this.containerDropdown.getValue()) {
        this.loadLogs();
      }
    } else {
      document.getElementById('container-selector').style.display = 'none';
      if (this.state.log.socket) {
        this.state.log.socket.close();
        this.state.log.socket = null;
      }
      this.state.log.currentKey = null;
    }
  }

  // ---------- Main tabs ----------
  switchMainTab(tabName, target) {
    document.querySelectorAll('.nav-tab[data-main-tab]').forEach(t => t.classList.remove('active'));
    target.classList.add('active');
    if (tabName === 'dashboard') {
      document.getElementById('dashboard-view').style.display = 'grid';
      document.getElementById('topology-view').style.display = 'none';
    } else if (tabName === 'topology') {
      document.getElementById('dashboard-view').style.display = 'none';
      document.getElementById('topology-view').style.display = 'grid';
    }
    this.state.ui.activeMainTab = tabName;
  }

  // ---------- UI wiring ----------
  setupNamespaceDropdown() {
    this.namespaceDropdown = document.getElementById('namespace-dropdown');
    if (!this.namespaceDropdown) return;
    this.namespaceDropdown.setAttribute('searchable', 'true');
    this.namespaceDropdown.addEventListener('change', (e) => {
      if (e.detail && e.detail.value !== undefined) {
        this.setNamespace(e.detail.value);
      }
    });
  }

  setupContextDropdown() {
    this.contextDropdown = document.getElementById('context-dropdown');
    if (!this.contextDropdown) return;
    this.contextDropdown.setAttribute('searchable', 'true');
    this.contextDropdown.addEventListener('change', (e) => {
      if (e.detail && e.detail.value !== undefined) {
        this.switchContext(e.detail.value);
      }
    });
  }

  async fetchCurrentContext() {
    try {
      const response = await fetch('/api/context/current');
      const data = await response.json();
      const currentContext = data.context;

      console.log('[App] Current backend context:', currentContext);

      // Set dropdown to match backend state
      if (this.contextDropdown) {
        this.contextDropdown.setValue(currentContext);
      }
    } catch (err) {
      console.error('[App] Failed to fetch current context:', err);
    }
  }

  async fetchAndDisplayContexts() {
    try {
      const response = await fetch('/api/contexts');
      const data = await response.json();
      const contexts = data.contexts || [];

      console.log('[App] Available contexts:', contexts);

      // Update dropdown options
      if (this.contextDropdown) {
        const options = contexts.map(ctx => ({
          value: ctx.name,
          label: ctx.name
        }));

        this.contextDropdown.setOptions(options);
        // REMOVED: Don't set value here - fetchCurrentContext() handles it
      }
    } catch (err) {
      console.error('[App] Failed to fetch contexts:', err);
    }
  }

  async switchContext(newContext) {
    if (!newContext) return;

    console.log(`[App] Switching to context: ${newContext}`);

    try {
      const response = await fetch(`/api/context/switch?context=${encodeURIComponent(newContext)}`, {
        method: 'POST'
      });

      if (!response.ok) {
        throw new Error(`Failed to switch context: ${response.statusText}`);
      }

      const data = await response.json();
      console.log('[App] Context switched successfully:', data);

      // Clear all state and reconnect
      resetForNewConnection(this.state);
      this.renderResourceList();

      // Update context dropdown to show the new context
      if (this.contextDropdown) {
        this.contextDropdown.setValue(newContext);
      }

      // Reset namespace filter to "all" for new context
      this.state.filters.namespace = 'all';
      localStorage.setItem(LOCAL_STORAGE_KEYS.namespace, 'all');
      if (this.namespaceDropdown) {
        this.namespaceDropdown.setValue('all');
      }

      // Reconnect WebSocket (sync status will trigger data refresh when ready)
      this.wsManager.connect();

    } catch (err) {
      console.error('[App] Failed to switch context:', err);
      alert(`Failed to switch context: ${err.message}`);
    }
  }

  attachUIListeners() {
    document.querySelectorAll('.nav-tab[data-main-tab]').forEach(tab => {
      tab.addEventListener('click', () => this.switchMainTab(tab.dataset.mainTab, tab));
    });

    document.getElementById('events-toggle').addEventListener('click', () => this.toggleEventsDrawer());
    document.getElementById('events-close').addEventListener('click', () => this.toggleEventsDrawer());
    document.getElementById('debug-close').addEventListener('click', () => this.toggleDebugDrawer());
    document.getElementById('detail-close').addEventListener('click', () => this.closeDetail());
    document.getElementById('detail-fullscreen').addEventListener('click', () => this.toggleFullscreen());

    document.querySelectorAll('.tab[data-tab]').forEach(tab => {
      tab.addEventListener('click', () => this.switchTab(tab.dataset.tab, tab));
    });

    this.containerDropdown = document.getElementById('container-dropdown');
    if (this.containerDropdown) {
      this.containerDropdown.setAttribute('searchable', 'true');
      this.containerDropdown.addEventListener('change', () => {
        if (this.state.ui.activeDetailTab === 'logs') {
          this.loadLogs();
        }
      });
    }

    // Note: Log mode button click handlers are attached in renderLogModeButtons()
  }
}

document.addEventListener('DOMContentLoaded', () => {
  const app = new App();
  app.init();
});
