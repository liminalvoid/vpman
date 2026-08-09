package repository

import (
	"context"
	"fmt"
	"time"

	"vpnmanager/internal/model"
)

type Subscription struct {
	conn conn
}

func NewSubscription(conn conn) *Subscription { return &Subscription{conn} }

func (r *Subscription) Create(ctx context.Context, userID int64, quantity int) (*model.Subscription, error) {
	const q = `
		INSERT INTO subscriptions (
			user_id,
			quantity
		)
		VALUES ($1, $2)
		RETURNING
			id,
			user_id,
			status,
			current_period_end,
			canceled_at,
			created_at,
			updated_at,
			quantity`

	var result model.Subscription
	err := r.conn.
		QueryRow(
			ctx,
			q,
			userID,
			quantity,
		).
		Scan(
			&result.ID,
			&result.UserID,
			&result.Status,
			&result.CurrentPeriodEnd,
			&result.CanceledAt,
			&result.CreatedAt,
			&result.UpdatedAt,
			&result.Quantity,
		)
	if err != nil {
		return nil, fmt.Errorf("creating subscription for user %d: %w", userID, err)
	}

	return &result, nil
}

func (r *Subscription) ChangeKeysLimitCount(
	ctx context.Context,
	count int,
	userID int64,
) (int64, error) {
	const q =`
		UPDATE subscriptions
		SET quantity = $1
		WHERE user_id = $2`

	tag, err := r.conn.Exec(ctx, q, count, userID)
	if err != nil {
		return 0, fmt.Errorf("changing keys limit value: %w", err)
	}

	return tag.RowsAffected(), nil
}

func (r *Subscription) SetEndPeriodForAll(
	ctx context.Context,
	timestamp time.Time,
) (int64, error) {
	const q = `
		UPDATE subscriptions
		SET current_period_end = $1, updated_at = now()
		WHERE status = 'active'`
	
	tag, err := r.conn.Exec(ctx, q, timestamp)
	if err != nil {
		return 0, fmt.Errorf("setting end period: %w", err)
	}

	return tag.RowsAffected(), nil
}
