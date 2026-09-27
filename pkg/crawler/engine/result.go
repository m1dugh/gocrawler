package engine

import "time"

// CrawlResult is the raw output produced by an Engine for a single url,
// before any output transformation is applied.
type CrawlResult struct {
	URL       string
	Content   string
	Status    int
	Timestamp time.Time
}
