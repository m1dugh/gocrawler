# gocrawler

A web crawler in Go, with support for basic (HTTP) and browser-based
crawling, scope-validated queueing, url visit-state tracking, and a
configurable result transformation pipeline.

See `ADVANCEMENT.md` for the current project status.

## Usage

```sh
gocrawler [flags] <seed-url>...
gocrawler fields          # fields a record can hold, and which are active
                          # (takes --config and the transformer flags)
gocrawler transformers    # available transformers
gocrawler sinks           # available sinks and their options
gocrawler --version          # or -v
```

From source: `go run ./cmd/gocrawler --scope scope.txt https://example.com/`.

A crawl needs at least one seed url and a scope. Records go to stdout, one
JSON object per line. Logs and errors go to stderr, so the output can be
piped.

Long flags use `--`; the short ones are `-S`, `-H` and `-l`.

| Flag | Meaning |
|---|---|
| `<seed-url>...` | Start urls. |
| `-l, --seeds-file <file>` | Read seeds from a file, one per line. `-` reads stdin. |
| `--scope <file>` | Scope rules file. |
| `-S, --scope-rule <regex>` | Inline scope rule, repeatable. Combined with `--scope`. |
| `--engine http\|browser` | Crawl engine. Default `http`. |
| `--browser-url <url>` | Use an existing browser. Implies `--engine browser`. |
| `--browser-path <path>` | Browser binary to spawn. Default: auto-detect. |
| `--headless[=false]` | Run a spawned browser headless. Default `true`. |
| `--browser-flag <flag>` | Extra browser flag, repeatable. |
| `-H, --header "Name: value"` | Request header, repeatable. Works with both engines. |
| `--timeout <duration>` | Per-fetch timeout. Default `30s`. |
| `--workers N` | Concurrent crawl workers. Default `4`. |
| `--transformer <name>[:key=value,...]` | Enable a transformer, repeatable. |
| `--no-transformer <name>` | Disable a default transformer, repeatable. |
| `--no-default-transformers` | Start from an empty pipeline. |
| `--sink <name>[:key=value,...]` | Select a sink, repeatable. Default: stdout only. |
| `--fields <id,id,...>` | Fields every sink writes. Default: every provided field. |
| `--config <file>` | YAML config file. Flags override its keys. |

Exit status is `0` when the crawl completed, `1` on a runtime error (an
unreadable file, an unreachable browser) and `2` on a usage or
configuration error (a missing scope, an unknown sink, a field no
transformer provides).

### Scope

Each line of a scope file is a regex matched as a prefix against candidate
urls. A line prefixed with `!` excludes matching urls instead of including
them. At least one include rule is required.

```
^https://example\.com/
!^https://example\.com/logout
```

```sh
gocrawler --scope scope.txt https://example.com/
gocrawler -S '^https://example\.com/' -S '!^https://example\.com/logout' https://example.com/
```

### Engines

The default engine is a plain HTTP client. The browser engine renders pages
in an actual chrome/chromium instance, for pages that need JS to produce
their final content.

```sh
# spawn a local chromium
gocrawler --engine browser --scope scope.txt https://spa.example.com/

# use an existing browser through its remote debugging endpoint
gocrawler --browser-url http://127.0.0.1:9222 --scope scope.txt https://spa.example.com/

# an endpoint behind a reverse proxy, with a header and a visible window
gocrawler --browser-url https://browser-cdp.example.com -H 'X-Bounty: me' \
  --scope scope.txt https://example.com/
gocrawler --engine browser --headless=false --browser-flag --window-size=1280,800 \
  --scope scope.txt https://example.com/
```

`--browser-url` takes the http(s) address of the debugging endpoint (the
websocket address is looked up at `/json/version`), or directly a
`ws://`/`wss://` address. With an existing browser, gocrawler works in a tab
it opens itself and closes only that tab when it exits. It never attaches to,
navigates or closes a tab that was already open. Whenever chrome is spawned, `--password-store=basic`
is passed, because without it chrome tries to reach the gnome keyring, which
fails or hangs in headless/sandboxed environments. Do the same when you start
a browser yourself:

```sh
chromium --headless --remote-debugging-port=9222 \
  --remote-debugging-address=127.0.0.1 --password-store=basic
```

### Transformers

Transformers fill the output record. Each one declares the fields it
provides, and the fields it requires from earlier steps. Only the default
transformers run unless you enable more.

| Transformer | Provides | Default |
|---|---|---|
| `full_url` | `full_url`: url as crawled | yes |
| `status` | `status`: http status code | yes |
| `content_length` | `content_length`: length of the content | yes |
| `partial_url` | `partial_url`: url without its query string | no |
| `content_hash` | `content_hash`: hash of the page content | no |
| `timestamp` | `timestamp`: time the page was fetched | no |
| `content` | `content`: full page content | no |

```sh
gocrawler --scope scope.txt --transformer content_hash https://example.com/
gocrawler --scope scope.txt --no-transformer content_length https://example.com/
gocrawler --scope scope.txt --no-default-transformers \
  --transformer full_url --transformer status https://example.com/
```

Enabled transformers run after the default ones, in the order given. A
transformer whose required field is no longer provided makes the crawl fail
at startup.

### Sinks and fields

A sink delivers records. Select one with `--sink name[:key=value,...]`,
repeatable. With no `--sink`, records go to stdout only; selecting any sink
replaces that default, and every record is written to each selected sink.
Every sink accepts a `format` option: `json` (the default) or `tsv`.

| Sink | Options |
|---|---|
| `stdout` | `format` |
| `file` | `path` (required), `format`. Appends one record per line. |

```sh
gocrawler --scope scope.txt --sink file:path=crawl.jsonl https://example.com/
gocrawler --scope scope.txt --sink stdout --sink file:path=crawl.jsonl https://example.com/
```

`--fields` chooses which fields the sinks write, and applies to every sink.
Each field must be provided by an active transformer; otherwise the error
names the transformer that provides it.

```sh
$ gocrawler --scope scope.txt --fields full_url,status https://example.com/
{"full_url":"https://example.com/","status":200}

$ gocrawler --scope scope.txt --fields content https://example.com/
gocrawler: field "content" is provided by transformer "content", which is not enabled
```

The `tsv` format writes one line of tab-separated values per record, with no
header. Columns follow the alphabetical order of the field names (so
`full_url` comes before `status`), and tabs, newlines, carriage returns and
backslashes inside a value are escaped as `\t`, `\n`, `\r` and `\\`. Times are
written in RFC 3339, like in JSON.

```sh
$ gocrawler --scope scope.txt --sink stdout:format=tsv --fields full_url,status https://example.com/
https://example.com/	200
```

Option values cannot contain a comma on the command line; use the config
file for those.

### Config file

`--config <file>` reads a YAML file in which every key mirrors a flag. Flags
override the file: a flag that is set replaces the same key, except seeds,
`--transformer` and `--no-transformer`, which are added to the file's. Unknown
keys are an error.

```yaml
seeds:
  - https://example.com/
scope:
  file: scope.txt
  rules:
    - '!^https://example\.com/admin'
engine: browser            # http (default) or browser
browser:
  url: https://browser-cdp.example.com
  # path: /usr/bin/chromium
  # headless: false
  # flags: ["--window-size=1280,800"]
headers:
  - "Authorization: Bearer abc123"
timeout: 15s
workers: 8
transformers:
  defaults: true           # false starts from an empty pipeline
  enable:
    - content_hash
    - name: word_count     # a name with options. NOT AVAILABLE: custom
                           # transformers are planned, remove it to run as is
      options:
        min_length: 3
  disable:
    - content_length
output:
  fields: [full_url, status, content_hash]   # for sinks without their own
  sinks:                   # omit for the default: stdout only
    - type: stdout
    - type: file
      path: crawl.jsonl
      format: json
      fields: [full_url, status, content_hash, word_count]   # drop word_count too
```

Fields are set per sink with the sink's own `fields`. For each sink they
resolve in this order: the sink's own `fields`, then `output.fields`, then
every field provided by the active transformers. The `--fields` flag
overrides all of them and applies to every sink, and `--sink` replaces the
file's sinks list.

## Browser Engine API

The browser engine (`internal/crawler/engine/browser`) renders pages over
the Chrome DevTools Protocol. Both flavors implement `engine.Engine`
(`Fetch(url) (content, status, err)` plus `Close() error`):

```go
// spawns a local chrome/chromium
e := browser.NewLocal(browser.LocalConfig{
    Headless: true,
    Headers:  map[string][]string{"X-Bounty": {"me"}},
})
defer e.Close()
content, status, err := e.Fetch("https://example.com/")

// connects to a running instance (http(s):// or ws(s):// address). It
// works in a tab of its own, and Close closes only that tab
r, err := browser.NewRemote(browser.RemoteConfig{
    WebsocketURL: "http://127.0.0.1:9222",
})
```

## Extending the crawler

> [!WARNING]
> ## THIS FEATURE IS NOT AVAILABLE YET. IT IS PLANNED.
>
> **You cannot extend the `gocrawler` binary with your own transformers,
> sinks or formats today.** There is no `--plugin` flag, and nothing is loaded
> from Go plugins (`.so` files). Loading them is a planned goal, not an
> implemented feature.
>
> What exists is the registry API described below, which a future plugin
> loader will use. It is only usable by writing your own Go program that
> imports this module and builds its own registry. The examples in this
> section show that API and are not a supported way to extend the released
> binary.

Transformers, sinks and serialization formats are registered by name in an
`output.Registry` (`pkg/crawler/output`), the way the built-in ones are.
Anything registered in a registry can be selected by name, like a built-in.

### Custom transformers

A transformer implements `output.Transformer`:

```go
type Transformer interface {
    Name() string
    Provides() []string // fields Transform may return
    Requires() []string // fields read from earlier steps
    Transform(result engine.CrawlResult, acc engine.Record) (engine.Record, error)
}
```

`Transform` returns the fields it contributes. `acc` holds what earlier steps
produced, so a step can build on them as long as it declares the field in
`Requires`. Returning a field not listed in `Provides` is an error.

```go
type wordCount struct{ minLength int }

func (wordCount) Name() string       { return "word_count" }
func (wordCount) Provides() []string { return []string{"word_count"} }
func (wordCount) Requires() []string { return nil }

func (w wordCount) Transform(r engine.CrawlResult, _ engine.Record) (engine.Record, error) {
    n := 0
    for _, word := range strings.Fields(r.Content) {
        if len(word) >= w.minLength {
            n++
        }
    }
    return engine.Record{"word_count": n}, nil
}

reg.RegisterTransformer(output.TransformerSpec{
    Name:        "word_count",
    Description: "number of words in the page",
    Options:     []output.OptionSpec{{Name: "min_length", Description: "shortest word counted"}},
    New: func(opts output.Options) (output.Transformer, error) {
        return wordCount{minLength: 1}, nil // parse opts["min_length"] here
    },
})
```

Set `Default: true` in the spec to make a transformer active unless disabled.
Unknown and missing options are rejected by the registry before `New` runs.

### Custom sinks

A sink implements `output.Sink`. It is given the serializer picked by the
`format` option, which it can use or ignore.

```go
type printSink struct{ prefix string }

func (s printSink) Write(r engine.Record) error { fmt.Println(s.prefix, r); return nil }
func (printSink) Close() error                  { return nil }

reg.RegisterSink(output.SinkSpec{
    Name:    "print",
    Options: []output.OptionSpec{{Name: "prefix"}},
    New: func(opts output.Options, _ output.Serializer) (output.Sink, error) {
        return printSink{prefix: opts["prefix"]}, nil
    },
})
```

Once registered in a registry, it is selected by name, for example
`--sink print:prefix=crawled` or `- type: print` in the config file. This only
works in a program that builds its own registry: the released `gocrawler`
binary cannot load it yet (see the warning at the top of this section).
Serialization formats are added with `reg.RegisterSerializer(name, ser)`.
