#!/usr/bin/env bash
set -eu

# ── Architecture ─────────────────────────────────────────────────────────────
case "${TARGETARCH}" in
  amd64) TW_ARCH="x64";   HUGO_ARCH="Linux-64bit" ;;
  arm64) TW_ARCH="arm64";  HUGO_ARCH="linux-arm64" ;;
  *)     echo "unsupported: ${TARGETARCH}" >&2; exit 1 ;;
esac

# ── System packages ──────────────────────────────────────────────────────────
apt-get update && apt-get install -y --no-install-recommends libgit2-dev pkgconf
rm -rf /var/lib/apt/lists/*

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

# ── Air (hot-reload) ────────────────────────────────────────────────────────
go install github.com/air-verse/air@${AIR_VERSION}

# ── sqlc ─────────────────────────────────────────────────────────────────────
go install github.com/sqlc-dev/sqlc/cmd/sqlc@v${SQLC_VERSION}

# ── pgschema ─────────────────────────────────────────────────────────────────
curl -fsSLo /usr/local/bin/pgschema \
  "https://github.com/pgplex/pgschema/releases/download/v${PGSCHEMA_VERSION}/pgschema-${PGSCHEMA_VERSION}-linux-${TARGETARCH}"
chmod +x /usr/local/bin/pgschema

# ── golangci-lint ────────────────────────────────────────────────────────────
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh \
  | sh -s -- -b /usr/local/bin ${GOLANGCI_LINT_VERSION}
