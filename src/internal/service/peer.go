package service

import (
	"context"
	"fmt"

	"vpnmanager/internal/configgen"
	"vpnmanager/internal/executor"
	"vpnmanager/internal/model"
	"vpnmanager/internal/repository"
)

type PeerService struct {
	configRepository  *repository.Config
	containerRegistry *executor.Registry
}

func NewPeer(
	configRepository  *repository.Config,
	containerRegistry *executor.Registry,
) *PeerService {
	return &PeerService{configRepository, containerRegistry}
}

func (s *PeerService) CreatePeer(
	ctx context.Context,
	ownerID int64,
	name, location string,
) (string, error) {
	cfg, err := configgen.New(&configgen.Awg{}, s.containerRegistry)
	if err != nil {
		return "", fmt.Errorf("generating config: %w", err)
	}

	cfgData := model.Config{
		Name:       name,
		UserID:     ownerID,
		ProtocolID: 0,
		ServerID:   0,
	}
	err = s.configRepository.Create(ctx, &cfgData)
	if err != nil {
		return "", fmt.Errorf("writing config info to DB: %w", err)
	}

	return cfg, nil
}

func (s *PeerService) ListPeers(
	ctx context.Context,
	ownerID int64,
) error {
	return nil
}

func (s *PeerService) GetAvailableProtocols(ctx context.Context, serverId string) ([]string, error) {
	ex, err := s.containerRegistry.Get(ctx, serverId)
	if err != nil {
		return nil, fmt.Errorf("getting %q from container registry: %w", serverId, err)
	}

	containers, err := ex.ListContainers(ctx)
	if err != nil {
		return nil, fmt.Errorf("")
	}

	for _, c := range containers.Items {
	}
}
