import { delegate, find, findAll, registerUpgrade } from '../lib/dom.js';
import { postJSON } from '../lib/fetch.js';
import { bufferToBase64url, base64urlToBuffer } from '../lib/encoding.js';

// Feature detection upgrade: show passkey UI elements when WebAuthn is available
// and the server has enabled passkey routes (data-passkey-enabled on the form).
registerUpgrade(function(scope) {
  if (!window.PublicKeyCredential) {
    return;
  }
  // Only activate when server has passkey routes enabled.
  var enabledEl = (scope.matches && scope.matches('[data-passkey-enabled]') ? scope : null) ||
    (scope.querySelector ? scope.querySelector('[data-passkey-enabled]') : null) ||
    (scope.closest ? scope.closest('[data-passkey-enabled]') : null);
  if (!enabledEl) {
    return;
  }
  findAll(scope, '[data-passkey-visible]').forEach(function(el) {
    el.classList.remove('hidden');
  });

});

// Authentication flow: sign in with passkey button.
delegate('click', '[data-passkey-authenticate]', function(event, el) {
  event.preventDefault();
  authenticate(el);
});

// Registration flow: add passkey button.
delegate('click', '[data-passkey-register]', function(event, el) {
  event.preventDefault();
  register(el);
});

function showPasskeyError(message) {
  var errEl = find(document, '[data-passkey-error]');
  if (!errEl) {
    return;
  }
  if (!message) {
    errEl.classList.add('hidden');
    errEl.textContent = '';
    return;
  }
  errEl.textContent = message;
  errEl.classList.remove('hidden');
}

function mapWebAuthnError(err) {
  if (err.name === 'NotAllowedError') return 'Cancelled by user or device.';
  if (err.name === 'InvalidStateError') return 'This passkey is already registered.';
  if (err.name === 'AbortError') return null;
  return 'Something went wrong. Try signing in with email.';
}

function authenticate(el) {
  var beginUrl  = el.dataset.beginUrl;
  var finishUrl = el.dataset.finishUrl;

  postJSON(beginUrl, {}).then(function(options) {
    var pk = options.publicKey;

    pk.challenge = base64urlToBuffer(pk.challenge);

    if (pk.allowCredentials) {
      pk.allowCredentials = pk.allowCredentials.map(function(c) {
        return Object.assign({}, c, { id: base64urlToBuffer(c.id) });
      });
    }

    return navigator.credentials.get({ publicKey: pk });
  }).then(function(credential) {
    var resp = credential.response;
    var body = {
      id: credential.id,
      rawId: bufferToBase64url(credential.rawId),
      type: credential.type,
      response: {
        authenticatorData: bufferToBase64url(resp.authenticatorData),
        clientDataJSON:    bufferToBase64url(resp.clientDataJSON),
        signature:         bufferToBase64url(resp.signature),
        userHandle:        resp.userHandle ? bufferToBase64url(resp.userHandle) : null,
      },
    };
    return postJSON(finishUrl, body);
  }).then(function(result) {
    window.location.href = result.redirect;
  }).catch(function(err) {
    var msg = mapWebAuthnError(err);
    if (msg) {
      showPasskeyError(msg);
    }
  });
}

function register(el) {
  var beginUrl  = el.dataset.beginUrl;
  var finishUrl = el.dataset.finishUrl;
  var listUrl   = el.dataset.listUrl;

  postJSON(beginUrl, {}).then(function(options) {
    var pk = options.publicKey;

    pk.challenge = base64urlToBuffer(pk.challenge);
    pk.user.id   = base64urlToBuffer(pk.user.id);

    if (pk.excludeCredentials) {
      pk.excludeCredentials = pk.excludeCredentials.map(function(c) {
        return Object.assign({}, c, { id: base64urlToBuffer(c.id) });
      });
    }

    return navigator.credentials.create({ publicKey: pk });
  }).then(function(credential) {
    var resp = credential.response;
    var name = 'Passkey ' + new Date().toISOString().slice(0, 10);
    var body = {
      id:   credential.id,
      rawId: bufferToBase64url(credential.rawId),
      type: credential.type,
      name: name,
      response: {
        attestationObject: bufferToBase64url(resp.attestationObject),
        clientDataJSON:    bufferToBase64url(resp.clientDataJSON),
        transports:        resp.getTransports ? resp.getTransports() : [],
      },
    };
    return postJSON(finishUrl, body);
  }).then(function() {
    if (window.htmx) {
      htmx.ajax('GET', listUrl, { target: '#passkeys-list-region', swap: 'innerHTML' });
    } else {
      location.reload();
    }
  }).catch(function(err) {
    console.error('[passkey] registration error:', err.name, '|', err.message);
    var msg = mapWebAuthnError(err);
    if (msg) {
      showPasskeyError(msg);
    }
  });
}
