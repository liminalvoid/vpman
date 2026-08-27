package model

import "time"

type SubscriptionStatus string

const (
	SubscriptionStatusActive  SubscriptionStatus = "active"
	SubscriptionStatusPending SubscriptionStatus = "pending"
	SubscriptionStatusPastDue SubscriptionStatus = "past_due"
)

type Subscription struct {
	ID               int64
	UserID           int64
	Status           SubscriptionStatus
	CurrentPeriodEnd *time.Time
	CanceledAt       *time.Time
	CreatedAt        time.Time
	UpdatedAt        *time.Time
	Quantity         int32
}
