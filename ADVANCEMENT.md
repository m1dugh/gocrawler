# Advancement

## Done

- URL de-duplication and visited-state tracking (including detection of
  parametrized/injectable urls via content-hash diffing)
- Scope matching (include/exclude regex rules)
- Crawl queue (scope-validated, thread-safe)
- Crawl engine strategy abstraction (interface that basic and
  browser-based engines will implement)
- Basic (HTTP) crawling engine, with support for custom request headers

## In progress

- Nothing currently being worked on.

## Remaining

- Browser-based crawling engine (remote and locally-spawned chrome/chromium)
- Worker pool orchestration
- Result processing / output
- CI publishing via goreleaser
- Nix package + dev shell
