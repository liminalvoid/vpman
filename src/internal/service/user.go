package service

import (
	"context"
	"errors"
	"fmt"

	"vpnmanager/internal/model"
	"vpnmanager/internal/repository"
)

type UserService struct {
	userRepository         *repository.User
	subscriptionRepository *repository.Subscription
}

func NewUser(
	userRepository *repository.User,
	subscriptionRepository *repository.Subscription,
) *UserService {
	return &UserService{
		userRepository,
		subscriptionRepository,
	}
}

func (s *UserService) CreateWithAccessCode(
	ctx context.Context,
	telegramID int64,
	code string,
) error {
	err := s.userRepository.CreateWithAccessCode(ctx, telegramID, code)
	if err != nil {
		return fmt.Errorf("marking code as used: %w", err)
	}

	return nil
}

func (s *UserService) GetByTelegramID(
	ctx context.Context,
	userID int64,
) (*model.User, error) {
	user, err := s.userRepository.FindByTelegramID(ctx, userID)
	if errors.Is(err, repository.ErrUserNotExist) {
		return nil, err
	} else if err != nil {
		return nil, fmt.Errorf("retrieving user by Telegram ID: %w", err)
	}

	return user, nil
}
