package urlstate

import "strings"

// URL represents a crawled URL split into its path (without query string)
// and its raw query string.
type URL struct {
	Path  string
	Query string
}

// ParseURL splits raw on the first '?' into a path and a query string.
// If raw has no query string, Query is empty.
func ParseURL(raw string) URL {
	path, query, _ := strings.Cut(raw, "?")
	return URL{Path: path, Query: query}
}
