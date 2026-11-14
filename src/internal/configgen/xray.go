package configgen

import (
	"context"
	"crypto/ecdh"
	"crypto/rand"
	"embed"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"text/template"

	"github.com/google/uuid"

	"vpnmanager/internal/executor"
)

type Xray struct{}

//go:embed templates/xray/inbound_client.json templates/xray/client.link
var xrayTemplateFS embed.FS

const (
	xrayConfigPath    = "/opt/xray/config.json"
	xrayStagingPath   = "/opt/xray/config.staging.json"
	xraySuspendedPath = "/opt/xray/suspended.json"

	xrayVisionFlow         = "xtls-rprx-vision"
	xrayDefaultFingerprint = "chrome"

	// xrayEmailBytes is sized so the hex-encoded email fits the 15-char
	// configs.ip_address column, which stores the client's identifier.
	xrayEmailBytes = 7
)

// xrayInstallScript validates the staged config with Xray itself before
// installing it, so a bad edit never replaces a working config. The file is
// overwritten in place rather than moved, keeping bind-mounted inodes intact.
const xrayInstallScript = `set -e
staging="$1"
conf="$2"
trap 'rm -f "$staging"' EXIT
cat > "$staging"
xray run -test -c "$staging" >&2
cat "$staging" > "$conf"
`

// xrayLocks serializes read-modify-write cycles per container. Xray keeps all
// clients in one JSON file, so concurrent edits would otherwise lose writes.
var xrayLocks sync.Map

func lockXray(cli *executor.ContainerExecutor, containerID string) func() {
	m, _ := xrayLocks.LoadOrStore(cli.HostAddr+"/"+containerID, &sync.Mutex{})
	mu := m.(*sync.Mutex)
	mu.Lock()

	return mu.Unlock
}

// xrayReality holds what a client needs to reach a VLESS + Reality inbound.
type xrayReality struct {
	Port       string
	Network    string
	ServerName string
	PublicKey  string
	ShortID    string
	Flow       string
}

func (x *Xray) create(
	ctx context.Context,
	cli *executor.ContainerExecutor,
	containerID, serverPublicAddr string,
) (string, string, error) {
	clientTemplate, err := template.ParseFS(xrayTemplateFS, "templates/xray/inbound_client.json")
	if err != nil {
		return "", "", fmt.Errorf("failed to parse inbound client template: %s", err)
	}
	linkTemplate, err := template.ParseFS(xrayTemplateFS, "templates/xray/client.link")
	if err != nil {
		return "", "", fmt.Errorf("failed to parse client link template: %s", err)
	}

	defer lockXray(cli, containerID)()

	cfg, err := loadXrayJSON(ctx, cli, containerID, xrayConfigPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to read server's config: %w", err)
	}
	if cfg == nil {
		return "", "", fmt.Errorf("server's config %s is empty", xrayConfigPath)
	}

	inbound, err := findRealityInbound(cfg)
	if err != nil {
		return "", "", err
	}

	reality, err := parseReality(inbound)
	if err != nil {
		return "", "", fmt.Errorf("failed to read reality settings: %w", err)
	}

	clients, err := inboundClients(inbound)
	if err != nil {
		return "", "", err
	}

	email, err := newXrayEmail(clients)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate client's email: %w", err)
	}
	clientID := uuid.NewString()

	// Build the new client entry and add it to the inbound
	var sb strings.Builder
	clientData := map[string]string{
		"ClientID": clientID,
		"Email":    email,
		"Flow":     reality.Flow,
	}
	if err := clientTemplate.Execute(&sb, clientData); err != nil {
		return "", "", fmt.Errorf("failed to build inbound client from template: %s", err)
	}

	var client map[string]any
	if err := json.Unmarshal([]byte(sb.String()), &client); err != nil {
		return "", "", fmt.Errorf("failed to decode inbound client: %s", err)
	}

	setInboundClients(inbound, append(clients, client))

	if err := saveXrayConfig(ctx, cli, containerID, cfg); err != nil {
		return "", "", fmt.Errorf("failed to write client to server's config: %w", err)
	}

	sb.Reset()

	linkData := map[string]string{
		"ClientID":    clientID,
		"Endpoint":    net.JoinHostPort(serverPublicAddr, reality.Port),
		"Network":     reality.Network,
		"Flow":        reality.Flow,
		"ServerName":  reality.ServerName,
		"Fingerprint": xrayDefaultFingerprint,
		"PublicKey":   reality.PublicKey,
		"ShortID":     reality.ShortID,
		"Remark":      serverPublicAddr,
	}
	if err := linkTemplate.Execute(&sb, linkData); err != nil {
		return "", "", fmt.Errorf("failed to build client link from template: %s", err)
	}

	return strings.TrimSpace(sb.String()), email, nil
}

func (x *Xray) revoke(
	ctx context.Context,
	cli *executor.ContainerExecutor,
	containerID string,
	email string,
) error {
	defer lockXray(cli, containerID)()

	cfg, inbound, clients, err := loadRealityClients(ctx, cli, containerID)
	if err != nil {
		return fmt.Errorf("removing config from server: %w", err)
	}

	if rest, removed := removeXrayClient(clients, email); removed != nil {
		setInboundClients(inbound, rest)
		if err := saveXrayConfig(ctx, cli, containerID, cfg); err != nil {
			return fmt.Errorf("removing config from server: %w", err)
		}
	}

	// A suspended client lives only in the side file, so drop it from there too.
	suspended, err := loadSuspended(ctx, cli, containerID)
	if err != nil {
		return fmt.Errorf("removing suspended config: %w", err)
	}
	if _, ok := suspended[email]; ok {
		delete(suspended, email)
		if err := writeXrayJSON(ctx, cli, containerID, xraySuspendedPath, suspended); err != nil {
			return fmt.Errorf("removing suspended config: %w", err)
		}
	}

	return nil
}

// suspend moves the client out of the inbound into a side file, keeping its
// UUID so the same link works again after resume.
func (x *Xray) suspend(
	ctx context.Context,
	cli *executor.ContainerExecutor,
	containerID, email string,
) error {
	defer lockXray(cli, containerID)()

	cfg, inbound, clients, err := loadRealityClients(ctx, cli, containerID)
	if err != nil {
		return fmt.Errorf("suspending config on server: %w", err)
	}

	rest, removed := removeXrayClient(clients, email)
	if removed == nil {
		suspended, err := loadSuspended(ctx, cli, containerID)
		if err != nil {
			return fmt.Errorf("suspending config on server: %w", err)
		}
		if _, ok := suspended[email]; ok {
			return nil
		}

		return fmt.Errorf("no client for %q", email)
	}

	suspended, err := loadSuspended(ctx, cli, containerID)
	if err != nil {
		return fmt.Errorf("suspending config on server: %w", err)
	}
	suspended[email] = removed

	// Persist the client before dropping it from the inbound, so a failure
	// in between never loses its credentials.
	if err := writeXrayJSON(ctx, cli, containerID, xraySuspendedPath, suspended); err != nil {
		return fmt.Errorf("saving suspended config: %w", err)
	}

	setInboundClients(inbound, rest)
	if err := saveXrayConfig(ctx, cli, containerID, cfg); err != nil {
		return fmt.Errorf("suspending config on server: %w", err)
	}

	return nil
}

func (x *Xray) resume(
	ctx context.Context,
	cli *executor.ContainerExecutor,
	containerID, email string,
) error {
	defer lockXray(cli, containerID)()

	suspended, err := loadSuspended(ctx, cli, containerID)
	if err != nil {
		return fmt.Errorf("resuming config on server: %w", err)
	}

	client, ok := suspended[email]
	if !ok {
		return fmt.Errorf("no suspended client for %q", email)
	}

	cfg, inbound, clients, err := loadRealityClients(ctx, cli, containerID)
	if err != nil {
		return fmt.Errorf("resuming config on server: %w", err)
	}

	if _, existing := removeXrayClient(clients, email); existing == nil {
		setInboundClients(inbound, append(clients, client))
		if err := saveXrayConfig(ctx, cli, containerID, cfg); err != nil {
			return fmt.Errorf("resuming config on server: %w", err)
		}
	}

	delete(suspended, email)
	if err := writeXrayJSON(ctx, cli, containerID, xraySuspendedPath, suspended); err != nil {
		return fmt.Errorf("clearing suspended config: %w", err)
	}

	return nil
}

func loadRealityClients(
	ctx context.Context,
	cli *executor.ContainerExecutor,
	containerID string,
) (map[string]any, map[string]any, []any, error) {
	cfg, err := loadXrayJSON(ctx, cli, containerID, xrayConfigPath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("reading server's config: %w", err)
	}
	if cfg == nil {
		return nil, nil, nil, fmt.Errorf("server's config %s is empty", xrayConfigPath)
	}

	inbound, err := findRealityInbound(cfg)
	if err != nil {
		return nil, nil, nil, err
	}

	clients, err := inboundClients(inbound)
	if err != nil {
		return nil, nil, nil, err
	}

	return cfg, inbound, clients, nil
}

// loadXrayJSON reads a JSON file from the container. A missing or empty file
// yields a nil map. Numbers are kept as json.Number so re-encoding doesn't
// alter them.
func loadXrayJSON(
	ctx context.Context,
	cli *executor.ContainerExecutor,
	containerID, path string,
) (map[string]any, error) {
	readCmd := []string{"sh", "-c", `[ ! -f "$1" ] || cat "$1"`, "sh", path}
	output, err := cli.Run(ctx, containerID, readCmd, nil)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(output) == "" {
		return nil, nil
	}

	dec := json.NewDecoder(strings.NewReader(output))
	dec.UseNumber()

	var v map[string]any
	if err := dec.Decode(&v); err != nil {
		return nil, fmt.Errorf("decoding %s: %w", path, err)
	}

	return v, nil
}

func writeXrayJSON(
	ctx context.Context,
	cli *executor.ContainerExecutor,
	containerID, path string,
	v any,
) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding %s: %w", path, err)
	}

	writeCmd := []string{"sh", "-c", `cat > "$1"`, "sh", path}
	if _, err := cli.Run(ctx, containerID, writeCmd, append(data, '\n')); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}

	return nil
}

// saveXrayConfig validates and installs the server's config, then restarts the
// container: Xray reads its config only at startup.
func saveXrayConfig(
	ctx context.Context,
	cli *executor.ContainerExecutor,
	containerID string,
	cfg map[string]any,
) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding server's config: %w", err)
	}

	installCmd := []string{"sh", "-c", xrayInstallScript, "sh", xrayStagingPath, xrayConfigPath}
	if _, err := cli.Run(ctx, containerID, installCmd, append(data, '\n')); err != nil {
		return fmt.Errorf("installing server's config: %w", err)
	}

	if err := cli.Restart(ctx, containerID); err != nil {
		return fmt.Errorf("applying server's config: %w", err)
	}

	return nil
}

func loadSuspended(
	ctx context.Context,
	cli *executor.ContainerExecutor,
	containerID string,
) (map[string]any, error) {
	suspended, err := loadXrayJSON(ctx, cli, containerID, xraySuspendedPath)
	if err != nil {
		return nil, err
	}
	if suspended == nil {
		suspended = make(map[string]any)
	}

	return suspended, nil
}

// findRealityInbound returns the first VLESS inbound secured with Reality.
func findRealityInbound(cfg map[string]any) (map[string]any, error) {
	inbounds, _ := cfg["inbounds"].([]any)
	for _, raw := range inbounds {
		inbound, ok := raw.(map[string]any)
		if !ok || inbound["protocol"] != "vless" {
			continue
		}

		stream, _ := inbound["streamSettings"].(map[string]any)
		if stream != nil && stream["security"] == "reality" {
			return inbound, nil
		}
	}

	return nil, errors.New("no VLESS + Reality inbound in server's config")
}

func inboundClients(inbound map[string]any) ([]any, error) {
	settings, ok := inbound["settings"].(map[string]any)
	if !ok {
		return nil, errors.New("VLESS inbound has no settings")
	}

	switch clients := settings["clients"].(type) {
	case nil:
		return nil, nil
	case []any:
		return clients, nil
	default:
		return nil, fmt.Errorf("unexpected VLESS clients type %T", clients)
	}
}

func setInboundClients(inbound map[string]any, clients []any) {
	if clients == nil {
		clients = []any{}
	}

	inbound["settings"].(map[string]any)["clients"] = clients
}

// removeXrayClient returns the clients without the one tagged with email, and
// that client, or nil when there is none.
func removeXrayClient(clients []any, email string) ([]any, any) {
	for i, raw := range clients {
		client, ok := raw.(map[string]any)
		if ok && client["email"] == email {
			rest := append(clients[:i:i], clients[i+1:]...)
			return rest, client
		}
	}

	return clients, nil
}

func newXrayEmail(clients []any) (string, error) {
	taken := make(map[string]struct{}, len(clients))
	for _, raw := range clients {
		if client, ok := raw.(map[string]any); ok {
			if email, ok := client["email"].(string); ok {
				taken[email] = struct{}{}
			}
		}
	}

	buf := make([]byte, xrayEmailBytes)
	for range 8 {
		if _, err := rand.Read(buf); err != nil {
			return "", err
		}

		email := hex.EncodeToString(buf)
		if _, ok := taken[email]; !ok {
			return email, nil
		}
	}

	return "", errors.New("could not find a free email")
}

func parseReality(inbound map[string]any) (*xrayReality, error) {
	stream := inbound["streamSettings"].(map[string]any)
	settings, ok := stream["realitySettings"].(map[string]any)
	if !ok {
		return nil, errors.New("inbound has no realitySettings")
	}

	port := fmt.Sprint(inbound["port"])
	if port == "" || port == "<nil>" {
		return nil, errors.New("inbound has no port")
	}

	network, _ := stream["network"].(string)
	if network == "" {
		network = "tcp"
	}

	privKey, _ := settings["privateKey"].(string)
	pubKey, err := xrayPublicKey(privKey)
	if err != nil {
		return nil, fmt.Errorf("deriving public key: %w", err)
	}

	serverNames := stringList(settings["serverNames"])
	if len(serverNames) == 0 {
		return nil, errors.New("realitySettings has no serverNames")
	}

	// Prefer a non-empty short ID; an empty one is valid only when listed.
	shortIDs := stringList(settings["shortIds"])
	if len(shortIDs) == 0 {
		return nil, errors.New("realitySettings has no shortIds")
	}
	shortID := shortIDs[0]
	for _, id := range shortIDs {
		if id != "" {
			shortID = id
			break
		}
	}

	return &xrayReality{
		Port:       port,
		Network:    network,
		ServerName: serverNames[0],
		PublicKey:  pubKey,
		ShortID:    shortID,
		Flow:       xrayFlow(inbound, network),
	}, nil
}

// xrayFlow matches the flow of the clients already on the inbound, so a new
// client behaves like the rest. Vision is the default and only works over raw
// TCP.
func xrayFlow(inbound map[string]any, network string) string {
	clients, _ := inboundClients(inbound)
	for _, raw := range clients {
		if client, ok := raw.(map[string]any); ok {
			flow, _ := client["flow"].(string)
			return flow
		}
	}

	if network == "tcp" || network == "raw" {
		return xrayVisionFlow
	}

	return ""
}

// xrayPublicKey derives the Reality public key from the server's X25519
// private key, both in Xray's unpadded base64url encoding.
func xrayPublicKey(privKey string) (string, error) {
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(privKey, "="))
	if err != nil {
		return "", fmt.Errorf("decoding private key: %w", err)
	}

	key, err := ecdh.X25519().NewPrivateKey(raw)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(key.PublicKey().Bytes()), nil
}

func stringList(v any) []string {
	items, _ := v.([]any)
	out := make([]string, 0, len(items))
	for _, item := range items {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}

	return out
}
