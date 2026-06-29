package configgen

import (
	"context"
	"fmt"
	"net/netip"
	"os"
	"text/template"
	"strings"

	"manager/internal/sshclient"
)

type Awg struct{}

func (l *Awg) create(ctx context.Context, ssh sshclient.Client, cg *ConfigGenerator) (string, error) {
	// Generate keys via SSH
	genPrivKeyCmd := sshclient.Command{
		Name: "wg",
		Args: []string{"genkey"},
	}
	privKeyRes, err := ssh.Run(ctx, genPrivKeyCmd)
	if err != nil {
		return "", fmt.Errorf("failed private key generation: %s", err)
	}
	genPubKeyCmd := sshclient.Command{
		Name: "wg",
		Args: []string{fmt.Sprintf("%s | awg pubkey", privKeyRes.Stdout)},
	}
	pubKeyRes, err := ssh.Run(ctx, genPubKeyCmd)
	if err != nil {
		return "", fmt.Errorf("failed public key generation: %s", err)
	}
	genPresharedKeyCmd := sshclient.Command{
		Name: "wg",
		Args: []string{"genpsk"},
	}
	presharedKeyRes, err := ssh.Run(ctx, genPresharedKeyCmd)
	if err != nil {
		return "", fmt.Errorf("failed preshared key generation: %s", err)
	}

	// Obtain server's credentials
	getServerPubKeyCmd := sshclient.Command{
		Name: "wg",
		Args: []string{"show", "wg0", "public-key"},
	}
	serverPubKeyRes, err := ssh.Run(ctx, getServerPubKeyCmd)
	if err != nil {
		return "", fmt.Errorf("failed to obtain server's public key: %s", err)
	}
	getServerPortCmd := sshclient.Command{
		Name: "wg",
		Args: []string{"show", "wg0", "listen-port"},
	}
	serverPortRes, err := ssh.Run(ctx, getServerPortCmd)
	if err != nil {
		return "", fmt.Errorf("failed to obtain server's port: %s", err)
	}

	// Get the latest added peer's IP
	getLatestPeerAddrCmd := sshclient.Command{
		Name: "awk",
		Args: []string{`'/\[Peer\]/{p=""} {p=p $0 ORS} END{printf "%s", p}'`, "/etc/wireguard/wg0.conf"},
	}
	latestPeerAddrRes, err := ssh.Run(ctx, getLatestPeerAddrCmd)
	if err != nil {
		return "", fmt.Errorf("failed to obtain latest peer's IP address: %s", err)
	}
	latestPeerAddr, err := netip.ParseAddr(latestPeerAddrRes.Stdout)
	if err != nil {
		return "", fmt.Errorf("failed to parse latest peer address: %s", err)
	}
	newPeerAddr := latestPeerAddr.Next()

	// Read config template file
	confTemplateFile, err := os.ReadFile("/manager/awg/template.conf")
	if err != nil {
		return "", fmt.Errorf("failed to read config template: %s", err)
	}

	// TODO: Add a peer creation in interface config
	//
	// here
	//

	// Create config file
	var sb strings.Builder
	confTemplate, err := template.New("config").Parse(string(confTemplateFile))
	if err != nil {
		return "", fmt.Errorf("failed to create config template: %s", err)
	}
	configData := struct {
		ClientAddr         string
		ClientPrivKey      string
		ServerPubKey       string
		ClientPresharedKey string
		ServerAddr         string
		ServerPort         string
	}{
		ClientAddr:         newPeerAddr.String(),
		ClientPrivKey:      serverPubKeyRes.Stdout,
		ServerPubKey:       serverPubKeyRes.Stdout,
		ClientPresharedKey: presharedKeyRes.Stdout,
		ServerAddr:         ssh.Host,
		ServerPort:         serverPortRes.Stdout,
	}
	if err := confTemplate.Execute(&sb, configData); err != nil {
		return "", fmt.Errorf("failed to build config from template: %s", err)
	}

	return sb.String(), nil
}

func (l *Awg) revoke(cg *ConfigGenerator) {}
