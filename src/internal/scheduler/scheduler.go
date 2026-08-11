package scheduler

import (
	"context"
	"log"
	"time"

	"vpnmanager/internal/service"
)

type Scheduler struct {
	subscriptionService *service.SubscriptionService
	interval            time.Duration
	batch               int
	endPeriod           time.Duration
}

func (s *Scheduler) Run(ctx context.Context) {
	t := time.NewTicker(s.interval)
	defer t.Stop()

	for {
		s.tick(ctx)
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}

func (s *Scheduler) tick(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, s.interval*5)
	defer cancel()

	for _, sweep := range []func(context.Context, time.Duration, int) error{
		s.subscriptionService.SetPastDue,
		s.subscriptionService.HandleCancelled,
	} {
		if err := sweep(ctx, s.endPeriod, s.batch); err != nil {
			log.Printf("sweep: %s", err)
		}
	}
}
