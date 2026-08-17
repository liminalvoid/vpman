package service

import (
	"context"
	"fmt"
	"time"

	"vpnmanager/internal/configgen"
	"vpnmanager/internal/executor"
	"vpnmanager/internal/repository"
	"vpnmanager/internal/tgbot/newsletter"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

type SubscriptionService struct {
	containerRegistry      *executor.Registry
	newsletter             *newsletter.Newsletter
	subscriptionRepository *repository.Subscription
	userRepository         *repository.User
	configRepository       *repository.Config
	serverRepository       *repository.Server
}

func NewSubscription(
	containerRegistry *executor.Registry,
	newsletter *newsletter.Newsletter,
	subscriptionRepository *repository.Subscription,
	userRepository *repository.User,
	configRepository *repository.Config,
	serverRepository *repository.Server,
) *SubscriptionService {
	return &SubscriptionService{
		containerRegistry,
		newsletter,
		subscriptionRepository,
		userRepository,
		configRepository,
		serverRepository,
	}
}

func (s *SubscriptionService) AddSubscription(
	ctx context.Context,
	telegramID int64,
	configQuantity int,
) error {
	user, err := s.userRepository.FindByTelegramID(ctx, telegramID)
	if err != nil {
		return fmt.Errorf("finding user %d: %w", telegramID, err)
	}

	_, err = s.subscriptionRepository.Create(ctx, user.ID, configQuantity)
	if err != nil {
		return fmt.Errorf("creating subscription: %w", err)
	}

	return nil
}

func (s *SubscriptionService) SetPeriodEnd(
	ctx context.Context,
	endDate time.Time,
) (int64, error) {
	rowsAffected, err := s.subscriptionRepository.SetEndPeriodForAll(ctx, endDate)
	if err != nil {
		return 0, fmt.Errorf("setting period end for each user: %w", err)
	}

	users, err := s.userRepository.ListAll(ctx)
	if err != nil {
		return 0, fmt.Errorf("retrieving all users: %w", err)
	}

	var recipients []telego.ChatID
	for _, u := range users {
		recipients = append(recipients, telego.ChatID{ID: *u.TelegramID})
	}

	s.newsletter.BroadcastMessage(ctx.(*th.Context), recipients, "You need to pay")

	return rowsAffected, nil
}

func (s *SubscriptionService) SetPastDue(
	ctx context.Context,
	period time.Duration,
	batch int,
) error {
	_, err := s.subscriptionRepository.SetPastDue(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (s *SubscriptionService) HandleCancelled(
	ctx context.Context,
	period time.Duration,
	batch int,
) error {
	staleConfigs, err := s.configRepository.ListStaleConfigs(ctx, period, batch)
	if err != nil {
		return fmt.Errorf("retrieving stale configs: %w", err)
	}

	if len(staleConfigs) == 0 {
		return nil
	}

	servers, err := s.serverRepository.ListAll(ctx)
	if err != nil {
		return fmt.Errorf("retrieving servers: %w", err)
	}

	serverMap := make(map[string]string, 0)
	for _, srv := range servers {
		publicID := srv.PublicID.String()
		if _, ok := serverMap[publicID]; !ok {
			serverMap[publicID] = srv.InternalAddress
		}
	}

	for _, c := range staleConfigs {
		addr, ok := serverMap[c.ServerPublicID.String()]
		if !ok {
			continue
		}

		ex, err := s.containerRegistry.GetOrCreate(ctx, &executor.Config{
			Name:     c.ServerPublicID.String(),
			User:     "root",
			HostAddr: addr,
			HostPort: "2375",
		})
		if err != nil {
			continue
		}

		ce, ok := ex.(*executor.ContainerExecutor)
		if !ok {
			continue
		}

		protocol, err := configgen.GetProtocol(c.Protocol)
		if err != nil {
			continue
		}

		cg := configgen.New(protocol, ce)

		err = cg.Revoke(ctx, c.Protocol, c.IPAddress)
		if err != nil {
			continue
		}
	}

	_, err = s.subscriptionRepository.DeleteStale(ctx, period)
	if err != nil {
		return fmt.Errorf("deleting stale subscriptions: %w", err)
	}

	return nil
}
