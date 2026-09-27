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

## In progress

- Nothing currently being worked on.

## Remaining

- Browser-based crawling engine (remote and locally-spawned chrome/chromium)
- Worker pool orchestration
- Wiring the result-processing goroutine (channel -> transform pipeline -> sinks)
- Driving the transform pipeline builder from the crawler's configuration
- Loading custom transformers as Go plugins, so anyone can extend the
  transform pipeline with their own steps without modifying this project
- CI publishing via goreleaser
- Nix package + dev shell
