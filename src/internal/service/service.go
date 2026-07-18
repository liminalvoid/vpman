package service

import (
	"context"
	"fmt"

	"vpnmanager/internal/model"
	"vpnmanager/internal/repository"
)

type ServerService struct {
	serverRepository *repository.Server
}

func NewServer(sr *repository.Server) *ServerService {
	return &ServerService{serverRepository: sr}
}

func (s *ServerService) ListAll(ctx context.Context) ([]model.Server, error) {
	servers, err := s.serverRepository.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing servers: %w", err)
	}

	return servers, nil
}
