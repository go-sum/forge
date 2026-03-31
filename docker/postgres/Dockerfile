FROM cgr.dev/chainguard/wolfi-base:latest

ARG PG_VERSION

RUN addgroup -S postgres && adduser -S -G postgres -D -H postgres

RUN apk update && apk add --no-cache \
      postgresql-${PG_VERSION} \
      postgresql-${PG_VERSION}-client \
      postgresql-${PG_VERSION}-contrib \
      postgresql-${PG_VERSION}-oci-entrypoint \
      gosu \
      su-exec

ENV PGDATA=/var/lib/postgresql/data
ENV LANG=C.UTF-8

# Wolfi installs extension .so files in /usr/lib/postgresql<ver>/ but
# PostgreSQL expects them in /usr/lib/postgresql<ver>/lib/.
RUN mkdir -p /usr/lib/postgresql${PG_VERSION}/lib && \
    ln -sf /usr/lib/postgresql${PG_VERSION}/*.so /usr/lib/postgresql${PG_VERSION}/lib/

RUN mkdir -p /var/lib/postgresql/data /var/run/postgresql /var/lib/postgres/initdb && \
    chown -R postgres:postgres /var/lib/postgresql /var/run/postgresql /var/lib/postgres

VOLUME /var/lib/postgresql/data
EXPOSE 5432

# Start as root — the entrypoint handles chown and drops to postgres via su-exec.
ENTRYPOINT ["/usr/bin/docker-entrypoint.sh"]
CMD ["postgres"]
