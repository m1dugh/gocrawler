// Package browser implements the browser-based crawling engine, which
// fetches urls by rendering them in an actual google-chrome/chromium
// instance over the Chrome DevTools Protocol (CDP).
//
// Two engines share the fetch logic in this file: RemoteEngine, which
// connects to an existing chrome/chromium instance over a websocket url,
// and LocalEngine, which spawns a local chrome/chromium instance.
package browser

import (
	"context"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

// fetch navigates to url in the browser reachable through ctx (a
// chromedp context created by either RemoteEngine or LocalEngine),
// returning the rendered page's html content and the http status of its
// main document response.
func fetch(ctx context.Context, url string) (content string, status int, err error) {
	var statusCode int64

	// The main document's response is the last Document-typed response
	// seen (there may be more than one on redirects, in which case the
	// last one is the page that actually got rendered).
	chromedp.ListenTarget(ctx, func(ev any) {
		e, ok := ev.(*network.EventResponseReceived)
		if !ok {
			return
		}
		if e.Type == network.ResourceTypeDocument {
			statusCode = e.Response.Status
		}
	})

	err = chromedp.Run(ctx,
		network.Enable(),
		chromedp.Navigate(url),
		chromedp.OuterHTML("html", &content, chromedp.ByQuery),
	)
	if err != nil {
		return "", 0, err
	}

	return content, int(statusCode), nil
}
