// Package config describes the crawler's configuration and loads it from a
// YAML file. Every key mirrors a command-line flag.
package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config is the whole crawler configuration. Zero values mean "not set",
// so defaults can be applied after the file and the command-line flags
// have been merged.
type Config struct {
	Seeds        []string      `yaml:"seeds"`
	Scope        Scope         `yaml:"scope"`
	Engine       string        `yaml:"engine"`
	Browser      Browser       `yaml:"browser"`
	Headers      []string      `yaml:"headers"`
	Timeout      time.Duration `yaml:"timeout"`
	Workers      int           `yaml:"workers"`
	Transformers Transformers  `yaml:"transformers"`
	Output       Output        `yaml:"output"`
}

// Scope holds the scope rules, from a file and/or given inline.
type Scope struct {
	File  string   `yaml:"file"`
	Rules []string `yaml:"rules"`
}

// Browser configures the browser engine.
type Browser struct {
	// URL is the remote debugging url of an existing browser.
	URL string `yaml:"url"`
	// Path is the browser binary to spawn.
	Path string `yaml:"path"`
	// Headless defaults to true when nil.
	Headless *bool    `yaml:"headless"`
	Flags    []string `yaml:"flags"`
}

// Transformers selects the transformers of the pipeline.
type Transformers struct {
	// Defaults defaults to true when nil; false starts from an empty
	// pipeline.
	Defaults *bool `yaml:"defaults"`
	// Enable lists transformers to run after the default ones.
	Enable []Transformer `yaml:"enable"`
	// Disable lists default transformers to turn off.
	Disable []string `yaml:"disable"`
}

// Transformer is one enabled transformer: a name and its options. In YAML
// it is either a bare name or a mapping with "name" and "options".
type Transformer struct {
	Name    string            `yaml:"name"`
	Options map[string]string `yaml:"options"`
}

// UnmarshalYAML accepts both the bare-name and the mapping form.
func (t *Transformer) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind == yaml.ScalarNode {
		t.Name = n.Value
		return nil
	}
	type plain Transformer
	return n.Decode((*plain)(t))
}

// Output configures what is written and where.
type Output struct {
	// Fields are written by sinks that have no fields of their own. Empty
	// means every field provided by the active transformers.
	Fields []string `yaml:"fields"`
	// Sinks empty means stdout only.
	Sinks []Sink `yaml:"sinks"`
}

// Sink is one selected sink. Every key other than "type" and "fields" is
// an option of the sink.
type Sink struct {
	Type    string
	Fields  []string
	Options map[string]string
}

// UnmarshalYAML reads "type" and "fields", keeping the rest as options.
func (s *Sink) UnmarshalYAML(n *yaml.Node) error {
	var raw struct {
		Type    string            `yaml:"type"`
		Fields  []string          `yaml:"fields"`
		Options map[string]string `yaml:",inline"`
	}
	if err := n.Decode(&raw); err != nil {
		return err
	}
	if raw.Type == "" {
		return fmt.Errorf("line %d: sink has no type", n.Line)
	}
	s.Type, s.Fields, s.Options = raw.Type, raw.Fields, raw.Options
	return nil
}

// Load reads and parses the YAML config file at path. Unknown keys are an
// error, so typos do not go unnoticed.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&cfg); err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return &cfg, nil
}

// ParseSpec splits "name[:key=value,key=value]" into a name and options.
func ParseSpec(spec string) (name string, opts map[string]string, err error) {
	name, rest, hasOpts := strings.Cut(spec, ":")
	if name == "" {
		return "", nil, fmt.Errorf("%q: missing name", spec)
	}
	opts = map[string]string{}
	if !hasOpts || rest == "" {
		return name, opts, nil
	}
	for _, kv := range strings.Split(rest, ",") {
		k, v, ok := strings.Cut(kv, "=")
		if !ok || k == "" {
			return "", nil, fmt.Errorf("%q: option %q is not key=value", spec, kv)
		}
		opts[k] = v
	}
	return name, opts, nil
}
