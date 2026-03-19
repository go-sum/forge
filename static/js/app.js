// Vanilla JS interaction handlers.
// Uses event delegation (no inline onclick attrs) for CSP 'script-src self' compliance.
// All hooks use data-* attributes; no Alpine.js required.
(function () {

  function sidebarPanelID(id) {
    return id + '-panel';
  }

  function sidebarBackdropID(id) {
    return id + '-backdrop';
  }

  // --- Theme ---
  function cycleTheme() {
    const order = ['light', 'dark', 'system'];
    const html = document.documentElement;
    const cur = html.dataset.themePreference || 'system';
    const next = order[(order.indexOf(cur) + 1) % order.length];
    html.dataset.themePreference = next;
    localStorage.setItem('themePreference', next);
    html.classList.toggle('dark',
      next === 'dark' || (next === 'system' && matchMedia('(prefers-color-scheme: dark)').matches));
  }

  // --- Tabs ---
  function activateTab(container, trigger, focus) {
    const triggers = Array.from(container.querySelectorAll('[role="tab"]'));
    const panels = Array.from(container.querySelectorAll('[role="tabpanel"]'));
    const activeClasses = ['bg-background', 'text-foreground', 'shadow'];

    triggers.forEach(function (candidate) {
      const active = candidate === trigger;
      candidate.setAttribute('aria-selected', active ? 'true' : 'false');
      candidate.setAttribute('tabindex', active ? '0' : '-1');
      activeClasses.forEach(function (className) {
        candidate.classList.toggle(className, active);
      });
    });

    panels.forEach(function (panel) {
      panel.hidden = panel.dataset.tab !== trigger.dataset.tab;
    });

    if (focus) {
      trigger.focus();
    }
  }

  function initTabs(container) {
    if (container.dataset.tabsInitialized === 'true') {
      return;
    }

    const triggers = Array.from(container.querySelectorAll('[role="tab"]'));
    if (!triggers.length) {
      return;
    }

    container.dataset.tabsInitialized = 'true';
    const selected = triggers.find(function (trigger) {
      return trigger.getAttribute('aria-selected') === 'true';
    }) || triggers[0];

    activateTab(container, selected, false);

    triggers.forEach(function (trigger, index) {
      trigger.addEventListener('click', function () {
        activateTab(container, trigger, false);
      });

      trigger.addEventListener('keydown', function (event) {
        let nextIndex = index;
        if (event.key === 'ArrowRight') {
          nextIndex = (index + 1) % triggers.length;
        } else if (event.key === 'ArrowLeft') {
          nextIndex = (index - 1 + triggers.length) % triggers.length;
        } else if (event.key === 'Home') {
          nextIndex = 0;
        } else if (event.key === 'End') {
          nextIndex = triggers.length - 1;
        } else {
          return;
        }

        event.preventDefault();
        activateTab(container, triggers[nextIndex], true);
      });
    });
  }

  document.querySelectorAll('[data-tabs]').forEach(initTabs);
  document.addEventListener('htmx:afterSettle', function (e) {
    if (e.detail.elt && e.detail.elt.querySelectorAll) {
      e.detail.elt.querySelectorAll('[data-tabs]').forEach(initTabs);
    }
  });

  // --- Delegated click handler ---
  document.addEventListener('click', function (e) {
    // Dismiss alert / toast — removes element from DOM.
    var dismiss = e.target.closest('[data-dismiss]');
    if (dismiss) {
      var dismissible = dismiss.closest('[data-dismissible]');
      if (dismissible) dismissible.remove();
      return;
    }

    // Toast trigger — clone <template> content into #toast-container with auto-dismiss.
    var toastBtn = e.target.closest('[data-toast-trigger]');
    if (toastBtn) {
      var tmpl = document.getElementById(toastBtn.dataset.toastTrigger);
      var container = document.getElementById('toast-container');
      if (tmpl && container) {
        var node = tmpl.content.firstElementChild.cloneNode(true);
        container.appendChild(node);
        setTimeout(function () {
          if (!node.parentNode) return;
          node.style.transition = 'opacity 300ms ease-out';
          node.style.opacity = '0';
          setTimeout(function () { node.remove(); }, 300);
        }, 5000);
      }
      return;
    }

    // Theme toggle button.
    if (e.target.closest('[data-theme-toggle]')) {
      cycleTheme();
      return;
    }

    // Open native <dialog>.
    var opener = e.target.closest('[data-dialog-open]');
    if (opener) {
      var dlg = document.getElementById(opener.dataset.dialogOpen);
      if (dlg) dlg.showModal();
      return;
    }

    // Close native <dialog> — walks up to the nearest <dialog> ancestor.
    if (e.target.closest('[data-dialog-close]')) {
      var dlgClose = e.target.closest('dialog');
      if (dlgClose) dlgClose.close();
      return;
    }

    // Sidebar toggle (mobile hamburger).
    var sidebarToggle = e.target.closest('[data-sidebar-toggle]');
    if (sidebarToggle) {
      var sidebarID = sidebarToggle.dataset.sidebarToggle || 'sidebar';
      var sidebar = document.getElementById(sidebarPanelID(sidebarID));
      var backdrop = document.getElementById(sidebarBackdropID(sidebarID));
      if (sidebar) sidebar.classList.toggle('-translate-x-full');
      if (backdrop) backdrop.classList.toggle('hidden');
      return;
    }

    // Sidebar close (backdrop click).
    var sidebarCloseTrigger = e.target.closest('[data-sidebar-close]');
    if (sidebarCloseTrigger) {
      var closeID = sidebarCloseTrigger.dataset.sidebarClose || 'sidebar';
      var sidebarClose = document.getElementById(sidebarPanelID(closeID));
      var backdropClose = document.getElementById(sidebarBackdropID(closeID));
      if (sidebarClose) sidebarClose.classList.add('-translate-x-full');
      if (backdropClose) backdropClose.classList.add('hidden');
      return;
    }

    // Click outside an open popover/dropdown (<details data-popover>) to close it.
    document.querySelectorAll('details[data-popover][open]').forEach(function (d) {
      if (!d.contains(e.target)) d.open = false;
    });
  });

  // --- File upload ---
  function updateFileName(zone, input) {
    var display = zone.querySelector('[data-file-name]');
    if (!display) return;
    var files = input.files;
    if (files.length === 0) {
      display.textContent = '';
    } else if (files.length === 1) {
      display.textContent = files[0].name;
    } else {
      display.textContent = files.length + ' files selected';
    }
  }

  document.addEventListener('dragover', function (e) {
    var zone = e.target.closest('[data-file-upload]');
    if (!zone) return;
    e.preventDefault();
    zone.dataset.dragging = '';
  });

  document.addEventListener('dragleave', function (e) {
    var zone = e.target.closest('[data-file-upload]');
    if (!zone || zone.contains(e.relatedTarget)) return;
    delete zone.dataset.dragging;
  });

  document.addEventListener('drop', function (e) {
    var zone = e.target.closest('[data-file-upload]');
    if (!zone) return;
    e.preventDefault();
    delete zone.dataset.dragging;
    var input = zone.querySelector('input[type="file"]');
    if (!input || input.disabled) return;
    var dt = new DataTransfer();
    var incoming = e.dataTransfer.files;
    var limit = input.multiple ? incoming.length : 1;
    for (var i = 0; i < limit; i++) dt.items.add(incoming[i]);
    input.files = dt.files;
    updateFileName(zone, input);
  });

  document.addEventListener('change', function (e) {
    var input = e.target.closest('input[type="file"]');
    if (!input) return;
    var zone = input.closest('[data-file-upload]');
    if (!zone) return;
    updateFileName(zone, input);
  });

})();
