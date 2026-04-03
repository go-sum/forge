# check=skip=InvalidDefaultArgInFrom
# ── Version pins ──────────────────────────────────────────────────────────────
# Build via Makefile (make tools-dev / make tools-prod) to ensure all
# args are forwarded automatically.
ARG AIR_VERSION
ARG GO_VERSION
ARG GOLANGCI_LINT_VERSION
ARG HUGO_VERSION
ARG PGSCHEMA_VERSION
ARG SQLC_VERSION
ARG TAILWIND_VERSION

# ───────────────────────────────────────────────────────────── go_toolchain ───
FROM golang:${GO_VERSION}-bookworm AS go_toolchain

# ────────────────────────────────────────────────────── tailwind_fetcher ───
FROM debian:bookworm-slim AS tailwind_fetcher
ARG TAILWIND_VERSION
ARG TARGETARCH
RUN apt-get update && apt-get install -y --no-install-recommends curl ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && case "${TARGETARCH}" in \
         amd64) TW_ARCH="x64" ;; \
         arm64) TW_ARCH="arm64" ;; \
         *) echo "unsupported: ${TARGETARCH}" >&2; exit 1 ;; \
       esac \
    && curl -fsSLo /usr/local/bin/tailwindcss \
         "https://github.com/tailwindlabs/tailwindcss/releases/download/v${TAILWIND_VERSION}/tailwindcss-linux-${TW_ARCH}" \
    && chmod +x /usr/local/bin/tailwindcss

# ────────────────────────────────────────────────────────── hugo_fetcher ───
FROM debian:bookworm-slim AS hugo_fetcher
ARG HUGO_VERSION
ARG TARGETARCH
RUN apt-get update && apt-get install -y --no-install-recommends curl ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && case "${TARGETARCH}" in \
         amd64) HUGO_ARCH="Linux-64bit" ;; \
         arm64) HUGO_ARCH="linux-arm64" ;; \
         *) echo "unsupported: ${TARGETARCH}" >&2; exit 1 ;; \
       esac \
    && curl -fsSLo /tmp/hugo.tar.gz \
         "https://github.com/gohugoio/hugo/releases/download/v${HUGO_VERSION}/hugo_${HUGO_VERSION}_${HUGO_ARCH}.tar.gz" \
    && tar -xzf /tmp/hugo.tar.gz -C /tmp hugo \
    && mv /tmp/hugo /usr/local/bin/hugo \
    && chmod +x /usr/local/bin/hugo \
    && rm -f /tmp/hugo.tar.gz

# ================================================================= targets ===
# ── Dev toolchain ────────────────────────────────────────────────────────────
# Used by: make lint, make test, make db-gen, make assets, make build
# Also provides air + tailwindcss binaries for dev_target via COPY --from
FROM go_toolchain AS dev
ARG AIR_VERSION
ARG SQLC_VERSION
ARG PGSCHEMA_VERSION
ARG GOLANGCI_LINT_VERSION
ARG TARGETARCH

ENV GOPATH=/root/go
ENV PATH="${GOPATH}/bin:/usr/local/go/bin:${PATH}"

RUN git config --global --add safe.directory /app

COPY --from=tailwind_fetcher /usr/local/bin/tailwindcss /usr/local/bin/tailwindcss
COPY --from=hugo_fetcher /usr/local/bin/hugo /usr/local/bin/hugo

RUN apt-get update && apt-get install -y --no-install-recommends \
      libgit2-dev pkgconf \
    && rm -rf /var/lib/apt/lists/*

RUN go install github.com/air-verse/air@${AIR_VERSION}
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@v${SQLC_VERSION}
RUN curl -fsSLo /usr/local/bin/pgschema \
      "https://github.com/pgplex/pgschema/releases/download/v${PGSCHEMA_VERSION}/pgschema-${PGSCHEMA_VERSION}-linux-${TARGETARCH}" && \
    chmod +x /usr/local/bin/pgschema
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh \
    | sh -s -- -b /usr/local/bin ${GOLANGCI_LINT_VERSION}

WORKDIR /app

# ── Prod toolchain ───────────────────────────────────────────────────────────
# Used by: assets_builder, healthcheck_builder, production_builder
# Minimal: Go runtime + tailwindcss + hugo for asset generation
FROM go_toolchain AS prod

ENV GOPATH=/root/go
ENV PATH="${GOPATH}/bin:/usr/local/go/bin:${PATH}"

RUN git config --global --add safe.directory /app

COPY --from=tailwind_fetcher /usr/local/bin/tailwindcss /usr/local/bin/tailwindcss
COPY --from=hugo_fetcher /usr/local/bin/hugo /usr/local/bin/hugo

WORKDIR /app
