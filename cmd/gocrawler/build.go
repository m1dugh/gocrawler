package main

import (
	"errors"
	"fmt"
	"net/textproto"
	"os"
	"strings"
	"time"

	"github.com/m1dugh/gocrawler/internal/crawler/config"
	browserengine "github.com/m1dugh/gocrawler/internal/crawler/engine/browser"
	httpengine "github.com/m1dugh/gocrawler/internal/crawler/engine/http"
	"github.com/m1dugh/gocrawler/internal/crawler/output/pipeline"
	"github.com/m1dugh/gocrawler/internal/crawler/queue"
	"github.com/m1dugh/gocrawler/internal/crawler/scope"
	"github.com/m1dugh/gocrawler/internal/crawler/urlstate"
	"github.com/m1dugh/gocrawler/pkg/crawler/engine"
	"github.com/m1dugh/gocrawler/pkg/crawler/output"
)

const (
	defaultWorkers = 4
	defaultTimeout = 30 * time.Second
)

// build turns a merged configuration into everything Run needs. The
// returned function releases the engine and the sinks.
func build(c *config.Config, reg *output.Registry) (Config, func(), error) {
	if len(c.Seeds) == 0 {
		return Config{}, nil, usagef("no seed url given")
	}
	if c.Scope.File == "" && len(c.Scope.Rules) == 0 {
		return Config{}, nil, usagef("no scope given: use --scope or --scope-rule")
	}

	s, err := buildScope(c.Scope)
	if err != nil {
		return Config{}, nil, err
	}
	q := queue.New(s)
	for _, seed := range c.Seeds {
		if err := q.Push(seed); err != nil {
			fmt.Fprintf(os.Stderr, "gocrawler: seed %q: %v\n", seed, err)
		}
	}

	pipe, err := buildPipeline(c.Transformers, reg)
	if err != nil {
		return Config{}, nil, err
	}

	bindings, err := buildSinks(c, pipe, reg)
	if err != nil {
		return Config{}, nil, err
	}
	closeSinks := func() {
		for _, b := range bindings {
			b.Sink.Close()
		}
	}

	eng, err := buildEngine(c)
	if err != nil {
		closeSinks()
		return Config{}, nil, err
	}

	workers := c.Workers
	if workers == 0 {
		workers = defaultWorkers
	}
	if workers < 0 {
		closeSinks()
		eng.Close()
		return Config{}, nil, usagef("--workers must be positive, got %d", workers)
	}

	return Config{
			Queue:    q,
			Store:    urlstate.NewStore(),
			Engine:   eng,
			Pipeline: pipe,
			Sinks:    bindings,
			Workers:  workers,
		}, func() {
			eng.Close()
			closeSinks()
		}, nil
}

// buildScope reads the scope file, if any, then adds the inline rules.
func buildScope(sc config.Scope) (*scope.Scope, error) {
	var rules []string
	if sc.File != "" {
		data, err := os.ReadFile(sc.File)
		if err != nil {
			return nil, err
		}
		rules = append(rules, strings.Split(string(data), "\n")...)
	}
	rules = append(rules, sc.Rules...)

	s, err := scope.New(rules)
	if err != nil {
		return nil, usageError{err}
	}
	return s, nil
}

// activeTransformers resolves the transformers to run, in order: the
// default ones that are not disabled, then the enabled ones.
func activeTransformers(tc config.Transformers, reg *output.Registry) ([]output.Transformer, error) {
	disabled := map[string]bool{}
	for _, name := range tc.Disable {
		if _, ok := reg.Transformer(name); !ok {
			return nil, usagef("--no-transformer: unknown transformer %q", name)
		}
		disabled[name] = true
	}

	var list []output.Transformer
	index := map[string]int{}
	add := func(name string, opts output.Options) error {
		t, err := reg.NewTransformer(name, opts)
		if err != nil {
			return wrapInvalid(err)
		}
		if i, ok := index[name]; ok {
			list[i] = t
			return nil
		}
		index[name] = len(list)
		list = append(list, t)
		return nil
	}

	if tc.Defaults == nil || *tc.Defaults {
		for _, spec := range reg.Transformers() {
			if spec.Default && !disabled[spec.Name] {
				if err := add(spec.Name, nil); err != nil {
					return nil, err
				}
			}
		}
	}
	for _, e := range tc.Enable {
		opts := output.Options(e.Options)
		if _, ok := index[e.Name]; ok && len(opts) == 0 {
			continue
		}
		if err := add(e.Name, opts); err != nil {
			return nil, err
		}
	}
	return list, nil
}

func buildPipeline(tc config.Transformers, reg *output.Registry) (*pipeline.Pipeline, error) {
	steps, err := activeTransformers(tc, reg)
	if err != nil {
		return nil, err
	}
	b := pipeline.NewBuilder()
	for _, t := range steps {
		b.Add(t)
	}
	p, err := b.Build()
	if err != nil {
		return nil, usageError{err}
	}
	return p, nil
}

// buildSinks creates the selected sinks (stdout when none is selected)
// and resolves the fields each one writes: its own, else the output-wide
// ones, else every field the pipeline provides (nil).
func buildSinks(c *config.Config, pipe *pipeline.Pipeline, reg *output.Registry) ([]SinkBinding, error) {
	entries := c.Output.Sinks
	if len(entries) == 0 {
		entries = []config.Sink{{Type: "stdout"}}
	}

	provided := pipe.Provided()
	var bindings []SinkBinding
	closeAll := func() {
		for _, b := range bindings {
			b.Sink.Close()
		}
	}

	for _, e := range entries {
		fields := e.Fields
		if len(fields) == 0 {
			fields = c.Output.Fields
		}
		if err := checkFields(fields, provided, c.Transformers, reg); err != nil {
			closeAll()
			return nil, err
		}

		sink, err := reg.NewSink(e.Type, output.Options(e.Options))
		if err != nil {
			closeAll()
			return nil, wrapInvalid(err)
		}
		bindings = append(bindings, SinkBinding{Sink: sink, Fields: fields})
	}
	return bindings, nil
}

// checkFields makes sure every wanted field is provided by an active
// transformer, naming the transformer that would provide a missing one.
func checkFields(wanted, provided []string, tc config.Transformers, reg *output.Registry) error {
	have := map[string]bool{}
	for _, f := range provided {
		have[f] = true
	}

	for _, f := range wanted {
		if have[f] {
			continue
		}
		for _, spec := range reg.Transformers() {
			t, err := reg.NewTransformer(spec.Name, nil)
			if err != nil || !containsString(t.Provides(), f) {
				continue
			}
			state := "not enabled"
			if containsString(tc.Disable, spec.Name) {
				state = "disabled"
			}
			return usagef("field %q is provided by transformer %q, which is %s", f, spec.Name, state)
		}
		return usagef("field %q is not provided by any transformer", f)
	}
	return nil
}

func buildEngine(c *config.Config) (engine.Engine, error) {
	headers, err := parseHeaders(c.Headers)
	if err != nil {
		return nil, err
	}
	timeout := c.Timeout
	if timeout == 0 {
		timeout = defaultTimeout
	}

	kind := c.Engine
	if kind == "" {
		kind = "http"
		if c.Browser.URL != "" {
			kind = "browser"
		}
	}

	switch kind {
	case "http":
		if c.Browser.URL != "" {
			return nil, usagef("--browser-url requires the browser engine, not %q", kind)
		}
		return httpengine.New(httpengine.Config{Headers: headers, Timeout: timeout}), nil
	case "browser":
		if c.Browser.URL != "" {
			return browserengine.NewRemote(browserengine.RemoteConfig{
				WebsocketURL: c.Browser.URL,
				Headers:      headers,
				Timeout:      timeout,
			})
		}
		headless := c.Browser.Headless == nil || *c.Browser.Headless
		return browserengine.NewLocal(browserengine.LocalConfig{
			BinaryPath: c.Browser.Path,
			Headless:   headless,
			Flags:      c.Browser.Flags,
			Headers:    headers,
			Timeout:    timeout,
		}), nil
	default:
		return nil, usagef("unknown engine %q: want http or browser", kind)
	}
}

// parseHeaders turns "Name: value" lines into a header map.
func parseHeaders(lines []string) (map[string][]string, error) {
	if len(lines) == 0 {
		return nil, nil
	}
	headers := map[string][]string{}
	for _, line := range lines {
		name, value, ok := strings.Cut(line, ":")
		name = strings.TrimSpace(name)
		if !ok || name == "" {
			return nil, usagef("header %q is not \"Name: value\"", line)
		}
		key := textproto.CanonicalMIMEHeaderKey(name)
		headers[key] = append(headers[key], strings.TrimSpace(value))
	}
	return headers, nil
}

// wrapInvalid marks registry selection mistakes as usage errors.
func wrapInvalid(err error) error {
	var ie *output.InvalidError
	if errors.As(err, &ie) {
		return usageError{err}
	}
	return err
}

func containsString(list []string, s string) bool {
	for _, e := range list {
		if e == s {
			return true
		}
	}
	return false
}
