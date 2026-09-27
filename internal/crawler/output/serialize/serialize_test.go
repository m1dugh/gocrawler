package serialize_test

import (
	"encoding/json"
	"reflect"
	"testing"

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
