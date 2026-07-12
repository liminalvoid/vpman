package configgen

import (
	"context"
	"fmt"

	"vpnmanager/internal/executor"
)

type Protocol interface {
	create(ctx context.Context, cli *executor.ContainerExecutor) (string, error)
	revoke(cg *ConfigGenerator)
}

type ConfigGenerator struct {
	protocol Protocol
}

func New(p Protocol) *ConfigGenerator {
	return &ConfigGenerator{
		protocol: p,
	}
}

func (cg *ConfigGenerator) SetProtocol(p Protocol) {
	cg.protocol = p
}

func (cg *ConfigGenerator) Create(ctx context.Context, cli *executor.ContainerExecutor) {
	cfg, err := cg.protocol.create(ctx, cli)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to create config: %s", err))

		return
	}

	fmt.Println("all great", cfg)
}

func (cg *ConfigGenerator) Revoke() {}
