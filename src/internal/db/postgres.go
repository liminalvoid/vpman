package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type Config struct {
	Username     string
	Password     string
	HostAddr     string
	HostPort     uint16
	DatabaseName string
}

type DB struct {
	conn *pgx.Conn
}

func New(ctx context.Context, cfg *Config) (*DB, error) {
	connectionStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		cfg.Username,
		cfg.Password,
		cfg.HostAddr,
		cfg.HostPort,
		cfg.DatabaseName,
	)
	conn, err := pgx.Connect(ctx, connectionStr)
	if err != nil {
		return nil, fmt.Errorf("init DB: %w", err)
	}

	return &DB{conn: conn}, nil
}

func (d *DB) Close(ctx context.Context) { d.conn.Close(ctx) }
