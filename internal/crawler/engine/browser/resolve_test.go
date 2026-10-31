package browser

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestResolveWebsocketURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/json/version" {
			http.NotFound(w, r)
			return
		}
		// The browser advertises its own internal address.
		fmt.Fprint(w, `{"webSocketDebuggerUrl":"ws://127.0.0.1:9222/devtools/browser/abc"}`)
	}))
	defer srv.Close()
	host := strings.TrimPrefix(srv.URL, "http://")

	got, err := resolveWebsocketURL(srv.URL)
	if err != nil {
		t.Fatalf("resolveWebsocketURL() error = %v", err)
	}
	if want := "ws://" + host + "/devtools/browser/abc"; got != want {
		t.Errorf("resolveWebsocketURL() = %q, want %q (the host given by the user, not the advertised one)", got, want)
	}

	got, err = resolveWebsocketURL(srv.URL + "/")
	if err != nil || !strings.HasSuffix(got, "/devtools/browser/abc") {
		t.Errorf("with a trailing slash: %q, %v", got, err)
	}

	if got, _ := resolveWebsocketURL("wss://example.com/devtools/browser/x"); got != "wss://example.com/devtools/browser/x" {
		t.Errorf("a ws url must be used as is, got %q", got)
	}
}

func TestResolveWebsocketURLErrors(t *testing.T) {
	notChrome := httptest.NewServer(http.NotFoundHandler())
	defer notChrome.Close()
	empty := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, `{}`) }))
	defer empty.Close()

	for name, endpoint := range map[string]string{
		"not a debugging endpoint": notChrome.URL,
		"no websocket url":         empty.URL,
		"unsupported scheme":       "ftp://example.com",
		"unreachable":              "http://127.0.0.1:1",
	} {
		if _, err := resolveWebsocketURL(endpoint); err == nil {
			t.Errorf("%s: expected error, got nil", name)
		}
	}
}
