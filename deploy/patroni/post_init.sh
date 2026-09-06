#!/usr/bin/env bash
# Runs once, on the member that bootstraps the cluster. Patroni passes a libpq
# connection string as $1 (a postgres:// URI over the local socket). Everything
# done here reaches the other members through streaming replication, so it must
# not be repeated by hand on the replicas.
set -euo pipefail

APP_PASSWORD="$(cat "${APP_PASSWORD_FILE:-/run/secrets/pg_app_password}")"

# One psql session: bootstrap.sql switches databases with \c, which reuses this
# connection's socket and user instead of us trying to rewrite the URI.
exec psql "$1" \
	--no-psqlrc \
	-v ON_ERROR_STOP=1 \
	-v app_user="${APP_DB_USER:-vpn}" \
	-v app_db="${APP_DB_NAME:-vpn}" \
	-v app_password="$APP_PASSWORD" \
	-f /etc/patroni/bootstrap.sql
