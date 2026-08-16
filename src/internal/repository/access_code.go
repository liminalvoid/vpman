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

func (r *AccessCode) Delete(ctx context.Context, code string) (int64, error) {
	const q = `
		DELETE FROM access_codes
		WHERE code = $1`

	tag, err := r.conn.Exec(ctx, q, code)
	if err != nil {
		return 0, fmt.Errorf("marking code as used: %w", err)
	}
	
	return tag.RowsAffected(), nil
}
