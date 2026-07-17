package repository

import (
	"context"

	"vpnmanager/internal/model"
)

type Plan struct {
	conn conn
}

func NewPlan(conn conn) *Plan { return &Plan{conn} }

func (r *Plan) Create(ctx context.Context, plan *model.Plan) error {
	const q = `
		INSERT INTO plans (code, name, amount_cents, currency, interval, interval_count, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, code, name, amount_cents, currency, interval, interval_count, is_active`
	
	return r.conn.
		QueryRow(
			ctx,
			q,
			plan.Code,
			plan.Name,
			plan.AmountCents,
			plan.Currency,
			plan.Interval,
			plan.IntervalCount,
			plan.IsActive,
		).
		Scan(
			&plan.ID,
			&plan.Code,
			&plan.Name,
			&plan.AmountCents,
			&plan.Currency,
			&plan.Interval,
			&plan.IntervalCount,
			&plan.IsActive,
		)
}
