package proxy

import (
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/den/nix-cache-discovery/internal/registry"
)

type sourceEntry struct {
	url     string
	expires time.Time
}

func (proxier *Proxier) recordSource(hashPart, source string) {
	if source == "" {
		return
	}

	proxier.mu.Lock()
	defer proxier.mu.Unlock()

	now := time.Now()

	if proxier.sources == nil {
		proxier.sources = make(map[string]sourceEntry)
	}

	for hash, entry := range proxier.sources {
		if now.After(entry.expires) {
			delete(proxier.sources, hash)
		}
	}

	proxier.sources[hashPart] = sourceEntry{url: source, expires: now.Add(proxier.PositiveTTL)}
}

func (proxier *Proxier) Sources() map[string]string {
	proxier.mu.Lock()
	defer proxier.mu.Unlock()

	now := time.Now()
	sources := make(map[string]string)

	for hash, entry := range proxier.sources {
		if now.After(entry.expires) {
			delete(proxier.sources, hash)
			continue
		}

		sources[hash] = entry.url
	}

	return sources
}

func peerSource(peer registry.Peer) string {
	parsed, err := url.Parse(peer.URL)

	if err != nil || peer.Hostname == "" {
		return peer.URL
	}

	if port := parsed.Port(); port != "" {
		parsed.Host = net.JoinHostPort(peer.Hostname, port)
	} else {
		parsed.Host = peer.Hostname
	}

	return parsed.String()
}

func fallbackSource(fallback string) string {
	return strings.TrimSuffix(fallback, "/")
}
