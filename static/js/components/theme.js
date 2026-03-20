import { delegate } from '../lib/dom.js';

function cycleTheme() {
  var order = ['light', 'dark', 'system'];
  var html = document.documentElement;
  var current = html.dataset.themePreference || 'system';
  var next = order[(order.indexOf(current) + 1) % order.length];

  html.dataset.themePreference = next;
  localStorage.setItem('themePreference', next);
  html.classList.toggle(
    'dark',
    next === 'dark' || (next === 'system' && matchMedia('(prefers-color-scheme: dark)').matches)
  );
}

delegate('click', '[data-theme-toggle]', function () {
  cycleTheme();
});
