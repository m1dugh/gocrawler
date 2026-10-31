package pipeline_test

import (
	"errors"
	"testing"

	"github.com/m1dugh/gocrawler/internal/crawler/output/pipeline"
	"github.com/m1dugh/gocrawler/pkg/crawler/engine"
	"github.com/m1dugh/gocrawler/pkg/crawler/output"
)

// echoTransformer provides one field, id, holding val. It records the
// accumulator it was given in seenPrev when set.
type echoTransformer struct {
	id       string
	val      interface{}
	requires []string
	seenPrev *engine.Record
}

func (e echoTransformer) Name() string       { return e.id }
func (e echoTransformer) Provides() []string { return []string{e.id} }
func (e echoTransformer) Requires() []string { return e.requires }

func (e echoTransformer) Transform(_ engine.CrawlResult, acc engine.Record) (engine.Record, error) {
	if e.seenPrev != nil {
		snapshot := make(engine.Record, len(acc))
		for k, v := range acc {
			snapshot[k] = v
		}
		*e.seenPrev = snapshot
	}
	return engine.Record{e.id: e.val}, nil
}

var errBoom = errors.New("boom")

type erroringTransformer struct{ echoTransformer }

func (erroringTransformer) Transform(engine.CrawlResult, engine.Record) (engine.Record, error) {
	return nil, errBoom
}

// rogueTransformer returns a field it did not declare.
type rogueTransformer struct{ echoTransformer }

func (rogueTransformer) Transform(engine.CrawlResult, engine.Record) (engine.Record, error) {
	return engine.Record{"undeclared": 1}, nil
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
	if record["a"] != 1 || record["b"] != 2 {
		t.Errorf("record = %+v, want a=1 b=2", record)
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
	p := pipeline.New([]output.Transformer{erroringTransformer{echoTransformer{id: "a"}}})

	if _, err := p.Run(engine.CrawlResult{}); !errors.Is(err, errBoom) {
		t.Fatalf("Run() error = %v, want it to wrap errBoom", err)
	}
}

func TestPipelineRejectsUndeclaredField(t *testing.T) {
	p := pipeline.New([]output.Transformer{rogueTransformer{echoTransformer{id: "a"}}})

	if _, err := p.Run(engine.CrawlResult{}); err == nil {
		t.Fatal("Run() with a transformer returning an undeclared field: expected error, got nil")
	}
}

func TestPipelineProvided(t *testing.T) {
	p := pipeline.New([]output.Transformer{
		echoTransformer{id: "a"},
		echoTransformer{id: "b"},
	})

	got := p.Provided()
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("Provided() = %v, want [a b]", got)
	}
}
