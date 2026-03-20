(() => {
  // static/js/lib/dom.js
  var delegates = /* @__PURE__ */ new Map();
  var upgradeHandlers = [];
  var delegatesBound = false;
  function toElement(node) {
    if (!node) {
      return null;
    }
    if (node.nodeType === Node.ELEMENT_NODE) {
      return node;
    }
    return node.parentElement || null;
  }
  function find(root, selector) {
    var scope = root || document;
    if (!scope || !selector) {
      return null;
    }
    if (scope.matches && scope.matches(selector)) {
      return scope;
    }
    if (!scope.querySelector) {
      return null;
    }
    return scope.querySelector(selector);
  }
  function findAll(root, selector) {
    var scope = root || document;
    if (!scope || !selector || !scope.querySelectorAll) {
      return [];
    }
    var matches = Array.from(scope.querySelectorAll(selector));
    if (scope.matches && scope.matches(selector)) {
      matches.unshift(scope);
    }
    return matches;
  }
  function closest(target, selector) {
    var element = toElement(target);
    if (!element || !selector || !element.closest) {
      return null;
    }
    return element.closest(selector);
  }
  function remove(node) {
    if (node && typeof node.remove === "function") {
      node.remove();
    }
  }
  function toggleClass(node, name, on) {
    if (node && node.classList) {
      node.classList.toggle(name, on);
    }
  }
  function data(node, key) {
    if (!node || !node.dataset) {
      return "";
    }
    return node.dataset[key] || "";
  }
  function delegate(eventName, selector, handler) {
    var handlers = delegates.get(eventName) || [];
    handlers.push({ selector, handler });
    delegates.set(eventName, handlers);
  }
  function registerUpgrade(handler) {
    upgradeHandlers.push(handler);
  }
  function bindDelegates() {
    if (delegatesBound) {
      return;
    }
    delegatesBound = true;
    delegates.forEach(function(handlers, eventName) {
      document.addEventListener(eventName, function(event) {
        handlers.forEach(function(entry) {
          if (!entry.selector) {
            entry.handler(event, null);
            return;
          }
          var match = closest(event.target, entry.selector);
          if (match) {
            entry.handler(event, match);
          }
        });
      });
    });
  }
  function upgrade(root) {
    bindDelegates();
    var scope = root || document;
    upgradeHandlers.forEach(function(handler) {
      handler(scope);
    });
  }

  // static/js/components/dialog.js
  delegate("click", "[data-dialog-open]", function(event, opener) {
    var dialog = document.getElementById(opener.dataset.dialogOpen);
    if (dialog) {
      dialog.showModal();
    }
  });
  delegate("click", "[data-dialog-close]", function(event, closer) {
    var dialog = closer.closest("dialog");
    if (dialog) {
      dialog.close();
    }
  });

  // static/js/components/feedback.js
  function queueToastRemoval(node) {
    setTimeout(function() {
      if (!node.parentNode) {
        return;
      }
      node.style.transition = "opacity 300ms ease-out";
      node.style.opacity = "0";
      setTimeout(function() {
        remove(node);
      }, 300);
    }, 5e3);
  }
  delegate("click", "[data-dismiss]", function(event, dismiss) {
    var dismissible = dismiss.closest("[data-dismissible]");
    if (dismissible) {
      remove(dismissible);
    }
  });
  delegate("click", "[data-toast-trigger]", function(event, trigger) {
    var tmpl = document.getElementById(trigger.dataset.toastTrigger);
    var container = document.getElementById("toast-container");
    var content = tmpl && tmpl.content ? tmpl.content.firstElementChild : null;
    if (!container || !content) {
      return;
    }
    var node = content.cloneNode(true);
    container.appendChild(node);
    queueToastRemoval(node);
  });

  // static/js/components/fileupload.js
  function updateFileName(zone, input) {
    var display = find(zone, "[data-file-name]");
    if (!display) {
      return;
    }
    var files = input.files;
    if (files.length === 0) {
      display.textContent = "";
    } else if (files.length === 1) {
      display.textContent = files[0].name;
    } else {
      display.textContent = files.length + " files selected";
    }
  }
  function clearDragging(zone) {
    delete zone.dataset.dragging;
  }
  delegate("dragover", "[data-file-upload]", function(event, zone) {
    event.preventDefault();
    zone.dataset.dragging = "";
  });
  delegate("dragleave", "[data-file-upload]", function(event, zone) {
    if (zone.contains(event.relatedTarget)) {
      return;
    }
    clearDragging(zone);
  });
  delegate("drop", "[data-file-upload]", function(event, zone) {
    event.preventDefault();
    clearDragging(zone);
    var input = find(zone, 'input[type="file"]');
    if (!input || input.disabled) {
      return;
    }
    var incoming = event.dataTransfer.files;
    var dt = new DataTransfer();
    var limit = input.multiple ? incoming.length : 1;
    for (var i = 0; i < limit; i++) {
      dt.items.add(incoming[i]);
    }
    input.files = dt.files;
    updateFileName(zone, input);
  });
  delegate("change", 'input[type="file"]', function(event, input) {
    var zone = closest(input, "[data-file-upload]");
    if (!zone) {
      return;
    }
    updateFileName(zone, input);
  });

  // static/js/components/popover.js
  delegate("click", null, function(event) {
    document.querySelectorAll("details[data-popover][open]").forEach(function(details) {
      if (!details.contains(event.target)) {
        details.open = false;
      }
    });
  });

  // static/js/components/tabs.js
  var activeClasses = ["bg-background", "text-foreground", "shadow"];
  function activateTab(container, trigger, focus) {
    var triggers = findAll(container, '[role="tab"]');
    var panels = findAll(container, '[role="tabpanel"]');
    var activeTab = data(trigger, "tab");
    triggers.forEach(function(candidate) {
      var active = candidate === trigger;
      candidate.setAttribute("aria-selected", active ? "true" : "false");
      candidate.setAttribute("tabindex", active ? "0" : "-1");
      activeClasses.forEach(function(className) {
        toggleClass(candidate, className, active);
      });
    });
    panels.forEach(function(panel) {
      panel.hidden = data(panel, "tab") !== activeTab;
    });
    if (focus) {
      trigger.focus();
    }
  }
  function mount(container) {
    if (container.dataset.tabsInitialized === "true") {
      return;
    }
    var triggers = findAll(container, '[role="tab"]');
    if (!triggers.length) {
      return;
    }
    container.dataset.tabsInitialized = "true";
    var selected = triggers.find(function(trigger) {
      return trigger.getAttribute("aria-selected") === "true";
    }) || triggers[0];
    activateTab(container, selected, false);
    triggers.forEach(function(trigger, index) {
      trigger.addEventListener("click", function() {
        activateTab(container, trigger, false);
      });
      trigger.addEventListener("keydown", function(event) {
        var nextIndex = index;
        if (event.key === "ArrowRight") {
          nextIndex = (index + 1) % triggers.length;
        } else if (event.key === "ArrowLeft") {
          nextIndex = (index - 1 + triggers.length) % triggers.length;
        } else if (event.key === "Home") {
          nextIndex = 0;
        } else if (event.key === "End") {
          nextIndex = triggers.length - 1;
        } else {
          return;
        }
        event.preventDefault();
        activateTab(container, triggers[nextIndex], true);
      });
    });
  }
  registerUpgrade(function(root) {
    findAll(root, "[data-tabs]").forEach(mount);
  });

  // static/js/components/theme.js
  function cycleTheme() {
    var order = ["light", "dark", "system"];
    var html = document.documentElement;
    var current = html.dataset.themePreference || "system";
    var next = order[(order.indexOf(current) + 1) % order.length];
    html.dataset.themePreference = next;
    localStorage.setItem("themePreference", next);
    html.classList.toggle(
      "dark",
      next === "dark" || next === "system" && matchMedia("(prefers-color-scheme: dark)").matches
    );
  }
  delegate("click", "[data-theme-toggle]", function() {
    cycleTheme();
  });

  // static/js/components.js
  function init() {
    if (document.documentElement.dataset.componentsInitialized === "true") {
      return;
    }
    document.documentElement.dataset.componentsInitialized = "true";
    upgrade(document);
    document.addEventListener("htmx:afterSettle", function(event) {
      if (event.detail && event.detail.elt) {
        upgrade(event.detail.elt);
      }
    });
  }

  // static/js/app.js
  init();
})();
