package service

import (
	"context"
	"fmt"

	"vpnmanager/internal/configgen"
	"vpnmanager/internal/model"
	"vpnmanager/internal/repository"
)

type PeerService struct {
	cg *configgen.ConfigGenerator
	cr *repository.Config
}

func NewPeer(
	cg *configgen.ConfigGenerator,
	cr *repository.Config,
) *PeerService {
	return &PeerService{cg: cg, cr: cr}
}

func (s *PeerService) CreatePeer(
	ctx context.Context,
	ownerID int64,
	name, location string,
) (string, error) {
	cfg, err := s.cg.Create(ctx)
	if err != nil {
		return "", fmt.Errorf("generating config: %w", err)
	}

	cfgData := model.Config{
		Name: name,
		UserID: ownerID,
		ProtocolID: 0,
		Location: location,
	}
	err = s.cr.Create(ctx, &cfgData)
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
