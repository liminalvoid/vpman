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

	for name, sweep := range map[string](func(context.Context, time.Duration, int) (int64, error)){
		"past_due":     s.subscriptionService.SetPastDue,
		"remove_stale": s.configService.RemoveStale,
	} {
		affectedEntities, err := sweep(ctx, s.endPeriod, s.batch)
		if err != nil {
			s.log.ErrorContext(ctx, "sweep failed", "err", err)
		}

		s.log.DebugContext(
			ctx,
			"successful sweep",
			"func_name", name,
			"affected_entities", affectedEntities,
		)
	}
}
