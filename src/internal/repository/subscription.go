package repository

import (
	"context"
	"fmt"
	"time"

	"vpnmanager/internal/model"

	"github.com/google/uuid"
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

func (r *Subscription) GetByTelegramID(
	ctx context.Context,
	telegramID int64,
) (*model.Subscription, error) {
	const q = `
		SELECT s.id, s.user_id, s.status, s.current_period_end, s.canceled_at, s.created_at, s.updated_at, s.quantity, s.payment_claimed_at
		FROM subscriptions s
		JOIN users u ON s.user_id = u.id
		WHERE u.telegram_id = $1`

	var result model.Subscription
	err := r.conn.
		QueryRow(ctx, q, telegramID).
		Scan(
			&result.ID,
			&result.UserID,
			&result.Status,
			&result.CurrentPeriodEnd,
			&result.CanceledAt,
			&result.CreatedAt,
			&result.UpdatedAt,
			&result.Quantity,
			&result.PaymentClaimedAt,
		)
	if err != nil {
		return nil, fmt.Errorf("retrieving subscription by Telegram ID: %w", err)
	}

	return &result, nil
}

func (r *Subscription) GetByPublicID(
	ctx context.Context,
	publicID uuid.UUID,
) (*model.Subscription, error) {
	const q = `
		SELECT s.id, s.user_id, s.status, s.current_period_end, s.canceled_at, s.created_at, s.updated_at, s.quantity, s.payment_claimed_at
		FROM subscriptions s
		JOIN users u ON s.user_id = u.id
		WHERE u.public_id = $1`

	var result model.Subscription
	err := r.conn.
		QueryRow(ctx, q, publicID).
		Scan(
			&result.ID,
			&result.UserID,
			&result.Status,
			&result.CurrentPeriodEnd,
			&result.CanceledAt,
			&result.CreatedAt,
			&result.UpdatedAt,
			&result.Quantity,
			&result.PaymentClaimedAt,
		)
	if err != nil {
		return nil, fmt.Errorf("retrieving subscription by public ID: %w", err)
	}

	return &result, nil
}

func (r *Subscription) IsLimitSurpassedByTelegramID(
	ctx context.Context,
	telegramID int64,
) (*bool, error) {
	const q = `
		SELECT (
			SELECT COUNT(c.id) FROM configs c WHERE c.subscription_id = s.id AND c.status = $2
		) >= s.quantity AS is_over_limit
		FROM subscriptions s
		JOIN users u ON s.user_id = u.id
		WHERE u.telegram_id = $1`

	var isOverLimit bool
	err := r.conn.QueryRow(ctx, q, telegramID, model.ConfigStatusActive).Scan(&isOverLimit)
	if err != nil {
		return nil, fmt.Errorf("calculating the limit overrun: %w", err)
	}

	return &isOverLimit, nil
}

func (r *Subscription) ChangeKeysLimitCount(
	ctx context.Context,
	count int,
	userID int64,
) (int64, error) {
	const q = `
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
		WHERE status = $2`

	tag, err := r.conn.Exec(
		ctx,
		q,
		timestamp,
		model.SubscriptionStatusActive,
	)
	if err != nil {
		return 0, fmt.Errorf("setting end period: %w", err)
	}

	return tag.RowsAffected(), nil
}

func (r *Subscription) SetPastDueAndSuspendConfigs(
	ctx context.Context,
) ([]model.ConfigListView, error) {
	const q = `
		WITH past_due AS (
			UPDATE subscriptions
			SET status = $1, updated_at = now()
			WHERE now() > current_period_end AND status = $2
			RETURNING id
		)
		UPDATE configs c
		SET status = $3
		FROM past_due pd, protocols p, servers s
		WHERE c.subscription_id = pd.id
		AND c.protocol_id = p.id
		AND c.server_id = s.id
		AND c.status = $4
		RETURNING c.id, c.public_id, c.name, c.ip_address, p.name, s.location, s.public_id, c.status, c.created_at`

	rows, err := r.conn.Query(
		ctx,
		q,
		model.SubscriptionStatusPastDue,
		model.SubscriptionStatusActive,
		model.ConfigStatusSuspended,
		model.ConfigStatusActive,
	)
	if err != nil {
		return nil, fmt.Errorf("setting past due status: %w", err)
	}

	var configs []model.ConfigListView
	for rows.Next() {
		var c model.ConfigListView
		if err := rows.Scan(
			&c.ID,
			&c.PublicID,
			&c.Name,
			&c.IPAddress,
			&c.Protocol,
			&c.Location,
			&c.ServerPublicID,
			&c.Status,
			&c.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning config list view: %w", err)
		}

		configs = append(configs, c)
	}

	return configs, nil
}

func (r *Subscription) ClaimPayment(
	ctx context.Context,
	telegramID int64,
) error {
	const q = `
		UPDATE subscriptions s
		SET status = $1, updated_at = now(), payment_claimed_at = now()
		FROM users u
		WHERE s.user_id = u.id AND u.telegram_id = $2`

	_, err := r.conn.Exec(ctx, q, model.SubscriptionStatusPending, telegramID)
	if err != nil {
		return fmt.Errorf("claiming payment: %w", err)
	}

	return nil
}

func (r *Subscription) DeleteStale(
	ctx context.Context,
	period time.Duration,
) (int64, error) {
	const q = `
		DELETE FROM subscriptions
		WHERE now() > current_period_end + $1`

	tag, err := r.conn.Exec(ctx, q, period)
	if err != nil {
		return 0, fmt.Errorf("retrieving configs from stale subscriptions: %w", err)
	}

	return tag.RowsAffected(), nil
}
