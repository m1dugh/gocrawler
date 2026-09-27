package sink_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/m1dugh/gocrawler/internal/crawler/output/serialize"
	"github.com/m1dugh/gocrawler/internal/crawler/output/sink"
	"github.com/m1dugh/gocrawler/pkg/crawler/engine"
)

func TestFileWrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.jsonl")

	f, err := sink.NewFile(path, serialize.JSON{})
	if err != nil {
		t.Fatalf("NewFile() error = %v", err)
	}
	defer f.Close()

	if err := f.Write(engine.Record{"full_url": "https://example.com/a"}); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := f.Write(engine.Record{"full_url": "https://example.com/b"}); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	f.Close()

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading output file: %v", err)
	}

	lines := bytes.Count(got, []byte("\n"))
	if lines != 2 {
		t.Errorf("wrote %d lines, want 2 (content: %q)", lines, got)
	}
}

func TestFileAppendsAcrossOpens(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.jsonl")

	f1, err := sink.NewFile(path, serialize.JSON{})
	if err != nil {
		t.Fatalf("NewFile() error = %v", err)
	}
	if err := f1.Write(engine.Record{"full_url": "first"}); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	f1.Close()

	f2, err := sink.NewFile(path, serialize.JSON{})
	if err != nil {
		t.Fatalf("NewFile() error = %v", err)
	}
	if err := f2.Write(engine.Record{"full_url": "second"}); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	f2.Close()

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading output file: %v", err)
	}
	if bytes.Count(got, []byte("\n")) != 2 {
		t.Errorf("expected appended content with 2 lines, got %q", got)
	}
}

func TestStdoutWrite(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}

	stdout := sink.NewStdout(serialize.JSON{})
	stdout.SetWriter(w)

	if err := stdout.Write(engine.Record{"full_url": "https://example.com"}); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	w.Close()

	got, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("reading pipe: %v", err)
	}
	if !bytes.Contains(got, []byte("https://example.com")) {
		t.Errorf("output = %q, want it to contain the record url", got)
	}
}
