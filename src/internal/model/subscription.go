package model

import "time"

type Subscription struct {
	ID                 int64
	UserID             int64
	Status             string
	CurrentPeriodEnd   time.Time
	CanceledAt         time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
	Quantity           int32
}
