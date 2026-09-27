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
  of small single-purpose transformer steps, each contributing a field to
  the output record, with later steps able to build on earlier ones
- Output sinks (file, stdout), with pluggable serialization formats
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
  rendered content and the main document's http status

## In progress

- Nothing currently being worked on.

## Remaining

- Wiring the browser engine into `cmd/gocrawler` as a selectable engine
  (currently only exercised directly, not from the CLI)
- Driving the transform pipeline builder from the crawler's configuration
- Loading custom transformers as Go plugins, so anyone can extend the
  transform pipeline with their own steps without modifying this project
- CI publishing via goreleaser
- Nix package + dev shell
