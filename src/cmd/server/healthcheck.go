package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// runHealthcheck lets the binary probe itself: `/server healthcheck`. The
// production image is FROM scratch, so there is no curl and no shell for
// Docker's HEALTHCHECK to call — the server has to be its own client.
//
// It deliberately probes /healthz (liveness) rather than /readyz. /readyz
// pings Postgres, and killing the bot because the database is briefly
// unreachable only adds a restart loop on top of an outage it cannot fix.
func runHealthcheck() int {
	addr := os.Getenv("HEALTH_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	// ":8080" and "0.0.0.0:8080" are both listen addresses, not dialable hosts.
	if host, port, found := strings.Cut(addr, ":"); found && (host == "" || host == "0.0.0.0") {
		addr = "127.0.0.1:" + port
	}

	client := &http.Client{Timeout: 3 * time.Second}

	resp, err := client.Get("http://" + addr + "/healthz")
	if err != nil {
		fmt.Fprintf(os.Stderr, "healthcheck: %s\n", err)

		return 1
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "healthcheck: status %d\n", resp.StatusCode)

		return 1
	}

	return 0
}
