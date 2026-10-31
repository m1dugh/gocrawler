package browser

import (
	"context"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

// LocalConfig configures a LocalEngine.
type LocalConfig struct {
	// BinaryPath is the path to the chrome/chromium executable to spawn.
	// Empty lets chromedp locate one on the system.
	BinaryPath string

	// Headless controls whether the spawned browser runs headless. The
	// zero value (false) spawns a visible browser; set true to run
	// headless.
	Headless bool

	// Flags are extra chrome command-line flags, as "--name" or
	// "--name=value".
	Flags []string

	// Headers are sent with every request the browser makes.
	Headers map[string][]string

	// Timeout bounds each Fetch call. Zero means no timeout.
	Timeout time.Duration
}

// LocalEngine is a crawler.Engine implementation that fetches urls by
// spawning a local chrome/chromium instance and rendering pages in it.
type LocalEngine struct {
	timeout time.Duration
	headers map[string][]string
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewLocal spawns a local chrome/chromium instance configured by cfg and
// returns a LocalEngine that renders pages on it.
func NewLocal(cfg LocalConfig) *LocalEngine {
	opts := append([]chromedp.ExecAllocatorOption{}, chromedp.DefaultExecAllocatorOptions[:]...)
	opts = append(opts,
		chromedp.Flag("headless", cfg.Headless),
		// Avoid chrome trying (and failing/hanging) to reach the gnome
		// keyring for credential storage in headless/sandboxed setups.
		chromedp.Flag("password-store", "basic"),
	)
	for _, f := range cfg.Flags {
		name, value, hasValue := strings.Cut(strings.TrimPrefix(f, "--"), "=")
		if hasValue {
			opts = append(opts, chromedp.Flag(name, value))
		} else {
			opts = append(opts, chromedp.Flag(name, true))
		}
	}
	if cfg.BinaryPath != "" {
		opts = append(opts, chromedp.ExecPath(cfg.BinaryPath))
	}

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancel := chromedp.NewContext(allocCtx)

	return &LocalEngine{
		timeout: cfg.Timeout,
		headers: cfg.Headers,
		ctx:     ctx,
		cancel: func() {
			cancel()
			allocCancel()
		},
	}
}

// Fetch renders url in the local browser and returns its content and
// main document http status.
func (e *LocalEngine) Fetch(url string) (content string, status int, err error) {
	ctx := e.ctx
	if e.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, e.timeout)
		defer cancel()
	}
	return fetch(ctx, url, e.headers)
}

// Close terminates the spawned browser process.
func (e *LocalEngine) Close() error {
	e.cancel()
	return nil
}
