package model

import "time"

type Subscription struct {
	ID                 int64
	UserID             int64
	PlanID             int32
	Status             string
	CurrentPeriodStart time.Time
	CurrentPeriodEnd   time.Time
	CancelAtPeriodEnd  bool
	CanceledAt         time.Time
	TrialEnd           time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
