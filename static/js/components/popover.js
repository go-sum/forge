import { delegate, remove } from '../lib/dom.js';

// Close any open popover when clicking outside it.
delegate('click', null, function (event) {
  document.querySelectorAll('details[data-popover][open]').forEach(function (details) {
    if (!details.contains(event.target)) {
      details.open = false;
    }
  });
});

// Close the containing popover when a menu item (button or link) is activated,
// then return focus to the <summary> trigger so keyboard users are not stranded.
delegate('click', 'details[data-popover] a, details[data-popover] button:not(summary)', function (event, item) {
  var details = item.closest('details[data-popover]');
  if (!details) return;
  details.open = false;
  var summary = details.querySelector('summary');
  if (summary) summary.focus();
});

// Escape closes the topmost open popover and returns focus to its trigger.
delegate('keydown', null, function (event) {
  if (event.key !== 'Escape') return;
  var open = document.querySelector('details[data-popover][open]');
  if (!open) return;
  event.preventDefault();
  open.open = false;
  var summary = open.querySelector('summary');
  if (summary) summary.focus();
});
