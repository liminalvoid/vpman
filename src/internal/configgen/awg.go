package configgen

import (
	"context"
	"embed"
	"fmt"
	"net/netip"
	"strings"
	"text/template"

	"vpnmanager/internal/executor"
)

type Awg struct{}

//go:embed templates/awg/peer.conf templates/awg/client.conf
var awgTemplateFS embed.FS

func (w *Awg) create(
	ctx context.Context,
	cli *executor.ContainerExecutor,
	containerID string,
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
	output, err := cli.Run(ctx, containerID, keysGenerationCmd, []byte{})
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
	output, err = cli.Run(ctx, containerID, getServerPubKeyCmd, []byte{})
	if err != nil {
		return "", fmt.Errorf("failed to get server's public key: %s", err)
	}
	serverPubKey := output

	getServerPortCmd := []string{"awg", "show", "awg0", "listen-port"}
	output, err = cli.Run(ctx, containerID, getServerPortCmd, []byte{})
	if err != nil {
		return "", fmt.Errorf("failed to get server's port: %s", err)
	}
	serverPort := output

	// Obtain the AmneziaWG obfuscation params (Jc, Jmin, Jmax, S1-S4, H1-H4, I1-I5)
	// from the [Interface] section of the server's config, since `awg show` doesn't
	// expose them and the client must match the server's values exactly.
	getObfuscationParamsCmd := []string{
		"awk",
		`/^\[Interface\]/{f=1; next} /^\[/{f=0} f && match($0, /^(Jc|Jmin|Jmax|S1|S2|S3|S4|H1|H2|H3|H4|I1|I2|I3|I4|I5)[ \t]*=/) { split($0, kv, "="); k=kv[1]; v=kv[2]; gsub(/^[ \t]+|[ \t]+$/, "", k); gsub(/^[ \t]+|[ \t]+$/, "", v); print k"="v }`,
		"/opt/amnezia/awg/awg0.conf",
	}
	output, err = cli.Run(ctx, containerID, getObfuscationParamsCmd, []byte{})
	if err != nil {
		return "", fmt.Errorf("failed to get obfuscation params: %s", err)
	}

	obfuscationParams := make(map[string]string)
	for line := range strings.SplitSeq(strings.TrimSpace(output), "\n") {
		if line == "" {
			continue
		}

		kv := strings.SplitN(line, "=", 2)
		if len(kv) != 2 {
			continue
		}

		obfuscationParams[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
	}

	// Get the latest added peer's IP
	getLatestPeerAddrCmd := []string{
		"awk",
		`/\[Peer\]/{ip=""} /AllowedIPs/{ip=$0} END{sub(/.*= */,"",ip); sub(/\/.*/,"",ip); printf "%s", ip}`,
		"/opt/amnezia/awg/awg0.conf",
	}
	output, err = cli.Run(ctx, containerID, getLatestPeerAddrCmd, []byte{})
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
	output, err = cli.Run(ctx, containerID, createNewPeerCmd, []byte{})
	if err != nil {
		return "", fmt.Errorf("failed to create new peer: %s", err)
	}

	var sb strings.Builder
	peerData := map[string]string{
		"ClientPubKey":       pubKey,
		"ClientPresharedKey": presKey,
		"ClientAddr":         newPeerAddr.String(),
	}
	if err := peerConfTemplate.Execute(&sb, peerData); err != nil {
		return "", fmt.Errorf("failed to build peer config from template: %s", err)
	}

	writePeerCmd := []string{"tee", "-a", "/opt/amnezia/awg/awg0.conf"}
	_, err = cli.Run(ctx, containerID, writePeerCmd, []byte(sb.String()))
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
		"Jc":                 obfuscationParams["Jc"],
		"Jmin":               obfuscationParams["Jmin"],
		"Jmax":               obfuscationParams["Jmax"],
		"S1":                 obfuscationParams["S1"],
		"S2":                 obfuscationParams["S2"],
		"S3":                 obfuscationParams["S3"],
		"S4":                 obfuscationParams["S4"],
		"H1":                 obfuscationParams["H1"],
		"H2":                 obfuscationParams["H2"],
		"H3":                 obfuscationParams["H3"],
		"H4":                 obfuscationParams["H4"],
		"I1":                 obfuscationParams["I1"],
		"I2":                 obfuscationParams["I2"],
		"I3":                 obfuscationParams["I3"],
		"I4":                 obfuscationParams["I4"],
		"I5":                 obfuscationParams["I5"],
	}
	if err := confTemplate.Execute(&sb, configData); err != nil {
		return "", fmt.Errorf("failed to build config from template: %s", err)
	}

	return sb.String(), nil
}

func (w *Awg) revoke(
	ctx context.Context,
	cli *executor.ContainerExecutor,
	containerID string,
	addr string,
) error {
	const script = `set -e
allowed="%s/32"
conf="/config/wg_confs/wg0.conf"

pubkey=$(wg show wg0 allowed-ips | awk -v ip="$allowed" '$2 == ip {print $1; exit}')
if [ -n "$pubkey" ]; then
	wg set wg0 peer "$pubkey" remove
fi

awk -v ip="$allowed" '
	function flush(  i) {
		if (n > 0 && !drop) { for (i = 0; i < n; i++) print buf[i] }
		n = 0; drop = 0
	}
	/^\[/ { flush() }
	{ buf[n++] = $0 }
	$1 == "AllowedIPs" && $3 == ip { drop = 1 }
	END { flush() }
' "$conf" > "$conf.tmp"
cat "$conf.tmp" > "$conf"
rm -f "$conf.tmp"
`

	removeConfigCmd := []string{
		"sh", "-c",
		fmt.Sprintf(script, addr),
	}
	_, err := cli.Run(ctx, containerID, removeConfigCmd, nil)
	if err != nil {
		return fmt.Errorf("removing config from server: %w", err)
	}

	return nil
}

func (w *Awg) suspend(
	ctx context.Context,
	cli *executor.ContainerExecutor,
	containerID, addr string,
) error {
	return nil
}

func (w *Awg) resume(
	ctx context.Context,
	cli *executor.ContainerExecutor,
	containerID, addr string,
) error {
	return nil
}
