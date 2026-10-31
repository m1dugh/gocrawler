// Package transform provides the built-in transformers of the result
// pipeline. Each one is small and single-purpose: it provides exactly one
// field, named after itself, and reads straight from the raw crawl result.
package transform

import (
	"strings"

	"github.com/m1dugh/gocrawler/internal/crawler/hash"
	"github.com/m1dugh/gocrawler/pkg/crawler/engine"
	"github.com/m1dugh/gocrawler/pkg/crawler/output"
)

// single is a transformer providing one field computed from the raw crawl
// result alone.
type single struct {
	name string
	fn   func(engine.CrawlResult) interface{}
}

func (s single) Name() string       { return s.name }
func (s single) Provides() []string { return []string{s.name} }
func (s single) Requires() []string { return nil }

func (s single) Transform(r engine.CrawlResult, _ engine.Record) (engine.Record, error) {
	return engine.Record{s.name: s.fn(r)}, nil
}

// FullURL provides "full_url": the url as crawled.
func FullURL() output.Transformer {
	return single{"full_url", func(r engine.CrawlResult) interface{} { return r.URL }}
}

// PartialURL provides "partial_url": the url without its query string.
func PartialURL() output.Transformer {
	return single{"partial_url", func(r engine.CrawlResult) interface{} {
		partialURL, _, _ := strings.Cut(r.URL, "?")
		return partialURL
	}}
}

// ContentHash provides "content_hash": the hash of the page content.
func ContentHash() output.Transformer {
	return single{"content_hash", func(r engine.CrawlResult) interface{} { return hash.Content(r.Content) }}
}

// Content provides "content": the full page content.
func Content() output.Transformer {
	return single{"content", func(r engine.CrawlResult) interface{} { return r.Content }}
}

// Status provides "status": the http status code.
func Status() output.Transformer {
	return single{"status", func(r engine.CrawlResult) interface{} { return r.Status }}
}

// ContentLength provides "content_length": the length of the content.
func ContentLength() output.Transformer {
	return single{"content_length", func(r engine.CrawlResult) interface{} { return len(r.Content) }}
}

// Timestamp provides "timestamp": the time the page was fetched.
func Timestamp() output.Transformer {
	return single{"timestamp", func(r engine.CrawlResult) interface{} { return r.Timestamp }}
}
