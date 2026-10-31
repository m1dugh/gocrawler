package browser_test

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/m1dugh/gocrawler/internal/crawler/engine/browser"
)

// debugTab is a tab as listed by the browser's /json/list endpoint.
type debugTab struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	URL  string `json:"url"`
}

// httpBase turns the debugging address given for the test into the http(s)
// base url of the DevTools HTTP endpoints.
func httpBase(t *testing.T, endpoint string) string {
	t.Helper()
	u, err := url.Parse(endpoint)
	if err != nil {
		t.Fatalf("parsing %q: %v", endpoint, err)
	}
	switch u.Scheme {
	case "ws":
		u.Scheme = "http"
	case "wss":
		u.Scheme = "https"
	}
	u.Path, u.RawQuery = "", ""
	return u.String()
}

func listTabs(t *testing.T, base string) []debugTab {
	t.Helper()
	resp, err := http.Get(base + "/json/list")
	if err != nil {
		t.Fatalf("listing tabs: %v", err)
	}
	defer resp.Body.Close()
	var all []debugTab
	if err := json.NewDecoder(resp.Body).Decode(&all); err != nil {
		t.Fatalf("decoding the tab list: %v", err)
	}
	var pages []debugTab
	for _, tab := range all {
		if tab.Type == "page" {
			pages = append(pages, tab)
		}
	}
	return pages
}

func findTab(tabs []debugTab, id string) (debugTab, bool) {
	for _, tab := range tabs {
		if tab.ID == id {
			return tab, true
		}
	}
	return debugTab{}, false
}

// TestRemoteEngineLeavesOtherTabsAlone checks the engine never attaches to,
// navigates or closes a tab it did not open, and cleans up its own.
func TestRemoteEngineLeavesOtherTabsAlone(t *testing.T) {
	endpoint := requireRemoteBrowser(t)
	base := httpBase(t, endpoint)

	// A tab that belongs to someone else, with a url the engine would
	// change if it navigated it.
	const bystanderURL = "about:blank#gocrawler-test-bystander"
	req, err := http.NewRequest(http.MethodPut, base+"/json/new?"+url.QueryEscape(bystanderURL), nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("opening the bystander tab: %v", err)
	}
	var bystander debugTab
	err = json.NewDecoder(resp.Body).Decode(&bystander)
	resp.Body.Close()
	if err != nil || bystander.ID == "" {
		t.Skipf("the browser does not allow opening a tab through /json/new (%v)", err)
	}
	defer func() {
		if r, err := http.Get(base + "/json/close/" + bystander.ID); err == nil {
			r.Body.Close()
		}
	}()

	before := listTabs(t, base)

	e, err := browser.NewRemote(browser.RemoteConfig{WebsocketURL: endpoint, Timeout: 30 * time.Second})
	if err != nil {
		t.Fatalf("NewRemote() error = %v", err)
	}

	during := listTabs(t, base)
	if len(during) != len(before)+1 {
		t.Errorf("%d tabs while the engine is open, want %d: it must open exactly one tab of its own", len(during), len(before)+1)
	}

	// Fetch twice: neither may touch the other tabs.
	for i := 0; i < 2; i++ {
		if _, _, err := e.Fetch("about:blank"); err != nil {
			t.Fatalf("Fetch() error = %v", err)
		}
	}
	if tab, ok := findTab(listTabs(t, base), bystander.ID); !ok {
		t.Fatal("the bystander tab was closed while the engine was open")
	} else if !strings.HasSuffix(tab.URL, "#gocrawler-test-bystander") {
		t.Errorf("the bystander tab was navigated to %q", tab.URL)
	}

	if err := e.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	// Closing the tab is asynchronous on the browser side: give it a moment.
	var after []debugTab
	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); time.Sleep(100 * time.Millisecond) {
		if after = listTabs(t, base); len(after) == len(before) {
			break
		}
	}

	tab, ok := findTab(after, bystander.ID)
	if !ok {
		t.Fatal("the bystander tab was closed by Close()")
	}
	if !strings.HasSuffix(tab.URL, "#gocrawler-test-bystander") {
		t.Errorf("the bystander tab was navigated to %q", tab.URL)
	}
	if len(after) != len(before) {
		t.Errorf("%d tabs after Close(), want the %d from before: the engine must close its own tab", len(after), len(before))
	}
	for _, b := range before {
		if _, ok := findTab(after, b.ID); !ok {
			t.Errorf("tab %s (%s) that was open before is gone", b.ID, b.URL)
		}
	}
}
