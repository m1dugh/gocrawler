package browser

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

// RemoteConfig configures a RemoteEngine.
type RemoteConfig struct {
	// WebsocketURL is the websocket address of an already-running
	// chrome/chromium instance's remote debugging endpoint, e.g.
	// "ws://127.0.0.1:9222/" or "http://127.0.0.1:9222/".
	WebsocketURL string

	// Headers are sent with every request the browser makes.
	Headers map[string][]string

	// Timeout bounds each Fetch call. Zero means no timeout.
	Timeout time.Duration
}

// RemoteEngine is a crawler.Engine implementation that fetches urls by
// rendering them in an existing chrome/chromium instance, reached over a
// websocket url.
type RemoteEngine struct {
	timeout time.Duration
	headers map[string][]string
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
	wsURL, err := resolveWebsocketURL(cfg.WebsocketURL)
	if err != nil {
		return nil, err
	}
	allocCtx, allocCancel := chromedp.NewRemoteAllocator(context.Background(), wsURL, chromedp.NoModifyURL)

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
		headers: cfg.Headers,
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
	return fetch(ctx, url, e.headers)
}

// Close disconnects from the remote browser. It does not terminate the
// remote browser process itself, only this engine's connection to it.
func (e *RemoteEngine) Close() error {
	e.cancel()
	return nil
}

// resolveWebsocketURL turns the debugging endpoint given by the user into
// the websocket url of the browser. A ws:// or wss:// url is used as is. For
// an http:// or https:// url the browser is asked for its websocket url at
// /json/version; the scheme and host of the url given by the user are kept
// (https gives wss), since the browser advertises its own internal address,
// which is wrong behind a proxy or without an explicit port.
func resolveWebsocketURL(endpoint string) (string, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("browser url %q: %w", endpoint, err)
	}

	switch u.Scheme {
	case "ws", "wss":
		return endpoint, nil
	case "http", "https":
	default:
		return "", fmt.Errorf("browser url %q: want an http(s) or ws(s) url", endpoint)
	}

	versionURL := *u
	versionURL.Path = strings.TrimSuffix(u.Path, "/") + "/json/version"
	versionURL.RawQuery = ""
	resp, err := http.Get(versionURL.String())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("browser url %q: %s answered %s, is it a chrome remote debugging endpoint?", endpoint, versionURL.Path, resp.Status)
	}

	var version struct {
		WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&version); err != nil || version.WebSocketDebuggerURL == "" {
		return "", fmt.Errorf("browser url %q: no webSocketDebuggerUrl in the /json/version answer", endpoint)
	}

	ws, err := url.Parse(version.WebSocketDebuggerURL)
	if err != nil {
		return "", err
	}
	ws.Scheme = "ws"
	if u.Scheme == "https" {
		ws.Scheme = "wss"
	}
	ws.Host = u.Host
	return ws.String(), nil
}
