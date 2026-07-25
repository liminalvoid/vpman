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
	configRepository *repository.Config,
	containerRegistry *executor.Registry,
) *PeerService {
	return &PeerService{configRepository, containerRegistry}
}

func (s *PeerService) CreatePeer(
	ctx context.Context,
	ownerID int64,
	name, server, protocol string,
) (string, error) {
	p, err := configgen.GetProtocol(protocol)
	if err != nil {
		return "", fmt.Errorf("looking up for %q protocol: %w", name, err)
	}

	ex, err := s.containerRegistry.Get(ctx, server)
	if err != nil {
		return "", fmt.Errorf("getting server %q from registry: %w", server, err)
	}

	ce, ok := ex.(*executor.ContainerExecutor)
	if !ok {
		return "", fmt.Errorf("unexpected executor type %T", ex)
	}

	cg := configgen.New(p, ce)
	resultCfg, err := cg.Create(ctx)
	if err != nil {
		return "", fmt.Errorf("creating config: %w", err)
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

	return resultCfg, nil
}

func (s *PeerService) ListPeers(
	ctx context.Context,
	ownerID int64,
) error {
	return nil
}

func (s *PeerService) GetAvailableProtocols(ctx context.Context, serverId string) (map[string]string, error) {
	ex, err := s.containerRegistry.Get(ctx, serverId)
	if err != nil {
		return nil, fmt.Errorf("getting %q from container registry: %w", serverId, err)
	}

	containers, err := ex.ListContainers(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving available containers: %w", err)
	}

	protocols := configgen.ListProtocols()

	protocolToContainerMap := make(map[string]string, 0)
	for _, c := range containers.Items {
		if _, ok := protocols[c.Names[0]]; ok {
			protocolToContainerMap[c.Names[0]] = c.ID
		}
	}

	return protocolToContainerMap, nil
}
