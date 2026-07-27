package service

import (
	"context"
	"fmt"
	"strings"

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

func (s *PeerService) GetAvailableProtocols(ctx context.Context, server model.Server) (map[string]string, error) {
	var ex executor.Executor
	ex, err := s.containerRegistry.Get(ctx, server.PublicID.String())
	fmt.Println(server.PublicID.String())
	if err != nil {
		cfg := executor.Config{
			Name: server.PublicID.String(),	
			User: "root",
			HostAddr: server.InternalAddress,
			HostPort: "2375",
		}
		ex, err = s.containerRegistry.Add(ctx, &cfg)
		if err != nil {
			return nil, fmt.Errorf("connecting to host %q: %w", server.InternalAddress, err)
		}
	}

	containers, err := ex.ListContainers(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving available containers: %w", err)
	}

	protocols := configgen.ListProtocols()

	protocolToContainerMap := make(map[string]string, 0)
	for _, c := range containers.Items {
		containerName := strings.ReplaceAll(c.Names[0], "/", "")
		fmt.Println(containerName)
		if _, ok := protocols[containerName]; ok {
			protocolToContainerMap[containerName] = c.ID
		}
	}

	return protocolToContainerMap, nil
}
