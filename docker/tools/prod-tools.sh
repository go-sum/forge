#!/usr/bin/env bash
set -eu

# ── Architecture ─────────────────────────────────────────────────────────────
case "${TARGETARCH}" in
  amd64) TW_ARCH="x64";   HUGO_ARCH="Linux-64bit" ;;
  arm64) TW_ARCH="arm64";  HUGO_ARCH="linux-arm64" ;;
  *)     echo "unsupported: ${TARGETARCH}" >&2; exit 1 ;;
esac

# ── Tailwind CSS ─────────────────────────────────────────────────────────────
curl -fsSLo /usr/local/bin/tailwindcss \
  "https://github.com/tailwindlabs/tailwindcss/releases/download/v${TAILWIND_VERSION}/tailwindcss-linux-${TW_ARCH}"
chmod +x /usr/local/bin/tailwindcss

# ── Hugo ─────────────────────────────────────────────────────────────────────
curl -fsSLo /tmp/hugo.tar.gz \
  "https://github.com/gohugoio/hugo/releases/download/v${HUGO_VERSION}/hugo_${HUGO_VERSION}_${HUGO_ARCH}.tar.gz"
tar -xzf /tmp/hugo.tar.gz -C /tmp hugo
mv /tmp/hugo /usr/local/bin/hugo
chmod +x /usr/local/bin/hugo
rm -f /tmp/hugo.tar.gz
