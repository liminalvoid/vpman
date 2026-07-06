package executor

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/client"
)

type ContainerExecutor struct {
	cli         *client.Client
	ContainerID string
	HostAddr    string
}

func (ce *ContainerExecutor) Run(ctx context.Context, cmd []string) (string, error) {
	execResp, err := ce.cli.ExecCreate(ctx, ce.ContainerID, client.ExecCreateOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return "", fmt.Errorf("failed command execution: %s", err)
	}

	attachResp, err := ce.cli.ExecAttach(ctx, execResp.ID, client.ExecAttachOptions{})
	if err != nil {
		return "", fmt.Errorf("failed attach connection: %s", err)
	}
	defer attachResp.Close()

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
