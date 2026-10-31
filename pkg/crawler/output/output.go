// Package output defines the extension points of the result-processing
// side of the crawler: transformers fill the output record, serializers
// turn a record into bytes, and sinks deliver records somewhere. Anything
// here can be implemented outside this module and registered in a Registry.
package output

import "github.com/m1dugh/gocrawler/pkg/crawler/engine"

// Transformer is one step of the transform pipeline. It reads the raw
// crawl result, plus the record accumulated by earlier steps, and returns
// the fields it contributes to the record.
type Transformer interface {
	// Name is the unique identifier used to enable or disable the
	// transformer.
	Name() string

	// Provides lists the fields Transform may return. Transform must not
	// return a field that is not listed here.
	Provides() []string

	// Requires lists the fields Transform reads from the accumulated
	// record. Each one must be provided by an earlier step.
	Requires() []string

	// Transform returns the fields this step contributes for result. acc
	// holds the fields produced by earlier steps and must not be modified.
	Transform(result engine.CrawlResult, acc engine.Record) (engine.Record, error)
}

// Serializer turns a Record into the bytes written by a sink.
type Serializer interface {
	Serialize(engine.Record) ([]byte, error)
}

// Sink receives output records. Write may be called concurrently with
// itself only if the implementation allows it; the crawler calls it from a
// single goroutine.
type Sink interface {
	Write(engine.Record) error

	// Close releases any resource held by the sink. It is called once,
	// after the last Write.
	Close() error
}
