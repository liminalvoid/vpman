package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"vpnmanager/internal/configgen"
	"vpnmanager/internal/executor"
	"vpnmanager/internal/lib"
	"vpnmanager/internal/model"
	"vpnmanager/internal/repository"
)

type ConfigService struct {
	log                    *slog.Logger
	configRepository       *repository.Config
	protocolRepository     *repository.Protocol
	userRepository         *repository.User
	subscriptionRepository *repository.Subscription
	serverRepository       *repository.Server
	containerRegistry      *executor.Registry
}

func NewConfig(
	log *slog.Logger,
	configRepository *repository.Config,
	protocolRepository *repository.Protocol,
	userRepository *repository.User,
	subscriptionRepository *repository.Subscription,
	serverRepository *repository.Server,
	containerRegistry *executor.Registry,
) *ConfigService {
	return &ConfigService{
		log,
		configRepository,
		protocolRepository,
		userRepository,
		subscriptionRepository,
		serverRepository,
		containerRegistry,
	}
}

func (s *ConfigService) CreateConfig(
	ctx context.Context,
	ownerID int64,
	server *model.Server,
	name, protocol, containerID, format string,
) (string, error) {
	subscription, err := s.subscriptionRepository.GetByTelegramID(ctx, ownerID)
	if err != nil {
		return "", fmt.Errorf("retrieving subscription: %w", err)
	}

	configs, err := s.configRepository.ListActiveByTelegramID(ctx, ownerID)
	if err != nil {
		return "", fmt.Errorf("retrieving configs: %w", err)
	}

	if len(configs) == int(subscription.Quantity) {
		return "", &model.SubscriptionLimitError{Limit: int(subscription.Quantity)}
	}

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
	resultCfg, ipAddr, err := cg.Create(ctx, containerID, server.PublicAddress)
	if err != nil {
		return "", fmt.Errorf("creating config: %w", err)
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

	cfgData := model.Config{
		Name:           name,
		IPAddress:      ipAddr,
		ProtocolID:     protocolStruct.ID,
		ServerID:       server.ID,
		SubscriptionID: subscription.ID,
		Status:         model.ConfigStatusActive,
	}
	err = s.configRepository.Create(ctx, &cfgData)
	if err != nil {
		return "", fmt.Errorf("writing config info to DB: %w", err)
	}

	// Xray configs are vless:// links, which AmneziaVPN imports as-is.
	if _, isXray := p.(*configgen.Xray); format == "amnezia" && !isXray {
		resultCfg, err = lib.CompressConfigForAmnezia([]byte(resultCfg), 8)
		if err != nil {
			return "", fmt.Errorf("compressing config: %w", err)
		}
	}

	s.log.InfoContext(
		ctx,
		"config successfully created",
		"protocol", protocol,
		"format", format,
		"addr", ipAddr,
	)

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

	s.log.InfoContext(
		ctx,
		"config's name updated",
		"config_id", publicID,
	)

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

	s.log.InfoContext(
		ctx,
		"config successfully revoked",
		"config_id", config.PublicID.String(),
	)

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

	s.log.InfoContext(
		ctx,
		"config successfully suspended",
		"config_id", config.PublicID.String(),
	)

	return nil
}

func (s *ConfigService) ListConfigs(
	ctx context.Context,
	ownerID int64,
) ([]model.ConfigListView, error) {
	configs, err := s.configRepository.ListActiveByTelegramID(ctx, ownerID)
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

func (s *ConfigService) RemoveStale(
	ctx context.Context,
	period time.Duration,
	batch int,
) (int64, error) {
	staleConfigs, err := s.configRepository.ListStaleConfigs(ctx, period, batch)
	if err != nil {
		return 0, fmt.Errorf("retrieving stale configs: %w", err)
	}

	if len(staleConfigs) == 0 {
		return 0, nil
	}

	servers, err := s.serverRepository.ListAll(ctx)
	if err != nil {
		return 0, fmt.Errorf("requesting servers: %w", err)
	}

	serverAddrMap := make(map[string]string, 0)
	for _, svr := range servers {
		if _, ok := serverAddrMap[svr.PublicID.String()]; !ok {
			serverAddrMap[svr.PublicID.String()] = svr.InternalAddress
		}
	}

	var staleCount int64
	for _, c := range staleConfigs {
		privateAddr, ok := serverAddrMap[c.ServerPublicID.String()]
		if !ok {
			s.log.WarnContext(ctx, "server not found for config", "config_id", c.PublicID, "server_id", c.ServerPublicID)
			continue
		}

		ex, err := s.containerRegistry.GetOrCreate(ctx, &executor.Config{
			Name:     c.ServerPublicID.String(),
			User:     "root",
			HostAddr: privateAddr,
			HostPort: "2375",
		})
		if err != nil {
			s.log.ErrorContext(ctx, "retrieving executor", "server_id", c.ServerPublicID, "err", err)
			continue
		}

		ce, ok := ex.(*executor.ContainerExecutor)
		if !ok {
			return staleCount, fmt.Errorf("unexpected executor type %T", ex)
		}

		p, err := configgen.GetProtocol(c.Protocol)
		if err != nil {
			s.log.ErrorContext(ctx, "retrieving protocol", "config_id", c.PublicID, "protocol", c.Protocol, "err", err)
			continue
		}
		cg := configgen.New(p, ce)

		// Revoke on the server before deleting the row, so a failed revoke
		// leaves the config listed for the next sweep.
		err = cg.Revoke(ctx, c.Protocol, c.IPAddress)
		if err != nil {
			s.log.ErrorContext(ctx, "revoking config on server", "config_id", c.PublicID, "ip", c.IPAddress, "err", err)
			continue
		}

		err = s.configRepository.Delete(ctx, c.PublicID.String())
		if err != nil {
			s.log.ErrorContext(ctx, "deleting stale config", "config_id", c.PublicID, "err", err)
			continue
		}

		staleCount += 1
	}

	return staleCount, nil
}
