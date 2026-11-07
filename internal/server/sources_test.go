package server

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/den/nix-cache-discovery/internal/proxy"
	"github.com/den/nix-cache-discovery/internal/registry"
)

func testServer() *Server {
	return New(Options{
		Proxier: proxy.New(registry.New(), "/nix/store", "", "self", nil, nil, false, time.Second, 0, "http", slog.Default()),
	})
}

func TestSourcesRejectsNonLoopbackClients(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, sourcesPath, nil)
	request.RemoteAddr = "192.168.1.10:1234"
	response := httptest.NewRecorder()

	testServer().handleSources(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden, got %d", response.Code)
	}
}

func TestSourcesAllowsLoopbackClients(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, sourcesPath, nil)
	request.RemoteAddr = "127.0.0.1:1234"
	response := httptest.NewRecorder()

	testServer().handleSources(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected ok, got %d", response.Code)
	}
}

func TestHostnameIsReported(t *testing.T) {
	if testServer().hostname() == "" {
		t.Fatal("expected a hostname")
	}
}
