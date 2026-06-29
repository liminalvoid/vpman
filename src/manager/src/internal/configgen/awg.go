package configgen

import (
	"context"

	"manager/internal/sshclient"
)

type Awg struct {}

func (l *Awg) create(ctx context.Context, ssh sshclient.Client, cg *ConfigGenerator) (error) {
	generateKeyCmd := sshclient.Command{
		Name: "sh",
		Args: []string{"-c", "private=$(wg genkey); printf '%s\n%s' \"$private\""},
	}

	res, err := ssh.Run(ctx, generateKeyCmd)
	if err != nil {
		return err
	}
}

func (l *Awg) revoke(cg *ConfigGenerator) {}
