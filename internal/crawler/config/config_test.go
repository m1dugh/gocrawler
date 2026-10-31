package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/m1dugh/gocrawler/internal/crawler/config"
)

func write(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "gocrawler.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadFull(t *testing.T) {
	path := write(t, `
seeds:
  - https://example.com/
scope:
  file: scope.txt
  rules:
    - '!^https://example\.com/logout'
engine: browser
browser:
  url: http://127.0.0.1:9222
  headless: false
  flags: ["--window-size=1280,800"]
headers:
  - "Authorization: Bearer abc"
timeout: 15s
workers: 8
transformers:
  defaults: false
  enable:
    - content
    - name: word_count
      options:
        min_length: 3
  disable:
    - status
output:
  fields: [full_url, status]
  sinks:
    - type: stdout
    - type: file
      path: crawl.jsonl
      format: json
      fields: [full_url]
`)

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Seeds[0] != "https://example.com/" || cfg.Scope.File != "scope.txt" || len(cfg.Scope.Rules) != 1 {
		t.Errorf("seeds/scope = %v / %+v", cfg.Seeds, cfg.Scope)
	}
	if cfg.Engine != "browser" || cfg.Browser.URL != "http://127.0.0.1:9222" ||
		cfg.Browser.Headless == nil || *cfg.Browser.Headless {
		t.Errorf("engine/browser = %q / %+v", cfg.Engine, cfg.Browser)
	}
	if cfg.Timeout != 15*time.Second || cfg.Workers != 8 {
		t.Errorf("timeout/workers = %v / %d", cfg.Timeout, cfg.Workers)
	}

	tc := cfg.Transformers
	if tc.Defaults == nil || *tc.Defaults {
		t.Errorf("Defaults = %v, want false", tc.Defaults)
	}
	if len(tc.Enable) != 2 || tc.Enable[0].Name != "content" ||
		tc.Enable[1].Name != "word_count" || tc.Enable[1].Options["min_length"] != "3" {
		t.Errorf("Enable = %+v", tc.Enable)
	}
	if len(tc.Disable) != 1 || tc.Disable[0] != "status" {
		t.Errorf("Disable = %v", tc.Disable)
	}

	out := cfg.Output
	if len(out.Fields) != 2 || len(out.Sinks) != 2 {
		t.Fatalf("output = %+v", out)
	}
	file := out.Sinks[1]
	if file.Type != "file" || file.Options["path"] != "crawl.jsonl" || file.Options["format"] != "json" ||
		len(file.Fields) != 1 || file.Fields[0] != "full_url" {
		t.Errorf("file sink = %+v", file)
	}
	if _, ok := file.Options["fields"]; ok {
		t.Errorf("file sink options = %v, fields must not be an option", file.Options)
	}
}

func TestLoadEmptyFile(t *testing.T) {
	if _, err := config.Load(write(t, "")); err != nil {
		t.Fatalf("Load() of an empty file error = %v", err)
	}
}

func TestLoadRejectsUnknownKey(t *testing.T) {
	if _, err := config.Load(write(t, "wokers: 3\n")); err == nil {
		t.Fatal("Load() with a misspelled key: expected error, got nil")
	}
}

func TestLoadRejectsSinkWithoutType(t *testing.T) {
	if _, err := config.Load(write(t, "output:\n  sinks:\n    - path: x\n")); err == nil {
		t.Fatal("Load() with a sink lacking a type: expected error, got nil")
	}
}

func TestLoadMissingFile(t *testing.T) {
	if _, err := config.Load(filepath.Join(t.TempDir(), "nope.yaml")); err == nil {
		t.Fatal("Load() of a missing file: expected error, got nil")
	}
}

func TestParseSpec(t *testing.T) {
	tests := []struct {
		spec     string
		wantName string
		wantOpts map[string]string
		wantErr  bool
	}{
		{"stdout", "stdout", map[string]string{}, false},
		{"file:path=a.jsonl", "file", map[string]string{"path": "a.jsonl"}, false},
		{"file:path=a.jsonl,format=json", "file", map[string]string{"path": "a.jsonl", "format": "json"}, false},
		{"file:", "file", map[string]string{}, false},
		{"file:path=a=b", "file", map[string]string{"path": "a=b"}, false},
		{":path=x", "", nil, true},
		{"file:path", "", nil, true},
		{"file:=x", "", nil, true},
	}
	for _, tt := range tests {
		name, opts, err := config.ParseSpec(tt.spec)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseSpec(%q) error = %v, wantErr %v", tt.spec, err, tt.wantErr)
			continue
		}
		if err != nil {
			continue
		}
		if name != tt.wantName || len(opts) != len(tt.wantOpts) {
			t.Errorf("ParseSpec(%q) = %q, %v", tt.spec, name, opts)
			continue
		}
		for k, v := range tt.wantOpts {
			if opts[k] != v {
				t.Errorf("ParseSpec(%q) opts[%q] = %q, want %q", tt.spec, k, opts[k], v)
			}
		}
	}
}
