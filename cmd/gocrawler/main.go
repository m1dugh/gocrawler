// Command gocrawler orchestrates a crawl: it wires together the queue,
// url visit-state tracking, crawl engine, transform pipeline and output
// sinks, and drives the worker pool described in CLAUDE.md.
package main

import (
	"errors"
	"fmt"
	"os"
)

// version is set at release time through -ldflags.
var version = "dev"

// usageError is a mistake in how gocrawler was invoked or configured, as
// opposed to a failure while running. It exits with status 2.
type usageError struct{ err error }

func (e usageError) Error() string { return e.err.Error() }
func (e usageError) Unwrap() error { return e.err }

func usagef(format string, args ...any) error {
	return usageError{fmt.Errorf(format, args...)}
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "gocrawler:", err)

		var ue usageError
		if errors.As(err, &ue) {
			fmt.Fprintln(os.Stderr, "run 'gocrawler --help' for usage")
			os.Exit(2)
		}
		os.Exit(1)
	}
}
