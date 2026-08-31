package queue

import (
	"sync"
	"testing"

	"github.com/m1dugh/gocrawler/internal/crawler/scope"
)

func mustScope(t *testing.T, rules ...string) *scope.Scope {
	t.Helper()
	s, err := scope.New(rules)
	if err != nil {
		t.Fatalf("failed to build scope: %v", err)
	}
	return s
}

func TestPopEmpty(t *testing.T) {
	q := New(mustScope(t, `https://example\.com`))

	if _, err := q.Pop(); err != ErrEmpty {
		t.Fatalf("expected ErrEmpty, got %v", err)
	}
}

func TestPushOutOfScope(t *testing.T) {
	q := New(mustScope(t, `https://example\.com`))

	if err := q.Push("https://other.com"); err != ErrOutOfScope {
		t.Fatalf("expected ErrOutOfScope, got %v", err)
	}
	if q.Length() != 0 {
		t.Fatalf("expected length 0, got %d", q.Length())
	}
}

func TestPushPopFIFOOrder(t *testing.T) {
	q := New(mustScope(t, `https://example\.com`))

	urls := []string{
		"https://example.com/a",
		"https://example.com/b",
		"https://example.com/c",
	}
	for _, u := range urls {
		if err := q.Push(u); err != nil {
			t.Fatalf("unexpected push error: %v", err)
		}
	}

	if got := q.Length(); got != len(urls) {
		t.Fatalf("expected length %d, got %d", len(urls), got)
	}

	for _, want := range urls {
		got, err := q.Pop()
		if err != nil {
			t.Fatalf("unexpected pop error: %v", err)
		}
		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	}

	if q.Length() != 0 {
		t.Fatalf("expected length 0 after draining, got %d", q.Length())
	}
	if _, err := q.Pop(); err != ErrEmpty {
		t.Fatalf("expected ErrEmpty after draining, got %v", err)
	}
}

func TestPushAfterDrainReusesTail(t *testing.T) {
	q := New(mustScope(t, `https://example\.com`))

	if err := q.Push("https://example.com/a"); err != nil {
		t.Fatal(err)
	}
	if _, err := q.Pop(); err != nil {
		t.Fatal(err)
	}
	// queue is now empty with head/tail both nil; pushing again must work.
	if err := q.Push("https://example.com/b"); err != nil {
		t.Fatal(err)
	}
	got, err := q.Pop()
	if err != nil {
		t.Fatal(err)
	}
	if got != "https://example.com/b" {
		t.Fatalf("expected b, got %q", got)
	}
}

func TestConcurrentPushPop(t *testing.T) {
	q := New(mustScope(t, `https://example\.com`))

	const n = 200
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_ = q.Push("https://example.com/x")
		}(i)
	}
	wg.Wait()

	if got := q.Length(); got != n {
		t.Fatalf("expected length %d after concurrent pushes, got %d", n, got)
	}

	popped := 0
	var mu sync.Mutex
	wg = sync.WaitGroup{}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := q.Pop(); err == nil {
				mu.Lock()
				popped++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if popped != n {
		t.Fatalf("expected %d successful pops, got %d", n, popped)
	}
	if q.Length() != 0 {
		t.Fatalf("expected length 0, got %d", q.Length())
	}
}
