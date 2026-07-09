package configgen

import (
	"context"
	"embed"
	"fmt"
	"net/netip"
	"strings"
	"text/template"

	"manager/internal/executor"
)

type Awg struct{}

//go:embed templates/awg/peer.conf templates/awg/client.conf
var awgTemplateFS embed.FS

func (w *Awg) create(
	ctx context.Context,
	cli *executor.ContainerExecutor,
) (string, error) {
	// Read templates
	peerConfTemplate, err := template.ParseFS(awgTemplateFS, "templates/awg/peer.conf")
	if err != nil {
		return "", fmt.Errorf("failed to parse peer's config template: %s", err) 
	}
	confTemplate, err := template.ParseFS(awgTemplateFS, "templates/awg/client.conf")
	if err != nil {
		return "", fmt.Errorf("failed to parse client's config template: %s", err) 
	}

	// Generate keys
	keysGenerationCmd := []string{
		"sh", "-c",
		`privKey=$(awg genkey); pubKey=$(echo $privKey | awg pubkey); presKey=$(awg genpsk); echo "$privKey|$pubKey|$presKey"`,
	}
	output, err := cli.Run(ctx, keysGenerationCmd, []byte{})
	if err != nil {
		return "", fmt.Errorf("failed keys generation: %s", err)
	}

	keys := strings.SplitN(strings.TrimSpace(output), "|", 3)
	if len(keys) != 3 {
		return "", fmt.Errorf("unexpected keys generation output: %q", output)
	}
	privKey, pubKey, presKey := keys[0], keys[1], keys[2]

	// Obtain server's credentials
	getServerPubKeyCmd := []string{"awg", "show", "awg0", "public-key"}
	output, err = cli.Run(ctx, getServerPubKeyCmd, []byte{})
	if err != nil {
		return "", fmt.Errorf("failed to get server's public key: %s", err)
	}
	serverPubKey := output

	getServerPortCmd := []string{"awg", "show", "awg0", "listen-port"}
	output, err = cli.Run(ctx, getServerPortCmd, []byte{})
	if err != nil {
		return "", fmt.Errorf("failed to get server's port: %s", err)
	}
	serverPort := output

	// Get the latest added peer's IP
	getLatestPeerAddrCmd := []string{
		"awk",
		`/\[Peer\]/{ip=""} /AllowedIPs/{ip=$0} END{sub(/.*= */,"",ip); sub(/\/.*/,"",ip); printf "%s", ip}`,
		"/opt/amnezia/awg/awg0.conf",
	}
	output, err = cli.Run(ctx, getLatestPeerAddrCmd, []byte{})
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
		"awg",
		"set",
		"awg0",
		"peer",
		pubKey,
		"preshared-key",
		presKey,
		"allowed-ips",
		newPeerAddr.String(),
		"endpoint",
		fmt.Sprintf("%s:%s", cli.HostAddr, serverPort),
		"persistent-keepalive 25",
	}
	output, err = cli.Run(ctx, createNewPeerCmd, []byte{})
	if err != nil {
		return "", fmt.Errorf("failed to create new peer: %s", err)
	}

	var sb strings.Builder
	peerData := map[string]string{
		"ClientPubKey": pubKey,
		"ClientPresharedKey": presKey,
		"ClientAddr":   newPeerAddr.String(),
	}
	if err := peerConfTemplate.Execute(&sb, peerData); err != nil {
		return "", fmt.Errorf("failed to build peer config from template: %s", err)
	}

	writePeerCmd := []string{"tee", "-a", "/opt/amnezia/awg/awg0.conf"}
	_, err = cli.Run(ctx, writePeerCmd, []byte(sb.String()))
	if err != nil {
		return "", fmt.Errorf("failed to write peer to server's config: %s", err)
	}

	sb.Reset()

	configData := map[string]string{
		"ClientAddr":         newPeerAddr.String(),
		"ClientPrivKey":      privKey,
		"ServerPubKey":       serverPubKey,
		"ClientPresharedKey": presKey,
		"ServerAddr":         cli.HostAddr,
		"ServerPort":         serverPort,
	}
	if err := confTemplate.Execute(&sb, configData); err != nil {
		return "", fmt.Errorf("failed to build config from template: %s", err)
	}

	return sb.String(), nil
}

func (w *Awg) revoke(cg *ConfigGenerator) {}
