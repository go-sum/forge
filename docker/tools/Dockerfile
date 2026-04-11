# check=skip=InvalidDefaultArgInFrom
# ── Dev toolchain ────────────────────────────────────────────────────────────
# Pre-built image for development commands: make lint, test, db-gen, assets, build.
# Also provides air + tailwindcss binaries for dev_target via COPY --from.
ARG GO_VERSION
ARG AIR_VERSION
ARG GOLANGCI_LINT_VERSION
ARG HUGO_VERSION
ARG MKCERT_VERSION
ARG PGSCHEMA_VERSION
ARG SQLC_VERSION
ARG TAILWIND_VERSION

# ──────────────────────────────────────────────────────────────── dev_tools ───
FROM golang:${GO_VERSION}-bookworm AS dev_tools
ARG AIR_VERSION
ARG GOLANGCI_LINT_VERSION
ARG HUGO_VERSION
ARG MKCERT_VERSION
ARG PGSCHEMA_VERSION
ARG SQLC_VERSION
ARG TAILWIND_VERSION
ARG TARGETARCH

ENV GOPATH=/root/go
ENV PATH="${GOPATH}/bin:/usr/local/go/bin:${PATH}"

RUN git config --global --add safe.directory /app

COPY dev-tools.sh /tmp/
RUN bash /tmp/dev-tools.sh && rm /tmp/dev-tools.sh

WORKDIR /app

# ── Production tools (Hugo + Tailwind only) ──────────────────────────────────
# Pre-built image for production builds: avoids re-downloading Hugo + Tailwind
# on every deploy. Mirrors the dev_tools caching strategy.
FROM golang:${GO_VERSION}-bookworm AS production_tools
ARG HUGO_VERSION
ARG MKCERT_VERSION
ARG TAILWIND_VERSION
ARG TARGETARCH

ENV GOPATH=/root/go
ENV PATH="${GOPATH}/bin:/usr/local/go/bin:${PATH}"

RUN git config --global --add safe.directory /app

COPY prod-tools.sh /tmp/
RUN bash /tmp/prod-tools.sh && rm /tmp/prod-tools.sh

WORKDIR /app
