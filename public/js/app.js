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
  delegate("click", "details[data-popover] a, details[data-popover] button:not(summary)", function(event, item) {
    var details = item.closest("details[data-popover]");
    if (!details) return;
    details.open = false;
    var summary = details.querySelector("summary");
    if (summary) summary.focus();
  });
  delegate("keydown", null, function(event) {
    if (event.key !== "Escape") return;
    var open = document.querySelector("details[data-popover][open]");
    if (!open) return;
    event.preventDefault();
    open.open = false;
    var summary = open.querySelector("summary");
    if (summary) summary.focus();
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

  // static/js/lib/fetch.js
  function csrfToken() {
    var meta = document.querySelector('meta[name="csrf-token"]');
    return meta ? meta.getAttribute("content") : "";
  }
  function postJSON(url, body) {
    return fetch(url, {
      method: "POST",
      headers: { "Content-Type": "application/json", "X-CSRF-Token": csrfToken() },
      body: JSON.stringify(body || {}),
      credentials: "same-origin"
    }).then(handleResponse);
  }
  function handleResponse(res) {
    if (!res.ok) {
      return res.json().catch(function() {
        return {};
      }).then(function(data2) {
        var err = new Error(data2.message || res.statusText);
        err.status = res.status;
        throw err;
      });
    }
    var ct = res.headers.get("content-type") || "";
    if (ct.indexOf("application/json") !== -1) return res.json();
    return res.text();
  }

  // static/js/lib/encoding.js
  function bufferToBase64url(buffer) {
    var bytes = new Uint8Array(buffer);
    var binary = "";
    for (var i = 0; i < bytes.length; i++) binary += String.fromCharCode(bytes[i]);
    return btoa(binary).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");
  }
  function base64urlToBuffer(str) {
    var base64 = str.replace(/-/g, "+").replace(/_/g, "/");
    while (base64.length % 4) base64 += "=";
    var binary = atob(base64);
    var bytes = new Uint8Array(binary.length);
    for (var i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i);
    return bytes.buffer;
  }

  // static/js/components/passkeys.js
  registerUpgrade(function(scope) {
    if (!window.PublicKeyCredential) {
      return;
    }
    var enabledEl = (scope.matches && scope.matches("[data-passkey-enabled]") ? scope : null) || (scope.querySelector ? scope.querySelector("[data-passkey-enabled]") : null) || (scope.closest ? scope.closest("[data-passkey-enabled]") : null);
    if (!enabledEl) {
      return;
    }
    findAll(scope, "[data-passkey-visible]").forEach(function(el) {
      el.classList.remove("hidden");
    });
  });
  delegate("click", "[data-passkey-authenticate]", function(event, el) {
    event.preventDefault();
    authenticate(el);
  });
  delegate("click", "[data-passkey-register]", function(event, el) {
    event.preventDefault();
    register(el);
  });
  function showPasskeyError(message) {
    var errEl = find(document, "[data-passkey-error]");
    if (!errEl) {
      return;
    }
    if (!message) {
      errEl.classList.add("hidden");
      errEl.textContent = "";
      return;
    }
    errEl.textContent = message;
    errEl.classList.remove("hidden");
  }
  function mapWebAuthnError(err) {
    if (err.name === "NotAllowedError") return "Cancelled by user or device.";
    if (err.name === "InvalidStateError") return "This passkey is already registered.";
    if (err.name === "AbortError") return null;
    return "Something went wrong. Try signing in with email.";
  }
  function authenticate(el) {
    var beginUrl = el.dataset.beginUrl;
    var finishUrl = el.dataset.finishUrl;
    postJSON(beginUrl, {}).then(function(options) {
      var pk = options.publicKey;
      pk.challenge = base64urlToBuffer(pk.challenge);
      if (pk.allowCredentials) {
        pk.allowCredentials = pk.allowCredentials.map(function(c) {
          return Object.assign({}, c, { id: base64urlToBuffer(c.id) });
        });
      }
      return navigator.credentials.get({ publicKey: pk });
    }).then(function(credential) {
      var resp = credential.response;
      var body = {
        id: credential.id,
        rawId: bufferToBase64url(credential.rawId),
        type: credential.type,
        response: {
          authenticatorData: bufferToBase64url(resp.authenticatorData),
          clientDataJSON: bufferToBase64url(resp.clientDataJSON),
          signature: bufferToBase64url(resp.signature),
          userHandle: resp.userHandle ? bufferToBase64url(resp.userHandle) : null
        }
      };
      return postJSON(finishUrl, body);
    }).then(function(result) {
      window.location.href = result.redirect;
    }).catch(function(err) {
      var msg = mapWebAuthnError(err);
      if (msg) {
        showPasskeyError(msg);
      }
    });
  }
  function register(el) {
    var beginUrl = el.dataset.beginUrl;
    var finishUrl = el.dataset.finishUrl;
    var listUrl = el.dataset.listUrl;
    postJSON(beginUrl, {}).then(function(options) {
      var pk = options.publicKey;
      pk.challenge = base64urlToBuffer(pk.challenge);
      pk.user.id = base64urlToBuffer(pk.user.id);
      if (pk.excludeCredentials) {
        pk.excludeCredentials = pk.excludeCredentials.map(function(c) {
          return Object.assign({}, c, { id: base64urlToBuffer(c.id) });
        });
      }
      return navigator.credentials.create({ publicKey: pk });
    }).then(function(credential) {
      var resp = credential.response;
      var name = "Passkey " + (/* @__PURE__ */ new Date()).toISOString().slice(0, 10);
      var body = {
        id: credential.id,
        rawId: bufferToBase64url(credential.rawId),
        type: credential.type,
        name,
        response: {
          attestationObject: bufferToBase64url(resp.attestationObject),
          clientDataJSON: bufferToBase64url(resp.clientDataJSON),
          transports: resp.getTransports ? resp.getTransports() : []
        }
      };
      return postJSON(finishUrl, body);
    }).then(function() {
      if (window.htmx) {
        htmx.ajax("GET", listUrl, { target: "#passkeys-list-region", swap: "innerHTML" });
      } else {
        location.reload();
      }
    }).catch(function(err) {
      console.error("[passkey] registration error:", err.name, "|", err.message);
      var msg = mapWebAuthnError(err);
      if (msg) {
        showPasskeyError(msg);
      }
    });
  }

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
