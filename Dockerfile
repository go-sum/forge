# ── Version pins ────────────────────────────────
# All version strings live here. Update once; every stage picks up the change.
ARG GO_VERSION=1.26
ARG ALPINE_VERSION=3.20
ARG PGSCHEMA_VERSION=1.7.4
ARG TAILWIND_VERSION=4.1.3
ARG HTMX_VERSION=2.0.4
ARG ALPINEJS_VERSION=3.14.8
# Dev-only tools. Pin to a semver (e.g. v1.61.7) to improve reproducibility.
ARG AIR_VERSION=latest
ARG SQLC_VERSION=latest
ARG GOLANGCI_LINT_VERSION=   # empty = latest; set to e.g. v2.1.6 to pin

# ── Dev target ──────────────────────────────────
FROM golang:${GO_VERSION}-alpine AS dev
ARG AIR_VERSION
ARG SQLC_VERSION
ARG PGSCHEMA_VERSION
ARG TAILWIND_VERSION
ARG GOLANGCI_LINT_VERSION

RUN apk add --no-cache curl libstdc++ gcc musl-dev

RUN go install github.com/air-verse/air@${AIR_VERSION}
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@${SQLC_VERSION}
RUN ARCH=$(uname -m | sed 's/x86_64/amd64/' | sed 's/aarch64/arm64/') && \
    curl -fsSLo /usr/local/bin/pgschema \
      "https://github.com/pgplex/pgschema/releases/download/v${PGSCHEMA_VERSION}/pgschema-${PGSCHEMA_VERSION}-linux-${ARCH}" && \
    chmod +x /usr/local/bin/pgschema
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh \
    | sh -s -- -b /usr/local/bin ${GOLANGCI_LINT_VERSION}
RUN ARCH=$(uname -m | sed 's/x86_64/x64/' | sed 's/aarch64/arm64/') && \
    curl -fsSLo /usr/local/bin/tailwindcss \
      "https://github.com/tailwindlabs/tailwindcss/releases/download/v${TAILWIND_VERSION}/tailwindcss-linux-${ARCH}-musl" && \
    chmod +x /usr/local/bin/tailwindcss

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
# Source mounted as volume — not copied
CMD ["air", "-c", ".air.toml"]

# ── Assets stage ────────────────────────────────
# Downloads vendor JS and compiles Tailwind CSS independently of the Go build.
FROM alpine:${ALPINE_VERSION} AS assets
ARG HTMX_VERSION
ARG ALPINEJS_VERSION
ARG TAILWIND_VERSION

RUN apk add --no-cache curl
WORKDIR /app

# Download vendor JS libraries
RUN mkdir -p public/js && \
    curl -fsSL "https://unpkg.com/htmx.org@${HTMX_VERSION}/dist/htmx.min.js" -o public/js/htmx.min.js && \
    curl -fsSL "https://unpkg.com/@alpinejs/csp@${ALPINEJS_VERSION}/dist/cdn.min.js" -o public/js/alpine.min.js

# Copy hand-written JS (from static/ — the canonical source location)
COPY static/js/ public/js/

# Install Tailwind standalone CLI and compile CSS
RUN ARCH=$(uname -m | sed 's/x86_64/x64/' | sed 's/aarch64/arm64/') && \
    curl -fsSLo /usr/local/bin/tailwindcss \
      "https://github.com/tailwindlabs/tailwindcss/releases/download/v${TAILWIND_VERSION}/tailwindcss-linux-${ARCH}-musl" && \
    chmod +x /usr/local/bin/tailwindcss
# Copy full context so @source can scan internal/view/**/*.go when the view layer exists
COPY . .
RUN mkdir -p public/css && \
    tailwindcss -i ./static/css/tailwind.css -o ./public/css/app.css --minify

# ── Production target ──────────────────────────
FROM golang:${GO_VERSION}-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

FROM alpine:${ALPINE_VERSION} AS production
RUN apk add --no-cache ca-certificates
COPY --from=builder /server /server
COPY --from=assets /app/public/ /public/
COPY config/ /config/
EXPOSE 8080
CMD ["/server"]
