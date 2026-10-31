package builtin_test

import (
	"reflect"
	"testing"

	"github.com/m1dugh/gocrawler/internal/crawler/output/builtin"
)

func TestDefaultTransformers(t *testing.T) {
	var got []string
	for _, spec := range builtin.NewRegistry().Transformers() {
		if spec.Default {
			got = append(got, spec.Name)
		}
	}

	want := []string{"full_url", "status", "content_length"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("default transformers = %v, want %v", got, want)
	}
}

func TestBuiltinSinksAndFormats(t *testing.T) {
	r := builtin.NewRegistry()

	if _, err := r.NewSink("stdout", nil); err != nil {
		t.Errorf("NewSink(stdout) error = %v", err)
	}
	if _, err := r.NewSink("file", nil); err == nil {
		t.Error("NewSink(file) without path: expected error, got nil")
	}
	if got := r.Formats(); !reflect.DeepEqual(got, []string{"json", "tsv"}) {
		t.Errorf("Formats() = %v, want [json tsv]", got)
	}
}
