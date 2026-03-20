import { delegate } from '../lib/dom.js';

delegate('click', '[data-dialog-open]', function (event, opener) {
  var dialog = document.getElementById(opener.dataset.dialogOpen);
  if (dialog) {
    dialog.showModal();
  }
});

delegate('click', '[data-dialog-close]', function (event, closer) {
  var dialog = closer.closest('dialog');
  if (dialog) {
    dialog.close();
  }
});
