-- Creates the application role and database, then loads the schema as that
-- role so it owns its own objects. Every step is guarded: Patroni only runs
-- post_init once, at bootstrap, but a retry after a half-finished bootstrap
-- should not die on "already exists".
-- CREATE DATABASE cannot run inside a transaction block, hence no
-- --single-transaction on the psql call.

SELECT format('CREATE ROLE %I LOGIN PASSWORD %L', :'app_user', :'app_password')
WHERE NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'app_user')
\gexec

SELECT format('CREATE DATABASE %I OWNER %I', :'app_db', :'app_user')
WHERE NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = :'app_db')
\gexec

\c :"app_db"

SET ROLE :"app_user";

SELECT NOT EXISTS (
	SELECT 1 FROM pg_class
	WHERE relname = 'protocols' AND relnamespace = 'public'::regnamespace
) AS need_schema \gset

\if :need_schema
\i /etc/patroni/schema.sql
\else
\echo 'schema already present, skipping'
\endif
