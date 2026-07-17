package service

import (
	"context"
	"fmt"

	"vpnmanager/internal/configgen"
	"vpnmanager/internal/model"
	"vpnmanager/internal/repository"
)

type PeerService struct {
	configGenerator  *configgen.ConfigGenerator
	configRepository *repository.Config
}

func NewPeer(
	configGenerator *configgen.ConfigGenerator,
	configRepository *repository.Config,
) *PeerService {
	return &PeerService{configGenerator, configRepository}
}

func (s *PeerService) CreatePeer(
	ctx context.Context,
	ownerID int64,
	name, location string,
) (string, error) {
	cfg, err := s.configGenerator.Create(ctx)
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
