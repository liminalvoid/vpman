package repository

import (
	"context"
	"fmt"

	"vpnmanager/internal/model"
)

type AccessCode struct {
	conn conn
}

func NewAccessCode(conn conn) *AccessCode { return &AccessCode{conn: conn} }

func (r *AccessCode) Create(ctx context.Context, code string) (*model.AccessCode, error) {
	const q = `
		INSERT INTO access_codes (code)
		VALUES ($1)
		RETURNING id, code, created_at, used_at`

	var result model.AccessCode
	err := r.conn.
		QueryRow(ctx, q, code).
		Scan(&result.ID, &result.Code, &result.CreatedAt, &result.UsedAt)
	if err != nil {
		return nil, fmt.Errorf("creating access code: %w", err)
	}

	return &result, nil
}

func (r *AccessCode) MarkUsed(ctx context.Context, code string) (*model.AccessCode, error) {
	const q = `
		UPDATE access_codes
		SET used_at = CURRENT_TIMESTAMP
		WHERE code = $1 AND used_at IS NULL
		RETURNING id, code, created_at, used_at`

	var ac model.AccessCode
	err := r.conn.
		QueryRow(ctx, q, code).
		Scan(&ac.ID, &ac.Code, &ac.CreatedAt, &ac.UsedAt)
	if err != nil {
		return nil, fmt.Errorf("marking access code used: %w", err)
	}
	
	return &ac, nil
}
