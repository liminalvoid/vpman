package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"vpnmanager/internal/model"
)

type conn interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type Config struct {
	conn conn
}

func NewConfig(conn conn) *Config { return &Config{conn: conn} }

func (r *Config) Create(ctx context.Context, config *model.Config) error {
	const q = `
		INSERT INTO configs (public_id, name, user_id, protocol_id, location)
		VALUES (gen_random_uuid(), $1, $2, $3, $4)
		RETURNING id, public_id, status, created_at`

	return r.conn.
		QueryRow(ctx, q, config.Name, config.UserID, config.ProtocolID, config.Location).
		Scan(&config.ID, &config.PublicID, &config.Status, &config.CreatedAt)
}

func (r *Config) ListByUserID(ctx context.Context, userID int64) ([]model.Config, error) {
	const q = `
		SELECT id, public_id, name, user_id, protocol_id, status, location, created_at
		FROM configs WHERE user_id = $1`

	rows, err := r.conn.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("querying configs: %w", err)
	}
	defer rows.Close()

	var out []model.Config
	for rows.Next() {
		var c model.Config
		if err := rows.Scan(
			&c.ID,
			&c.PublicID,
			&c.Name,
			&c.UserID,
			&c.ProtocolID,
			&c.Status,
			&c.Location,
			&c.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning config: %w", err)
		}

		out = append(out, c)
	}

	return out, rows.Err()
}
