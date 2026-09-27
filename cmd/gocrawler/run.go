package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/m1dugh/gocrawler/internal/crawler/hash"
	"github.com/m1dugh/gocrawler/internal/crawler/linkextract"
	"github.com/m1dugh/gocrawler/internal/crawler/output/pipeline"
	"github.com/m1dugh/gocrawler/internal/crawler/queue"
	"github.com/m1dugh/gocrawler/internal/crawler/urlstate"
	"github.com/m1dugh/gocrawler/pkg/crawler/engine"
)

// outputSink is the subset of output.Sink this command writes results
// to (satisfied by both sink.File and sink.Stdout).
type outputSink interface {
	Write(engine.Record) error
}

// Config holds everything the orchestrator needs to run a crawl. All the
// actual crawling/transform/output logic lives behind these components;
// Run itself only wires them together per CLAUDE.md's methodology.
type Config struct {
	Queue    *queue.Queue
	Store    *urlstate.Store
	Engine   engine.Engine
	Pipeline *pipeline.Pipeline
	Sinks    []outputSink
	Workers  int
}

// Run drives the worker pool: n crawl workers pull urls from Queue,
// fetch and process them, and push newly discovered in-scope links back
// onto the Queue; one result-processing goroutine reads crawl results
// off a channel, runs them through Pipeline, and writes the resulting
// Record to every Sink. Run blocks until the crawl is complete.
func Run(cfg Config) {
	results := make(chan engine.CrawlResult)

	var resultWG sync.WaitGroup
	resultWG.Add(1)
	go func() {
		defer resultWG.Done()
		processResults(cfg, results)
	}()

	tasks := &taskTracker{}

	var workerWG sync.WaitGroup
	for i := 0; i < cfg.Workers; i++ {
		workerWG.Add(1)
		go func() {
			defer workerWG.Done()
			runWorker(cfg, results, tasks)
		}()
	}

	workerWG.Wait()
	close(results)
	resultWG.Wait()
}

// taskTracker guards the queue's "is there still work to do" question:
// a url popped off the queue counts as in flight until it has finished
// processing (including pushing any links it discovers), so a worker
// never mistakes "queue momentarily empty" for "crawl finished".
type taskTracker struct {
	mu       sync.Mutex
	inFlight int
}

// tryPop pops the next url from q, if any, and marks it in flight. done
// must be called exactly once the url has finished processing.
func (t *taskTracker) tryPop(q *queue.Queue) (url string, done func(), ok bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	u, err := q.Pop()
	if err != nil {
		return "", nil, false
	}

	t.inFlight++
	return u, t.taskDone, true
}

func (t *taskTracker) taskDone() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.inFlight--
}

// idle reports whether the queue is empty and no popped url is still
// being processed, i.e. no more work can ever appear.
func (t *taskTracker) idle(q *queue.Queue) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.inFlight == 0 && q.Length() == 0
}

// runWorker repeatedly pops a url from the queue and processes it, until
// the queue is empty and no other worker has work in flight.
func runWorker(cfg Config, results chan<- engine.CrawlResult, tasks *taskTracker) {
	for {
		url, done, ok := tasks.tryPop(cfg.Queue)
		if !ok {
			if tasks.idle(cfg.Queue) {
				return
			}
			time.Sleep(10 * time.Millisecond)
			continue
		}

		processURL(cfg, url, results)
		done()
	}
}

// processURL implements the per-url steps from CLAUDE.md's "Url
// processing" section: skip if already visited, fetch, extract and
// enqueue links, record visit state, and hand the result off for output
// processing.
func processURL(cfg Config, url string, results chan<- engine.CrawlResult) {
	u := urlstate.ParseURL(url)
	if !cfg.Store.ShouldCrawl(u) {
		return
	}

	content, status, err := cfg.Engine.Fetch(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gocrawler: fetch %q: %v\n", url, err)
		return
	}

	for _, link := range linkextract.ExtractLinks(content, url) {
		_ = cfg.Queue.Push(link)
	}

	cfg.Store.RecordVisit(u, hash.Content(content))

	results <- engine.CrawlResult{
		URL:       url,
		Content:   content,
		Status:    status,
		Timestamp: time.Now(),
	}
}

// processResults reads crawl results until the channel is closed,
// transforming each one through the pipeline and writing it to every
// configured sink.
func processResults(cfg Config, results <-chan engine.CrawlResult) {
	for result := range results {
		record, err := cfg.Pipeline.Run(result)
		if err != nil {
			fmt.Fprintf(os.Stderr, "gocrawler: transform %q: %v\n", result.URL, err)
			continue
		}

		for _, s := range cfg.Sinks {
			if err := s.Write(record); err != nil {
				fmt.Fprintf(os.Stderr, "gocrawler: write output: %v\n", err)
			}
		}
	}
}
