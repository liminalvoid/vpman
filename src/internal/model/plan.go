package model

import (
	"github.com/google/uuid"
)

type PlanInterval string

const (
	PlanIntervalDay   PlanInterval = "day"
	PlanIntervalWeek  PlanInterval = "week"
	PlanIntervalMonth PlanInterval = "month"
	PlanIntervalYear  PlanInterval = "year"
)

type Plan struct {
	ID              int32
	PublicID        uuid.UUID
	Name            string
	UnitAmountMinor int32
	Currency        string
	Interval        PlanInterval
	IntervalCount 	int32
	IsActive        bool
}
