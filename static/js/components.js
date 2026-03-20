// Shared runtime for pkg/components.
// Focused behavior modules live under static/js/components/, while this file remains the shared entrypoint.

import { upgrade } from './lib/dom.js';
import './components/dialog.js';
import './components/feedback.js';
import './components/fileupload.js';
import './components/popover.js';
import './components/tabs.js';
import './components/theme.js';

export { closest, data, delegate, find, findAll, remove, toggleClass, upgrade } from './lib/dom.js';

export function init() {
  if (document.documentElement.dataset.componentsInitialized === 'true') {
    return;
  }

  document.documentElement.dataset.componentsInitialized = 'true';
  upgrade(document);

  document.addEventListener('htmx:afterSettle', function (event) {
    if (event.detail && event.detail.elt) {
      upgrade(event.detail.elt);
    }
  });
}
