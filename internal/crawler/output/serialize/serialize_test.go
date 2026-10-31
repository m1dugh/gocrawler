package serialize_test

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/m1dugh/gocrawler/internal/crawler/output/serialize"
	"github.com/m1dugh/gocrawler/pkg/crawler/engine"
)

func TestJSON(t *testing.T) {
	r := engine.Record{"full_url": "https://example.com", "status": float64(200)}

	b, err := serialize.JSON{}.Serialize(r)
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	var got engine.Record
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if !reflect.DeepEqual(got, r) {
		t.Errorf("round-tripped record = %+v, want %+v", got, r)
	}
}

func TestTSV(t *testing.T) {
	r := engine.Record{
		"status":    200,
		"full_url":  "https://example.com/a",
		"timestamp": time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		"ok":        true,
		"missing":   nil,
		"tags":      []string{"a", "b"},
	}

	b, err := serialize.TSV{}.Serialize(r)
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	// Alphabetical by field name: full_url, missing, ok, status, tags, timestamp.
	want := "https://example.com/a\t\ttrue\t200\t[\"a\",\"b\"]\t2026-01-01T00:00:00Z"
	if string(b) != want {
		t.Errorf("Serialize() = %q, want %q", b, want)
	}
}

func TestTSVEscapesSeparators(t *testing.T) {
	r := engine.Record{"content": "a\tb\nc\r\\d"}

	b, err := serialize.TSV{}.Serialize(r)
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}
	if want := `a\tb\nc\r\\d`; string(b) != want {
		t.Errorf("Serialize() = %q, want %q", b, want)
	}
	if bytes.ContainsAny(b, "\t\n\r") {
		t.Errorf("Serialize() = %q, must hold no raw tab or line break", b)
	}
}

func TestTSVStableAcrossRecords(t *testing.T) {
	a, _ := serialize.TSV{}.Serialize(engine.Record{"b": 1, "a": 2})
	b, _ := serialize.TSV{}.Serialize(engine.Record{"a": 2, "b": 1})
	if string(a) != string(b) || string(a) != "2\t1" {
		t.Errorf("Serialize() = %q and %q, want both %q", a, b, "2\t1")
	}
}

func TestTSVEmptyRecord(t *testing.T) {
	b, err := serialize.TSV{}.Serialize(engine.Record{})
	if err != nil || len(b) != 0 {
		t.Errorf("Serialize() = %q, %v, want an empty line", b, err)
	}
}

func TestTSVUnserializableValue(t *testing.T) {
	if _, err := (serialize.TSV{}).Serialize(engine.Record{"f": make(chan int)}); err == nil {
		t.Error("Serialize() with a channel value: expected error, got nil")
	}
}
