package service

import (
	"context"
	"fmt"

	"vpnmanager/internal/executor"
	"vpnmanager/internal/model"
	"vpnmanager/internal/repository"
)

type ServerService struct {
	serverRepository  *repository.Server
	containerExecutor *executor.ContainerExecutor	
}

func NewServer(sr *repository.Server, ce *executor.ContainerExecutor) *ServerService {
	return &ServerService{serverRepository: sr, containerExecutor: ce}
}

func (s *ServerService) ListAll(ctx context.Context) ([]model.Server, error) {
	servers, err := s.serverRepository.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing servers: %w", err)
	}

	return servers, nil
}

func (s *ServerService) GetAvailableProtocols(server *model.Server) {
	ceCfg := executor.Config{
		User:        "root",
		HostAddr:    server.InternalAddress,
		HostPort:    "22",
	}
	s.containerExecutor.
}
