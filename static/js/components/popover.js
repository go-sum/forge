import { delegate, remove } from '../lib/dom.js';

delegate('click', null, function (event) {
  document.querySelectorAll('details[data-popover][open]').forEach(function (details) {
    if (!details.contains(event.target)) {
      details.open = false;
    }
  });
});
