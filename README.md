# gocrawler

A web crawler in Go, with support for basic (HTTP) and browser-based
crawling, scope-validated queueing, url visit-state tracking, and a
configurable result transformation pipeline.

See `ADVANCEMENT.md` for the current project status.

## Custom Transforms

Crawl results go through a transform pipeline before being written to an
output sink: each step is a small `output.Transformer` that contributes
one field, keyed by its own identifier, to the final `engine.Record`.
Built-in steps (`internal/crawler/output/transform`) read straight from
the raw `engine.CrawlResult` — full url, partial url, content hash,
content, http status, content length, timestamp.

A custom transformer just implements `output.Transformer`:

```go
type Transformer interface {
    ID() string
    Transform(result engine.CrawlResult, acc engine.Record) (interface{}, error)
}
```

`Transform` receives both the raw crawl result and the `Record`
accumulated by prior steps in the pipeline, so a custom step can build on
what earlier steps already produced.

Example: building a pipeline of built-in steps plus a custom `word_count`
transformer, then serializing and writing the resulting record to stdout.

```go
package main

import (
	"fmt"
	"time"

	"github.com/m1dugh/gocrawler/internal/crawler/output/pipeline"
	"github.com/m1dugh/gocrawler/internal/crawler/output/serialize"
	"github.com/m1dugh/gocrawler/internal/crawler/output/sink"
	"github.com/m1dugh/gocrawler/internal/crawler/output/transform"
	"github.com/m1dugh/gocrawler/pkg/crawler/engine"
	"github.com/m1dugh/gocrawler/pkg/crawler/output"
)

// A custom transformer, the kind a plugin would eventually contribute.
// It reads content straight from the raw CrawlResult (not from another
// step's output) to compute a crude word count.
type wordCount struct{}

func (wordCount) ID() string { return "word_count" }

func (wordCount) Transform(r engine.CrawlResult, _ engine.Record) (interface{}, error) {
	return len(r.Content), nil
}

var _ output.Transformer = wordCount{}

func main() {
	// The Builder is the API the crawler's config module will drive:
	// one Add per configured step, in order, then Build() the pipeline.
	p := pipeline.NewBuilder().
		Add(transform.FullURL{}).
		Add(transform.PartialURL{}).
		Add(transform.ContentHash{}).
		Add(transform.Content{}).
		Add(transform.Status{}).
		Add(transform.ContentLength{}).
		Add(transform.Timestamp{}).
		Add(wordCount{}).
		Build()

	result := engine.CrawlResult{
		URL:       "https://example.com/page?id=1",
		Content:   "hello world this is a page",
		Status:    200,
		Timestamp: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	record, err := p.Run(result)
	if err != nil {
		panic(err)
	}

	ser := serialize.JSON{}

	stdoutSink := sink.NewStdout(ser)
	if err := stdoutSink.Write(record); err != nil {
		panic(err)
	}

	fmt.Println("record keys:", len(record))
}
```

Output:

```json
{"content":"hello world this is a page","content_hash":"cb84dcdc0ee40e9b4ec3592ad2e3085aa98342ce9dd3bf71e3a34262ada9fa7f","content_length":26,"full_url":"https://example.com/page?id=1","partial_url":"https://example.com/page","status":200,"timestamp":"2026-01-01T00:00:00Z","word_count":26}
record keys: 8
```

Custom transformers loaded as external Go plugins are a planned goal (see
`ADVANCEMENT.md`) — not yet implemented.
