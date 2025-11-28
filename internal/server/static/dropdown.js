class K8VDropdown extends HTMLElement {
  constructor() {
    super();
    this.options = [];
    this.filtered = [];
    this.value = '';
    this.highlightedIndex = -1;
    this.searchable = this.getAttribute('searchable') !== 'false';
  }

  connectedCallback() {
    this.classList.add('namespace-dropdown');
    this.render();
    this.attachEvents();
  }

  render() {
    this.innerHTML = `
      <div class="dropdown-selected">${this.getPlaceholder()}</div>
      <div class="dropdown-menu" style="display: none;">
        ${this.searchable ? '<input type="text" class="dropdown-search" placeholder="Search..." />' : ''}
        <div class="dropdown-options"></div>
      </div>
    `;
    this.selectedEl = this.querySelector('.dropdown-selected');
    this.menuEl = this.querySelector('.dropdown-menu');
    this.searchInput = this.querySelector('.dropdown-search');
    this.optionsContainer = this.querySelector('.dropdown-options');
    this.renderOptions();
  }

  getPlaceholder() {
    return this.getAttribute('placeholder') || 'Select...';
  }

  setOptions(options, selectedValue) {
    this.options = options || [];
    if (selectedValue !== undefined) {
      this.value = selectedValue;
    }
    this.filtered = this.options.slice();
    this.renderOptions();
    this.updateSelectedDisplay();
  }

  setValue(value) {
    this.value = value;
    this.updateSelectedDisplay();
    this.renderOptions();
  }

  getValue() {
    return this.value;
  }

  updateSelectedDisplay() {
    if (!this.selectedEl) return;
    const found = this.options.find(o => o.value === this.value);
    this.selectedEl.textContent = found ? found.label : this.getPlaceholder();
  }

  renderOptions() {
    if (!this.optionsContainer) return;
    const filterText = this.searchInput ? this.searchInput.value.toLowerCase() : '';
    this.filtered = this.options.filter(o => o.label.toLowerCase().includes(filterText));
    this.optionsContainer.innerHTML = '';
    this.filtered.forEach((opt, index) => {
      const el = document.createElement('div');
      let className = 'dropdown-option';
      if (opt.value === this.value) className += ' active';
      if (index === this.highlightedIndex) className += ' highlighted';
      el.className = className;
      el.textContent = opt.label;
      el.dataset.index = index;
      el.addEventListener('click', () => {
        this.selectValue(opt.value);
        this.close();
      });
      this.optionsContainer.appendChild(el);
    });
  }

  selectValue(value) {
    this.value = value;
    this.updateSelectedDisplay();
    this.renderOptions();
    this.dispatchEvent(new CustomEvent('change', { detail: { value } }));
  }

  open() {
    this.menuEl.style.display = 'flex';
    this.selectedEl.classList.add('open');
    this.highlightedIndex = -1;
    if (this.searchInput) {
      this.searchInput.value = '';
      this.searchInput.focus();
    }
    this.renderOptions();
  }

  close() {
    this.menuEl.style.display = 'none';
    this.selectedEl.classList.remove('open');
    this.highlightedIndex = -1;
  }

  isOpen() {
    return this.menuEl && this.menuEl.style.display !== 'none';
  }

  toggle() {
    if (this.menuEl.style.display === 'none') {
      this.open();
    } else {
      this.close();
    }
  }

  handleKeyboard(e) {
    if (this.menuEl.style.display === 'none') return;
    if (!this.filtered.length) return;

    if (e.key === 'ArrowDown') {
      e.preventDefault();
      this.highlightedIndex = Math.min(this.highlightedIndex + 1, this.filtered.length - 1);
      this.renderOptions();
      this.scrollToHighlighted();
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      this.highlightedIndex = Math.max(this.highlightedIndex - 1, 0);
      this.renderOptions();
      this.scrollToHighlighted();
    } else if (e.key === 'Enter') {
      e.preventDefault();
      if (this.highlightedIndex >= 0 && this.highlightedIndex < this.filtered.length) {
        this.selectValue(this.filtered[this.highlightedIndex].value);
        this.close();
      }
    } else if (e.key === 'Escape') {
      e.preventDefault();
      this.close();
    }
  }

  scrollToHighlighted() {
    const highlighted = this.optionsContainer.querySelector('.dropdown-option.highlighted');
    if (highlighted) {
      highlighted.scrollIntoView({ block: 'nearest', behavior: 'smooth' });
    }
  }

  attachEvents() {
    this.selectedEl.addEventListener('click', () => this.toggle());
    if (this.searchInput) {
      this.searchInput.addEventListener('input', () => {
        this.highlightedIndex = -1;
        this.renderOptions();
      });
      this.searchInput.addEventListener('keydown', (e) => this.handleKeyboard(e));
    }
    document.addEventListener('click', (e) => {
      if (!this.contains(e.target)) {
        this.close();
      }
    });
  }
}

customElements.define('k8v-dropdown', K8VDropdown);
