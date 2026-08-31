package scope

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// ErrNoIncludeRules is returned when a scope file contains no include
// rules. Without at least one include rule, every url would fall out of
// scope by default, which would either crawl nothing or (if the default
// were flipped to allow-all) crawl unboundedly.
var ErrNoIncludeRules = errors.New("scope: no include rules defined")

// Scope determines whether a url is allowed to be crawled, based on a set
// of prefix-matching regexes: normal rules include a url when they match,
// rules prefixed with '!' exclude a url when they match.
type Scope struct {
	includes []*regexp.Regexp
	excludes []*regexp.Regexp
}

// New builds a Scope from regex rules, one per line. A line prefixed with
// '!' is an exclude rule; any other non-blank line is an include rule.
// Each regex is anchored to match only a prefix of the url. Returns
// ErrNoIncludeRules if no include rule is present.
func New(rules []string) (*Scope, error) {
	s := &Scope{}

	for i, line := range rules {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		exclude := strings.HasPrefix(line, "!")
		pattern := line
		if exclude {
			pattern = strings.TrimPrefix(line, "!")
		}

		re, err := regexp.Compile("^(?:" + pattern + ")")
		if err != nil {
			return nil, fmt.Errorf("scope: invalid regex on line %d: %w", i+1, err)
		}

		if exclude {
			s.excludes = append(s.excludes, re)
		} else {
			s.includes = append(s.includes, re)
		}
	}

	if len(s.includes) == 0 {
		return nil, ErrNoIncludeRules
	}

	return s, nil
}

// NewFromReader builds a Scope by reading rules line by line from r.
func NewFromReader(r io.Reader) (*Scope, error) {
	var rules []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		rules = append(rules, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scope: reading rules: %w", err)
	}
	return New(rules)
}

// IsInScope reports whether url is in scope: at least one include rule
// matches as a prefix, and no exclude rule matches as a prefix.
func (s *Scope) IsInScope(url string) bool {
	included := false
	for _, re := range s.includes {
		if re.MatchString(url) {
			included = true
			break
		}
	}
	if !included {
		return false
	}

	for _, re := range s.excludes {
		if re.MatchString(url) {
			return false
		}
	}

	return true
}
