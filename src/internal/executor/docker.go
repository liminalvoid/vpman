package executor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/client"
)

type Config struct {
	User        string
	HostAddr    string
	HostPort    string
}

type ContainerExecutor struct {
	mu   sync.Mutex
	pool map[string]*client.Client
}

func New() *ContainerExecutor {
	return &ContainerExecutor{mu: sync.Mutex{}, pool: make(map[string]*client.Client)}
}

func (ce *ContainerExecutor) Connect(cfg *Config) error {
	dockerHost := fmt.Sprintf("ssh://%s@%s:%s", cfg.User, cfg.HostAddr, cfg.HostPort)
	c, err := client.New(client.WithHost(dockerHost))
	if err != nil {
		return fmt.Errorf("creating container executor: %w", err)
	}

	ce.pool[cfg.HostAddr] = c
}

func (ce *ContainerExecutor) Run(ctx context.Context, cmd []string, stdin []byte) (string, error) {
	execResp, err := ce.pool[].ExecCreate(ctx, ce.ContainerID, client.ExecCreateOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
		AttachStdin:  true,
	})
	if err != nil {
		return "", fmt.Errorf("failed command execution: %s", err)
	}

	attachResp, err := ce.cli.ExecAttach(ctx, execResp.ID, client.ExecAttachOptions{})
	if err != nil {
		return "", fmt.Errorf("failed attach connection: %s", err)
	}
	defer attachResp.Close()

	if len(stdin) > 0 {
		attachResp.Conn.Write(stdin)
		attachResp.CloseWrite()
	}

	var stdoutBuf bytes.Buffer
	if _, err := stdcopy.StdCopy(&stdoutBuf, io.Discard, attachResp.Reader); err != nil {
		return "", fmt.Errorf("failed de-multiplexing stdErr and stdOut: %s", err)
	}

	_, err = ce.cli.ExecInspect(ctx, execResp.ID, client.ExecInspectOptions{})
	if err != nil {
		return "", fmt.Errorf("failed process info retrieval: %s", err)
	}

	return stdoutBuf.String(), nil
}

func (ce *ContainerExecutor) ListContainers(ctx context.Context) (*client.ContainerListResult, error) {
	list, err := ce.cli.ContainerList(ctx, client.ContainerListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch containers list: %s", err)
	}

	return &list, nil
}
