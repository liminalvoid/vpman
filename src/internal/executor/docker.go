package executor

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/client"
)

type Config struct {
	User     string
	HostAddr string
	HostPort string
}

type Executor interface {
	Run(ctx context.Context, cmd []string, stdin []byte) (string, error)
	Close() error
}

type ContainerExecutor struct {
	cli         *client.Client
	hostAddr    string
	containerID string
}

type Factory func(ctx context.Context, cfg *Config) (Executor, error)

func NewExecutor(ctx context.Context, cfg *Config) (Executor, error) {
	dockerHost := fmt.Sprintf("ssh://%s@%s:%s", cfg.User, cfg.HostAddr, cfg.HostPort)
	c, err := client.New(client.WithHost(dockerHost))
	if err != nil {
		return nil, fmt.Errorf("creating container executor: %w", err)
	}

	return &ContainerExecutor{cli: c, hostAddr: cfg.HostAddr}, nil
}

func (ce *ContainerExecutor) Run(ctx context.Context, cmd []string, stdin []byte) (string, error) {
	execResp, err := ce.cli.ExecCreate(ctx, ce.containerID, client.ExecCreateOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
		AttachStdin:  true,
	})
	if err != nil {
		return "", fmt.Errorf("executing command: %w", err)
	}
	
	attachResp, err := ce.cli.ExecAttach(ctx, execResp.ID, client.ExecAttachOptions{})
	if err != nil {
		return "", fmt.Errorf("attaching connection: %w", err)
	}
	defer attachResp.Close()
	
	if len(stdin) > 0 {
		attachResp.Conn.Write(stdin)
		attachResp.CloseWrite()
	}
	
	var stdoutBuf bytes.Buffer
	if _, err := stdcopy.StdCopy(&stdoutBuf, io.Discard, attachResp.Reader); err != nil {
		return "", fmt.Errorf("de-multiplexing stdErr and stdOut: %w", err)
	}
	
	_, err = ce.cli.ExecInspect(ctx, execResp.ID, client.ExecInspectOptions{})
	if err != nil {
		return "", fmt.Errorf("retrieving process info: %w", err)
	}
	
	return stdoutBuf.String(), nil
}

func (ce *ContainerExecutor) Close() error {
	return ce.cli.Close()
}

func (ce *ContainerExecutor) ListContainers(ctx context.Context) (*client.ContainerListResult, error) {
	list, err := ce.cli.ContainerList(ctx, client.ContainerListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("fetching containers list: %w", err)
	}

	return &list, nil
}
