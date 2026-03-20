import { closest, data, delegate, find } from '../lib/dom.js';

function updateFileName(zone, input) {
  var display = find(zone, '[data-file-name]');
  if (!display) {
    return;
  }

  var files = input.files;
  if (files.length === 0) {
    display.textContent = '';
  } else if (files.length === 1) {
    display.textContent = files[0].name;
  } else {
    display.textContent = files.length + ' files selected';
  }
}

function clearDragging(zone) {
  delete zone.dataset.dragging;
}

delegate('dragover', '[data-file-upload]', function (event, zone) {
  event.preventDefault();
  zone.dataset.dragging = '';
});

delegate('dragleave', '[data-file-upload]', function (event, zone) {
  if (zone.contains(event.relatedTarget)) {
    return;
  }
  clearDragging(zone);
});

delegate('drop', '[data-file-upload]', function (event, zone) {
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

delegate('change', 'input[type="file"]', function (event, input) {
  var zone = closest(input, '[data-file-upload]');
  if (!zone) {
    return;
  }

  updateFileName(zone, input);
});
