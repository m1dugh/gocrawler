package pipeline_test

import (
	"testing"

	"github.com/m1dugh/gocrawler/internal/crawler/output/pipeline"
	"github.com/m1dugh/gocrawler/pkg/crawler/engine"
)

func TestBuilderAddBuildRunsInOrder(t *testing.T) {
	p, err := pipeline.NewBuilder().
		Add(echoTransformer{id: "a", val: 1}).
		Add(echoTransformer{id: "b", val: 2}).
		Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	record, err := p.Run(engine.CrawlResult{})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if record["a"] != 1 || record["b"] != 2 {
		t.Errorf("record = %+v, want a=1 b=2", record)
	}
}

func TestBuilderEmpty(t *testing.T) {
	p, err := pipeline.NewBuilder().Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	record, err := p.Run(engine.CrawlResult{})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(record) != 0 {
		t.Errorf("record = %+v, want empty", record)
	}
}

func TestBuilderAcceptsRequirementProvidedEarlier(t *testing.T) {
	_, err := pipeline.NewBuilder().
		Add(echoTransformer{id: "a"}).
		Add(echoTransformer{id: "b", requires: []string{"a"}}).
		Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
}

func TestBuilderRejectsMissingRequirement(t *testing.T) {
	_, err := pipeline.NewBuilder().
		Add(echoTransformer{id: "b", requires: []string{"a"}}).
		Build()
	if err == nil {
		t.Fatal("Build() with an unprovided requirement: expected error, got nil")
	}
}

func TestBuilderRejectsRequirementProvidedLater(t *testing.T) {
	_, err := pipeline.NewBuilder().
		Add(echoTransformer{id: "b", requires: []string{"a"}}).
		Add(echoTransformer{id: "a"}).
		Build()
	if err == nil {
		t.Fatal("Build() with a requirement provided only by a later step: expected error, got nil")
	}
}
