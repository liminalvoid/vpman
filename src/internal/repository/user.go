package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"vpnmanager/internal/model"
)

type User struct {
	conn conn
}

var ErrUserNotExist = errors.New("user doesn't exist")

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

var ErrNoAccessCode = errors.New("code doesn't exist")

func (r *User) CreateWithAccessCode(
	ctx context.Context,
	telegramID int64,
	code string,
) error {
	const q = `
		WITH redeemed AS (
			DELETE FROM access_codes WHERE code = $1
			RETURNING id
		)
		INSERT INTO users (telegram_id)
		SELECT $2 FROM redeemed
		RETURNING id, public_id, telegram_id`

	var user model.User
	err := r.conn.
		QueryRow(ctx, q, code, telegramID).
		Scan(&user.ID, &user.PublicID, &user.TelegramID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNoAccessCode
	}

	return err
}

func (r *User) FindByTelegramID(ctx context.Context, telegramID int64) (*model.User, error) {
	const q = `
		SELECT id, public_id, telegram_id, is_admin
		FROM users WHERE telegram_id = $1`
	
	var user model.User
	err := r.conn.	
		QueryRow(ctx, q, telegramID).
		Scan(&user.ID, &user.PublicID, &user.TelegramID, &user.IsAdmin)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotExist
	} else if err != nil {
		return nil, fmt.Errorf("looking for user: %w", err)
	}

	return &user, nil
}
