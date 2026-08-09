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
	userPublicID uuid.UUID,
	configQuantity int,
) error {
	user, err := s.userRepository.FindByTelegramID(ctx, userPublicID, configQuantity)
	if err != nil {
		return fmt.Errorf("finding user %q: %w", userPublicID, configQuantity)
	}

	_, err := s.subscriptionRepository.Create(ctx, int64(userPublicID[]))
}
