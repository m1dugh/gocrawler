package sink

import (
	"io"
	"os"
	"sync"

	"github.com/m1dugh/gocrawler/pkg/crawler/engine"
	"github.com/m1dugh/gocrawler/pkg/crawler/output"
)

// Stdout is an output.Sink that writes serialized records to an
// io.Writer, defaulting to os.Stdout.
type Stdout struct {
	mu  sync.Mutex
	w   io.Writer
	ser output.Serializer
}

// NewStdout returns a Stdout sink that serializes records with ser and
// writes them to os.Stdout.
func NewStdout(ser output.Serializer) *Stdout {
	return &Stdout{w: os.Stdout, ser: ser}
}

// SetWriter overrides the destination writer, which otherwise defaults to
// os.Stdout. Mainly useful for tests.
func (s *Stdout) SetWriter(w io.Writer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.w = w
}

// Write serializes r and writes it to the sink's writer, one record per
// line.
func (s *Stdout) Write(r engine.Record) error {
	b, err := s.ser.Serialize(r)
	if err != nil {
		return err
	}
	b = append(b, '\n')

	s.mu.Lock()
	defer s.mu.Unlock()
	_, err = s.w.Write(b)
	return err
}
