// Alpine.js component registrations.
//
// This file must be loaded (via defer) BEFORE alpine.min.js so that the
// 'alpine:init' listener is registered before Alpine fires the event.
// With defer, scripts execute in document order — app.js appears first in
// base.go, so it always runs first.
//
// The @alpinejs/csp build requires all x-data values to reference a named
// component registered here; inline object literals ("{ show: true }") are
// not supported by the CSP evaluator.

document.addEventListener('alpine:init', () => {
  // Used by pkg/ui/alert.go for dismissible alert banners.
  Alpine.data('dismissible', () => ({
    isHidden: false,
    alertClass: '',
    dismiss() {
      this.isHidden = true;
      this.alertClass = 'alert-hidden';
    },
  }));
});
