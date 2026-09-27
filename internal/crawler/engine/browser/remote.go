package browser

import (
	"context"
	"time"

	"github.com/chromedp/chromedp"
)

// RemoteConfig configures a RemoteEngine.
type RemoteConfig struct {
	// WebsocketURL is the websocket address of an already-running
	// chrome/chromium instance's remote debugging endpoint, e.g.
	// "ws://127.0.0.1:9222/" or "http://127.0.0.1:9222/".
	WebsocketURL string

	// Timeout bounds each Fetch call. Zero means no timeout.
	Timeout time.Duration
}

// RemoteEngine is a crawler.Engine implementation that fetches urls by
// rendering them in an existing chrome/chromium instance, reached over a
// websocket url.
type RemoteEngine struct {
	timeout time.Duration
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewRemote connects to the chrome/chromium instance at cfg.WebsocketURL
// and returns a RemoteEngine that renders pages on it.
//
// Some chrome/chromium builds fail to open a new tab via CDP's
// Target.createTarget when driven this way ("Failed to open new tab -
// no browser is open"). To work around this, NewRemote attaches to an
// existing page target when the browser already has one, only falling
// back to letting chromedp create a new target if none exists.
func NewRemote(cfg RemoteConfig) (*RemoteEngine, error) {
	allocCtx, allocCancel := chromedp.NewRemoteAllocator(context.Background(), cfg.WebsocketURL)

	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	targets, err := chromedp.Targets(browserCtx)
	if err != nil {
		browserCancel()
		allocCancel()
		return nil, err
	}

	var opts []chromedp.ContextOption
	for _, t := range targets {
		if t.Type == "page" {
			opts = append(opts, chromedp.WithTargetID(t.TargetID))
			break
		}
	}

	ctx, cancel := chromedp.NewContext(browserCtx, opts...)

	return &RemoteEngine{
		timeout: cfg.Timeout,
		ctx:     ctx,
		cancel: func() {
			cancel()
			browserCancel()
			allocCancel()
		},
	}, nil
}

// Fetch renders url in the remote browser and returns its content and
// main document http status.
func (e *RemoteEngine) Fetch(url string) (content string, status int, err error) {
	ctx := e.ctx
	if e.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, e.timeout)
		defer cancel()
	}
	return fetch(ctx, url)
}

// Close disconnects from the remote browser. It does not terminate the
// remote browser process itself, only this engine's connection to it.
func (e *RemoteEngine) Close() error {
	e.cancel()
	return nil
}
