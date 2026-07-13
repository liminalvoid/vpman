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
	executor *executor.ContainerExecutor
}

func New(p Protocol, ce *executor.ContainerExecutor) *ConfigGenerator {
	return &ConfigGenerator{
		protocol: p,
		executor: ce,
	}
}

func (cg *ConfigGenerator) SetProtocol(p Protocol) {
	cg.protocol = p
}

func (cg *ConfigGenerator) GetProtocol() (string, error) {
	switch cg.protocol {
	case &Awg{}:
		return "awg", nil
	case &Wg{}:
		return "wg", nil
	}

	return "", fmt.Errorf("failed to map protocol")
}

func (cg *ConfigGenerator) Create(ctx context.Context) {
	cfg, err := cg.protocol.create(ctx, cg.executor)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to create config: %s", err))

		return
	}

	fmt.Println("all great", cfg)
}

func (cg *ConfigGenerator) Revoke() {}
