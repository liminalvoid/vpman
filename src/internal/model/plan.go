package model

type Plan struct {
	ID            int32
	Code          string
	Name          string
	AmountCents   int32
	Currency      string
	Interval      string
	IntervalCount int32
	IsActive      bool
}
