# syntax=docker/dockerfile:1

# Patroni on top of the official Postgres image. Built here rather than pulled
# so the Postgres major version, the Patroni version and the config surface are
# all pinned by this repo instead of by somebody else's release cadence.
FROM postgres:17-bookworm

ARG PATRONI_VERSION=4.0.4

RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends \
        python3 python3-venv curl ca-certificates; \
    python3 -m venv /opt/patroni; \
    /opt/patroni/bin/pip install --no-cache-dir --upgrade pip; \
    /opt/patroni/bin/pip install --no-cache-dir \
        "patroni[etcd3]==${PATRONI_VERSION}" "psycopg[binary]"; \
    ln -s /opt/patroni/bin/patroni /usr/local/bin/patroni; \
    ln -s /opt/patroni/bin/patronictl /usr/local/bin/patronictl; \
    rm -rf /var/lib/apt/lists/*

COPY entrypoint.sh /usr/local/bin/patroni-entrypoint.sh
RUN chmod +x /usr/local/bin/patroni-entrypoint.sh

EXPOSE 5432 8008

ENTRYPOINT ["/usr/local/bin/patroni-entrypoint.sh"]
