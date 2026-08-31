package urlstate

import "testing"

func TestUnvisitedPathShouldCrawl(t *testing.T) {
	s := NewStore()
	u := ParseURL("https://example.com/a")

	if !s.ShouldCrawl(u) {
		t.Fatal("expected unvisited path to be crawlable")
	}
}

func TestNoQueryPathVisitedOnce(t *testing.T) {
	s := NewStore()
	u := ParseURL("https://example.com/a")

	s.RecordVisit(u, "hash1")

	if s.ShouldCrawl(u) {
		t.Fatal("expected no-query path to be Visited and not re-crawled")
	}

	e := s.entries[u.Path]
	if got := e.State(); got != Visited {
		t.Fatalf("expected state Visited, got %v", got)
	}
}

func TestQueryStringVisitedOnce(t *testing.T) {
	s := NewStore()
	u := ParseURL("https://example.com/a?x=1")

	s.RecordVisit(u, "hash1")

	e := s.entries[u.Path]
	if got := e.State(); got != VisitedOnce {
		t.Fatalf("expected state VisitedOnce, got %v", got)
	}

	if s.ShouldCrawl(u) {
		t.Fatal("expected same query string to be considered visited")
	}

	other := ParseURL("https://example.com/a?y=2")
	if !s.ShouldCrawl(other) {
		t.Fatal("expected a different query string to still be crawlable")
	}
}

func TestSameHashAcrossQueryStringsIsVisited(t *testing.T) {
	s := NewStore()
	u1 := ParseURL("https://example.com/a?x=1")
	u2 := ParseURL("https://example.com/a?y=2")

	s.RecordVisit(u1, "same-hash")
	s.RecordVisit(u2, "same-hash")

	e := s.entries[u1.Path]
	if got := e.State(); got != Visited {
		t.Fatalf("expected state Visited, got %v", got)
	}
	if s.ShouldCrawl(u1) {
		t.Fatal("expected Visited state to skip re-crawl")
	}
}

func TestDifferentHashAcrossQueryStringsIsParametrized(t *testing.T) {
	s := NewStore()
	u1 := ParseURL("https://example.com/a?x=1")
	u2 := ParseURL("https://example.com/a?y=2")

	s.RecordVisit(u1, "hash-a")
	s.RecordVisit(u2, "hash-b")

	e := s.entries[u1.Path]
	if got := e.State(); got != VisitedParametrized {
		t.Fatalf("expected state VisitedParametrized, got %v", got)
	}

	if s.ShouldCrawl(u1) {
		t.Fatal("expected VisitedParametrized path to not be re-crawled")
	}
}

func TestIsVisitedHash(t *testing.T) {
	s := NewStore()
	u := ParseURL("https://example.com/a?x=1")
	s.RecordVisit(u, "hash-a")

	e := s.entries[u.Path]

	if !e.IsVisitedHash(u, "hash-a") {
		t.Fatal("expected same hash for known query string to be visited")
	}
	if e.IsVisitedHash(u, "hash-b") {
		t.Fatal("expected different hash for known query string to not be visited")
	}

	unseen := ParseURL("https://example.com/a?z=3")
	if e.IsVisitedHash(unseen, "hash-a") {
		t.Fatal("expected unseen query string to not be visited")
	}
}
