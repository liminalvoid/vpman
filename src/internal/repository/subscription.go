package repository

import (
	"context"
	"fmt"

	"vpnmanager/internal/model"
)

type Subscription struct {
	conn conn
}

func NewSubscription(conn conn) *Subscription { return &Subscription{conn} }

func (r *Subscription) Create(ctx context.Context, subscription *model.Subscription) (*model.Subscription, error) {
	const q = `
		INSERT INTO subscriptions (user_id, status, current_period_start, current_period_end, cancel_at_period_end, plan_id, quantity, unit_amount_minor)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, status, current_period_start, current_period_end, cancel_at_period_end, canceled_at, created_at, updated_at, max_configs`

	var result model.Subscription
	err := r.conn.
		QueryRow(
			ctx,
			q,
			subscription.UserID,
			subscription.Status,
			subscription.CurrentPeriodStart,
			subscription.CurrentPeriodEnd,
			subscription.CancelAtPeriodEnd,
			subscription.PlanID,
			subscription.Quantity,
			subscription.UnitAmountMinor,
		).
		Scan(
			&result.ID,
			&result.UserID,
			&result.Status,
			&result.CurrentPeriodStart,
			&result.CurrentPeriodEnd,
			&result.CancelAtPeriodEnd,
			&result.CanceledAt,
			&result.CreatedAt,
			&result.UpdatedAt,
			&result.PlanID,
			&result.Quantity,
			&result.UnitAmountMinor,
		)
	if err != nil {
		return nil, fmt.Errorf("creating subscription %d: %w", subscription.ID, err)
	}

	return &result, nil
}
