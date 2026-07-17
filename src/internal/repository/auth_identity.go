package repository

import (
	"context"

	"vpnmanager/internal/model"
)

type AuthIdentity struct {
	conn conn
}

func NewAuthIdentity(conn conn) *AuthIdentity { return &AuthIdentity{conn: conn} }

func (r *AuthIdentity) Create(ctx context.Context, authIdentity *model.AuthIdentity) error {
	const q = `
		INSERT INTO auth_identities (user_id, provider_id, provider_user_id)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, provider_id, provider_user_id`

	return r.conn.
		QueryRow(ctx, q, authIdentity.UserID, authIdentity.ProviderID, authIdentity.ProviderUserID).
		Scan(&authIdentity.ID, &authIdentity.UserID, &authIdentity.ProviderID, &authIdentity.ProviderUserID)
}
