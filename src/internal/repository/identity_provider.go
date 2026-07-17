package repository

import (
	"context"

	"vpnmanager/internal/model"
)

type IdentityProvider struct {
	conn conn
}

func NewIdentityProvider(conn conn) *IdentityProvider { return &IdentityProvider{conn: conn} }

func (r *IdentityProvider) Create(ctx context.Context, identityProvider *model.IdentityProvider) error {
	const q = `
		INSERT INTO identity_providers (name)
		VALUES ($1)
		RETURNING id, name`

	return r.conn.
		QueryRow(ctx, q, identityProvider.Name).
		Scan(&identityProvider.ID, &identityProvider.Name)
}
