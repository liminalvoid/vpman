package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Username     string
	Password     string
	HostAddr     string
	HostPort     uint16
	DatabaseName string
}

type DB struct {
	pool *pgxpool.Pool
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
	dbpool, err := pgxpool.New(ctx, connectionStr)
	if err != nil {
		return nil, fmt.Errorf("init DB: %w", err)
	}

	return &DB{pool: dbpool}, nil
}

func (d *DB) Close() { d.pool.Close() }

func (d *DB) Pool() *pgxpool.Pool { return d.pool }
