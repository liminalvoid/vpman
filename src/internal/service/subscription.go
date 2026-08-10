package service

import (
	"context"
	"fmt"
	"time"

	"vpnmanager/internal/executor"
	"vpnmanager/internal/repository"
	"vpnmanager/internal/configgen"
)

type SubscriptionService struct {
	subscriptionRepository *repository.Subscription
	userRepository         *repository.User
	configRepository       *repository.Config
	serverRepository       *repository.Server
	containerRegistry      *executor.Registry
}

func NewSubscription(
	subscriptionRepository *repository.Subscription,
	userRepository         *repository.User,
	configRepository       *repository.Config,
	serverRepository       *repository.Server,
	containerRegistry      *executor.Registry,
) *SubscriptionService {
	return &SubscriptionService{
		subscriptionRepository,
		userRepository,
		configRepository,
		serverRepository,
		containerRegistry,
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

	return rowsAffected, nil
}

func (s *SubscriptionService) SetPastDue(
	ctx context.Context,
) (int64, error) {
	rowsAffected, err := s.subscriptionRepository.SetPastDue(ctx)
	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}

func (s *SubscriptionService) RemoveStale(
	ctx context.Context,
	batch int,
	period time.Duration,
) error {
	staleConfigs, err := s.configRepository.ListStaleConfigs(ctx, period, batch)
	if err != nil {
		return fmt.Errorf("retrieving stale configs: %w", err)
	}

	servers, err := s.serverRepository.ListAll(ctx)
	if err != nil {
		return fmt.Errorf("requesting servers: %w", err)
	}

	serverAddrMap := make(map[string]string, 0)
	for _, svr := range servers {
		if _, ok := serverAddrMap[svr.PublicID.String()]; !ok {
			serverAddrMap[svr.PublicID.String()] = svr.InternalAddress
		}
	}

	for _, c := range staleConfigs {
		privateAddr, ok := serverAddrMap[c.ServerPublicID.String()]
		if !ok {
			continue
		}

		ex, err := s.containerRegistry.GetOrCreate(ctx, &executor.Config{
			Name: c.ServerPublicID.String(),
			User: "root",
			HostAddr: privateAddr,
			HostPort: "2375",
		})
		if err != nil {
			return fmt.Errorf("retrieving container %q: %w", c.ServerPublicID, err)
		}

		ce, ok := ex.(*executor.ContainerExecutor)
		if !ok {
			return fmt.Errorf("unexpected executor type %T", ex)
		}

		p, err := configgen.GetProtocol(c.Protocol)
		if err != nil {
			return fmt.Errorf("retrieving protocol %q: %w", c.Protocol, err)
		}
		cg := configgen.New(p, ce)

		err = cg.Revoke(ctx, c.Protocol, c.IPAddress)
		if err != nil {
			return fmt.Errorf("revoking protocol with IP address %q: %w", c.IPAddress, err)
		}
	}

	return nil
}
