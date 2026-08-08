package configgen

import (
	"context"
	"fmt"

	"vpnmanager/internal/executor"
)

type Protocol interface {
	create(ctx context.Context, cli *executor.ContainerExecutor, containerID string) (string, error)
	revoke(ctx context.Context, cli *executor.ContainerExecutor, containerID string, addr string) error
	suspend(ctx context.Context, cli *executor.ContainerExecutor, containerID string, addr string) error
	resume(ctx context.Context, cli *executor.ContainerExecutor, containerID string, addr string) error
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

func (cg *ConfigGenerator) Create(ctx context.Context, containerID string) (string, error) {
	cfg, err := cg.protocol.create(ctx, cg.executor, containerID)
	if err != nil {
		return "", fmt.Errorf("creating config: %w", err)
	}

	return cfg, nil
}

func (cg *ConfigGenerator) Revoke(ctx context.Context, containerID string, addr string) error {
	err := cg.protocol.revoke(ctx, cg.executor, containerID, addr)
	if err != nil {
		return fmt.Errorf("revoking config: %w", err)
	}

	return nil
}

func (cg *ConfigGenerator) Suspend(
	ctx context.Context,
	containerID, addr string,
) error {
	err := cg.protocol.suspend(ctx, cg.executor, containerID, addr)
	if err != nil {
		return fmt.Errorf("suspending config: %w", err)
	}

	return nil
}

func (cg *ConfigGenerator) Resume(
	ctx context.Context,
	containerID, addr string,
) error {
	err := cg.protocol.resume(ctx, cg.executor, containerID, addr)
	if err != nil {
		return fmt.Errorf("resuming config: %w", err)
	}

	return nil
}
