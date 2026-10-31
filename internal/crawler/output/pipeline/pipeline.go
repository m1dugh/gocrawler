// Package pipeline runs an ordered sequence of transformers over a raw
// crawl result and gathers the fields they provide into a record.
package pipeline

import (
	"fmt"

	"github.com/m1dugh/gocrawler/pkg/crawler/engine"
	"github.com/m1dugh/gocrawler/pkg/crawler/output"
)

// Pipeline is an ordered sequence of transformers.
type Pipeline struct {
	steps []output.Transformer
}

// New returns a Pipeline running steps in order. It performs no
// validation; use a Builder to check that required fields are provided.
func New(steps []output.Transformer) *Pipeline {
	return &Pipeline{steps: steps}
}

// Provided lists the fields the pipeline produces, in step order.
func (p *Pipeline) Provided() []string {
	var fields []string
	seen := map[string]bool{}
	for _, step := range p.steps {
		for _, f := range step.Provides() {
			if !seen[f] {
				seen[f] = true
				fields = append(fields, f)
			}
		}
	}
	return fields
}

// Run executes every step in order and returns the resulting record. Each
// step sees the fields produced by the steps before it. A step returning a
// field it did not declare in Provides is an error.
func (p *Pipeline) Run(r engine.CrawlResult) (engine.Record, error) {
	acc := engine.Record{}
	for _, step := range p.steps {
		fields, err := step.Transform(r, acc)
		if err != nil {
			return nil, fmt.Errorf("pipeline: transformer %q: %w", step.Name(), err)
		}

		provides := step.Provides()
		for k, v := range fields {
			if !contains(provides, k) {
				return nil, fmt.Errorf("pipeline: transformer %q returned undeclared field %q", step.Name(), k)
			}
			acc[k] = v
		}
	}
	return acc, nil
}

func contains(list []string, s string) bool {
	for _, e := range list {
		if e == s {
			return true
		}
	}
	return false
}
