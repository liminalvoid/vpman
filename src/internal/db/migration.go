package db

import (
	"embed"
	"fmt"
	"io/fs"
	"context"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
)

//go:embed migrations/*.sql
var migrations embed.FS

func (d *DB) Migrate(ctx context.Context) error {
	fsys, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return err
	}
	locker, err := lock.NewPostgresSessionLocker()
	if err != nil {
		return err
	}
	p, err := goose.NewProvider(
		goose.DialectPostgres,
		stdlib.OpenDBFromPool(d.pool),
		fsys,
		goose.WithSessionLocker(locker),
	)
	if err != nil {
		return fmt.Errorf("migrations: %w", err)
	}
	_, err = p.Up(ctx)

	return err
}
