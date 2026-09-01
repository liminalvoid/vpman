package service

import (
	"context"
	"fmt"
	"time"

	"vpnmanager/internal/executor"
	"vpnmanager/internal/model"
	"vpnmanager/internal/repository"
	"vpnmanager/internal/tgbot/newsletter"

	"github.com/google/uuid"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
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
		if u.IsAdmin {
			continue
		}

		recipients = append(recipients, telego.ChatID{ID: *u.TelegramID})
	}

	confirmPaymentKb := tu.InlineKeyboard([][]telego.InlineKeyboardButton{
		tu.InlineKeyboardRow(
			tu.
				InlineKeyboardButton("Подтвердить оплату").
				WithCallbackData("confirm_payment"),
		),
	}...)

	s.newsletter.BroadcastMessage(
		ctx.(*th.Context),
		recipients,
		message,
		confirmPaymentKb,
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

	return nil
}
