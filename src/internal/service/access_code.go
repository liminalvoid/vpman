package service

import (
	"context"
	"fmt"

	"vpnmanager/internal/repository"
)

type AccessCodeService struct {
	accessCodeRepository *repository.AccessCode
}

func NewAccessCodeService(
	accessCodeRepository *repository.AccessCode,
) *AccessCodeService {
	return &AccessCodeService{
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
