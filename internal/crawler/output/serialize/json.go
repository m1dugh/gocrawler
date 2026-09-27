// Package serialize implements output.Serializer.
package serialize

import (
	"encoding/json"

	"github.com/m1dugh/gocrawler/pkg/crawler/engine"
)

// JSON serializes a Record as a single JSON object.
type JSON struct{}

// Serialize encodes r as JSON.
func (JSON) Serialize(r engine.Record) ([]byte, error) {
	return json.Marshal(r)
}
