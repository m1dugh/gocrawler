package pipeline_test

import (
	"errors"
	"testing"

	"github.com/m1dugh/gocrawler/internal/crawler/output/pipeline"
	"github.com/m1dugh/gocrawler/pkg/crawler/engine"
	"github.com/m1dugh/gocrawler/pkg/crawler/output"
)

// echoTransformer returns a fixed value and, if seenPrev is set, also
// records a snapshot of the accumulator it saw.
type echoTransformer struct {
	id       string
	val      interface{}
	seenPrev *engine.Record
}

func (e echoTransformer) ID() string { return e.id }

func (e echoTransformer) Transform(_ engine.CrawlResult, acc engine.Record) (interface{}, error) {
	if e.seenPrev != nil {
		snapshot := make(engine.Record, len(acc))
		for k, v := range acc {
			snapshot[k] = v
		}
		*e.seenPrev = snapshot
	}
	return e.val, nil
}

var errBoom = errors.New("boom")

type erroringTransformer struct{ id string }

func (e erroringTransformer) ID() string { return e.id }

func (e erroringTransformer) Transform(engine.CrawlResult, engine.Record) (interface{}, error) {
	return nil, errBoom
}

func TestPipelineRunAccumulates(t *testing.T) {
	p := pipeline.New([]output.Transformer{
		echoTransformer{id: "a", val: 1},
		echoTransformer{id: "b", val: 2},
	})

	record, err := p.Run(engine.CrawlResult{})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if record["a"] != 1 {
		t.Errorf(`record["a"] = %v, want 1`, record["a"])
	}
	if record["b"] != 2 {
		t.Errorf(`record["b"] = %v, want 2`, record["b"])
	}
}

func TestPipelineStepSeesPriorAccumulator(t *testing.T) {
	var seen engine.Record
	p := pipeline.New([]output.Transformer{
		echoTransformer{id: "a", val: 1},
		echoTransformer{id: "b", val: 2, seenPrev: &seen},
	})

	if _, err := p.Run(engine.CrawlResult{}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if seen["a"] != 1 {
		t.Errorf(`step "b" saw acc["a"] = %v, want 1`, seen["a"])
	}
	if _, ok := seen["b"]; ok {
		t.Errorf(`step "b" should not see its own not-yet-produced value`)
	}
}

func TestPipelineRunPropagatesError(t *testing.T) {
	p := pipeline.New([]output.Transformer{erroringTransformer{id: "a"}})

	if _, err := p.Run(engine.CrawlResult{}); err == nil {
		t.Fatal("Run() with erroring transformer: expected error, got nil")
	}
}
