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

// awgMirroredParams are the [Interface] settings that shape the wire format.
// They are copied verbatim from the server's config: a client that disagrees on
// any of them cannot complete a handshake. A param the server doesn't define is
// left empty and omitted from the generated config.
var awgMirroredParams = []string{
	"Jc", "Jmin", "Jmax",
	"S1", "S2", "S3", "S4",
	"H1", "H2", "H3", "H4",
	"HeaderProtectionKey",
	"ContentPaddingAddition",
	"RandomTrailers",
	"DisableCookies",
}

// awgTimingDefaults are the randomized timers each side applies locally, so they
// need not match the server. The server's value wins when it defines one,
// otherwise these defaults are used.
var awgTimingDefaults = map[string]string{
	"RekeyAfterTime":       "100-120",
	"RekeyTimeout":         "3-7",
	"RejectAfterTime":      "150-180",
	"KeepaliveTimeout":     "5-15",
	"MaxHandshakeAttempts": "15-20",
	"PersistentKeepalive":  "25-35",
}

// awgServerParamsScript prints "Key=Value" for every recognised tuning param in
// the [Interface] section. It splits on the first "=" only, so base64 values
// such as HeaderProtectionKey keep their padding.
const awgServerParamsScript = `
/^\[Interface\]/ { f = 1; next }
/^\[/ { f = 0 }
f {
	p = index($0, "=")
	if (p == 0) next
	k = substr($0, 1, p - 1)
	v = substr($0, p + 1)
	gsub(/^[ \t]+|[ \t]+$/, "", k)
	gsub(/^[ \t]+|[ \t]+$/, "", v)
	if (k ~ /^(Jc|Jmin|Jmax|S1|S2|S3|S4|H1|H2|H3|H4|HeaderProtectionKey|ContentPaddingAddition|RandomTrailers|DisableCookies|RekeyAfterTime|RekeyTimeout|RejectAfterTime|KeepaliveTimeout|MaxHandshakeAttempts|PersistentKeepalive)$/) print k "=" v
}
`

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
	serverPubKey := strings.TrimSpace(output)

	getServerPortCmd := []string{"awg", "show", "awg0", "listen-port"}
	output, err = cli.Run(ctx, containerID, getServerPortCmd, []byte{})
	if err != nil {
		return "", fmt.Errorf("failed to get server's port: %s", err)
	}
	serverPort := strings.TrimSpace(output)

	// Obtain the AmneziaWG tuning params from the [Interface] section of the
	// server's config, since `awg show` doesn't expose them and the client must
	// match the server's values exactly.
	getServerParamsCmd := []string{
		"awk",
		awgServerParamsScript,
		"/opt/amnezia/awg/awg0.conf",
	}
	output, err = cli.Run(ctx, containerID, getServerParamsCmd, []byte{})
	if err != nil {
		return "", fmt.Errorf("failed to get server's tuning params: %s", err)
	}

	serverParams := make(map[string]string)
	for line := range strings.SplitSeq(strings.TrimSpace(output), "\n") {
		if line == "" {
			continue
		}

		kv := strings.SplitN(line, "=", 2)
		if len(kv) != 2 {
			continue
		}

		serverParams[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
	}

	// Get the latest added peer's IP. A server without peers yet has no
	// AllowedIPs, so fall back to the interface's own address.
	getLatestPeerAddrCmd := []string{
		"awk",
		`/^\[/{s=$0} s=="[Interface]" && /^[ \t]*Address/{addr=$0} s=="[Peer]" && /^[ \t]*AllowedIPs/{ip=$0} END{if(ip=="")ip=addr; sub(/.*= */,"",ip); sub(/[\/,].*/,"",ip); printf "%s", ip}`,
		"/opt/amnezia/awg/awg0.conf",
	}
	output, err = cli.Run(ctx, containerID, getLatestPeerAddrCmd, []byte{})
	if err != nil {
		return "", fmt.Errorf("failed to get latest peer's IP: %s", err)
	}

	latestPeerAddr, err := netip.ParseAddr(strings.TrimSpace(output))
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
		"persistent-keepalive",
		"25",
	}
	output, err = cli.Run(ctx, containerID, createNewPeerCmd, []byte{})
	if err != nil {
		return "", fmt.Errorf("failed to create new peer: %s", err)
	}

	var sb strings.Builder
	peerData := map[string]string{
		"ClientPubKey":  pubKey,
		"ClientPresKey": presKey,
		"ClientAddr":    newPeerAddr.String(),
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
	}
	for _, k := range awgMirroredParams {
		configData[k] = serverParams[k]
	}
	for k, def := range awgTimingDefaults {
		if v, ok := serverParams[k]; ok {
			configData[k] = v
			continue
		}

		configData[k] = def
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
	const script = `set -e
allowed="%s/32"
conf="/config/wg_confs/wg0.conf"

pubkey=$(wg show wg0 allowed-ips | awk -v ip="$allowed" '$2 == ip {print $1; exit}')
if [ -n "$pubkey" ]; then
	wg set wg0 peer "$pubkey" remove
fi

awk -v ip="$allowed" '
	function flush(  i) {
		for (i = 0; i < n; i++) {
			if (hit && buf[i] !~ /^#/) print "#" buf[i]; else print buf[i]
		}
		n = 0; hit = 0
	}
	/^#?\[/ { flush() }
	{ buf[n++] = $0 }
	$0 ~ /AllowedIPs/ && $NF == ip { hit = 1 }
	END { flush() }
' "$conf" > "$conf.tmp"
cat "$conf.tmp" > "$conf"
rm -f "$conf.tmp"
`

	suspendConfigCmd := []string{
		"sh", "-c",
		fmt.Sprintf(script, addr),
	}
	_, err := cli.Run(ctx, containerID, suspendConfigCmd, nil)
	if err != nil {
		return fmt.Errorf("suspending config on server: %w", err)
	}

	return nil
}

func (w *Awg) resume(
	ctx context.Context,
	cli *executor.ContainerExecutor,
	containerID, addr string,
) error {
	const script = `set -e
allowed="%s/32"
conf="/config/wg_confs/wg0.conf"

awk -v ip="$allowed" '
	function flush(  i) {
		for (i = 0; i < n; i++) { if (hit) sub(/^#/, "", buf[i]); print buf[i] }
		n = 0; hit = 0
	}
	/^#?\[/ { flush() }
	{ buf[n++] = $0 }
	$0 ~ /AllowedIPs/ && $NF == ip { hit = 1 }
	END { flush() }
' "$conf" > "$conf.tmp"
cat "$conf.tmp" > "$conf"
rm -f "$conf.tmp"

set -- $(awk -v ip="$allowed" '
	/^\[/ { pk = ""; psk = "" }
	$1 == "PublicKey" { pk = $NF }
	$1 == "PresharedKey" { psk = $NF }
	$1 == "AllowedIPs" && $NF == ip { print pk, psk; exit }
' "$conf")
[ -n "$1" ] || { echo "no peer for $allowed" >&2; exit 1; }

if [ -n "$2" ]; then
	printf '%%s\n' "$2" | wg set wg0 peer "$1" preshared-key /dev/stdin allowed-ips "$allowed"
else
	wg set wg0 peer "$1" allowed-ips "$allowed"
fi
`

	resumeConfigCmd := []string{
		"sh", "-c",
		fmt.Sprintf(script, addr),
	}
	_, err := cli.Run(ctx, containerID, resumeConfigCmd, nil)
	if err != nil {
		return fmt.Errorf("resuming config on server: %w", err)
	}

	return nil
}
