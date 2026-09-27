package transform_test

import (
	"testing"
	"time"

	"github.com/m1dugh/gocrawler/internal/crawler/hash"
	"github.com/m1dugh/gocrawler/internal/crawler/output/transform"
	"github.com/m1dugh/gocrawler/pkg/crawler/engine"
)

func TestFullURL(t *testing.T) {
	r := engine.CrawlResult{URL: "https://example.com/path?q=1"}

	fu := transform.FullURL{}

	v, err := fu.Transform(r, engine.Record{})
	if err != nil {
		t.Fatalf("Transform() error = %v", err)
	}
	if v != r.URL {
		t.Errorf("Transform() = %v, want %q", v, r.URL)
	}
	if fu.ID() != "full_url" {
		t.Errorf("ID() = %q, want %q", fu.ID(), "full_url")
	}
}

func TestPartialURL(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"https://example.com/path?q=1", "https://example.com/path"},
		{"https://example.com/path", "https://example.com/path"},
	}

	for _, tt := range tests {
		v, err := transform.PartialURL{}.Transform(engine.CrawlResult{URL: tt.url}, engine.Record{})
		if err != nil {
			t.Fatalf("Transform() error = %v", err)
		}
		if v != tt.want {
			t.Errorf("Transform(%q) = %v, want %q", tt.url, v, tt.want)
		}
	}
}

func TestContentHash(t *testing.T) {
	r := engine.CrawlResult{Content: "hello world"}

	v, err := transform.ContentHash{}.Transform(r, engine.Record{})
	if err != nil {
		t.Fatalf("Transform() error = %v", err)
	}
	if want := hash.Content(r.Content); v != want {
		t.Errorf("Transform() = %v, want %q", v, want)
	}
}

func TestContent(t *testing.T) {
	r := engine.CrawlResult{Content: "hello world"}

	v, err := transform.Content{}.Transform(r, engine.Record{})
	if err != nil {
		t.Fatalf("Transform() error = %v", err)
	}
	if v != r.Content {
		t.Errorf("Transform() = %v, want %q", v, r.Content)
	}
}

func TestStatus(t *testing.T) {
	r := engine.CrawlResult{Status: 404}

	v, err := transform.Status{}.Transform(r, engine.Record{})
	if err != nil {
		t.Fatalf("Transform() error = %v", err)
	}
	if v != r.Status {
		t.Errorf("Transform() = %v, want %d", v, r.Status)
	}
}

func TestContentLength(t *testing.T) {
	r := engine.CrawlResult{Content: "hello world"}

	v, err := transform.ContentLength{}.Transform(r, engine.Record{})
	if err != nil {
		t.Fatalf("Transform() error = %v", err)
	}
	if v != len(r.Content) {
		t.Errorf("Transform() = %v, want %d", v, len(r.Content))
	}
}

func TestTimestamp(t *testing.T) {
	ts := time.Unix(1000, 0)
	r := engine.CrawlResult{Timestamp: ts}

	v, err := transform.Timestamp{}.Transform(r, engine.Record{})
	if err != nil {
		t.Fatalf("Transform() error = %v", err)
	}
	got, ok := v.(time.Time)
	if !ok || !got.Equal(ts) {
		t.Errorf("Transform() = %v, want %v", v, ts)
	}
}
