import { data, findAll, registerUpgrade, toggleClass } from '../lib/dom.js';

var activeClasses = ['bg-background', 'text-foreground', 'shadow'];

function activateTab(container, trigger, focus) {
  var triggers = findAll(container, '[role="tab"]');
  var panels = findAll(container, '[role="tabpanel"]');
  var activeTab = data(trigger, 'tab');

  triggers.forEach(function (candidate) {
    var active = candidate === trigger;
    candidate.setAttribute('aria-selected', active ? 'true' : 'false');
    candidate.setAttribute('tabindex', active ? '0' : '-1');
    activeClasses.forEach(function (className) {
      toggleClass(candidate, className, active);
    });
  });

  panels.forEach(function (panel) {
    panel.hidden = data(panel, 'tab') !== activeTab;
  });

  if (focus) {
    trigger.focus();
  }
}

function mount(container) {
  if (container.dataset.tabsInitialized === 'true') {
    return;
  }

  var triggers = findAll(container, '[role="tab"]');
  if (!triggers.length) {
    return;
  }

  container.dataset.tabsInitialized = 'true';
  var selected = triggers.find(function (trigger) {
    return trigger.getAttribute('aria-selected') === 'true';
  }) || triggers[0];

  activateTab(container, selected, false);

  triggers.forEach(function (trigger, index) {
    trigger.addEventListener('click', function () {
      activateTab(container, trigger, false);
    });

    trigger.addEventListener('keydown', function (event) {
      var nextIndex = index;
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

registerUpgrade(function (root) {
  findAll(root, '[data-tabs]').forEach(mount);
});
