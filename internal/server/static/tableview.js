import { getColumnsForType } from './config.js';

// ========== Data Extraction Helper Functions ==========
// These functions extract cell values from resource objects for table columns

function extractCellValue(resource, columnId) {
  switch (columnId) {
    case 'name':
      return resource.name;
    case 'namespace':
      return resource.namespace || '-';
    case 'type':
      return resource.type || '-';
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

// ========== Web Component ==========

class ResourceTable extends HTMLElement {
  constructor() {
    super();
    this.resources = new Map();
    this.filters = { type: 'Pod', search: '' };
    this.selectedRowIndex = -1;
  }

  connectedCallback() {
    this.classList.add('resource-table-component');
    this.render();
  }

  // ========== Public API ==========

  setData(resources, filters, selectedRowIndex = -1) {
    this.resources = resources;
    this.filters = filters;
    this.selectedRowIndex = selectedRowIndex;
    this.renderList();
  }

  selectRow(rowIndex) {
    const previousRow = this.querySelector('tbody tr.selected');
    if (previousRow) {
      previousRow.classList.remove('selected');
    }

    this.selectedRowIndex = rowIndex;
    const newRow = this.querySelector(`tbody tr[data-row-index="${rowIndex}"]`);
    if (newRow) {
      newRow.classList.add('selected');
      newRow.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
    }

    this.dispatchEvent(new CustomEvent('selectionChange', {
      detail: { rowIndex }
    }));
  }

  navigateUp() {
    const rows = this.querySelectorAll('tbody tr:not(.empty-state-row)');
    if (rows.length === 0) return;

    let newIndex = this.selectedRowIndex;
    if (newIndex <= 0) {
      newIndex = 0; // Stay at first row
    } else {
      newIndex--;
    }

    this.selectRow(newIndex);
  }

  navigateDown() {
    const rows = this.querySelectorAll('tbody tr:not(.empty-state-row)');
    if (rows.length === 0) return;

    let newIndex = this.selectedRowIndex;
    if (newIndex < 0) {
      newIndex = 0; // Start at first row if nothing selected
    } else if (newIndex >= rows.length - 1) {
      newIndex = rows.length - 1; // Stay at last row
    } else {
      newIndex++;
    }

    this.selectRow(newIndex);
  }

  getSelectedResourceId() {
    if (this.selectedRowIndex < 0) return null;
    const row = this.querySelector(`tbody tr[data-row-index="${this.selectedRowIndex}"]`);
    return row ? row.dataset.resourceId : null;
  }

  updateResource(resourceId, eventType) {
    const tbody = this.querySelector('tbody');
    const existingRow = tbody.querySelector(`[data-resource-id="${resourceId}"]`);

    if (eventType === 'DELETED') {
      if (existingRow) {
        const deletedIndex = parseInt(existingRow.dataset.rowIndex, 10);
        existingRow.remove();
        this.reindexRows();
        this.adjustSelectionAfterDelete(deletedIndex);
      }
      return;
    }

    const resource = this.resources.get(resourceId);
    if (!resource) return;

    const matchesTypeFilter = resource.type === this.filters.type;
    const matchesSearchFilter = !this.filters.search || resource.name.toLowerCase().includes(this.filters.search);
    const matchesFilter = matchesTypeFilter && matchesSearchFilter;

    if (!matchesFilter) {
      if (existingRow) {
        const deletedIndex = parseInt(existingRow.dataset.rowIndex, 10);
        existingRow.remove();
        this.reindexRows();
        this.adjustSelectionAfterDelete(deletedIndex);
      }
      return;
    }

    if (eventType === 'MODIFIED' && existingRow) {
      const rowIndex = parseInt(existingRow.dataset.rowIndex, 10);
      const newRow = this.createRow(resource, rowIndex);
      if (existingRow.classList.contains('selected')) {
        newRow.classList.add('selected');
      }
      existingRow.replaceWith(newRow);
    } else if (eventType === 'ADDED') {
      const allVisible = this.getFilteredList();
      const index = allVisible.findIndex(r => r.id === resourceId);
      const newRow = this.createRow(resource, index);

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

      this.reindexRows();
      this.adjustSelectionAfterAdd(index);
    }
  }

  // ========== Internal Methods ==========

  render() {
    this.innerHTML = `
      <table class="resource-table">
        <thead>
          <tr id="table-header-row"></tr>
        </thead>
        <tbody id="resource-table-body"></tbody>
      </table>
    `;
  }

  renderList() {
    const tbody = this.querySelector('tbody');
    tbody.innerHTML = '';

    const list = this.getFilteredList();

    // Render table header
    this.renderHeader();

    // Handle empty state
    if (!list.length) {
      tbody.appendChild(this.renderEmptyState());
      if (window.feather) {
        window.feather.replace();
      }
      return;
    }

    // Render rows
    list.forEach((resource, index) => {
      tbody.appendChild(this.createRow(resource, index));
    });

    // Reset selection if out of bounds
    if (this.selectedRowIndex >= list.length) {
      this.selectedRowIndex = -1;
    }
  }

  renderHeader() {
    const headerRow = this.querySelector('#table-header-row');
    headerRow.innerHTML = '';

    const columns = getColumnsForType(this.filters.type);

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

  createRow(resource, rowIndex) {
    const row = document.createElement('tr');
    row.className = resource.type.toLowerCase();
    row.dataset.resourceId = resource.id;
    row.dataset.rowIndex = rowIndex;

    // Mark as selected if this is the selected row
    if (rowIndex === this.selectedRowIndex) {
      row.classList.add('selected');
    }

    // Click handler
    row.addEventListener('click', () => {
      this.selectRow(rowIndex);
      this.dispatchEvent(new CustomEvent('rowClick', {
        detail: { resourceId: resource.id }
      }));
    });

    const columns = getColumnsForType(this.filters.type);

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

  renderEmptyState() {
    const row = document.createElement('tr');
    row.className = 'empty-state-row';
    const td = document.createElement('td');
    td.colSpan = getColumnsForType(this.filters.type).length;

    const resourceLabel = (() => {
      if (!this.filters.type) return 'resources';
      return this.filters.type.endsWith('s') ? this.filters.type : `${this.filters.type}s`;
    })();
    const emptyMessage = this.filters.search
      ? `No ${resourceLabel} matching "${this.filters.search}"`
      : `No ${resourceLabel} found`;
    const emptyDetail = this.filters.namespace !== 'all'
      ? `in namespace "${this.filters.namespace}"`
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
    return row;
  }

  reindexRows() {
    const tbody = this.querySelector('tbody');
    const rows = tbody.querySelectorAll('tr:not(.empty-state-row)');
    rows.forEach((row, index) => {
      row.dataset.rowIndex = index;
    });
  }

  getFilteredList() {
    return Array.from(this.resources.values())
      .filter(r => {
        const matchesType = r.type === this.filters.type;
        if (!matchesType) return false;
        if (this.filters.search && !r.name.toLowerCase().includes(this.filters.search)) {
          return false;
        }
        return true;
      })
      .sort((a, b) => a.name.localeCompare(b.name));
  }

  adjustSelectionAfterDelete(deletedIndex) {
    if (this.selectedRowIndex === deletedIndex) {
      this.selectedRowIndex = -1;
    } else if (this.selectedRowIndex > deletedIndex) {
      this.selectedRowIndex--;
    }
  }

  adjustSelectionAfterAdd(insertIndex) {
    if (this.selectedRowIndex >= insertIndex) {
      this.selectedRowIndex++;
    }
  }
}

customElements.define('resource-table', ResourceTable);
