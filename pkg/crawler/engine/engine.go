// Package engine defines the crawl engine strategy interface implemented
// by the basic (HTTP) and browser-based crawling strategies.
package engine

// Engine fetches the content of a url using a given crawling strategy
// (plain HTTP client, or a google-chrome/chromium instance).
type Engine interface {
	// Fetch retrieves the content of url and returns it alongside the
	// http status code corresponding to the request.
	Fetch(url string) (content string, status int, err error)

	// Close releases any resources held by the engine (e.g. a spawned
	// browser process or an open connection). Engines that hold no such
	// resources may implement it as a no-op.
	Close() error
}
