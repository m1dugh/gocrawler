package pipeline

import (
	"fmt"

	"github.com/m1dugh/gocrawler/pkg/crawler/output"
)

// Builder assembles a Pipeline one transformer at a time, in order.
type Builder struct {
	steps []output.Transformer
}

// NewBuilder returns an empty Builder.
func NewBuilder() *Builder {
	return &Builder{}
}

// Add appends a transformer to the pipeline under construction.
func (b *Builder) Add(t output.Transformer) *Builder {
	b.steps = append(b.steps, t)
	return b
}

// Build returns the Pipeline. It fails if a transformer requires a field
// that no earlier transformer provides.
func (b *Builder) Build() (*Pipeline, error) {
	provided := map[string]bool{}
	for _, step := range b.steps {
		for _, req := range step.Requires() {
			if !provided[req] {
				return nil, fmt.Errorf("transformer %q requires field %q, which no earlier transformer provides", step.Name(), req)
			}
		}
		for _, f := range step.Provides() {
			provided[f] = true
		}
	}

	steps := make([]output.Transformer, len(b.steps))
	copy(steps, b.steps)
	return New(steps), nil
}
