package proxy

import (
	"log/slog"
	"testing"
	"time"

	"github.com/den/nix-cache-discovery/internal/registry"
)

func testProxier() *Proxier {
	return New(registry.New(), "/nix/store", "", "self", nil, nil, false, time.Second, 0, "http", slog.Default())
}

func TestRecordSourceAndSources(t *testing.T) {
	proxier := testProxier()
	hash := "0123456789abcdef0123456789abcdef"

	proxier.recordSource(hash, "https://cache.nixos.org")

	sources := proxier.Sources()

	if sources[hash] != "https://cache.nixos.org" {
		t.Fatalf("expected recorded source, got %q", sources[hash])
	}
}

func TestSourcesDropsExpiredEntries(t *testing.T) {
	proxier := testProxier()
	proxier.PositiveTTL = -time.Second
	hash := "0123456789abcdef0123456789abcdef"

	proxier.recordSource(hash, "https://cache.nixos.org")

	if len(proxier.Sources()) != 0 {
		t.Fatal("expected expired source to be removed")
	}
}

func TestPeerSourceUsesHostname(t *testing.T) {
	peer := registry.Peer{
		URL:      "http://192.168.1.5:5055",
		Hostname: "buildbox",
	}

	if got := peerSource(peer); got != "http://buildbox:5055" {
		t.Fatalf("expected hostname source, got %q", got)
	}
}

func TestPeerSourceFallsBackToURL(t *testing.T) {
	peer := registry.Peer{URL: "http://192.168.1.5:5055"}

	if got := peerSource(peer); got != peer.URL {
		t.Fatalf("expected peer URL, got %q", got)
	}
}

func TestFallbackSourceTrimsTrailingSlash(t *testing.T) {
	if got := fallbackSource("https://cache.nixos.org/"); got != "https://cache.nixos.org" {
		t.Fatalf("unexpected fallback source %q", got)
	}
}
