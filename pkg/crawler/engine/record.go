package engine

// Record is the output-ready result of running a CrawlResult through a
// transformer pipeline: each entry is keyed by the identifier of the
// Transformer that produced it.
type Record map[string]interface{}
