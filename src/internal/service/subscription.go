package service

import (
	"context"
	"fmt"
	"time"

	"vpnmanager/internal/executor"
	"vpnmanager/internal/repository"
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
