package scope

import (
	"strings"
	"testing"
)

func TestNewNoIncludeRules(t *testing.T) {
	_, err := New([]string{`!https://evil\.example\.com`})
	if err != ErrNoIncludeRules {
		t.Fatalf("expected ErrNoIncludeRules, got %v", err)
	}
}

func TestNewEmptyRules(t *testing.T) {
	_, err := New(nil)
	if err != ErrNoIncludeRules {
		t.Fatalf("expected ErrNoIncludeRules, got %v", err)
	}
}

func TestPrefixMatchOnly(t *testing.T) {
	s, err := New([]string{`https://google\.com`})
	if err != nil {
		t.Fatal(err)
	}

	if !s.IsInScope("https://google.com/search") {
		t.Fatal("expected exact prefix match to be in scope")
	}
	if s.IsInScope("ahttps://google.com/search") {
		t.Fatal("expected non-prefix match to be out of scope")
	}
	if s.IsInScope("https://notgoogle.com") {
		t.Fatal("expected unrelated url to be out of scope")
	}
}

func TestExcludeOverridesInclude(t *testing.T) {
	s, err := New([]string{
		`https://example\.com`,
		`!https://example\.com/admin`,
	})
	if err != nil {
		t.Fatal(err)
	}

	if !s.IsInScope("https://example.com/home") {
		t.Fatal("expected non-excluded path to be in scope")
	}
	if s.IsInScope("https://example.com/admin/panel") {
		t.Fatal("expected excluded path to be out of scope even though it matches an include rule")
	}
}

func TestMultipleIncludeRules(t *testing.T) {
	s, err := New([]string{
		`https://a\.example\.com`,
		`https://b\.example\.com`,
	})
	if err != nil {
		t.Fatal(err)
	}

	if !s.IsInScope("https://a.example.com/x") {
		t.Fatal("expected first include rule to match")
	}
	if !s.IsInScope("https://b.example.com/x") {
		t.Fatal("expected second include rule to match")
	}
	if s.IsInScope("https://c.example.com/x") {
		t.Fatal("expected non-matching url to be out of scope")
	}
}

func TestBlankLinesAndInvalidRegexIgnoredOrErrored(t *testing.T) {
	if _, err := New([]string{"", "  ", `https://example\.com`}); err != nil {
		t.Fatalf("expected blank lines to be skipped, got error: %v", err)
	}

	if _, err := New([]string{`https://example\.com`, `[invalid(`}); err == nil {
		t.Fatal("expected invalid regex to produce an error")
	}
}

func TestNewFromReader(t *testing.T) {
	r := strings.NewReader("https://example\\.com\n!https://example\\.com/admin\n")
	s, err := NewFromReader(r)
	if err != nil {
		t.Fatal(err)
	}

	if !s.IsInScope("https://example.com/home") {
		t.Fatal("expected in-scope url to be included")
	}
	if s.IsInScope("https://example.com/admin") {
		t.Fatal("expected excluded url to be out of scope")
	}
}
