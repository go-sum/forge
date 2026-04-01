# check=skip=InvalidDefaultArgInFrom
# ── Version pins ──────────────────────────────────────────────────────────────
# Version strings are passed as --build-arg from .versions (project root).
# Build via Makefile (make docker-dev / make docker-build) to ensure all
# args are forwarded automatically.
ARG AIR_VERSION
ARG GO_VERSION
ARG GOLANGCI_LINT_VERSION
ARG HTMX_VERSION
ARG HUGO_VERSION
ARG PGSCHEMA_VERSION
ARG SQLC_VERSION
ARG TAILWIND_VERSION

# ───────────────────────────────────────────────────────────── go_toolchain ───
FROM golang:${GO_VERSION}-bookworm AS go_toolchain

# ──────────────────────────────────────────────────────────── dev_toolchain ───
FROM go_toolchain AS dev_toolchain
ARG AIR_VERSION
ARG TAILWIND_VERSION
ARG TARGETARCH

ENV GOPATH=/root/go
ENV PATH="${GOPATH}/bin:/usr/local/go/bin:${PATH}"

RUN go install github.com/air-verse/air@${AIR_VERSION}
RUN git config --global --add safe.directory /app
RUN case "${TARGETARCH}" in \
      amd64) TW_ARCH="x64" ;; \
      arm64) TW_ARCH="arm64" ;; \
      *) echo "unsupported architecture: ${TARGETARCH}" >&2; exit 1 ;; \
    esac && \
    curl -fsSLo /usr/local/bin/tailwindcss \
      "https://github.com/tailwindlabs/tailwindcss/releases/download/v${TAILWIND_VERSION}/tailwindcss-linux-${TW_ARCH}" && \
    chmod +x /usr/local/bin/tailwindcss

WORKDIR /app

# ──────────────────────────────────────────────────────────── cli_toolchain ───
FROM dev_toolchain AS cli_toolchain
ARG SQLC_VERSION
ARG PGSCHEMA_VERSION
ARG HUGO_VERSION
ARG GOLANGCI_LINT_VERSION
ARG TARGETARCH

RUN apt-get update && apt-get install -y --no-install-recommends \
      libgit2-dev pkgconf \
    && rm -rf /var/lib/apt/lists/*

RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@v${SQLC_VERSION}
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

# ===================================================================== dev ====
# ─────────────────────────────────────────────────────────────── dev_target ───
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

# =============================================================== production ===
# ─────────────────────────────────────────────────────────── assets_builder ───
FROM cli_toolchain AS assets_builder
ARG HTMX_VERSION
WORKDIR /app
COPY . .
RUN HTMX_VERSION=${HTMX_VERSION} go run ./cli/build assets --minify && \
    go run ./cli/build docs && \
    go run ./cli/build sprites

# ────────────────────────────────────────────────────── healthcheck_builder ───
FROM golang:${GO_VERSION}-bookworm AS healthcheck_builder
WORKDIR /build
COPY docker/app/healthcheck.go .
RUN CGO_ENABLED=0 go build -ldflags='-s -w' -o /healthcheck healthcheck.go

# ─────────────────────────────────────────────────────── production_builder ───
FROM cgr.dev/chainguard/wolfi-base:latest AS production_builder
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

# ──────────────────────────────────────────────────────── production_target ───
FROM cgr.dev/chainguard/static:latest AS production_target
WORKDIR /
ENV PUBLIC_DIR=public \
    PUBLIC_PREFIX=/public
COPY --from=production_builder --chown=nonroot:nonroot /bin/server /bin/server
COPY --from=healthcheck_builder --chown=nonroot:nonroot /healthcheck /healthcheck
COPY --from=assets_builder --chown=nonroot:nonroot /app/public/ /public/
COPY --from=production_builder --chown=nonroot:nonroot /app/config/ /config/
USER nonroot
EXPOSE 8080
CMD ["/bin/server"]
