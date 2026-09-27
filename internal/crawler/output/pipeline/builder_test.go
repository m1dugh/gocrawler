package pipeline_test

import (
	"testing"

	"github.com/m1dugh/gocrawler/internal/crawler/output/pipeline"
	"github.com/m1dugh/gocrawler/pkg/crawler/engine"
)

func TestBuilderAddBuildRunsInOrder(t *testing.T) {
	p := pipeline.NewBuilder().
		Add(echoTransformer{id: "a", val: 1}).
		Add(echoTransformer{id: "b", val: 2}).
		Build()

	record, err := p.Run(engine.CrawlResult{})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if record["a"] != 1 || record["b"] != 2 {
		t.Errorf("record = %+v, want a=1 b=2", record)
	}
}

func TestBuilderEmpty(t *testing.T) {
	p := pipeline.NewBuilder().Build()

	record, err := p.Run(engine.CrawlResult{})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(record) != 0 {
		t.Errorf("record = %+v, want empty", record)
	}
}
