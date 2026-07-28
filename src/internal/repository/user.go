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
		INSERT INTO users (public_id, username, telegram_id, password_hash)
		VALUES (get_random_uuid(), $1, $2, $3)
		RETURNING id, public_id, username`

	return r.conn.
		QueryRow(ctx, q, user.Username, user.TelegramID, user.PasswordHash).
		Scan(&user.ID, &user.PublicID, &user.Username, &user.TelegramID)
}

func (r *User) FindByTelegramID(ctx context.Context, user *model.User) error {
	const q = `
		SELECT public_id, username, telegram_id, is_admin
		FROM users WHERE telegram_id = $1`
	
	return r.conn.	
		QueryRow(ctx, q, user.TelegramID).
		Scan(&user.PublicID, &user.Username, &user.TelegramID, &user.IsAdmin)
}
