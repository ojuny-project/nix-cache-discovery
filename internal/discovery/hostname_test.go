package discovery

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/den/nix-cache-discovery/internal/registry"
)

func TestProbeReadsAdvertisedHostname(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != InfoPath {
			http.NotFound(w, r)
			return
		}

		fmt.Fprintln(w, "id=peer")
		fmt.Fprintln(w, "version=1")
		fmt.Fprintln(w, "port=5000")
		fmt.Fprintln(w, "hostname=buildbox")
	}))
	defer upstream.Close()

	parsed, err := url.Parse(upstream.URL)

	if err != nil {
		t.Fatal(err)
	}

	port, err := strconv.Atoi(parsed.Port())

	if err != nil {
		t.Fatal(err)
	}

	prober := New(nil, port, time.Second, 1, "http", "self", registry.New(), slog.Default())
	prober.client = upstream.Client()

	peer, err := prober.probeOnce(context.Background(), netip.MustParseAddr(parsed.Hostname()))

	if err != nil {
		t.Fatal(err)
	}

	if peer.Hostname != "buildbox" {
		t.Fatalf("expected advertised hostname, got %q", peer.Hostname)
	}
}
