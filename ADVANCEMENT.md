# Advancement

Tracks progress against the feature/architecture list in `CLAUDE.md`.

## Features

- [ ] Basic crawling (plain HTTP client, no JS rendering)
- [ ] Browser based crawling (remote websocket chrome/chromium)
- [ ] Browser based crawling (locally spawned chrome/chromium)

## Methodology components

- [ ] Queue (linked list + scope, `Length()`/`Pop()`/`Push()`, mutex-protected)
- [x] URL visit-state tracking (`internal/crawler/urlstate`): path/query split, `unvisited` / `visited once` / `visited` / `visited_parametrized` states, `Store.ShouldCrawl` (pre-fetch) / `Store.RecordVisit` (post-fetch)
- [x] Scope (`internal/crawler/scope`): regex file, `!` inversion, prefix match, `IsInScope(url) bool`; errors on zero include rules
- [ ] Crawl engine interface (strategy pattern, `Fetch(url) (content, status)`)
- [ ] Worker pool orchestration (main thread, n crawl workers, 1 result worker)
- [ ] Result processing / output (file, stdout, other configured outputs)

## Tooling

- [ ] CI publishing via goreleaser
- [ ] Nix package + dev shell in `flake.nix`

## Notes

- `go.mod` initialized: module `github.com/m1dugh/gocrawler`.
- `internal/crawler/urlstate` has unit + concurrency tests (`go test ./internal/crawler/urlstate/... -race`).
- `internal/crawler/scope` has unit tests (`go test ./internal/crawler/scope/... -race`). Combination rule: in scope if any include regex matches as a prefix AND no exclude regex matches; a scope with zero include rules returns `ErrNoIncludeRules` rather than defaulting to allow-all (avoids unbounded crawling).
