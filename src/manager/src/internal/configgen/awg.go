package configgen

import (
	"embed"
	"context"
	"fmt"
	"net/netip"
	// "os"
	"strings"
	"text/template"

	"manager/internal/executor"
)

type Awg struct{}

// TODO: Read configs directly from FS in runtime
//go:embed peer.conf client.conf
var templatesFS embed.FS

var (
	peerConfTemplate = template.Must(template.ParseFS(templatesFS, "peer.conf"))
	confTemplate     = template.Must(template.ParseFS(templatesFS, "client.conf"))
)

func (a *Awg) create(
	ctx context.Context,
	cli *executor.ContainerExecutor,
) (string, error) {
	// Generate keys
	keysGenerationCmd := []string{
		"sh", "-c",
		`privKey=$(wg genkey); pubKey=$(echo $privKey | wg pubkey); presKey=$(wg genpsk); echo "$privKey|$pubKey|$presKey"`,
	}
	output, err := cli.Run(ctx, keysGenerationCmd)
	if err != nil {
		return "", fmt.Errorf("failed keys generation: %s", err)
	}

	keys := strings.SplitN(strings.TrimSpace(output), "|", 3)
	if len(keys) != 3 {
		return "", fmt.Errorf("unexpected keys generation output: %q", output)
	}
	privKey, pubKey, presKey := keys[0], keys[1], keys[2]

	// Obtain server's credentials
	getServerPubKeyCmd := []string{"wg", "show", "wg0", "public-key"}
	output, err = cli.Run(ctx, getServerPubKeyCmd)
	if err != nil {
		return "", fmt.Errorf("failed to get server's public key: %s", err)
	}
	serverPubKey := output

	getServerPortCmd := []string{"wg", "show", "wg0", "listen-port"}
	output, err = cli.Run(ctx, getServerPortCmd)
	if err != nil {
		return "", fmt.Errorf("failed to get server's port: %s", err)
	}
	serverPort := output

	// Get the latest added peer's IP
	getLatestPeerAddrCmd := []string{
		"awk",
		`/\[Peer\]/{ip=""} /AllowedIPs/{ip=$0} END{sub(/.*= */,"",ip); sub(/\/.*/,"",ip); printf "%s", ip}`,
		"/config/wg_confs/wg0.conf",
	}
	output, err = cli.Run(ctx, getLatestPeerAddrCmd)
	if err != nil {
		return "", fmt.Errorf("failed to get latest peer's IP: %s", err)
	}

	latestPeerAddr, err := netip.ParseAddr(output)
	if err != nil {
		return "", fmt.Errorf("failed to parse latest peer's IP: %s", err)
	}
	newPeerAddr, err := latestPeerAddr.Next().Prefix(32)
	if err != nil {
		return "", fmt.Errorf("failed to add prefix to parsed IP: %s", err)
	}

	// Add peer to interface config
	createNewPeerCmd := []string{
		"wg",
		"set",
		"wg0",
		"peer",
		pubKey,
		"allowed-ips",
		newPeerAddr.String(),
		"endpoint",
		fmt.Sprintf("%s:%s", cli.HostAddr, serverPort),
		"persistent-keepalive 25",
	}
	output, err = cli.Run(ctx, createNewPeerCmd)
	if err != nil {
		return "", fmt.Errorf("failed to create new peer: %s", err)
	}

	writePeerCmd := []string{"tee", "-a", "/config/wg_confs/wg0.conf"}
	output, err = cli.Run(ctx, writePeerCmd)
	if err != nil {
		return "", fmt.Errorf("failed to write peer to server's config: %s", err)
	}

	// Parse peer template file
	// peerConfTemplateFile, err := os.ReadFile("peer.conf")
	// if err != nil {
	// 	return "", fmt.Errorf("failed to read peer config file: %s", err)
	// }
	// peerConfTemplate, err := template.New("peer").Parse(string(peerConfTemplateFile))
	// if err != nil {
	// 	return "", fmt.Errorf("failed to parse peer config template: %s", err)
	// }

	var sb strings.Builder
	peerData := map[string]string{
		"ClientPubKey": pubKey,
		"ClientAddr":   newPeerAddr.String(),
	}
	if err := peerConfTemplate.Execute(&sb, peerData); err != nil {
		return "", fmt.Errorf("failed to build peer config from template: %s", err)
	}
	_, err = cli.Run(ctx, writePeerCmd)

	sb.Reset()

	// Create config file
	// confTemplateFile, err := os.ReadFile("client.conf")
	// if err != nil {
	// 	return "", fmt.Errorf("failed to read config template: %s", err)
	// }
	// confTemplate, err := template.New("config").Parse(string(confTemplateFile))
	// if err != nil {
	// 	return "", fmt.Errorf("failed to create config template: %s", err)
	// }
	configData := map[string]string{
		"ClientAddr":         newPeerAddr.String(),
		"ClientPrivKey":      privKey,
		"ServerPubKey":       serverPubKey,
		"ClientPresharedKey": presKey,
		"ServerAddr":         "192.168.0.1",
		"ServerPort":         serverPort,
	}
	if err := confTemplate.Execute(&sb, configData); err != nil {
		return "", fmt.Errorf("failed to build config from template: %s", err)
	}

	return sb.String(), nil
}

func (a *Awg) revoke(cg *ConfigGenerator) {}
