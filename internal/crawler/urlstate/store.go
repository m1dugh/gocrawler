package urlstate

import "sync"

// Store tracks visit state for every path seen by the crawler, keyed by
// URL.Path. It is safe for concurrent use by multiple worker goroutines.
type Store struct {
	mu      sync.Mutex
	entries map[string]*Entry
}

// NewStore returns an empty Store.
func NewStore() *Store {
	return &Store{entries: make(map[string]*Entry)}
}

// ShouldCrawl implements the pre-fetch check: it reports whether u has not
// yet been visited (or should be visited again) according to the recorded
// state for its path.
func (s *Store) ShouldCrawl(u URL) bool {
	s.mu.Lock()
	entry, ok := s.entries[u.Path]
	s.mu.Unlock()

	if !ok {
		return true
	}
	return !entry.IsVisited(u)
}

// RecordVisit implements the post-fetch step: it creates the entry for u's
// path if absent, then records the (query string, hash) pair.
func (s *Store) RecordVisit(u URL, hash string) {
	s.mu.Lock()
	entry, ok := s.entries[u.Path]
	if !ok {
		entry = newEntry()
		s.entries[u.Path] = entry
	}
	s.mu.Unlock()

	entry.Visit(u, hash)
}
