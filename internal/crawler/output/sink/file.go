// Package sink implements output.Sink for different destinations.
package sink

import (
	"os"
	"sync"

	"github.com/m1dugh/gocrawler/pkg/crawler/engine"
	"github.com/m1dugh/gocrawler/pkg/crawler/output"
)

// File is an output.Sink that appends serialized records to a file.
type File struct {
	mu  sync.Mutex
	f   *os.File
	ser output.Serializer
}

// NewFile opens (creating/appending as needed) the file at path and
// returns a File sink that serializes records with ser.
func NewFile(path string, ser output.Serializer) (*File, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}
	return &File{f: f, ser: ser}, nil
}

// Write serializes r and appends it to the file, one record per line.
func (s *File) Write(r engine.Record) error {
	b, err := s.ser.Serialize(r)
	if err != nil {
		return err
	}
	b = append(b, '\n')

	s.mu.Lock()
	defer s.mu.Unlock()
	_, err = s.f.Write(b)
	return err
}

// Close closes the underlying file.
func (s *File) Close() error {
	return s.f.Close()
}
