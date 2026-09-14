package service

import (
	"context"
	"fmt"
	"log/slog"

	"vpnmanager/internal/repository"
)

type AccessCodeService struct {
	log                  *slog.Logger
	accessCodeRepository *repository.AccessCode
}

func NewAccessCodeService(
	log *slog.Logger,
	accessCodeRepository *repository.AccessCode,
) *AccessCodeService {
	return &AccessCodeService{
		log,
		accessCodeRepository,
	}
}

func (s *AccessCodeService) Create(
	ctx context.Context,
	code string,
) error {
	_, err := s.accessCodeRepository.Create(ctx, code)
	if err != nil {
		return fmt.Errorf("creating access code: %w", err)
	}

	return nil
}
