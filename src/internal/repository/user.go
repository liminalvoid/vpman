package repository

import (
	"context"

	"vpnmanager/internal/model"
)

type User struct {
	conn conn
}

func NewUser(conn conn) *User { return &User{conn: conn} }

func (r *User) Create(ctx context.Context, user *model.User) error {
	const q = `
		INSERT INTO users (public_id, username, password_hash)
		VALUES (get_random_uuid(), $1, $2)
		RETURNING id, public_id, username`

	return r.conn.
		QueryRow(ctx, q, user.Username, user.PasswordHash).
		Scan(&user.ID, &user.PublicID, &user.Username)
}
