package repository

import (
	"context"
	"fmt"
	"time"

	"vpnmanager/internal/model"
)

type Config struct {
	conn conn
}

func NewConfig(conn conn) *Config { return &Config{conn: conn} }

func (r *Config) Create(ctx context.Context, config *model.Config) error {
	const q = `
		INSERT INTO configs (name, ip_address, subscription_id, protocol_id, server_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, public_id, status, created_at`

	return r.conn.
		QueryRow(ctx, q, config.Name, config.IPAddress, config.SubscriptionID, config.ProtocolID, config.ServerID).
		Scan(&config.ID, &config.PublicID, &config.Status, &config.CreatedAt)
}

func (r *Config) GetByPublicID(
	ctx context.Context,
	publicID string,
) (*model.ConfigListView, error) {
	const q = `
		SELECT c.id, c.public_id, c.name, c.ip_address, p.name, s.location, s.public_id, c.status, c.created_at
		FROM configs c
		JOIN protocols p ON c.protocol_id = p.id
		JOIN servers s ON c.server_id = s.id
		WHERE c.public_id = $1`

	var resultConfig model.ConfigListView
	err := r.conn.
		QueryRow(ctx, q, publicID).
		Scan(
			&resultConfig.ID,
			&resultConfig.PublicID,
			&resultConfig.Name,
			&resultConfig.IPAddress,
			&resultConfig.Protocol,
			&resultConfig.Location,
			&resultConfig.ServerPublicID,
			&resultConfig.Status,
			&resultConfig.CreatedAt,
		)
	if err != nil {
		return nil, fmt.Errorf("retrieving config by public ID %q: %w", publicID, err)
	}

	return &resultConfig, nil
}

func (r *Config) ListByTelegramID(ctx context.Context, userID int64) ([]model.ConfigListView, error) {
	const q = `
		SELECT c.id, c.public_id, c.name, c.ip_address, p.name, s.location, c.status, c.created_at
		FROM configs c
		JOIN subscriptions sb ON c.subscription_id = sb.id
		JOIN protocols p ON c.protocol_id = p.id
		JOIN servers s ON c.server_id = s.id
		JOIN users u ON sb.user_id = u.id
		WHERE u.telegram_id = $1`

	rows, err := r.conn.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("listing configs by Telegram ID %d: %w", userID, err)
	}
	defer rows.Close()

	var out []model.ConfigListView
	for rows.Next() {
		var c model.ConfigListView
		if err := rows.Scan(
			&c.ID,
			&c.PublicID,
			&c.Name,
			&c.IPAddress,
			&c.Protocol,
			&c.Location,
			&c.Status,
			&c.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning config list view: %w", err)
		}

		out = append(out, c)
	}

	return out, rows.Err()
}

func (r *Config) ListStaleConfigs(
	ctx context.Context,
	period time.Duration,
	batch int,
) ([]model.ConfigListView, error) {
	const q = `
		SELECT c.id, c.public_id, c.name, c.ip_address, p.name, s.location, s.public_id, c.status, c.created_at
		FROM configs c
		JOIN protocols p ON c.protocol_id = p.id
		JOIN servers s ON c.server_id = s.id
		JOIN subscriptions sb ON c.subscription_id = sb.id
		WHERE now() > sb.current_period_end + $1 AND c.status = $3
		ORDER BY c.id
		LIMIT $2`
	
	rows, err := r.conn.Query(ctx, q, period, batch, model.ConfigStatusSuspended)
	if err != nil {
		return nil, fmt.Errorf("retrieving configs from stale subscriptions: %w", err)
	}
	defer rows.Close()

	var out []model.ConfigListView
	for rows.Next() {
		var c model.ConfigListView
		if err := rows.Scan(
			&c.ID,
			&c.PublicID,
			&c.Name,
			&c.IPAddress,
			&c.Protocol,
			&c.Location,
			&c.ServerPublicID,
			&c.Status,
			&c.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning config list view: %w", err)
		}

		out = append(out, c)
	}

	return out, rows.Err()
}

func (r *Config) UpdateName(
	ctx context.Context,
	publicID, name string,
) (int64, error) {
	const q = `
		UPDATE configs
		SET name = $1
		WHERE public_id = $2`

	tag, err := r.conn.Exec(ctx, q, name, publicID)
	if err != nil {
		return -1, fmt.Errorf("updating config %q: %w", publicID, err)
	}

	return tag.RowsAffected(), nil
}

func (r *Config) Delete(ctx context.Context, publicID string) error {
	const q = `DELETE FROM configs WHERE public_id = $1`

	_, err := r.conn.Exec(ctx, q, publicID)
	if err != nil {
		return fmt.Errorf("deleting config %q: %w", publicID, err)
	}

	return nil
}

func (r *Config) MarkSuspended(ctx context.Context, publicID string) (int64, error) {
	const q = `
		UPDATE configs
		SET status = $2
		WHERE public_id = $1`

	tag, err := r.conn.Exec(ctx, q, publicID, model.ConfigStatusSuspended)
	if err != nil {
		return -1, fmt.Errorf("marking config %q as suspended: %w", publicID, err)
	}

	return tag.RowsAffected(), nil
}

func (r *Config) MarkSuspendedNLatest(
	ctx context.Context,
	subscriptionID int64,
	count int,
) (int64, error) {
	const q = `
		UPDATE configs
		SET status = $1
		WHERE id IN (
			SELECT id
			FROM configs
			WHERE subscription_id = $2
			ORDER BY created_at DESC
			LIMIT $3
		)`
	
	tag, err := r.conn.Exec(ctx, q, model.ConfigStatusSuspended, subscriptionID, count)
	if err != nil {
		return 0, fmt.Errorf("marking %d latest configs as suspended: %w", count, err)
	}

	return tag.RowsAffected(), nil
}
