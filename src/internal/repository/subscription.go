package repository

import (
	"context"

	"vpnmanager/internal/model"
)

type Subscription struct {
	conn conn
}

func NewSubscription(conn conn) *Subscription { return &Subscription{conn} }

func (r *Subscription) Create(ctx context.Context, subscription *model.Subscription) error {
	const q = `
		INSERT INTO subscriptions (user_id, status, current_period_start, current_period_end, cancel_at_period_end, max_configs)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, status, current_period_start, current_period_end, cancel_at_period_end, canceled_at, created_at, updated_at, max_configs`

	return r.conn.
		QueryRow(
			ctx,
			q,
			subscription.UserID,
			subscription.Status,
			subscription.CurrentPeriodStart,
			subscription.CurrentPeriodEnd,
			subscription.CancelAtPeriodEnd,
			subscription.MaxConfigs,
		).
		Scan(
			&subscription.ID,
			&subscription.UserID,
			&subscription.Status,
			&subscription.CurrentPeriodStart,
			&subscription.CurrentPeriodEnd,
			&subscription.CancelAtPeriodEnd,
			&subscription.CanceledAt,
			&subscription.CreatedAt,
			&subscription.UpdatedAt,
			&subscription.MaxConfigs,
		)
}
