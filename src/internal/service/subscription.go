package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"vpnmanager/internal/repository"
)

type SubscriptionService struct {
	subscriptionRepository *repository.Subscription
	userRepository         *repository.User
}

func NewSubscription() *SubscriptionService {
	return &SubscriptionService{}
}

func (s *SubscriptionService) AddSubscription(
	ctx context.Context,
	telegramID int64,
	configQuantity int,
) error {
	user, err := s.userRepository.FindByTelegramID(ctx, telegramID)
	if err != nil {
		return fmt.Errorf("finding user %q: %w", telegramID, configQuantity)
	}

	_, err = s.subscriptionRepository.Create(ctx, user.ID, configQuantity)
	if err != nil {
		return fmt.Errorf("creating subscription: %w", err)
	}

	return nil
}
