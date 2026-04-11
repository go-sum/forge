export function csrfToken() {
  var meta = document.querySelector('meta[name="csrf-token"]');
  return meta ? meta.getAttribute('content') : '';
}

export function postJSON(url, body) {
  return fetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', 'X-CSRF-Token': csrfToken() },
    body: JSON.stringify(body || {}),
    credentials: 'same-origin',
  }).then(handleResponse);
}

export function deleteJSON(url) {
  return fetch(url, {
    method: 'DELETE',
    headers: { 'X-CSRF-Token': csrfToken() },
    credentials: 'same-origin',
  }).then(handleResponse);
}

function handleResponse(res) {
  if (!res.ok) {
    return res.json().catch(function() { return {}; }).then(function(data) {
      var err = new Error(data.message || res.statusText);
      err.status = res.status;
      throw err;
    });
  }
  var ct = res.headers.get('content-type') || '';
  if (ct.indexOf('application/json') !== -1) return res.json();
  return res.text();
}
