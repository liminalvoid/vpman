package configgen

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/moby/moby/client"
	"github.com/moby/moby/api/pkg/stdcopy"
)

type Awg struct{}

func (l *Awg) create(ctx context.Context, cli *client.Client, cg *ConfigGenerator, containerID string) (string, error) {
	execResp, err := cli.ExecCreate(ctx, containerID, client.ExecCreateOptions{
		Cmd: cmd,
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return "", fmt.Errorf("failed: %s", err)
	}

	attachResp, err := cli.ExecAttach(ctx, execResp.ID, client.ExecAttachOptions{})
	if err != nil {
		return "", fmt.Errorf("failed: %s", err)
	}
	defer attachResp.Close()

	var stdoutBuf bytes.Buffer
	if _, err := stdcopy.StdCopy(&stdoutBuf, io.Discard, attachResp.Reader); err != nil {
		return "", fmt.Errorf("failed: %s", err)
	}

	_, err = cli.ExecInspect(ctx, execResp.ID, client.ExecInspectOptions{})
	if err != nil {
		return "", fmt.Errorf("failed: %s", err)
	}

	return "", nil
	// // Generate keys via SSH
	// genPrivKeyCmd := sshclient.Command{
	// 	Name: "wg",
	// 	Args: []string{"genkey"},
	// }
	// privKeyRes, err := ssh.Run(ctx, genPrivKeyCmd, nil)
	// if err != nil {
	// 	return "", fmt.Errorf("failed private key generation: %s", err)
	// }
	// genPubKeyCmd := sshclient.Command{
	// 	Name: "wg",
	// 	Args: []string{fmt.Sprintf("%s | awg pubkey", privKeyRes.Stdout)},
	// }
	// pubKeyRes, err := ssh.Run(ctx, genPubKeyCmd, nil)
	// if err != nil {
	// 	return "", fmt.Errorf("failed public key generation: %s", err)
	// }
	// genPresharedKeyCmd := sshclient.Command{
	// 	Name: "wg",
	// 	Args: []string{"genpsk"},
	// }
	// presharedKeyRes, err := ssh.Run(ctx, genPresharedKeyCmd, nil)
	// if err != nil {
	// 	return "", fmt.Errorf("failed preshared key generation: %s", err)
	// }

	// // Obtain server's credentials
	// getServerPubKeyCmd := sshclient.Command{
	// 	Name: "wg",
	// 	Args: []string{"show", "wg0", "public-key"},
	// }
	// serverPubKeyRes, err := ssh.Run(ctx, getServerPubKeyCmd, nil)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to obtain server's public key: %s", err)
	// }
	// getServerPortCmd := sshclient.Command{
	// 	Name: "wg",
	// 	Args: []string{"show", "wg0", "listen-port"},
	// }
	// serverPortRes, err := ssh.Run(ctx, getServerPortCmd, nil)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to obtain server's port: %s", err)
	// }

	// // Get the latest added peer's IP
	// getLatestPeerAddrCmd := sshclient.Command{
	// 	Name: "awk",
	// 	Args: []string{`'/\[Peer\]/{p=""} {p=p $0 ORS} END{printf "%s", p}'`, "/etc/wireguard/wg0.conf"},
	// }
	// latestPeerAddrRes, err := ssh.Run(ctx, getLatestPeerAddrCmd, nil)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to obtain latest peer's IP address: %s", err)
	// }
	// latestPeerAddr, err := netip.ParseAddr(latestPeerAddrRes.Stdout)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to parse latest peer address: %s", err)
	// }
	// newPeerAddr := latestPeerAddr.Next()
	// newPeerAddrWithMask, err := newPeerAddr.Prefix(32)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to add prefix to new peer's IP address: %s", err)
	// }

	// // Add peer to interface config
	// addPeerCmd := sshclient.Command{
	// 	Name: "wg",
	// 	Args: []string{
	// 		"set",
	// 		"wg0",
	// 		"peer",
	// 		pubKeyRes.Stdout,
	// 		"allowed-ips",
	// 		newPeerAddrWithMask.String(),
	// 		"endpoint",
	// 		fmt.Sprintf("%s:%s", ssh.Host, serverPortRes.Stdout),
	// 		"persistent-keepalive 25",
	// 	},
	// }
	// _, err = ssh.Run(ctx, addPeerCmd, nil)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to write new peer to the interface: %s", err)
	// }
	// writePeerCmd := sshclient.Command{
	// 	Name: "tee",
	// 	Args: []string{"-a", "/etc/wireguard/wg0.conf", ">/dev/null"},
	// }
	// peerConfTemplateFile, err := os.ReadFile("./peer.conf")
	// peerConfTemplate, err := template.New("peer").Parse(string(peerConfTemplateFile))
	// if err != nil {
	// 	return "", fmt.Errorf("failed to parse peer config template: %s", err)
	// }
	// var sb strings.Builder
	// peerData := map[string]string{
	// 	"ClientPubKey": pubKeyRes.Stdout,
	// 	"ClientAddr": newPeerAddrWithMask.String(),
	// }
	// if err := peerConfTemplate.Execute(&sb, peerData); err != nil {
	// 	return "", fmt.Errorf("failed to build peer config from template: %s", err)
	// }
	// r := strings.NewReader(sb.String())
	// _, err = ssh.Run(ctx, writePeerCmd, r)

	// sb.Reset()

	// // Create config file
	// confTemplateFile, err := os.ReadFile("./client.conf")
	// if err != nil {
	// 	return "", fmt.Errorf("failed to read config template: %s", err)
	// }
	// confTemplate, err := template.New("config").Parse(string(confTemplateFile))
	// if err != nil {
	// 	return "", fmt.Errorf("failed to create config template: %s", err)
	// }
	// configData := map[string]string{
	// 	"ClientAddr":         newPeerAddr.String(),
	// 	"ClientPrivKey":      serverPubKeyRes.Stdout,
	// 	"ServerPubKey":       serverPubKeyRes.Stdout,
	// 	"ClientPresharedKey": presharedKeyRes.Stdout,
	// 	"ServerAddr":         ssh.Host,
	// 	"ServerPort":         serverPortRes.Stdout,
	// }
	// if err := confTemplate.Execute(&sb, configData); err != nil {
	// 	return "", fmt.Errorf("failed to build config from template: %s", err)
	// }

	// return sb.String(), nil
}

func (l *Awg) revoke(cg *ConfigGenerator) {}
