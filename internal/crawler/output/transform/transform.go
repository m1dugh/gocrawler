// Package transform implements the built-in output.Transformer steps.
package transform

import (
	"strings"

	"github.com/m1dugh/gocrawler/internal/crawler/hash"
	"github.com/m1dugh/gocrawler/pkg/crawler/engine"
)

// FullURL is a Transformer that contributes the full crawled url.
type FullURL struct{}

// ID returns "full_url".
func (FullURL) ID() string { return "full_url" }

// Transform returns r.URL.
func (FullURL) Transform(r engine.CrawlResult, _ engine.Record) (interface{}, error) {
	return r.URL, nil
}

// PartialURL is a Transformer that contributes the crawled url without
// its query string.
type PartialURL struct{}

// ID returns "partial_url".
func (PartialURL) ID() string { return "partial_url" }

// Transform returns r.URL with any query string removed.
func (PartialURL) Transform(r engine.CrawlResult, _ engine.Record) (interface{}, error) {
	partialURL, _, _ := strings.Cut(r.URL, "?")
	return partialURL, nil
}

// ContentHash is a Transformer that contributes the hash of the crawled
// content.
type ContentHash struct{}

// ID returns "content_hash".
func (ContentHash) ID() string { return "content_hash" }

// Transform returns the hash of r.Content.
func (ContentHash) Transform(r engine.CrawlResult, _ engine.Record) (interface{}, error) {
	return hash.Content(r.Content), nil
}

// Content is a Transformer that contributes the full crawled content.
type Content struct{}

// ID returns "content".
func (Content) ID() string { return "content" }

// Transform returns r.Content.
func (Content) Transform(r engine.CrawlResult, _ engine.Record) (interface{}, error) {
	return r.Content, nil
}

// Status is a Transformer that contributes the crawled http status code.
type Status struct{}

// ID returns "status".
func (Status) ID() string { return "status" }

// Transform returns r.Status.
func (Status) Transform(r engine.CrawlResult, _ engine.Record) (interface{}, error) {
	return r.Status, nil
}

// ContentLength is a Transformer that contributes the length of the
// crawled content.
type ContentLength struct{}

// ID returns "content_length".
func (ContentLength) ID() string { return "content_length" }

// Transform returns len(r.Content).
func (ContentLength) Transform(r engine.CrawlResult, _ engine.Record) (interface{}, error) {
	return len(r.Content), nil
}

// Timestamp is a Transformer that contributes the crawl timestamp.
type Timestamp struct{}

// ID returns "timestamp".
func (Timestamp) ID() string { return "timestamp" }

// Transform returns r.Timestamp.
func (Timestamp) Transform(r engine.CrawlResult, _ engine.Record) (interface{}, error) {
	return r.Timestamp, nil
}
