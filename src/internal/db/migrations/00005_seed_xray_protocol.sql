-- +goose Up
INSERT INTO protocols (name) VALUES ('xray')
ON CONFLICT (name) DO NOTHING;

-- +goose Down
DELETE FROM protocols WHERE name = 'xray';
