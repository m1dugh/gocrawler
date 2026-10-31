package serialize

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/m1dugh/gocrawler/pkg/crawler/engine"
)

// TSV serializes a Record as one line of tab-separated values, with no
// header. Columns follow the alphabetical order of the field names, so
// every record of a crawl has the same layout as long as it has the same
// fields. A value is escaped so it never breaks the line or the columns: a
// tab, a newline, a carriage return and a backslash become \t, \n, \r and
// \\.
type TSV struct{}

var tsvEscaper = strings.NewReplacer(`\`, `\\`, "\t", `\t`, "\n", `\n`, "\r", `\r`)

// Serialize encodes the values of r as a TSV line, without the trailing
// newline.
func (TSV) Serialize(r engine.Record) ([]byte, error) {
	keys := make([]string, 0, len(r))
	for k := range r {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	cols := make([]string, len(keys))
	for i, k := range keys {
		v, err := tsvValue(r[k])
		if err != nil {
			return nil, fmt.Errorf("tsv: field %q: %w", k, err)
		}
		cols[i] = tsvEscaper.Replace(v)
	}
	return []byte(strings.Join(cols, "\t")), nil
}

// tsvValue renders one value: text as is, a time in RFC 3339 like the JSON
// format, nothing for nil, and JSON for anything structured.
func tsvValue(v interface{}) (string, error) {
	switch v := v.(type) {
	case nil:
		return "", nil
	case string:
		return v, nil
	case time.Time:
		return v.Format(time.RFC3339Nano), nil
	case bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return fmt.Sprint(v), nil
	default:
		b, err := json.Marshal(v)
		return string(b), err
	}
}
