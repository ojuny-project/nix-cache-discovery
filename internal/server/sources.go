package server

import (
	"fmt"
	"net/http"
	"net/netip"
	"os"
	"sort"
)

const sourcesPath = "/.nix-cache-discovery/v1/sources"

func (server *Server) handleSources(w http.ResponseWriter, r *http.Request) {
	address, err := netip.ParseAddrPort(r.RemoteAddr)

	if err != nil || !address.Addr().Unmap().IsLoopback() {
		http.Error(w, "Forbidden.\n", http.StatusForbidden)
		return
	}

	sources := server.options.Proxier.Sources()
	hashes := make([]string, 0, len(sources))

	for hash := range sources {
		hashes = append(hashes, hash)
	}

	sort.Strings(hashes)

	w.Header().Set("Content-Type", "text/plain")

	for _, hash := range hashes {
		fmt.Fprintf(w, "%s %s\n", hash, sources[hash])
	}
}

func (server *Server) hostname() string {
	hostname, err := os.Hostname()

	if err != nil {
		return ""
	}

	return hostname
}
