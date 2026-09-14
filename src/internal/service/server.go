package service

import (
	"context"
	"fmt"
	"log/slog"

	"vpnmanager/internal/executor"
	"vpnmanager/internal/model"
	"vpnmanager/internal/repository"
)

type ServerService struct {
	log               *slog.Logger
	serverRepository  *repository.Server
	containerRegistry *executor.Registry
}

func NewServer(
	log *slog.Logger,
	serverRepository *repository.Server,
	containerRegistry *executor.Registry,
) *ServerService {
	return &ServerService{log, serverRepository, containerRegistry}
}

func (s *ServerService) ListAll(ctx context.Context) ([]model.Server, error) {
	servers, err := s.serverRepository.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing servers: %w", err)
	}

	return servers, nil
}

func (s *ServerService) GetByPublicID(
	ctx context.Context,
	publicID string,
) (*model.Server, error) {
	server, err := s.serverRepository.GetByPublicID(ctx, publicID)
	if err != nil {
		return nil, err
	}

	return server, nil
}
