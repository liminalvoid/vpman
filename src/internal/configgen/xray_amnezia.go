package configgen

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

// AmneziaVPN's own name for its Xray container, and the local SOCKS port its
// client app points tun2socks at. Both are fixed by the app.
const (
	amneziaXrayContainer = "amnezia-xray"
	amneziaXraySocksPort = 10808

	amneziaDNS1 = "1.1.1.1"
	amneziaDNS2 = "1.0.0.1"
)

// XrayLinkToAmnezia turns a VLESS + Reality share link into the JSON AmneziaVPN
// exports for its self-hosted Xray servers: a container entry whose
// last_config is a complete Xray client config. Compress the result with
// lib.CompressConfigForAmnezia to get a vpn:// key.
func XrayLinkToAmnezia(link, description string) ([]byte, error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, fmt.Errorf("parsing link: %w", err)
	}
	if u.Scheme != "vless" {
		return nil, fmt.Errorf("unsupported link scheme %q", u.Scheme)
	}

	id := u.User.Username()
	host := u.Hostname()
	if id == "" || host == "" {
		return nil, errors.New("link has no client id or host")
	}

	port, err := strconv.Atoi(u.Port())
	if err != nil {
		return nil, fmt.Errorf("link has invalid port %q", u.Port())
	}

	q := u.Query()
	if q.Get("security") != "reality" {
		return nil, fmt.Errorf("unsupported link security %q", q.Get("security"))
	}

	network := q.Get("type")
	if network == "" {
		network = "tcp"
	}

	user := map[string]any{
		"id":         id,
		"encryption": "none",
	}
	if flow := q.Get("flow"); flow != "" {
		user["flow"] = flow
	}

	clientCfg := map[string]any{
		"log": map[string]any{"loglevel": "error"},
		"inbounds": []any{map[string]any{
			"listen":   "127.0.0.1",
			"port":     amneziaXraySocksPort,
			"protocol": "socks",
			"settings": map[string]any{"udp": true},
		}},
		"outbounds": []any{map[string]any{
			"protocol": "vless",
			"settings": map[string]any{
				"vnext": []any{map[string]any{
					"address": host,
					"port":    port,
					"users":   []any{user},
				}},
			},
			"streamSettings": map[string]any{
				"network":  network,
				"security": "reality",
				"realitySettings": map[string]any{
					"fingerprint": q.Get("fp"),
					"serverName":  q.Get("sni"),
					"publicKey":   q.Get("pbk"),
					"shortId":     q.Get("sid"),
					"spiderX":     q.Get("spx"),
				},
			},
		}},
	}

	lastConfig, err := json.MarshalIndent(clientCfg, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("encoding client config: %w", err)
	}

	amneziaCfg := map[string]any{
		"containers": []any{map[string]any{
			"container": amneziaXrayContainer,
			"xray": map[string]any{
				"last_config":     string(lastConfig),
				"port":            strconv.Itoa(port),
				"transport_proto": "tcp",
			},
		}},
		"defaultContainer": amneziaXrayContainer,
		"description":      description,
		"dns1":             amneziaDNS1,
		"dns2":             amneziaDNS2,
		"hostName":         host,
	}

	return json.Marshal(amneziaCfg)
}
