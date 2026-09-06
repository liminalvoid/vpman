#!/usr/bin/env bash
# Patroni reads its passwords from the environment and has no _FILE convention,
# so the secrets get lifted off disk here before it starts.
set -euo pipefail

export_secret() {
	local var="$1" file="$2"

	if [ -r "$file" ]; then
		export "$var"="$(cat "$file")"
	else
		echo "patroni-entrypoint: missing secret $file" >&2
		exit 1
	fi
}

export_secret PATRONI_SUPERUSER_PASSWORD \
	"${SUPERUSER_PASSWORD_FILE:-/run/secrets/pg_superuser_password}"
export_secret PATRONI_REPLICATION_PASSWORD \
	"${REPLICATION_PASSWORD_FILE:-/run/secrets/pg_replication_password}"

export PATRONI_SUPERUSER_USERNAME="${PATRONI_SUPERUSER_USERNAME:-postgres}"
export PATRONI_REPLICATION_USERNAME="${PATRONI_REPLICATION_USERNAME:-replicator}"

# A freshly created named volume can land root-owned; initdb refuses that.
DATA_ROOT="${PATRONI_DATA_ROOT:-/var/lib/postgresql/data}"
mkdir -p "$DATA_ROOT"
chown -R postgres:postgres "$DATA_ROOT"
chmod 0700 "$DATA_ROOT"

exec gosu postgres patroni /etc/patroni/patroni.yml
