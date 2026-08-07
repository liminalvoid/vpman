package model

import "time"

type Subscription struct {
	ID                 int64
	UserID             int64
	Status             string
	CurrentPeriodStart time.Time
	CurrentPeriodEnd   time.Time
	CancelAtPeriodEnd  bool
	CanceledAt         time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
	PlanID             int32
	Quantity           int32
	UnitAmountMinor    int32
}
