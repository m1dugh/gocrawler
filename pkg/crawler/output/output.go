// Package output defines the transform pipeline and sink interfaces used
// to turn a raw crawl result into an output-ready record and write it to
// a destination (file, stdout, or other).
package output

import "github.com/m1dugh/gocrawler/pkg/crawler/engine"

// Transformer is one step of the transform pipeline. It receives the raw
// crawl result and the Record accumulated by prior steps, and returns the
// value it contributes under its own ID.
//
// Transformer is the extension point for custom transform steps,
// including ones loaded from external plugins (see ADVANCEMENT.md).
type Transformer interface {
	// ID identifies this transformer's contribution in the Record.
	ID() string

	// Transform computes this step's value from the raw crawl result and
	// whatever previous steps in the pipeline have already produced.
	Transform(result engine.CrawlResult, acc engine.Record) (interface{}, error)
}

// Serializer encodes a Record into bytes for a Sink to write, so output
// format is swappable independently of the destination.
type Serializer interface {
	Serialize(engine.Record) ([]byte, error)
}

// Sink writes a Record to a destination, one record per line.
type Sink interface {
	Write(engine.Record) error
}
