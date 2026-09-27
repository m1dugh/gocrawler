// Package linkextract extracts embedded links from crawled content.
package linkextract

import (
	"net/url"
	"regexp"
)

// linkPattern matches absolute http(s) urls and href/src attribute
// values found in html content.
var linkPattern = regexp.MustCompile(`https?://[^\s"'<>)]+|(?:href|src)\s*=\s*["']([^"']+)["']`)

// ExtractLinks returns every url embedded in content, resolved against
// baseURL (the url content was fetched from) so relative links (e.g.
// href="/path") become absolute. A link that fails to parse, or that
// resolves against an unparseable baseURL, is skipped.
func ExtractLinks(content, baseURL string) []string {
	base, err := url.Parse(baseURL)
	if err != nil {
		return nil
	}

	matches := linkPattern.FindAllStringSubmatch(content, -1)

	links := make([]string, 0, len(matches))
	for _, m := range matches {
		raw := m[1]
		if raw == "" {
			raw = m[0]
		}

		ref, err := url.Parse(raw)
		if err != nil {
			continue
		}

		links = append(links, base.ResolveReference(ref).String())
	}

	return links
}
