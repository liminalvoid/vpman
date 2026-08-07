package repository

import (
	"context"
	"fmt"

	"vpnmanager/internal/model"
)

type Plan struct {
	conn conn
}

func NewPlan(conn conn) *Plan { return &Plan{conn: conn} }

func (r *Plan) Create(ctx context.Context, plan *model.Plan) (*model.Plan, error) {
	const q = `
		INSERT INTO plans (name, unit_amount_minor, currency, interval, interval_count)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, public_id, name, unit_amount_minor, currency, interval, interval_count, is_active`

	var result model.Plan
	err := r.conn.
		QueryRow(ctx, q, plan.Name, plan.UnitAmountMinor, plan.Currency, plan.Interval, plan.IntervalCount).
		Scan(
			&result.ID,
			&result.PublicID,
			&result.Name,
			&result.UnitAmountMinor,
			&result.Currency,
			&result.Interval,
			&result.IntervalCount,
			&result.IsActive,
		)
	if err != nil {
		return nil, fmt.Errorf("creating new plan %q in DB: %w", result.PublicID.String(), err)
	}

	return &result, nil
}
