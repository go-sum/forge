# ── Version pins ─────────────────────────────────────────────────────────────
# All version strings live here. Update once; every stage picks up the change.
ARG GO_VERSION=1.26
ARG DEBIAN_VERSION=bookworm
ARG DEBIAN_STATIC=debian12
ARG PGSCHEMA_VERSION=1.7.4
ARG TAILWIND_VERSION=4.1.3
ARG HTMX_VERSION=2.0.4
ARG HUGO_VERSION=0.159.1
# Dev-only tools. Pin to a semver (e.g. v1.61.7) to improve reproducibility.
ARG AIR_VERSION=latest
ARG SQLC_VERSION=latest
# empty = latest; set to e.g. v2.1.6 to pin
ARG GOLANGCI_LINT_VERSION=

# ── Dev target ───────────────────────────────────────────────────────────────
FROM golang:${GO_VERSION}-bookworm AS dev_target
ARG AIR_VERSION
ARG SQLC_VERSION
ARG PGSCHEMA_VERSION
ARG TAILWIND_VERSION
ARG HUGO_VERSION
ARG GOLANGCI_LINT_VERSION
ARG TARGETARCH

RUN apt-get update && apt-get install -y --no-install-recommends \
      curl gcc git bash ca-certificates libgit2-dev pkg-config && \
    rm -rf /var/lib/apt/lists/*

RUN go install github.com/air-verse/air@${AIR_VERSION}
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@${SQLC_VERSION}
RUN git config --global --add safe.directory /app
RUN curl -fsSLo /usr/local/bin/pgschema \
      "https://github.com/pgplex/pgschema/releases/download/v${PGSCHEMA_VERSION}/pgschema-${PGSCHEMA_VERSION}-linux-${TARGETARCH}" && \
    chmod +x /usr/local/bin/pgschema
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh \
    | sh -s -- -b /usr/local/bin ${GOLANGCI_LINT_VERSION}
RUN TW_ARCH=$(echo "${TARGETARCH}" | sed 's/amd64/x64/') && \
    curl -fsSLo /usr/local/bin/tailwindcss \
      "https://github.com/tailwindlabs/tailwindcss/releases/download/v${TAILWIND_VERSION}/tailwindcss-linux-${TW_ARCH}" && \
    chmod +x /usr/local/bin/tailwindcss
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

WORKDIR /app
COPY go.mod go.sum go.work go.work.sum ./
COPY pkg/auth/go.mod pkg/auth/go.sum /app/pkg/auth/
COPY pkg/componentry/go.mod pkg/componentry/go.sum /app/pkg/componentry/
COPY pkg/kv/go.mod pkg/kv/go.sum /app/pkg/kv/
COPY pkg/security/go.mod pkg/security/go.sum /app/pkg/security/
COPY pkg/send/go.mod /app/pkg/send/
COPY pkg/server/go.mod pkg/server/go.sum /app/pkg/server/
COPY pkg/session/go.mod pkg/session/go.sum /app/pkg/session/
COPY pkg/site/go.mod pkg/server/go.sum /app/pkg/site/
RUN go mod download
# Source mounted as volume — not copied
CMD ["go", "run", "./cli/dev"]

# ── Assets stage ─────────────────────────────────────────────────────────────
# Builds the same asset pipeline used by dev so production does not drift.
FROM dev_target AS assets_stage
ARG HTMX_VERSION
WORKDIR /app
COPY . .
RUN HTMX_VERSION=${HTMX_VERSION} go run ./cli/build assets --minify

# ── Production target ───────────────────────────────────────────────────────
FROM golang:${GO_VERSION}-bookworm AS builder_stage
WORKDIR /app
# go.prod.mod has no replace directives — pkg/ modules resolve from GitHub.
# go.prod.sum locks exact checksums for a reproducible build.
# To regenerate: make prod-sync
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
    CGO_ENABLED=0 go build -o /server ./cmd/server

ARG DEBIAN_STATIC
FROM gcr.io/distroless/static-${DEBIAN_STATIC}:nonroot AS production_target
WORKDIR /
ENV PUBLIC_DIR=public \
    PUBLIC_PREFIX=/public
COPY --from=builder_stage --chown=nonroot:nonroot /server /server
COPY --from=assets_stage --chown=nonroot:nonroot /app/public/ /public/
COPY --from=builder_stage --chown=nonroot:nonroot /app/config/ /config/
EXPOSE 8080
CMD ["/server"]
