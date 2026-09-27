package browser_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/m1dugh/gocrawler/internal/crawler/engine/browser"
	"github.com/m1dugh/gocrawler/pkg/crawler/engine"
)

// requireChrome skips the test if no chrome/chromium binary is found on
// PATH, since LocalEngine needs one to spawn.
func requireChrome(t *testing.T) {
	t.Helper()
	for _, name := range []string{"google-chrome", "google-chrome-stable", "chromium", "chromium-browser"} {
		if _, err := exec.LookPath(name); err == nil {
			return
		}
	}
	t.Skip("no chrome/chromium binary found on PATH")
}

func TestLocalEngineFetch(t *testing.T) {
	requireChrome(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		fmt.Fprint(w, `<html><body><p>hello browser</p></body></html>`)
	}))
	defer srv.Close()

	e := browser.NewLocal(browser.LocalConfig{Headless: true, Timeout: 30 * time.Second})
	defer e.Close()

	content, status, err := e.Fetch(srv.URL)
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if status != http.StatusTeapot {
		t.Errorf("status = %d, want %d", status, http.StatusTeapot)
	}
	if !strings.Contains(content, "hello browser") {
		t.Errorf("content = %q, want it to contain %q", content, "hello browser")
	}
}

var _ engine.Engine = (*browser.LocalEngine)(nil)
