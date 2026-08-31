package urlstate

import (
	"fmt"
	"sync"
	"testing"
)

func TestStoreConcurrentAccess(t *testing.T) {
	s := NewStore()
	base := "https://example.com/a"

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			u := ParseURL(fmt.Sprintf("%s?q=%d", base, i))
			if s.ShouldCrawl(u) {
				s.RecordVisit(u, "hash")
			}
		}(i)
	}
	wg.Wait()

	u := ParseURL(base + "?q=0")
	e := s.entries[u.Path]
	if e == nil {
		t.Fatal("expected entry to exist after concurrent visits")
	}
	if got := e.State(); got != Visited {
		t.Fatalf("expected state Visited (all same hash), got %v", got)
	}
}
