package repository

import (
	"context"

	"vpnmanager/internal/model"
)

type Protocol struct {
	conn conn
}

func NewProtocol(conn conn) *Protocol { return &Protocol{conn: conn} }

func (r *Protocol) Create(ctx context.Context, protocol *model.Protocol) error {
	const q = `
		INSERT INTO protocols (name)
		VALUES ($1)
		RETURNING id, name`

	return r.conn.
		QueryRow(ctx, q, protocol.Name).
		Scan(&protocol.ID, &protocol.Name)
}

func (r *Protocol) Lookup(ctx context.Context, protocol *model.Protocol) error {
	const q = `
		SELECT id, name
		FROM protocols WHERE name = $1`

	return r.conn.
		QueryRow(ctx, q, protocol.Name).
		Scan(&protocol.ID, &protocol.Name)
}
