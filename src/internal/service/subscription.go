package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"text/template"
	"time"

	"vpnmanager/internal/executor"
	"vpnmanager/internal/model"
	"vpnmanager/internal/repository"
	"vpnmanager/internal/tgbot/newsletter"

	"github.com/google/uuid"
)

type SubscriptionService struct {
	log                    *slog.Logger
	containerRegistry      *executor.Registry
	newsletter             *newsletter.Newsletter
	subscriptionRepository *repository.Subscription
	userRepository         *repository.User
	configRepository       *repository.Config
	serverRepository       *repository.Server
}

func NewSubscription(
	log *slog.Logger,
	containerRegistry *executor.Registry,
	newsletter *newsletter.Newsletter,
	subscriptionRepository *repository.Subscription,
	userRepository *repository.User,
	configRepository *repository.Config,
	serverRepository *repository.Server,
) *SubscriptionService {
	return &SubscriptionService{
		log,
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

	s.log.InfoContext(
		ctx,
		"subscription successfully created",
		"keys_limit", configQuantity,
	)

	return nil
}

func (s *SubscriptionService) GetByTelegramID(
	ctx context.Context,
	telegramID int64,
) (*model.Subscription, error) {
	subscription, err := s.subscriptionRepository.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return nil, fmt.Errorf("retrieving subscription via user's Telegram ID: %w", err)
	}

	return subscription, nil
}

func (s *SubscriptionService) GetByPublicID(
	ctx context.Context,
	userPublicID uuid.UUID,
) (*model.Subscription, error) {
	subscription, err := s.subscriptionRepository.GetByPublicID(ctx, userPublicID)
	if err != nil {
		return nil, fmt.Errorf("retrieving subscription via user's public ID: %w", err)
	}

	return subscription, nil
}

func (s *SubscriptionService) IsLimitSurpassedByTelegramID(
	ctx context.Context,
	telegramID int64,
) (*bool, error) {
	isLimitSurpassed, err := s.subscriptionRepository.IsLimitSurpassedByTelegramID(ctx, telegramID)
	if err != nil {
		return nil, fmt.Errorf("retrieving configs limit overrun: %w", err)
	}

	return isLimitSurpassed, nil
}

func (s *SubscriptionService) SetPeriodEnd(
	ctx context.Context,
	endDate time.Time,
	message string,
	amount int,
) (int64, error) {
	rowsAffected, err := s.subscriptionRepository.SetEndPeriodForAll(ctx, endDate)
	if err != nil {
		return 0, fmt.Errorf("setting period end for each user: %w", err)
	}

	users, err := s.userRepository.ListAll(ctx)
	if err != nil {
		return 0, fmt.Errorf("retrieving all users: %w", err)
	}

	messageTemplate, err := template.New("message").Parse(message)
	if err != nil {
		return 0, fmt.Errorf("parsing message template: %w", err)
	}

	totalKeys := 0
	for _, u := range users {
		subscription, err := s.subscriptionRepository.GetByPublicID(ctx, u.PublicID)
		if err != nil {
			return 0, err
		}

		totalKeys += int(subscription.Quantity)
	}

	pricePerKey := amount / totalKeys

	for _, u := range users {
		if u.IsAdmin {
			continue
		}

		subscription, err := s.subscriptionRepository.GetByPublicID(ctx, u.PublicID)
		if err != nil {
			return 0, err
		}

		var buf strings.Builder
		err = messageTemplate.Execute(&buf, subscription.Quantity*int32(pricePerKey))
		if err != nil {
			return 0, fmt.Errorf("executing message template: %w", err)
		}

		telegramID := *u.TelegramID
		s.newsletter.SendPaymentNotification(
			ctx,
			telegramID,
			buf.String(),
		)
	}

	s.log.InfoContext(
		ctx,
		"end period has been set",
		"end_period", endDate,
	)

	return rowsAffected, nil
}

func (s *SubscriptionService) SetPastDue(
	ctx context.Context,
	period time.Duration,
	batch int,
) (int64, error) {
	rowsAffected, err := s.subscriptionRepository.SetPastDue(ctx)
	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}

func (s *SubscriptionService) SetKeysLimit(
	ctx context.Context,
	userPublicID uuid.UUID,
	limit int,
) error {
	subscription, err := s.subscriptionRepository.GetByPublicID(ctx, userPublicID)
	if err != nil {
		return fmt.Errorf("retrieving subscription via user ID: %w", err)
	}
	prevLimit := subscription.Quantity

	_, err = s.subscriptionRepository.ChangeKeysLimitCount(ctx, limit, subscription.UserID)
	if err != nil {
		return fmt.Errorf("trying change keys limit: %w", err)
	}

	if int(prevLimit) > limit {
		removalCount := int(prevLimit) - limit
		_, err := s.configRepository.MarkSuspendedNLatest(ctx, subscription.ID, removalCount)
		if err != nil {
			return fmt.Errorf("trying to suspend %d latest configs: %w", removalCount, err)
		}
	}

	s.log.InfoContext(
		ctx,
		"successfully set keys limit",
		"for", userPublicID,
		"new_limit", limit,
	)

	return nil
}

func (s *SubscriptionService) ClaimPayment(
	ctx context.Context,
	telegramID int64,
) error {
	err := s.subscriptionRepository.ClaimPayment(ctx, telegramID)
	if err != nil {
		return err
	}

	s.log.InfoContext(
		ctx,
		"payment has been claimed",
	)

	return nil
}
