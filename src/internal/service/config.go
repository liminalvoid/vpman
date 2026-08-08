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

type ConfigService struct {
	configRepository   *repository.Config
	protocolRepository *repository.Protocol
	userRepository     *repository.User
	containerRegistry  *executor.Registry
}

func NewConfig(
	configRepository *repository.Config,
	protocolRepository *repository.Protocol,
	userRepository *repository.User,
	containerRegistry *executor.Registry,
) *ConfigService {
	return &ConfigService{
		configRepository,
		protocolRepository,
		userRepository,
		containerRegistry,
	}
}

func (s *ConfigService) CreateConfig(
	ctx context.Context,
	ownerID int64,
	server *model.Server,
	name, protocol, containerID string,
) (string, error) {
	// TODO: Add revert of transaction and revocation of config on failure
	p, err := configgen.GetProtocol(protocol)
	if err != nil {
		return "", fmt.Errorf("looking up for %q protocol: %w", name, err)
	}

	cfg := &executor.Config{
		Name:     server.PublicID.String(),
		User:     "root",
		HostAddr: server.InternalAddress,
		HostPort: "2375",
	}
	ex, err := s.containerRegistry.GetOrCreate(ctx, cfg)
	if err != nil {
		return "", fmt.Errorf(
			"getting server %q from registry: %w",
			server.PublicID.String(),
			err,
		)
	}

	ce, ok := ex.(*executor.ContainerExecutor)
	if !ok {
		return "", fmt.Errorf("unexpected executor type %T", ex)
	}

	cg := configgen.New(p, ce)
	resultCfg, err := cg.Create(ctx, containerID)
	if err != nil {
		return "", fmt.Errorf("creating config: %w", err)
	}

	var ipAddr string
	switch p.(type) {
	case *configgen.Awg, *configgen.Wg:
		lookupStr := "Address = "
		ipAddrIdx := strings.Index(resultCfg, lookupStr)
		if ipAddrIdx == -1 {
			return "", fmt.Errorf("can't find config's IP address")
		}
		ipAddrIdx += len(lookupStr)
		ipAddr = resultCfg[ipAddrIdx:]
		bitmaskIdx := strings.Index(ipAddr, "/32")
		if bitmaskIdx == -1 {
			return "", fmt.Errorf("can't find EOL to extract IP address")
		}
		ipAddr = ipAddr[:bitmaskIdx]
	}

	protocolStruct := model.Protocol{
		Name: protocol,
	}
	err = s.protocolRepository.Find(
		ctx,
		&protocolStruct,
	)
	if err != nil {
		return "", fmt.Errorf("querying protocol %q: %w", protocolStruct.Name, err)
	}

	user := model.User{
		TelegramID: ownerID,
	}
	err = s.userRepository.FindByTelegramID(ctx, &user)
	if err != nil {
		return "", fmt.Errorf("querying user by Telegram ID (%d): %w", ownerID, err)
	}

	cfgData := model.Config{
		Name:       name,
		IPAddress:  ipAddr,
		UserID:     user.ID,
		ProtocolID: protocolStruct.ID,
		ServerID:   int64(server.ID),
		Status:     model.ConfigStatusActive,
	}
	err = s.configRepository.Create(ctx, &cfgData)
	if err != nil {
		return "", fmt.Errorf("writing config info to DB: %w", err)
	}

	return resultCfg, nil
}

func (s *ConfigService) UpdateConfigName(
	ctx context.Context,
	publicID, name string,
) error {
	_, err := s.configRepository.UpdateName(ctx, publicID, name)
	if err != nil {
		return fmt.Errorf("updating config's %q name: %w", publicID, err)
	}

	return nil
}

func (s *ConfigService) RevokeConfig(
	ctx context.Context,
	server *model.Server,
	config *model.ConfigListView,
) error {
	cfg := &executor.Config{
		Name:     server.PublicID.String(),
		User:     "root",
		HostAddr: server.InternalAddress,
		HostPort: "2375",
	}
	ex, err := s.containerRegistry.GetOrCreate(ctx, cfg)
	if err != nil {
		return fmt.Errorf(
			"retrieving server %q from registry: %w",
			server.PublicID.String(),
			err,
		)
	}

	ce, ok := ex.(*executor.ContainerExecutor)
	if !ok {
		return fmt.Errorf("unexpected executor type %T", ex)
	}

	p, err := configgen.GetProtocol(config.Protocol)
	if err != nil {
		return fmt.Errorf("mapping protocol %q to object: %w", config.Protocol, err)
	}

	cg := configgen.New(p, ce)

	err = cg.Revoke(ctx, config.Protocol, config.IPAddress)
	if err != nil {
		return fmt.Errorf("deleting peer: %w", err)
	}

	err = s.configRepository.Delete(ctx, config.PublicID.String())
	if err != nil {
		return fmt.Errorf("deleting config %q from DB: %w", config.PublicID, err)
	}

	return nil
}

func (s *ConfigService) SuspendConfig(
	ctx context.Context,
	server *model.Server,
	config *model.ConfigListView,
) error {
	cfg := &executor.Config{
		Name:     server.PublicID.String(),
		User:     "root",
		HostAddr: server.InternalAddress,
		HostPort: "2375",
	}
	ex, err := s.containerRegistry.GetOrCreate(ctx, cfg)
	if err != nil {
		return fmt.Errorf(
			"retrieving server %q from registry: %w",
			server.PublicID.String(),
			err,
		)
	}

	ce, ok := ex.(*executor.ContainerExecutor)
	if !ok {
		return fmt.Errorf("unexpected executor type %T", ex)
	}

	p, err := configgen.GetProtocol(config.Protocol)
	if err != nil {
		return fmt.Errorf("mapping protocol %q to object: %w", config.Protocol, err)
	}

	cg := configgen.New(p, ce)

	err = cg.Suspend(ctx, config.Protocol, config.IPAddress)
	if err != nil {
		return fmt.Errorf("suspending peer: %w", err)
	}

	_, err = s.configRepository.MarkSuspended(ctx, config.PublicID.String())
	if err != nil {
		return fmt.Errorf("marking config %q as suspended: %w", config.PublicID, err)
	}

	return nil
}

func (s *ConfigService) ListConfigs(
	ctx context.Context,
	ownerID int64,
) ([]model.ConfigListView, error) {
	configs, err := s.configRepository.ListByTelegramID(ctx, ownerID)
	if err != nil {
		return nil, fmt.Errorf("querying user's (%d) configs: %w", ownerID, err)
	}

	return configs, nil
}

func (s *ConfigService) GetConfigByPublicID(
	ctx context.Context,
	publicID string,
) (*model.ConfigListView, error) {
	cfg, err := s.configRepository.GetByPublicID(ctx, publicID)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (s *ConfigService) GetAvailableProtocols(ctx context.Context, server *model.Server) (map[string]struct{}, error) {
	cfg := &executor.Config{
		Name:     server.PublicID.String(),
		User:     "root",
		HostAddr: server.InternalAddress,
		HostPort: "2375",
	}

	ex, err := s.containerRegistry.GetOrCreate(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("retrieving executor from server %q: %w", server.InternalAddress, err)
	}

	containers, err := ex.ListContainers(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving available containers: %w", err)
	}

	protocols := configgen.ListProtocols()

	availableProtocols := make(map[string]struct{}, 0)
	for _, c := range containers.Items {
		containerName := strings.ReplaceAll(c.Names[0], "/", "")
		if _, ok := protocols[containerName]; ok {
			availableProtocols[containerName] = struct{}{}
		}
	}

	return availableProtocols, nil
}
