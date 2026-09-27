package http_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	engineiface "github.com/m1dugh/gocrawler/pkg/crawler/engine"

	httpengine "github.com/m1dugh/gocrawler/internal/crawler/engine/http"
)

var _ engineiface.Engine = (*httpengine.Engine)(nil)

func TestFetch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Custom"); got != "value" {
			t.Errorf("X-Custom header = %q, want %q", got, "value")
		}
		w.WriteHeader(http.StatusTeapot)
		_, _ = io.WriteString(w, "hello world")
	}))
	defer srv.Close()

	e := httpengine.New(httpengine.Config{
		Headers: map[string][]string{"X-Custom": {"value"}},
	})

	content, status, err := e.Fetch(srv.URL)
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if status != http.StatusTeapot {
		t.Errorf("status = %d, want %d", status, http.StatusTeapot)
	}
	if content != "hello world" {
		t.Errorf("content = %q, want %q", content, "hello world")
	}
}

func TestFetchInvalidURL(t *testing.T) {
	e := httpengine.New(httpengine.Config{})

	if _, _, err := e.Fetch("://bad-url"); err == nil {
		t.Fatal("Fetch() with invalid url: expected error, got nil")
	}
}
