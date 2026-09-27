// Package http implements the basic crawling engine, which fetches urls
// using a plain net/http client with no html/js rendering.
package http

import (
	"io"
	"net/http"
	"time"
)

// Config configures the basic HTTP crawling engine.
type Config struct {
	// Headers are custom headers sent with every request (e.g. User-Agent,
	// Authorization, Cookie).
	Headers map[string][]string

	// Timeout is the per-request timeout. Zero means no timeout.
	Timeout time.Duration
}

// Engine is a crawler.Engine implementation that fetches urls using a
// plain http client.
type Engine struct {
	client  *http.Client
	headers map[string][]string
}

// New returns an Engine configured with cfg.
func New(cfg Config) *Engine {
	return &Engine{
		client:  &http.Client{Timeout: cfg.Timeout},
		headers: cfg.Headers,
	}
}

// Fetch retrieves the content of url over HTTP and returns it alongside
// the response status code.
func (e *Engine) Fetch(url string) (content string, status int, err error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", 0, err
	}

	for key, values := range e.headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", resp.StatusCode, err
	}

	return string(body), resp.StatusCode, nil
}

// Close is a no-op: Engine holds no resources that need releasing.
func (e *Engine) Close() error {
	return nil
}
