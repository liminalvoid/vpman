package service

import (
	"context"
	"errors"
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

var ErrNoAccessCode = errors.New("code doesn't exist")

func (s *AccessCodeService) Delete(
	ctx context.Context,
	code string,
) error {
	count, err := s.accessCodeRepository.Delete(ctx, code)
	if err != nil {
		return fmt.Errorf("marking code as used: %w", err)
	}

	if count == 0 {
		return ErrNoAccessCode
	}

	return nil
}
