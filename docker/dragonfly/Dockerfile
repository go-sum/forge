# check=skip=InvalidDefaultArgInFrom
ARG DRAGONFLY_VERSION
ARG GO_VERSION

FROM docker.dragonflydb.io/dragonflydb/dragonfly:${DRAGONFLY_VERSION} AS upstream

FROM golang:${GO_VERSION}-bookworm AS healthcheck_builder
WORKDIR /build
COPY healthcheck.go .
RUN CGO_ENABLED=0 go build -ldflags='-s -w' -o /healthcheck healthcheck.go

FROM cgr.dev/chainguard/wolfi-base:latest

RUN apk update && apk add --no-cache \
      ca-certificates-bundle \
      libgcc \
      libstdc++ \
      libunwind \
      openssl \
      tini \
      zlib

COPY --from=upstream /usr/local/bin/dragonfly /usr/local/bin/dragonfly
COPY --from=healthcheck_builder /healthcheck /usr/local/bin/healthcheck

RUN addgroup -g 999 dfly && adduser -u 999 -G dfly -D -H dfly

EXPOSE 6379
USER dfly

ENTRYPOINT ["tini", "--"]
CMD ["dragonfly", "--logtostderr"]
