package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	httpengine "github.com/m1dugh/gocrawler/internal/crawler/engine/http"
	"github.com/m1dugh/gocrawler/internal/crawler/queue"
	"github.com/m1dugh/gocrawler/internal/crawler/scope"
	"github.com/m1dugh/gocrawler/internal/crawler/urlstate"
	"github.com/m1dugh/gocrawler/pkg/crawler/engine"
)

// recordingSink is an outputSink that stores every Record it receives,
// safe for concurrent use.
type recordingSink struct {
	mu      sync.Mutex
	records []engine.Record
}

func (s *recordingSink) Write(r engine.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records = append(s.records, r)
	return nil
}

func (s *recordingSink) urls(t *testing.T) []string {
	t.Helper()
	s.mu.Lock()
	defer s.mu.Unlock()

	urls := make([]string, 0, len(s.records))
	for _, r := range s.records {
		u, _ := r["full_url"].(string)
		urls = append(urls, u)
	}
	return urls
}

func TestRunCrawlsLinkedPages(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/a", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<a href="/b">b</a>`)
	})
	mux.HandleFunc("/b", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<a href="/c">c</a>`)
	})
	mux.HandleFunc("/c", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `no more links`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	s, err := scope.New([]string{"^" + srv.URL})
	if err != nil {
		t.Fatalf("scope.New() error = %v", err)
	}

	q := queue.New(s)
	if err := q.Push(srv.URL + "/a"); err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	rec := &recordingSink{}

	Run(Config{
		Queue:    q,
		Store:    urlstate.NewStore(),
		Engine:   httpengine.New(httpengine.Config{}),
		Pipeline: defaultPipeline(),
		Sinks:    []outputSink{rec},
		Workers:  2,
	})

	got := rec.urls(t)
	want := map[string]bool{
		srv.URL + "/a": true,
		srv.URL + "/b": true,
		srv.URL + "/c": true,
	}

	if len(got) != len(want) {
		t.Fatalf("crawled %d urls (%v), want %d", len(got), got, len(want))
	}
	for _, u := range got {
		if !want[u] {
			t.Errorf("unexpected crawled url %q", u)
		}
	}
}

func TestRunSkipsAlreadyVisited(t *testing.T) {
	var hits int
	var mu sync.Mutex

	mux := http.NewServeMux()
	mux.HandleFunc("/a", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		hits++
		mu.Unlock()
		fmt.Fprint(w, `<a href="/a">self</a>`) // links to itself
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	s, err := scope.New([]string{"^" + srv.URL})
	if err != nil {
		t.Fatalf("scope.New() error = %v", err)
	}

	q := queue.New(s)
	if err := q.Push(srv.URL + "/a"); err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	rec := &recordingSink{}

	Run(Config{
		Queue:    q,
		Store:    urlstate.NewStore(),
		Engine:   httpengine.New(httpengine.Config{}),
		Pipeline: defaultPipeline(),
		Sinks:    []outputSink{rec},
		Workers:  2,
	})

	mu.Lock()
	defer mu.Unlock()
	if hits != 1 {
		t.Errorf("server hit %d times, want 1 (self-link should be de-duplicated)", hits)
	}
}

func TestDefaultPipelineJSONRoundTrip(t *testing.T) {
	p := defaultPipeline()
	result := engine.CrawlResult{URL: "https://example.com", Content: "hi", Status: 200}

	record, err := p.Run(result)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if _, err := json.Marshal(record); err != nil {
		t.Fatalf("record not JSON-serializable: %v", err)
	}
}
