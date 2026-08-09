package repository

import (
	"context"
	"fmt"

	"vpnmanager/internal/model"
)

type User struct {
	conn conn
}

func NewUser(conn conn) *User { return &User{conn: conn} }

func (r *User) Create(ctx context.Context, user *model.User) error {
	const q = `
		INSERT INTO users (username, telegram_id, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, public_id, username`

	return r.conn.
		QueryRow(ctx, q, user.Username, user.TelegramID, user.PasswordHash).
		Scan(&user.ID, &user.PublicID, &user.Username)
}

func (r *User) FindByTelegramID(ctx context.Context, telegramID int64) (*model.User, error) {
	const q = `
		SELECT id, public_id, telegram_id, is_admin
		FROM users WHERE telegram_id = $1`
	
	var user model.User
	err := r.conn.	
		QueryRow(ctx, q, user.TelegramID).
		Scan(&user.ID, &user.PublicID, &user.TelegramID, &user.IsAdmin)
	if err != nil {
		return nil, fmt.Errorf("looking for user: %w", err)
	}

	return &user, nil
}
