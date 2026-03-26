# ── Version pins ─────────────────────────────────────────────────────────────
# All version strings live here. Update once; every stage picks up the change.
ARG GO_VERSION=1.26
ARG ALPINE_VERSION=3.20
ARG PGSCHEMA_VERSION=1.7.4
ARG TAILWIND_VERSION=4.1.3
ARG HTMX_VERSION=2.0.4
# Dev-only tools. Pin to a semver (e.g. v1.61.7) to improve reproducibility.
ARG AIR_VERSION=latest
ARG SQLC_VERSION=latest
# empty = latest; set to e.g. v2.1.6 to pin
ARG GOLANGCI_LINT_VERSION=

# ── Dev target ───────────────────────────────────────────────────────────────
FROM golang:${GO_VERSION}-alpine AS dev
ARG AIR_VERSION
ARG SQLC_VERSION
ARG PGSCHEMA_VERSION
ARG TAILWIND_VERSION
ARG GOLANGCI_LINT_VERSION

RUN apk add --no-cache curl libstdc++ gcc musl-dev openssl git

RUN go install github.com/air-verse/air@${AIR_VERSION}
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@${SQLC_VERSION}
RUN git config --global --add safe.directory /app
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
COPY go.mod go.sum go.work go.work.sum ./
COPY pkg/auth/go.mod pkg/auth/go.sum /app/pkg/auth/
COPY pkg/componentry/go.mod pkg/componentry/go.sum /app/pkg/componentry/
COPY pkg/security/go.mod pkg/security/go.sum /app/pkg/security/
COPY pkg/server/go.mod pkg/server/go.sum /app/pkg/server/
COPY pkg/site/go.mod pkg/server/go.sum /app/pkg/site/
COPY pkg/send/go.mod /app/pkg/send/
RUN go mod download
# Source mounted as volume — not copied
CMD ["go", "run", "./cli", "dev"]

# ── Assets stage ─────────────────────────────────────────────────────────────
# Builds the same asset pipeline used by dev so production does not drift.
FROM dev AS assets
ARG HTMX_VERSION
WORKDIR /app
COPY . .
RUN HTMX_VERSION=${HTMX_VERSION} go run ./cli build-assets --minify

# ── Production target ───────────────────────────────────────────────────────
FROM golang:${GO_VERSION}-alpine AS builder
WORKDIR /app
COPY go.mod go.sum go.work go.work.sum ./
COPY pkg/auth/go.mod pkg/auth/go.sum /app/pkg/auth/
COPY pkg/componentry/go.mod pkg/componentry/go.sum /app/pkg/componentry/
COPY pkg/security/go.mod pkg/security/go.sum /app/pkg/security/
COPY pkg/server/go.mod pkg/server/go.sum /app/pkg/server/
COPY pkg/site/go.mod pkg/server/go.sum /app/pkg/site/
COPY pkg/send/go.mod /app/pkg/send/
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
