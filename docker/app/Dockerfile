# check=skip=InvalidDefaultArgInFrom
# ── External images ──────────────────────────────────────────────────────────
# Pre-built toolchain images from docker/tools/Dockerfile.
# Build via Makefile (make tools-dev / make tools-prod) before building app.
ARG GO_VERSION
ARG HTMX_VERSION
ARG TOOLS_DEV_IMAGE=forge-tools:dev
ARG TOOLS_PROD_IMAGE=forge-tools:prod

FROM ${TOOLS_DEV_IMAGE}  AS tools_dev
FROM ${TOOLS_PROD_IMAGE} AS tools_prod

# ===================================================================== dev ====
# ─────────────────────────────────────────────────────────────── dev_target ───
FROM cgr.dev/chainguard/wolfi-base:latest AS dev_target

COPY --from=tools_dev /usr/local/go /usr/local/go
COPY --from=tools_dev /root/go/bin/air /usr/local/bin/air
COPY --from=tools_dev /usr/local/bin/tailwindcss /usr/local/bin/tailwindcss
RUN apk update && apk add --no-cache bash ca-certificates-bundle curl git libstdc++

ENV GOPATH=/root/go
ENV PATH="${GOPATH}/bin:/usr/local/go/bin:${PATH}"

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
FROM tools_prod AS assets_builder
ARG HTMX_VERSION
WORKDIR /app
COPY . .
RUN HTMX_VERSION=${HTMX_VERSION} go run ./cli/build assets --minify && \
    go run ./cli/build docs && \
    go run ./cli/build sprites

# ────────────────────────────────────────────────────── healthcheck_builder ───
# healthcheck.go uses only stdlib — no workspace modules needed.
FROM tools_prod AS healthcheck_builder
ARG GO_VERSION
WORKDIR /build
COPY docker/app/healthcheck.go main.go
RUN printf 'module healthcheck\ngo %s\n' "${GO_VERSION}" > go.mod && \
    CGO_ENABLED=0 go build -ldflags='-s -w' -o /healthcheck .

# ─────────────────────────────────────────────────────── production_builder ───
FROM cgr.dev/chainguard/wolfi-base:latest AS production_builder
COPY --from=tools_prod /usr/local/go /usr/local/go
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
ARG APP_FOLDER=/opt/server
WORKDIR ${APP_FOLDER}
ENV APP_FOLDER=${APP_FOLDER} \
    PUBLIC_DIR=${APP_FOLDER}/public \
    PUBLIC_PREFIX=/public
COPY --from=production_builder  --chown=nonroot:nonroot /bin/server       ${APP_FOLDER}/server
COPY --from=healthcheck_builder --chown=nonroot:nonroot /healthcheck       /opt/healthcheck
COPY --from=assets_builder      --chown=nonroot:nonroot /app/public/       ${APP_FOLDER}/public/
COPY --from=production_builder  --chown=nonroot:nonroot /app/config/*.yaml ${APP_FOLDER}/config/
USER nonroot
EXPOSE 8080
CMD ["/opt/server/server"]
