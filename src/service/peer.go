package service

import (
	"context"
	"fmt"

	"vpnmanager/internal/configgen"
)

type PeerService struct {
	cg *configgen.ConfigGenerator
}

func New(cg *configgen.ConfigGenerator) *PeerService {
	return &PeerService{cg: cg}
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

	return cfg, nil
}

func (s *PeerService) ListPeers(
	ctx context.Context,
	ownerID int64,
) error {
	return nil
}
