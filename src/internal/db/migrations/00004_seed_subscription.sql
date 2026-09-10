-- +goose Up
-- +goose ENVSUB ON
INSERT INTO subscriptions (user_id, quantity) VALUES (1, 5)
SELECT id, 5
FROM users
WHERE telegram_id = ${ADMIN_TELEGRAM_ID:?ADMIN_TELEGRAM_ID must be set}
ON CONFLICT (user_id) DO NOTHING;
-- +goose ENVSUB OFF
