# Advancement

## Done

- URL de-duplication and visited-state tracking (including detection of
  parametrized/injectable urls via content-hash diffing)
- Scope matching (include/exclude regex rules)
- Crawl queue (scope-validated, thread-safe)
- Crawl engine strategy abstraction (interface that basic and
  browser-based engines will implement)
- Basic (HTTP) crawling engine, with support for custom request headers
- Result transformation pipeline: an ordered, builder-assembled sequence
  of transformers, each declaring the fields it provides and requires and
  contributing them to the output record, with later steps able to build
  on earlier ones. Only a small set is active by default; the others are
  opt-in
- Registry of transformers, sinks and serialization formats selectable by
  name, the extension point custom ones will plug into
- Output sinks (file, stdout), with pluggable serialization formats (JSON,
  TSV), each record written to every selected sink
- Embedded link extraction from crawled content, resolved to absolute
  urls against the page they were found on
- Worker pool orchestration (`cmd/gocrawler`): concurrent crawl workers
  pulling from the queue and feeding discovered links back into it, and a
  result-processing goroutine running each result through the transform
  pipeline into the configured output sinks — a full working end-to-end
  crawl from the command line
- Browser-based crawling engine: renders pages in an actual
  google-chrome/chromium instance, either an existing one reached over a
  websocket url or one spawned locally by the crawler, and reports the
  rendered content and the main document's http status, with custom
  request headers and extra browser flags
- Command-line interface: seeds from arguments, a file or stdin; scope from
  a file and inline rules; engine selection (http, local browser, remote
  browser behind any debugging endpoint); headers, timeout and workers;
  transformer selection; multiple sinks; and per-run field selection
- YAML configuration file mirroring every flag, with per-sink field
  selection, overridden by the flags
- Container image: a multi-stage Alpine build that cross-compiles the
  binary, so one Dockerfile serves both amd64 and arm64. The image bundles
  no browser, to stay small: the browser engine is used from it only with
  an existing browser

## In progress

- Nothing currently being worked on.

## Remaining

- Loading custom transformers and sinks as Go plugins, so anyone can
  extend the crawler without modifying this project
- CI publishing via goreleaser
- CI building and publishing the multi-arch container image
- Nix package + dev shell
