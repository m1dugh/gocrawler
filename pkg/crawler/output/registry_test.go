package output_test

import (
	"errors"
	"testing"

	"github.com/m1dugh/gocrawler/pkg/crawler/engine"
	"github.com/m1dugh/gocrawler/pkg/crawler/output"
)

type fakeSerializer struct{}

func (fakeSerializer) Serialize(engine.Record) ([]byte, error) { return nil, nil }

type fakeSink struct {
	opts output.Options
	ser  output.Serializer
}

func (fakeSink) Write(engine.Record) error { return nil }
func (fakeSink) Close() error              { return nil }

func newRegistry() *output.Registry {
	r := output.NewRegistry()
	r.RegisterSerializer("json", fakeSerializer{})
	r.RegisterSerializer("csv", fakeSerializer{})
	r.RegisterSink(output.SinkSpec{
		Name:    "fake",
		Options: []output.OptionSpec{{Name: "path", Required: true}},
		New: func(opts output.Options, ser output.Serializer) (output.Sink, error) {
			return fakeSink{opts, ser}, nil
		},
	})
	return r
}

func isInvalid(err error) bool {
	var ie *output.InvalidError
	return errors.As(err, &ie)
}

func TestNewSinkPassesOptionsWithoutFormat(t *testing.T) {
	s, err := newRegistry().NewSink("fake", output.Options{"path": "x", "format": "csv"})
	if err != nil {
		t.Fatalf("NewSink() error = %v", err)
	}
	fs := s.(fakeSink)
	if fs.opts["path"] != "x" {
		t.Errorf("opts = %v, want path=x", fs.opts)
	}
	if _, ok := fs.opts["format"]; ok {
		t.Errorf("opts = %v, format must not be passed to the sink", fs.opts)
	}
}

func TestNewSinkDefaultFormat(t *testing.T) {
	if _, err := newRegistry().NewSink("fake", output.Options{"path": "x"}); err != nil {
		t.Fatalf("NewSink() with default format error = %v", err)
	}
}

func TestNewSinkInvalid(t *testing.T) {
	tests := map[string]struct {
		name string
		opts output.Options
	}{
		"unknown sink":   {"nope", nil},
		"missing option": {"fake", output.Options{}},
		"unknown option": {"fake", output.Options{"path": "x", "bogus": "1"}},
		"unknown format": {"fake", output.Options{"path": "x", "format": "xml"}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := newRegistry().NewSink(tt.name, tt.opts)
			if !isInvalid(err) {
				t.Errorf("NewSink() error = %v, want an InvalidError", err)
			}
		})
	}
}

func TestNewTransformerUnknown(t *testing.T) {
	_, err := newRegistry().NewTransformer("nope", nil)
	if !isInvalid(err) {
		t.Errorf("NewTransformer() error = %v, want an InvalidError", err)
	}
}

func TestRegisterDuplicatePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("registering a sink twice did not panic")
		}
	}()
	r := newRegistry()
	r.RegisterSink(output.SinkSpec{Name: "fake"})
}
