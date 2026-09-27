package browser_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/m1dugh/gocrawler/internal/crawler/engine/browser"
	"github.com/m1dugh/gocrawler/pkg/crawler/engine"
)

// requireRemoteBrowser skips the test unless GOCRAWLER_TEST_CDP_URL is
// set to a reachable chrome/chromium remote-debugging websocket
// address. There's no way to discover or spawn one portably from within
// a test without also depending on a local chrome binary being present
// and its debugging port being predictable, so this is left to be
// provided by whoever runs the test against a real browser.
func requireRemoteBrowser(t *testing.T) string {
	t.Helper()
	wsURL := os.Getenv("GOCRAWLER_TEST_CDP_URL")
	if wsURL == "" {
		t.Skip("GOCRAWLER_TEST_CDP_URL not set; skipping RemoteEngine test")
	}
	return wsURL
}

func TestRemoteEngineFetch(t *testing.T) {
	wsURL := requireRemoteBrowser(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<html><body><p>hello remote browser</p></body></html>`))
	}))
	defer srv.Close()

	e, err := browser.NewRemote(browser.RemoteConfig{WebsocketURL: wsURL, Timeout: 30 * time.Second})
	if err != nil {
		t.Fatalf("NewRemote() error = %v", err)
	}
	defer e.Close()

	content, status, err := e.Fetch(srv.URL)
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if status != http.StatusOK {
		t.Errorf("status = %d, want %d", status, http.StatusOK)
	}
	if !strings.Contains(content, "hello remote browser") {
		t.Errorf("content = %q, want it to contain %q", content, "hello remote browser")
	}
}

var _ engine.Engine = (*browser.RemoteEngine)(nil)
