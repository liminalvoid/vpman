package scheduler

import (
	"context"
	"log/slog"
	"time"

	"vpnmanager/internal/service"
)

type Scheduler struct {
	log                 *slog.Logger
	subscriptionService *service.SubscriptionService
	configService       *service.ConfigService
	interval            time.Duration
	batch               int
	endPeriod           time.Duration
}

func New(
	log *slog.Logger,
	subscriptionService *service.SubscriptionService,
	configService *service.ConfigService,
	interval time.Duration,
	batch int,
	endPeriod time.Duration,
) *Scheduler {
	return &Scheduler{
		log,
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
	ctx, cancel := context.WithTimeout(ctx, s.interval*5)
	defer cancel()

	// Order matters: configs are suspended before stale ones are removed.
	sweeps := []struct {
		name string
		fn   func(context.Context, time.Duration, int) (int64, error)
	}{
		{"past_due", s.subscriptionService.SetPastDue},
		{"remove_stale", s.configService.RemoveStale},
	}

	for _, sweep := range sweeps {
		affectedEntities, err := sweep.fn(ctx, s.endPeriod, s.batch)
		if err != nil {
			s.log.ErrorContext(ctx, "sweep failed", "func_name", sweep.name, "err", err)
			continue
		}

		s.log.DebugContext(
			ctx,
			"successful sweep",
			"func_name", sweep.name,
			"affected_entities", affectedEntities,
		)
	}
}
