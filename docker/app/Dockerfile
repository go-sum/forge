# check=skip=InvalidDefaultArgInFrom
# ── Version pins ─────────────────────────────────────────────────────────────
# Version strings are passed as --build-arg from .versions (project root).
# Build via Makefile (make docker-dev / make docker-build) to ensure all
# args are forwarded automatically.
ARG GO_VERSION
ARG PGSCHEMA_VERSION
ARG TAILWIND_VERSION
ARG HTMX_VERSION
ARG HUGO_VERSION
# Dev-only tools. Pin to a semver (e.g. v1.61.7) to improve reproducibility.
ARG AIR_VERSION
ARG SQLC_VERSION
# empty = latest; set to e.g. v2.1.6 to pin
ARG GOLANGCI_LINT_VERSION

# ── Go toolchain (pinned version from official image) ────────────────────────
FROM golang:${GO_VERSION}-bookworm AS go_toolchain

# ── Dev toolchain (bookworm — Go + air + tailwindcss) ────────────
# Used by dev_target (hot-reload) and as the base for cli_toolchain.
FROM go_toolchain AS dev_toolchain
ARG AIR_VERSION
ARG TAILWIND_VERSION
ARG TARGETARCH

ENV GOPATH=/root/go
ENV PATH="${GOPATH}/bin:/usr/local/go/bin:${PATH}"

RUN go install github.com/air-verse/air@${AIR_VERSION}
RUN git config --global --add safe.directory /app
RUN TW_ARCH=$(echo "${TARGETARCH}" | sed 's/amd64/x64/') && \
    curl -fsSLo /usr/local/bin/tailwindcss \
      "https://github.com/tailwindlabs/tailwindcss/releases/download/v${TAILWIND_VERSION}/tailwindcss-linux-${TW_ARCH}" && \
    chmod +x /usr/local/bin/tailwindcss

WORKDIR /app

# ── CLI toolchain (bookworm — sqlc, pgschema, lint, hugo, libgit2) ──
# Used for cli tool commands: lint, test, build, db-*, assets, package-*.
FROM dev_toolchain AS cli_toolchain
ARG SQLC_VERSION
ARG PGSCHEMA_VERSION
ARG HUGO_VERSION
ARG GOLANGCI_LINT_VERSION
ARG TARGETARCH

RUN apt-get update && apt-get install -y --no-install-recommends \
      libgit2-dev pkgconf \
    && rm -rf /var/lib/apt/lists/*

RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@${SQLC_VERSION}
RUN curl -fsSLo /usr/local/bin/pgschema \
      "https://github.com/pgplex/pgschema/releases/download/v${PGSCHEMA_VERSION}/pgschema-${PGSCHEMA_VERSION}-linux-${TARGETARCH}" && \
    chmod +x /usr/local/bin/pgschema
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh \
    | sh -s -- -b /usr/local/bin ${GOLANGCI_LINT_VERSION}
RUN case "${TARGETARCH}" in \
      amd64) HUGO_ARCH="Linux-64bit" ;; \
      arm64) HUGO_ARCH="linux-arm64" ;; \
      *) echo "unsupported architecture: ${TARGETARCH}" >&2; exit 1 ;; \
    esac && \
    curl -fsSLo /tmp/hugo.tar.gz \
      "https://github.com/gohugoio/hugo/releases/download/v${HUGO_VERSION}/hugo_${HUGO_VERSION}_${HUGO_ARCH}.tar.gz" && \
    tar -xzf /tmp/hugo.tar.gz -C /tmp hugo && \
    mv /tmp/hugo /usr/local/bin/hugo && \
    chmod +x /usr/local/bin/hugo && \
    rm -f /tmp/hugo.tar.gz

# ── Dev target (wolfi — minimal hot-reload container) ────────────────────────
# Used only for `make dev` via docker compose. Contains Go + air + tailwindcss.
# Heavy CLI tools (sqlc, pgschema, hugo, lint) run in cli_toolchain instead.
FROM cgr.dev/chainguard/wolfi-base:latest AS dev_target
ARG AIR_VERSION
ARG TARGETARCH

COPY --from=go_toolchain /usr/local/go /usr/local/go
COPY --from=dev_toolchain /usr/local/bin/tailwindcss /usr/local/bin/tailwindcss
RUN apk update && apk add --no-cache bash ca-certificates-bundle curl git libstdc++

ENV GOPATH=/root/go
ENV PATH="${GOPATH}/bin:/usr/local/go/bin:${PATH}"

RUN go install github.com/air-verse/air@${AIR_VERSION}
RUN git config --global --add safe.directory /app

WORKDIR /app
COPY go.mod go.sum go.work go.work.sum ./
COPY pkg/auth/go.mod pkg/auth/go.sum /app/pkg/auth/
COPY pkg/componentry/go.mod pkg/componentry/go.sum /app/pkg/componentry/
COPY pkg/kv/go.mod pkg/kv/go.sum /app/pkg/kv/
COPY pkg/security/go.mod pkg/security/go.sum /app/pkg/security/
COPY pkg/send/go.mod /app/pkg/send/
COPY pkg/server/go.mod pkg/server/go.sum /app/pkg/server/
COPY pkg/session/go.mod pkg/session/go.sum /app/pkg/session/
COPY pkg/site/go.mod pkg/site/go.sum /app/pkg/site/
RUN go mod download
# Source mounted as volume — not copied
CMD ["go", "run", "./cli/dev"]

# ── Assets stage ─────────────────────────────────────────────────────────────
# Builds the same asset pipeline used by dev so production does not drift.
# Uses cli_toolchain — needs both tailwindcss and hugo.
FROM cli_toolchain AS assets_stage
ARG HTMX_VERSION
WORKDIR /app
COPY . .
RUN HTMX_VERSION=${HTMX_VERSION} go run ./cli/build assets --minify

# ── App healthcheck binary ──────────────────────────────────────────────────
# Tiny static binary for Docker HEALTHCHECK in the distroless production image.
FROM golang:${GO_VERSION}-bookworm AS healthcheck_builder
WORKDIR /build
COPY docker/app/healthcheck.go .
RUN CGO_ENABLED=0 go build -ldflags='-s -w' -o /healthcheck healthcheck.go

# ── Builder stage ────────────────────────────────────────────────────────────
FROM cgr.dev/chainguard/wolfi-base:latest AS builder_stage
COPY --from=go_toolchain /usr/local/go /usr/local/go
RUN apk update && apk add --no-cache ca-certificates-bundle git
ENV PATH="/usr/local/go/bin:${PATH}"

WORKDIR /app
# go.prod.mod has no replace directives — pkg/ modules resolve from GitHub.
# go.prod.sum locks exact checksums for a reproducible build.
# To regenerate: make package-sync
COPY go.prod.mod go.mod
COPY go.prod.sum go.sum
# Private modules require GITHUB_ACCESS_TOKEN passed as a build secret.
# Usage: docker build --secret id=github_token,env=GITHUB_ACCESS_TOKEN ...
# The .netrc file persists auth across RUN steps but never reaches the final image.
RUN --mount=type=secret,id=github_token \
    TOKEN="$(cat /run/secrets/github_token 2>/dev/null)" && \
    if [ -n "${TOKEN}" ]; then \
      printf "machine github.com\nlogin x-access-token\npassword %s\n" "${TOKEN}" > /root/.netrc && \
      chmod 600 /root/.netrc; \
    fi && \
    GONOSUMDB='github.com/go-sum/*' GOPRIVATE='github.com/go-sum/*' go mod download
COPY cmd/ ./cmd/
COPY cli/ ./cli/
COPY config/ ./config/
RUN rm -f config/*.development.yaml
COPY db/ ./db/
COPY internal/ ./internal/
RUN GONOSUMDB='github.com/go-sum/*' GOPRIVATE='github.com/go-sum/*' \
    CGO_ENABLED=0 go build -o /bin/server ./cmd/server

# ── Production target ───────────────────────────────────────────────────────
FROM cgr.dev/chainguard/static:latest AS production_target
WORKDIR /
ENV PUBLIC_DIR=public \
    PUBLIC_PREFIX=/public
COPY --from=builder_stage --chown=nonroot:nonroot /bin/server /bin/server
COPY --from=healthcheck_builder --chown=nonroot:nonroot /healthcheck /healthcheck
COPY --from=assets_stage --chown=nonroot:nonroot /app/public/ /public/
COPY --from=builder_stage --chown=nonroot:nonroot /app/config/ /config/
USER nonroot
EXPOSE 8080
CMD ["/bin/server"]
