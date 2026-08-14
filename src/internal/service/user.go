package service

import (
	"context"
	"fmt"

	"vpnmanager/internal/model"
	"vpnmanager/internal/repository"
)

type UserService struct {
	userRepository *repository.User
}

func NewUser(
	userRepository *repository.User,
) *UserService {
	return &UserService{
		userRepository,
	}
}

func (s *UserService) GetByTelegramID(
	ctx context.Context,
	userID int64,
) (*model.User, error) {
	user, err := s.userRepository.FindByTelegramID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("retrieving user by Telegram ID: %w", err)
	}

	return user, nil
}
