package transform_test

import (
	"testing"
	"time"

	"github.com/m1dugh/gocrawler/internal/crawler/hash"
	"github.com/m1dugh/gocrawler/internal/crawler/output/transform"
	"github.com/m1dugh/gocrawler/pkg/crawler/engine"
	"github.com/m1dugh/gocrawler/pkg/crawler/output"
)

var ts = time.Unix(1000, 0)

func TestTransformers(t *testing.T) {
	r := engine.CrawlResult{
		URL:       "https://example.com/path?q=1",
		Content:   "hello world",
		Status:    404,
		Timestamp: ts,
	}

	tests := []struct {
		name string
		new  func() output.Transformer
		want interface{}
	}{
		{"full_url", transform.FullURL, r.URL},
		{"partial_url", transform.PartialURL, "https://example.com/path"},
		{"content_hash", transform.ContentHash, hash.Content(r.Content)},
		{"content", transform.Content, r.Content},
		{"status", transform.Status, 404},
		{"content_length", transform.ContentLength, len(r.Content)},
		{"timestamp", transform.Timestamp, ts},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := tt.new()
			if tr.Name() != tt.name {
				t.Errorf("Name() = %q, want %q", tr.Name(), tt.name)
			}
			if p := tr.Provides(); len(p) != 1 || p[0] != tt.name {
				t.Errorf("Provides() = %v, want [%s]", p, tt.name)
			}
			if len(tr.Requires()) != 0 {
				t.Errorf("Requires() = %v, want none", tr.Requires())
			}

			got, err := tr.Transform(r, engine.Record{})
			if err != nil {
				t.Fatalf("Transform() error = %v", err)
			}
			if len(got) != 1 || got[tt.name] != tt.want {
				t.Errorf("Transform() = %v, want {%s: %v}", got, tt.name, tt.want)
			}
		})
	}
}

func TestPartialURLWithoutQuery(t *testing.T) {
	got, err := transform.PartialURL().Transform(engine.CrawlResult{URL: "https://example.com/path"}, engine.Record{})
	if err != nil {
		t.Fatalf("Transform() error = %v", err)
	}
	if got["partial_url"] != "https://example.com/path" {
		t.Errorf("Transform() = %v", got)
	}
}
