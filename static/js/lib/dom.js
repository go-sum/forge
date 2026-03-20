var delegates = new Map();
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

export function find(root, selector) {
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

export function findAll(root, selector) {
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

export function closest(target, selector) {
  var element = toElement(target);
  if (!element || !selector || !element.closest) {
    return null;
  }
  return element.closest(selector);
}

export function remove(node) {
  if (node && typeof node.remove === 'function') {
    node.remove();
  }
}

export function toggleClass(node, name, on) {
  if (node && node.classList) {
    node.classList.toggle(name, on);
  }
}

export function data(node, key) {
  if (!node || !node.dataset) {
    return '';
  }
  return node.dataset[key] || '';
}

export function delegate(eventName, selector, handler) {
  var handlers = delegates.get(eventName) || [];
  handlers.push({ selector: selector, handler: handler });
  delegates.set(eventName, handlers);
}

export function registerUpgrade(handler) {
  upgradeHandlers.push(handler);
}

function bindDelegates() {
  if (delegatesBound) {
    return;
  }
  delegatesBound = true;

  delegates.forEach(function (handlers, eventName) {
    document.addEventListener(eventName, function (event) {
      handlers.forEach(function (entry) {
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

export function upgrade(root) {
  bindDelegates();

  var scope = root || document;
  upgradeHandlers.forEach(function (handler) {
    handler(scope);
  });
}
