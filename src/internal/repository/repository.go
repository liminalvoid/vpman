package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type conn interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type Repositories struct {
	AccessCode   *AccessCode
	Config       *Config
	Protocol     *Protocol
	Server       *Server
	Subscription *Subscription
	User         *User
}

func New(c conn) *Repositories {
	return &Repositories{
		AccessCode:   NewAccessCode(c),
		Config:       NewConfig(c),
		Protocol:     NewProtocol(c),
		Server:       NewServer(c),
		Subscription: NewSubscription(c),
		User:         NewUser(c),
	}
}
