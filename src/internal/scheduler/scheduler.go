package scheduler

import (
	"context"
	"log"
	"time"

	"vpnmanager/internal/service"
)

type Scheduler struct {
	subs     *service.SubscriptionService
	interval time.Duration
	batch    int
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
	ctx, cancel := context.WithTimeout(ctx, s.interval * 5)
	defer cancel()

	for _, sweep := range []func(context.Context, int) error{
		s.subs.ExpireDueSubscriptions,
		s.subs.RetryDunning,
		s.subs.SuspendExpiredGrace,
	} {
		if err := sweep(ctx, s.batch); err != nil {
			log.Printf("sweep: %s", err)
		}
	}
}
