package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"vpnmanager/internal/model"
	"vpnmanager/internal/repository"
)

type UserService struct {
	log           *slog.Logger
	users         *repository.User
	subscriptions *repository.Subscription
}

func NewUser(
	log *slog.Logger,
	users *repository.User,
	subs *repository.Subscription,
) *UserService {
	return &UserService{
		log: log,
		users: users,
		subscriptions: subs,
	}
}

func (s *UserService) CreateWithAccessCode(
	ctx context.Context,
	telegramID int64,
	code string,
) error {
	user, err := s.users.CreateWithAccessCode(ctx, telegramID, code)
	if err != nil {
		return fmt.Errorf("marking code as used: %w", err)
	}

	_, err = s.subscriptions.Create(ctx, user.ID, 5)
	if err != nil {
		return fmt.Errorf("creating subscription for new user: %w", err)
	}

	return nil
}

func (s *UserService) ListAll(ctx context.Context) ([]model.User, error) {
	users, err := s.users.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving users: %w", err)
	}

	return users, nil
}

func (s *UserService) GetByTelegramID(
	ctx context.Context,
	userID int64,
) (*model.User, error) {
	user, err := s.users.FindByTelegramID(ctx, userID)
	if errors.Is(err, repository.ErrUserNotExist) {
		return nil, err
	} else if err != nil {
		return nil, fmt.Errorf("retrieving user by Telegram ID: %w", err)
	}

	return user, nil
}
