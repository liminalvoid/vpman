package scheduler

import (
	"context"
	"log"
	"time"

	"vpnmanager/internal/service"
)

type Scheduler struct {
	subscriptionService *service.SubscriptionService
	configService       *service.ConfigService
	interval            time.Duration
	batch               int
	endPeriod           time.Duration
}

func New(
	subscriptionService *service.SubscriptionService,
	configService *service.ConfigService,
	interval time.Duration,
	batch int,
	endPeriod time.Duration,
) *Scheduler {
	return &Scheduler{
		subscriptionService,
		configService,
		interval,
		batch,
		endPeriod,
	}
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

	for name, sweep := range map[string](func(context.Context, time.Duration, int) (int64, error)){
		"past_due": s.subscriptionService.SetPastDue,
		"remove_stale": s.configService.RemoveStale,
	} {
		affectedEntities, err := sweep(ctx, s.endPeriod, s.batch)
		if err != nil {
			log.Printf("sweep: %s", err)
		}

		log.Printf("successful sweep: %s, affected %d entities", name, affectedEntities)
	}
}
