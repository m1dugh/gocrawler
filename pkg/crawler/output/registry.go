package output

import (
	"fmt"
	"sort"
)

// DefaultFormat is the serialization format used by a sink when none is
// given through its "format" option.
const DefaultFormat = "json"

// InvalidError reports a selection that cannot be satisfied: an unknown
// transformer, sink or format, or a bad option. It is a mistake in the
// configuration rather than a runtime failure.
type InvalidError struct{ msg string }

func (e *InvalidError) Error() string { return e.msg }

func invalidf(format string, args ...any) error {
	return &InvalidError{fmt.Sprintf(format, args...)}
}

// Options are the key=value options given to a transformer or a sink.
type Options map[string]string

// OptionSpec documents one option accepted by a transformer or a sink.
type OptionSpec struct {
	Name        string
	Description string
	Required    bool
}

// TransformerSpec describes a transformer that can be enabled by name.
type TransformerSpec struct {
	Name        string
	Description string
	// Default marks a transformer that is active unless disabled.
	Default bool
	Options []OptionSpec
	New     func(Options) (Transformer, error)
}

// SinkSpec describes a sink that can be selected by name. Every sink also
// accepts the "format" option, handled by the Registry.
type SinkSpec struct {
	Name        string
	Description string
	Options     []OptionSpec
	New         func(opts Options, ser Serializer) (Sink, error)
}

// Registry holds the transformers, sinks and serializers available to a
// crawl, in registration order for transformers.
type Registry struct {
	transformers []TransformerSpec
	sinks        map[string]SinkSpec
	serializers  map[string]Serializer
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		sinks:       map[string]SinkSpec{},
		serializers: map[string]Serializer{},
	}
}

// RegisterTransformer adds a transformer. It panics on a duplicate name, as
// registration is a programming step.
func (r *Registry) RegisterTransformer(s TransformerSpec) {
	if _, ok := r.Transformer(s.Name); ok {
		panic(fmt.Sprintf("output: transformer %q registered twice", s.Name))
	}
	r.transformers = append(r.transformers, s)
}

// RegisterSink adds a sink. It panics on a duplicate name.
func (r *Registry) RegisterSink(s SinkSpec) {
	if _, ok := r.sinks[s.Name]; ok {
		panic(fmt.Sprintf("output: sink %q registered twice", s.Name))
	}
	r.sinks[s.Name] = s
}

// RegisterSerializer adds a serialization format. It panics on a duplicate
// name.
func (r *Registry) RegisterSerializer(name string, s Serializer) {
	if _, ok := r.serializers[name]; ok {
		panic(fmt.Sprintf("output: serializer %q registered twice", name))
	}
	r.serializers[name] = s
}

// Transformers lists the registered transformers in registration order.
func (r *Registry) Transformers() []TransformerSpec {
	return append([]TransformerSpec(nil), r.transformers...)
}

// Transformer returns the transformer registered under name.
func (r *Registry) Transformer(name string) (TransformerSpec, bool) {
	for _, s := range r.transformers {
		if s.Name == name {
			return s, true
		}
	}
	return TransformerSpec{}, false
}

// Sinks lists the registered sinks sorted by name.
func (r *Registry) Sinks() []SinkSpec {
	out := make([]SinkSpec, 0, len(r.sinks))
	for _, s := range r.sinks {
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// Formats lists the registered serialization formats, sorted.
func (r *Registry) Formats() []string {
	out := make([]string, 0, len(r.serializers))
	for name := range r.serializers {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// NewTransformer builds the transformer registered under name from opts,
// rejecting unknown options and missing required ones.
func (r *Registry) NewTransformer(name string, opts Options) (Transformer, error) {
	spec, ok := r.Transformer(name)
	if !ok {
		return nil, invalidf("unknown transformer %q", name)
	}
	if err := checkOptions(spec.Options, opts, nil); err != nil {
		return nil, invalidf("transformer %q: %v", name, err)
	}
	return spec.New(opts)
}

// NewSink builds the sink registered under name from opts. The "format"
// option picks the serializer and defaults to DefaultFormat; it is not
// passed on to the sink's own options.
func (r *Registry) NewSink(name string, opts Options) (Sink, error) {
	spec, ok := r.sinks[name]
	if !ok {
		return nil, invalidf("unknown sink %q", name)
	}
	if err := checkOptions(spec.Options, opts, []string{"format"}); err != nil {
		return nil, invalidf("sink %q: %v", name, err)
	}

	format := opts["format"]
	if format == "" {
		format = DefaultFormat
	}
	ser, ok := r.serializers[format]
	if !ok {
		return nil, invalidf("sink %q: unknown format %q", name, format)
	}

	own := Options{}
	for k, v := range opts {
		if k != "format" {
			own[k] = v
		}
	}
	return spec.New(own, ser)
}

func checkOptions(specs []OptionSpec, opts Options, extra []string) error {
	known := map[string]bool{}
	for _, e := range extra {
		known[e] = true
	}
	for _, s := range specs {
		known[s.Name] = true
		if s.Required && opts[s.Name] == "" {
			return fmt.Errorf("missing required option %q", s.Name)
		}
	}
	keys := make([]string, 0, len(opts))
	for k := range opts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if !known[k] {
			return fmt.Errorf("unknown option %q", k)
		}
	}
	return nil
}
