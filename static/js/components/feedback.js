import { delegate, find, remove } from '../lib/dom.js';

function queueToastRemoval(node) {
  setTimeout(function () {
    if (!node.parentNode) {
      return;
    }

    node.style.transition = 'opacity 300ms ease-out';
    node.style.opacity = '0';
    setTimeout(function () {
      remove(node);
    }, 300);
  }, 5000);
}

delegate('click', '[data-dismiss]', function (event, dismiss) {
  var dismissible = dismiss.closest('[data-dismissible]');
  if (dismissible) {
    remove(dismissible);
  }
});

// Example/demo helper: clone a toast template into the shared toast container.
delegate('click', '[data-toast-trigger]', function (event, trigger) {
  var tmpl = document.getElementById(trigger.dataset.toastTrigger);
  var container = document.getElementById('toast-container');
  var content = tmpl && tmpl.content ? tmpl.content.firstElementChild : null;
  if (!container || !content) {
    return;
  }

  var node = content.cloneNode(true);
  container.appendChild(node);
  queueToastRemoval(node);
});
