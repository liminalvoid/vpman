-- +goose UP
INSERT INTO protocols (name) VALUES ('awg'), ('wg')
ON CONFLICT (name) DO NOTHING;

-- +goose Down
DELETE FROM protocols WHERE name IN ('awg', 'wg');
