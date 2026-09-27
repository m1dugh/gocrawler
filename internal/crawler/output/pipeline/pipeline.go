// Package pipeline runs a configured sequence of output.Transformer steps
// over a crawl result, accumulating each step's contribution into a
// Record keyed by transformer ID.
package pipeline

import (
	"fmt"

	"github.com/m1dugh/gocrawler/pkg/crawler/engine"
	"github.com/m1dugh/gocrawler/pkg/crawler/output"
)

// Pipeline runs an ordered list of output.Transformer steps.
type Pipeline struct {
	steps []output.Transformer
}

// New returns a Pipeline that runs steps in order.
func New(steps []output.Transformer) *Pipeline {
	return &Pipeline{steps: steps}
}

// Run passes r through every step of the pipeline in order, accumulating
// each step's contribution into the returned Record under its ID.
func (p *Pipeline) Run(r engine.CrawlResult) (engine.Record, error) {
	acc := engine.Record{}
	for _, step := range p.steps {
		v, err := step.Transform(r, acc)
		if err != nil {
			return nil, fmt.Errorf("pipeline: transformer %q: %w", step.ID(), err)
		}
		acc[step.ID()] = v
	}
	return acc, nil
}
