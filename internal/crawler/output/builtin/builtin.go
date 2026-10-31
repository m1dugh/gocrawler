// Package builtin builds the registry holding everything the crawler ships
// with: the built-in transformers, sinks and serialization formats.
package builtin

import (
	"github.com/m1dugh/gocrawler/internal/crawler/output/serialize"
	"github.com/m1dugh/gocrawler/internal/crawler/output/sink"
	"github.com/m1dugh/gocrawler/internal/crawler/output/transform"
	"github.com/m1dugh/gocrawler/pkg/crawler/output"
)

// NewRegistry returns a Registry with all built-ins registered. Default
// transformers are listed first, in the order they run.
func NewRegistry() *output.Registry {
	r := output.NewRegistry()

	for _, t := range []struct {
		fn          func() output.Transformer
		description string
		def         bool
	}{
		{transform.FullURL, "url as crawled", true},
		{transform.Status, "http status code", true},
		{transform.ContentLength, "length of the content", true},
		{transform.PartialURL, "url without its query string", false},
		{transform.ContentHash, "hash of the page content", false},
		{transform.Timestamp, "time the page was fetched", false},
		{transform.Content, "full page content", false},
	} {
		fn := t.fn
		r.RegisterTransformer(output.TransformerSpec{
			Name:        fn().Name(),
			Description: t.description,
			Default:     t.def,
			New:         func(output.Options) (output.Transformer, error) { return fn(), nil },
		})
	}

	r.RegisterSerializer("json", serialize.JSON{})

	r.RegisterSink(output.SinkSpec{
		Name:        "stdout",
		Description: "write one record per line to standard output",
		New: func(_ output.Options, ser output.Serializer) (output.Sink, error) {
			return sink.NewStdout(ser), nil
		},
	})
	r.RegisterSink(output.SinkSpec{
		Name:        "file",
		Description: "append one record per line to a file",
		Options:     []output.OptionSpec{{Name: "path", Description: "output file", Required: true}},
		New: func(opts output.Options, ser output.Serializer) (output.Sink, error) {
			return sink.NewFile(opts["path"], ser)
		},
	})

	return r
}
