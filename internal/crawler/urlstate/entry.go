package urlstate

import "sync"

// VisitState describes how much is known about a path's visit history.
type VisitState int

const (
	Unvisited VisitState = iota
	VisitedOnce
	Visited
	VisitedParametrized
)

// Entry tracks, for a single path, the content hash observed for each
// query string seen so far.
type Entry struct {
	mu      sync.Mutex
	hashes  map[string]string // query string -> content hash ("" key = no query string)
	hasDiff bool              // true once two different query strings produced different hashes
}

func newEntry() *Entry {
	return &Entry{hashes: make(map[string]string)}
}

// State computes the current VisitState for this entry.
func (e *Entry) State() VisitState {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.stateLocked()
}

func (e *Entry) stateLocked() VisitState {
	if len(e.hashes) == 0 {
		return Unvisited
	}
	if _, noQuery := e.hashes[""]; noQuery {
		return Visited
	}
	if len(e.hashes) == 1 {
		return VisitedOnce
	}
	if e.hasDiff {
		return VisitedParametrized
	}
	return Visited
}

// IsVisited reports whether u should be skipped, based on what is already
// recorded for its path.
func (e *Entry) IsVisited(u URL) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	switch e.stateLocked() {
	case Unvisited:
		return false
	case VisitedOnce:
		_, known := e.hashes[u.Query]
		return known
	default: // Visited, VisitedParametrized
		return true
	}
}

// IsVisitedHash reports whether u is already visited with the given content
// hash, additionally treating an unseen (query string, hash) combination as
// not visited even when the path is otherwise known.
func (e *Entry) IsVisitedHash(u URL, hash string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	if known, ok := e.hashes[u.Query]; ok {
		return known == hash
	}
	return false
}

// Visit records the (query string, hash) pair for u.
func (e *Entry) Visit(u URL, hash string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	for query, existing := range e.hashes {
		if query != u.Query && existing != hash {
			e.hasDiff = true
			break
		}
	}
	e.hashes[u.Query] = hash
}
