-- +goose Up
-- +goose ENVSUB ON
INSERT INTO users (telegram_id, is_admin)
VALUES (${ADMIN_TELEGRAM_ID?ADMIN_TELEGRAM_ID must be set}, true)
ON CONFLICT (telegram_id) DO UPDATE SET is_admin = true;
-- +goose ENVSUB OFF
