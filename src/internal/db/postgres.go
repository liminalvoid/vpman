package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	pool *pgxpool.Pool
}

// New opens a pool against the given connection string. The string may list
// several hosts (pg1:5432,pg2:5432,...) together with
// target_session_attrs=read-write, in which case pgx tries each host on every
// new connection and keeps only the one accepting writes. That is what lets
// the bot follow a Patroni failover without a proxy in front of the cluster.
func New(ctx context.Context, dsn string) (*DB, error) {
	dbpool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("init DB: %w", err)
	}

	if err := dbpool.Ping(ctx); err != nil {
		dbpool.Close()

		return nil, fmt.Errorf("ping DB: %w", err)
	}

	return &DB{pool: dbpool}, nil
}

func (d *DB) Close() { d.pool.Close() }

func (d *DB) Pool() *pgxpool.Pool { return d.pool }
