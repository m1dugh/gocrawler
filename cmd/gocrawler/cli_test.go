package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/m1dugh/gocrawler/internal/crawler/config"
	"github.com/m1dugh/gocrawler/internal/crawler/output/builtin"
)

func names(t *testing.T, tc config.Transformers) []string {
	t.Helper()
	steps, err := activeTransformers(tc, builtin.NewRegistry())
	if err != nil {
		t.Fatalf("activeTransformers() error = %v", err)
	}
	var got []string
	for _, s := range steps {
		got = append(got, s.Name())
	}
	return got
}

func ptr[T any](v T) *T { return &v }

func TestActiveTransformers(t *testing.T) {
	tests := map[string]struct {
		tc   config.Transformers
		want []string
	}{
		"defaults": {
			config.Transformers{},
			[]string{"full_url", "status", "content_length"},
		},
		"disable one default": {
			config.Transformers{Disable: []string{"status"}},
			[]string{"full_url", "content_length"},
		},
		"enable runs after the defaults": {
			config.Transformers{Enable: []config.Transformer{{Name: "content"}, {Name: "timestamp"}}},
			[]string{"full_url", "status", "content_length", "content", "timestamp"},
		},
		"no defaults": {
			config.Transformers{Defaults: ptr(false), Enable: []config.Transformer{{Name: "status"}, {Name: "full_url"}}},
			[]string{"status", "full_url"},
		},
		"enabling a default keeps it in place": {
			config.Transformers{Enable: []config.Transformer{{Name: "status"}}},
			[]string{"full_url", "status", "content_length"},
		},
		"disabling a non-default is a no-op": {
			config.Transformers{Disable: []string{"content"}},
			[]string{"full_url", "status", "content_length"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := names(t, tt.tc); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("active = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestActiveTransformersUnknown(t *testing.T) {
	reg := builtin.NewRegistry()
	for name, tc := range map[string]config.Transformers{
		"unknown disable": {Disable: []string{"nope"}},
		"unknown enable":  {Enable: []config.Transformer{{Name: "nope"}}},
		"bad option":      {Enable: []config.Transformer{{Name: "content", Options: map[string]string{"x": "1"}}}},
	} {
		_, err := activeTransformers(tc, reg)
		var ue usageError
		if !errors.As(err, &ue) {
			t.Errorf("%s: error = %v, want a usage error", name, err)
		}
	}
}

func TestCheckFieldsExplainsTheProvider(t *testing.T) {
	reg := builtin.NewRegistry()
	tests := []struct {
		name string
		tc   config.Transformers
		want string
	}{
		{"disabled default", config.Transformers{Disable: []string{"status"}},
			`field "status" is provided by transformer "status", which is disabled`},
		{"not enabled", config.Transformers{},
			`field "content" is provided by transformer "content", which is not enabled`},
	}

	for _, tt := range tests {
		pipe, err := buildPipeline(tt.tc, reg)
		if err != nil {
			t.Fatal(err)
		}
		field := strings.Split(tt.want, `"`)[1]
		err = checkFields([]string{field}, pipe.Provided(), tt.tc, reg)
		if err == nil || err.Error() != tt.want {
			t.Errorf("%s: error = %v, want %q", tt.name, err, tt.want)
		}
	}

	pipe, _ := buildPipeline(config.Transformers{}, reg)
	if err := checkFields([]string{"bogus"}, pipe.Provided(), config.Transformers{}, reg); err == nil {
		t.Error("an unknown field: expected error, got nil")
	}
	if err := checkFields([]string{"full_url", "status"}, pipe.Provided(), config.Transformers{}, reg); err != nil {
		t.Errorf("provided fields: error = %v", err)
	}
}

func TestParseHeaders(t *testing.T) {
	h, err := parseHeaders([]string{"authorization: Bearer abc", "X-A: 1", "X-A: 2"})
	if err != nil {
		t.Fatalf("parseHeaders() error = %v", err)
	}
	if h["Authorization"][0] != "Bearer abc" || !reflect.DeepEqual(h["X-A"], []string{"1", "2"}) {
		t.Errorf("headers = %v", h)
	}
	if _, err := parseHeaders([]string{"no colon"}); err == nil {
		t.Error("a header without a colon: expected error, got nil")
	}
}

// run executes the root command with args, returning what it wrote to
// stdout and the error.
func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func isUsage(err error) bool {
	var ue usageError
	return errors.As(err, &ue)
}

func TestUsageErrors(t *testing.T) {
	tests := map[string][]string{
		"no seed":             {"--scope-rule", "^https://example.com"},
		"no scope":            {"https://example.com/"},
		"unknown flag":        {"--bogus", "https://example.com/"},
		"unknown engine":      {"--engine", "ftp", "-S", "^https://example.com", "https://example.com/"},
		"unknown sink":        {"--sink", "nope", "-S", "^https://example.com", "https://example.com/"},
		"bad sink option":     {"--sink", "file:bogus=1", "-S", "^https://example.com", "https://example.com/"},
		"sink missing path":   {"--sink", "file", "-S", "^https://example.com", "https://example.com/"},
		"malformed sink":      {"--sink", "file:path", "-S", "^https://example.com", "https://example.com/"},
		"unprovided field":    {"--fields", "content", "-S", "^https://example.com", "https://example.com/"},
		"bad header":          {"-H", "nocolon", "-S", "^https://example.com", "https://example.com/"},
		"bad scope regex":     {"-S", "(", "https://example.com/"},
		"browser url, http":   {"--engine", "http", "--browser-url", "http://127.0.0.1:1", "-S", "^https://example.com", "https://example.com/"},
		"negative workers":    {"--workers=-1", "-S", "^https://example.com", "https://example.com/"},
		"unknown disable":     {"--no-transformer", "nope", "-S", "^https://example.com", "https://example.com/"},
		"unknown transformer": {"--transformer", "nope", "-S", "^https://example.com", "https://example.com/"},
	}
	for name, args := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := run(t, args...); !isUsage(err) {
				t.Errorf("error = %v, want a usage error", err)
			}
		})
	}
}

func TestMissingConfigFileIsNotAUsageError(t *testing.T) {
	_, err := run(t, "--config", filepath.Join(t.TempDir(), "nope.yaml"))
	if err == nil || isUsage(err) {
		t.Errorf("error = %v, want a non-usage error", err)
	}
}

func site(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/a", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<a href="/b">b</a>`)
	})
	mux.HandleFunc("/b", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `leaf`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func lines(t *testing.T, path string) []string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	return strings.Split(strings.TrimSpace(string(data)), "\n")
}

func TestCrawlWritesSelectedSinksAndFields(t *testing.T) {
	srv := site(t)
	dir := t.TempDir()
	all, urls := filepath.Join(dir, "all.jsonl"), filepath.Join(dir, "urls.jsonl")

	_, err := run(t, "-S", "^"+srv.URL,
		"--sink", "file:path="+all,
		"--sink", "file:path="+urls,
		srv.URL+"/a")
	if err != nil {
		t.Fatalf("crawl error = %v", err)
	}

	// Default transformers only: full_url, status, content_length.
	for _, l := range lines(t, all) {
		for _, f := range []string{`"full_url"`, `"status"`, `"content_length"`} {
			if !strings.Contains(l, f) {
				t.Errorf("record %s lacks %s", l, f)
			}
		}
		if strings.Contains(l, `"content_hash"`) || strings.Contains(l, `"timestamp"`) {
			t.Errorf("record %s holds a field of a non-default transformer", l)
		}
	}
	if got := len(lines(t, all)); got != 2 {
		t.Errorf("wrote %d records, want 2", got)
	}
}

func TestCrawlFieldsAndTransformerFlags(t *testing.T) {
	srv := site(t)
	out := filepath.Join(t.TempDir(), "out.jsonl")

	_, err := run(t, "-S", "^"+srv.URL,
		"--transformer", "content_hash",
		"--no-transformer", "content_length",
		"--sink", "file:path="+out,
		"--fields", "full_url,content_hash",
		srv.URL+"/a")
	if err != nil {
		t.Fatalf("crawl error = %v", err)
	}

	for _, l := range lines(t, out) {
		if !strings.Contains(l, `"content_hash"`) || strings.Contains(l, `"status"`) {
			t.Errorf("record %s, want only full_url and content_hash", l)
		}
	}
}

func TestConfigFileWithPerSinkFieldsAndFlagOverride(t *testing.T) {
	srv := site(t)
	dir := t.TempDir()
	wide, narrow := filepath.Join(dir, "wide.jsonl"), filepath.Join(dir, "narrow.jsonl")
	cfgPath := filepath.Join(dir, "gocrawler.yaml")

	yaml := fmt.Sprintf(`
seeds: ["%[1]s/a"]
scope:
  rules: ["^%[1]s"]
transformers:
  enable: [content_hash]
output:
  fields: [full_url, status]
  sinks:
    - type: file
      path: %[2]s
      fields: [full_url, content_hash]
    - type: file
      path: %[3]s
`, srv.URL, wide, narrow)
	if err := os.WriteFile(cfgPath, []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := run(t, "--config", cfgPath); err != nil {
		t.Fatalf("crawl error = %v", err)
	}
	for _, l := range lines(t, wide) {
		if !strings.Contains(l, `"content_hash"`) || strings.Contains(l, `"status"`) {
			t.Errorf("wide sink record %s, want its own fields full_url, content_hash", l)
		}
	}
	for _, l := range lines(t, narrow) {
		if !strings.Contains(l, `"status"`) || strings.Contains(l, `"content_hash"`) {
			t.Errorf("narrow sink record %s, want the output-wide fields full_url, status", l)
		}
	}

	// --fields overrides every sink's own fields.
	os.Remove(wide)
	os.Remove(narrow)
	if _, err := run(t, "--config", cfgPath, "--fields", "full_url"); err != nil {
		t.Fatalf("crawl error = %v", err)
	}
	for _, path := range []string{wide, narrow} {
		for _, l := range lines(t, path) {
			if strings.Contains(l, `"status"`) || strings.Contains(l, `"content_hash"`) {
				t.Errorf("%s record %s, want only full_url", filepath.Base(path), l)
			}
		}
	}
}

func TestSinkFormatTSV(t *testing.T) {
	srv := site(t)
	dir := t.TempDir()
	tsv, js := filepath.Join(dir, "out.tsv"), filepath.Join(dir, "out.jsonl")

	_, err := run(t, "-S", "^"+srv.URL,
		"--sink", "file:path="+tsv+",format=tsv",
		"--sink", "file:path="+js,
		"--fields", "full_url,status",
		srv.URL+"/b")
	if err != nil {
		t.Fatalf("crawl error = %v", err)
	}

	// Alphabetical by field name: full_url, then status.
	if got, want := lines(t, tsv), []string{srv.URL + "/b\t200"}; !reflect.DeepEqual(got, want) {
		t.Errorf("tsv sink = %q, want %q", got, want)
	}
	if got := lines(t, js); len(got) != 1 || !strings.HasPrefix(got[0], "{") {
		t.Errorf("json sink = %q, want one JSON object", got)
	}

	_, err = run(t, "-S", "^"+srv.URL, "--sink", "stdout:format=xml", srv.URL+"/b")
	if !isUsage(err) {
		t.Errorf("unknown format: error = %v, want a usage error", err)
	}
}

func TestSeedsFileAndStdin(t *testing.T) {
	srv := site(t)
	dir := t.TempDir()
	out := filepath.Join(dir, "out.jsonl")
	seeds := filepath.Join(dir, "seeds.txt")
	os.WriteFile(seeds, []byte("\n"+srv.URL+"/b\n\n"), 0o644)

	if _, err := run(t, "-S", "^"+srv.URL, "-l", seeds, "--sink", "file:path="+out); err != nil {
		t.Fatalf("crawl error = %v", err)
	}
	if got := lines(t, out); len(got) != 1 || !strings.Contains(got[0], "/b") {
		t.Errorf("records = %v, want one for /b", got)
	}

	cmd := newRootCmd()
	cmd.SetIn(strings.NewReader(srv.URL + "/b\n"))
	out2 := filepath.Join(dir, "out2.jsonl")
	cmd.SetArgs([]string{"-S", "^" + srv.URL, "-l", "-", "--sink", "file:path=" + out2})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("crawl from stdin error = %v", err)
	}
	if got := lines(t, out2); len(got) != 1 {
		t.Errorf("records from stdin = %v, want one", got)
	}
}

func TestListingSubcommands(t *testing.T) {
	out, err := run(t, "transformers")
	if err != nil {
		t.Fatalf("transformers error = %v", err)
	}
	for _, want := range []string{"full_url", "content_hash", "NAME", "DEFAULT"} {
		if !strings.Contains(out, want) {
			t.Errorf("transformers output lacks %q:\n%s", want, out)
		}
	}

	out, err = run(t, "sinks")
	if err != nil {
		t.Fatalf("sinks error = %v", err)
	}
	for _, want := range []string{"stdout", "file", "path", "(required)", "format"} {
		if !strings.Contains(out, want) {
			t.Errorf("sinks output lacks %q:\n%s", want, out)
		}
	}

	out, err = run(t, "fields", "--transformer", "content", "--no-transformer", "status")
	if err != nil {
		t.Fatalf("fields error = %v", err)
	}
	for field, active := range map[string]string{"content": "yes", "status": "no", "full_url": "yes", "timestamp": "no"} {
		found := false
		for _, l := range strings.Split(out, "\n") {
			if f := strings.Fields(l); len(f) == 3 && f[0] == field {
				found = true
				if f[2] != active {
					t.Errorf("field %s active = %s, want %s", field, f[2], active)
				}
			}
		}
		if !found {
			t.Errorf("fields output lacks %s:\n%s", field, out)
		}
	}
}

func TestVersion(t *testing.T) {
	out, err := run(t, "--version")
	if err != nil || !strings.HasPrefix(out, "gocrawler ") {
		t.Errorf("--version = %q, %v", out, err)
	}
}
