// Command gocrawler orchestrates a crawl: it wires together the queue,
// url visit-state tracking, crawl engine, transform pipeline and output
// sinks, and drives the worker pool described in CLAUDE.md.
package main

import (
	"flag"
	"fmt"
	"os"

	httpengine "github.com/m1dugh/gocrawler/internal/crawler/engine/http"
	"github.com/m1dugh/gocrawler/internal/crawler/output/pipeline"
	"github.com/m1dugh/gocrawler/internal/crawler/output/serialize"
	"github.com/m1dugh/gocrawler/internal/crawler/output/sink"
	"github.com/m1dugh/gocrawler/internal/crawler/output/transform"
	"github.com/m1dugh/gocrawler/internal/crawler/queue"
	"github.com/m1dugh/gocrawler/internal/crawler/scope"
	"github.com/m1dugh/gocrawler/internal/crawler/urlstate"
)

func main() {
	scopeFile := flag.String("scope", "", "path to the scope rules file (required)")
	outFile := flag.String("out", "", "path to the output file (defaults to stdout only)")
	workers := flag.Int("workers", 4, "number of concurrent crawl workers")
	flag.Parse()

	seeds := flag.Args()
	if *scopeFile == "" || len(seeds) == 0 {
		fmt.Fprintln(os.Stderr, "usage: gocrawler -scope <file> [-out <file>] [-workers N] <seed-url> [seed-url ...]")
		os.Exit(2)
	}

	s, err := loadScope(*scopeFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "gocrawler:", err)
		os.Exit(1)
	}

	q := queue.New(s)
	for _, seed := range seeds {
		if err := q.Push(seed); err != nil {
			fmt.Fprintf(os.Stderr, "gocrawler: seed %q: %v\n", seed, err)
		}
	}

	sinks := []outputSink{sink.NewStdout(serialize.JSON{})}
	if *outFile != "" {
		fileSink, err := sink.NewFile(*outFile, serialize.JSON{})
		if err != nil {
			fmt.Fprintln(os.Stderr, "gocrawler:", err)
			os.Exit(1)
		}
		defer fileSink.Close()
		sinks = append(sinks, fileSink)
	}

	eng := httpengine.New(httpengine.Config{})
	defer eng.Close()

	cfg := Config{
		Queue:    q,
		Store:    urlstate.NewStore(),
		Engine:   eng,
		Pipeline: defaultPipeline(),
		Sinks:    sinks,
		Workers:  *workers,
	}

	Run(cfg)
}

// loadScope reads scope rules from path and builds a Scope.
func loadScope(path string) (*scope.Scope, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return scope.NewFromReader(f)
}

// defaultPipeline builds the built-in transform pipeline (full url,
// partial url, content hash, status, content length, timestamp).
//
// This will eventually be assembled from the crawler's configuration
// instead of hardcoded here.
func defaultPipeline() *pipeline.Pipeline {
	return pipeline.NewBuilder().
		Add(transform.FullURL{}).
		Add(transform.PartialURL{}).
		Add(transform.ContentHash{}).
		Add(transform.Status{}).
		Add(transform.ContentLength{}).
		Add(transform.Timestamp{}).
		Build()
}
