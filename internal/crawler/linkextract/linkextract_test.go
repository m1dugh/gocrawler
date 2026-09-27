package linkextract_test

import (
	"reflect"
	"sort"
	"testing"

	"github.com/m1dugh/gocrawler/internal/crawler/linkextract"
)

func TestExtractLinksAbsoluteURL(t *testing.T) {
	content := `check out https://example.com/page?q=1 for more`

	got := linkextract.ExtractLinks(content, "https://example.com/")
	want := []string{"https://example.com/page?q=1"}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("ExtractLinks() = %v, want %v", got, want)
	}
}

func TestExtractLinksHrefAndSrcRelative(t *testing.T) {
	content := `<a href="/relative/path">link</a><img src="/img.png">`

	got := linkextract.ExtractLinks(content, "https://example.com/base/page")
	sort.Strings(got)

	want := []string{
		"https://example.com/img.png",
		"https://example.com/relative/path",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ExtractLinks() = %v, want %v", got, want)
	}
}

func TestExtractLinksRelativeWithoutLeadingSlash(t *testing.T) {
	content := `<a href="sibling.html">sibling</a>`

	got := linkextract.ExtractLinks(content, "https://example.com/base/page.html")
	want := []string{"https://example.com/base/sibling.html"}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("ExtractLinks() = %v, want %v", got, want)
	}
}

func TestExtractLinksMixed(t *testing.T) {
	content := `<a href="https://example.com/a">a</a> plain https://example.com/b text`

	got := linkextract.ExtractLinks(content, "https://example.com/")
	sort.Strings(got)

	want := []string{"https://example.com/a", "https://example.com/b"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ExtractLinks() = %v, want %v", got, want)
	}
}

func TestExtractLinksNone(t *testing.T) {
	got := linkextract.ExtractLinks("no links in here", "https://example.com/")
	if len(got) != 0 {
		t.Errorf("ExtractLinks() = %v, want empty", got)
	}
}

func TestExtractLinksInvalidBase(t *testing.T) {
	got := linkextract.ExtractLinks(`<a href="/a">a</a>`, "://not-a-url")
	if got != nil {
		t.Errorf("ExtractLinks() with invalid base = %v, want nil", got)
	}
}
