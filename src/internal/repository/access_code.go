package repository

import (
	"context"

	"vpnmanager/internal/model"
)

type AccessCode struct {
	conn conn
}

func NewAccessCode(conn conn) *AccessCode { return &AccessCode{conn: conn} }

func (r *AccessCode) Create(ctx context.Context, accessCode *model.AccessCode) error {
	const q = `
		INSERT INTO access_codes (code)
		VALUES ($1)
		RETURNING id, code, created_at, used_at`

	return r.conn.
		QueryRow(ctx, q, accessCode.Code).
		Scan(&accessCode.ID, &accessCode.Code, &accessCode.CreatedAt, &accessCode.UsedAt)
}
