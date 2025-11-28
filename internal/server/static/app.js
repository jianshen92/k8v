import { API_PATHS, EVENTS_LIMIT, LOCAL_STORAGE_KEYS, RELATIONSHIP_TYPES, RESOURCE_TYPES } from './config.js';
import { createInitialState, resetForNewConnection } from './state.js';
import { createResourceSocket } from './ws.js';
import './dropdown.js';

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
    this.setupContextDropdown();
    this.setupNamespaceDropdown();
    this.setupSearchFilter();
    this.fetchAndDisplayContexts();
    this.fetchNamespaces();

    await this.fetchAndDisplayStats();
    this.wsManager.connect();
    feather.replace();
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

  handleGlobalKeydown(event) {
    const activeElement = document.activeElement;
    const isInputFocused = activeElement && (
      activeElement.tagName === 'INPUT' ||
      activeElement.tagName === 'TEXTAREA' ||
      activeElement.isContentEditable
    );

    if (event.key === '/' && !isInputFocused) {
      event.preventDefault();
      this.activateSearch();
      return;
    }

    if (event.key === 'Escape') {
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
      document.getElementById('resource-count').textContent = `${counts.total || 0} resources`;

      console.log('[Stats] Loaded counts:', counts);
    } catch (error) {
      console.error('[Stats] Failed to fetch stats:', error);
    }
  }

  // ---------- Resources rendering ----------
  typeToClass(t) { return t.toLowerCase(); }

  createResourceElement(r) {
    const item = document.createElement('div');
    item.className = `resource-item ${this.typeToClass(r.type)}`;
    item.dataset.resourceId = r.id;
    item.addEventListener('click', () => this.showResource(r.id));

    const icon = document.createElement('div');
    icon.className = `resource-icon ${this.typeToClass(r.type)}`;
    icon.textContent = r.type[0];

    const header = document.createElement('div');
    header.className = 'resource-header';

    const info = document.createElement('div');
    info.className = 'resource-info';

    const name = document.createElement('div');
    name.className = 'resource-name';
    name.textContent = r.name;

    const sub = document.createElement('div');
    sub.className = 'resource-subtitle';
    sub.textContent = `${r.namespace || '-'} • ${r.type}`;

    info.appendChild(name);
    info.appendChild(sub);

    const statusDot = document.createElement('div');
    statusDot.className = `resource-status ${r.health || 'unknown'}`;

    header.appendChild(info);
    header.appendChild(statusDot);

    item.appendChild(icon);
    item.appendChild(header);

    return item;
  }

  renderResourceList() {
    const container = document.getElementById('resource-list');
    container.innerHTML = '';

    const list = Array.from(this.state.resources.values())
      .filter(r => {
        if (r.type !== this.state.filters.type) return false;
        if (this.state.filters.search && !r.name.toLowerCase().includes(this.state.filters.search)) return false;
        return true;
      })
      .sort((a, b) => a.name.localeCompare(b.name));

    if (!list.length) {
      const empty = document.createElement('div');
      empty.className = 'empty-state';
      const emptyMessage = this.state.filters.search
        ? `No ${this.state.filters.type}s matching "${this.state.filters.search}"`
        : `No ${this.state.filters.type}s found`;
      const emptyDetail = this.state.filters.namespace !== 'all'
        ? `in namespace "${this.state.filters.namespace}"`
        : 'in cluster';

      empty.innerHTML = `
        <div><i data-feather="inbox" style="width: 48px; height: 48px; stroke-width: 1.5;"></i></div>
        <div style="font-size: 16px; margin: 12px 0 8px;">${emptyMessage}</div>
        <div style="font-size: 13px;">${emptyDetail}</div>
      `;
      container.appendChild(empty);
      feather.replace();
      return;
    }

    for (const r of list) {
      container.appendChild(this.createResourceElement(r));
    }
  }

  updateResourceInList(resourceId, eventType) {
    const container = document.getElementById('resource-list');
    const existingElement = container.querySelector(`[data-resource-id="${resourceId}"]`);

    if (eventType === 'DELETED') {
      if (existingElement) existingElement.remove();
      return;
    }

    const resource = this.state.resources.get(resourceId);
    if (!resource) return;

    const matchesTypeFilter = resource.type === this.state.filters.type;
    const matchesSearchFilter = !this.state.filters.search || resource.name.toLowerCase().includes(this.state.filters.search);
    const matchesFilter = matchesTypeFilter && matchesSearchFilter;

    if (!matchesFilter) {
      if (existingElement) existingElement.remove();
      return;
    }

    if (eventType === 'MODIFIED' && existingElement) {
      const newElement = this.createResourceElement(resource);
      existingElement.replaceWith(newElement);
    } else if (eventType === 'ADDED') {
      const allVisible = Array.from(this.state.resources.values())
        .filter(r => r.type === this.state.filters.type)
        .filter(r => !this.state.filters.search || r.name.toLowerCase().includes(this.state.filters.search))
        .sort((a, b) => a.name.localeCompare(b.name));

      const index = allVisible.findIndex(r => r.id === resourceId);
      const newElement = this.createResourceElement(resource);

      if (index === allVisible.length - 1 || container.children.length === 0) {
        container.appendChild(newElement);
      } else {
        const nextResource = allVisible[index + 1];
        if (nextResource) {
          const nextElement = container.querySelector(`[data-resource-id="${nextResource.id}"]`);
          if (nextElement) {
            container.insertBefore(newElement, nextElement);
          } else {
            container.appendChild(newElement);
          }
        } else {
          container.appendChild(newElement);
        }
      }
    }
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
    const wsUrl = `${wsProtocol}//${window.location.host}${API_PATHS.logsWs}?namespace=${resource.namespace}&pod=${resource.name}&container=${container}`;

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

  showResource(resourceId) {
    const resource = this.getResourceById(resourceId);
    if (!resource) return;

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
        document.getElementById('logs-content').textContent = 'Please select a container to view logs.';
        document.getElementById('logs-loading').style.display = 'none';
        document.getElementById('logs-error').style.display = 'none';
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

  async fetchAndDisplayContexts() {
    try {
      const response = await fetch('/api/contexts');
      const data = await response.json();
      const contexts = data.contexts || [];

      // Find current context
      const currentCtx = contexts.find(c => c.current);
      const currentContext = currentCtx ? currentCtx.name : '';

      console.log('[App] Available contexts:', contexts, 'Current:', currentContext);

      // Update dropdown
      if (this.contextDropdown) {
        const options = contexts.map(ctx => ({
          value: ctx.name,
          label: ctx.name + (ctx.current ? ' (current)' : '')
        }));

        this.contextDropdown.setOptions(options);
        this.contextDropdown.setValue(currentContext);
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

      // Wait a bit for server to complete the switch
      await new Promise(resolve => setTimeout(resolve, 1000));

      // Refresh namespaces and stats for new context
      this.fetchNamespaces();
      await this.fetchAndDisplayStats();

      // Reconnect WebSocket
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
  }
}

document.addEventListener('DOMContentLoaded', () => {
  const app = new App();
  app.init();
});
