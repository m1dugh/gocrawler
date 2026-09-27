package hash_test

import (
	"testing"

	"github.com/m1dugh/gocrawler/internal/crawler/hash"
)

func TestContentDeterministic(t *testing.T) {
	if hash.Content("hello") != hash.Content("hello") {
		t.Fatal("Content() is not deterministic for identical input")
	}
}

func TestContentDiffers(t *testing.T) {
	if hash.Content("hello") == hash.Content("world") {
		t.Fatal("Content() produced the same hash for different input")
	}
}
