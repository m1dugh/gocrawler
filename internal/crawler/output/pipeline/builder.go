package pipeline

import "github.com/m1dugh/gocrawler/pkg/crawler/output"

// Builder incrementally assembles the ordered list of transformer steps
// that make up a Pipeline. It is the API the crawler's configuration
// module will drive to build a Pipeline from config (not yet
// implemented): one Add call per configured transformer step, in order.
type Builder struct {
	steps []output.Transformer
}

// NewBuilder returns an empty Builder.
func NewBuilder() *Builder {
	return &Builder{}
}

// Add appends t as the next step of the pipeline being built, and
// returns the Builder for chaining.
func (b *Builder) Add(t output.Transformer) *Builder {
	b.steps = append(b.steps, t)
	return b
}

// Build returns a Pipeline that runs the accumulated steps in the order
// they were added.
func (b *Builder) Build() *Pipeline {
	steps := make([]output.Transformer, len(b.steps))
	copy(steps, b.steps)
	return New(steps)
}
