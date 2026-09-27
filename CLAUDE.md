# Go Crawler

This project is a web crawler in golang.

## Features

**IMPORTANT: every functional advancement or technical refactor MUST be
marked in ADVANCEMENT.md as it happens.** Minor changes (small fixes,
tweaks, non-structural edits) do not need to be marked. A functional
advancement is a capability the project now has (e.g. "the crawl queue
is implemented", "scope matching is implemented") — record that it
exists, not how it was implemented (no package paths, method
signatures, error names, or test commands). ADVANCEMENT.md should
reflect the global advancement of the project from a functional point
of view, or note significant technical refactors — not a technical
changelog of everything that has been done. It should mark what is
currently being worked on and what remains, not implementation details
about work already completed.

### Basic crawling

The first mode of crawling is basic crawling, that is fetching files using http
client with no html/js rendering.

### Browser based crawling

The second mode of crawling is browser crawling. In this mode it can use the
following things to crawl :

1. An existing google-chrome/chromium instance provided through a websocket url
2. A local install of google-chrome/chromium instance that can be spawned by the
   crawler.

Whenever a chrome/chromium instance is spawned (locally, for testing, or
otherwise), always pass `--password-store=basic` as a command-line flag.
Without it, chrome tries to use the gnome keyring for credential storage,
which fails or hangs in headless/sandboxed environments.

## Methodology

Here is the expected methodology for crawling :

1. The user provides a set of urls to start from alongside with a scope that determines
   which urls are valid to keep on crawling and which ones are not
2. The program starts a queue containing all the provided urls by the user.
3. Until the queue is empty and all threads are done working, a thread pool
   pulls a url from the queue and processes it.

### Url processing

The processing of an url comes in several steps :

1. Has the url already been marked as seen ? If yes, skip this url and stop the processing here.
2. Mark the url as seen.
3. Using the Crawling engine, retrieve the content.
4. Once the full content has been retrieved, using a regex, extract all the
   embeded links in the resource. For each of the links, push them to the queue
   if they are in scope only.
5. Feed the content of the page alongside with the http code (if applicable),
   and the url in a channel that will be used for result processing.
6. Another go routine reads the channel containing the full object result of the
   crawl and processes it depending on the config. This go routine is supposed
   to do the following stuff :
   - Create an object corresponding to the configuration of the crawler (it may
     contain full url, partial url (without path params), full content,
     hashed content, http status, content length, timestamp, etc.)
   - Push it in the appropriate outputs based on the configuration : file,
     stdout, or other

Here is a schema of what is happening :

```
main thread orchestrates worker threads based on queue and result thread
|                       A
|                       | pushes found links to queue
|                       |
|--> n worker threads in charge of actual crawling
|                       |
|                       | pushes to worker thread output 
|                       V
|--> 1 worker thread in charge of results/output
```

### Result transformation pipeline

The result-processing go routine turns a raw crawl result into an
output-ready record through a **transform pipeline**: an ordered sequence
of transformer steps, each contributing one field.

- Each transformer step has a unique string identifier and a function
  that takes the raw crawl result and the record accumulated by previous
  steps, and returns an arbitrary value (`interface{}`) for this step.
- The pipeline runs the steps in order, storing each step's result in the
  output record under its own identifier. The result is a map from
  transformer identifier to the value it produced, not a fixed struct.
- Because each step receives what earlier steps have already produced,
  later steps can build on earlier ones (e.g. a step could reuse a field
  another step already computed instead of recomputing it).
- Built-in transformer steps are small and single-purpose (one per
  output field, e.g. full url, partial url, content hash, http status,
  content length, timestamp), each reading straight from the raw crawl
  result rather than from a bundled intermediate object.
- The pipeline is assembled through a builder: steps are added one at a
  time, in order, then built into a runnable pipeline. The crawler's
  configuration module will eventually drive this builder to assemble
  the pipeline from config (not yet implemented), so the set and order
  of steps is driven by config rather than hardcoded.

**Goal: custom transformers as Go plugins.** The transformer pipeline is
designed so that anyone can eventually write and load their own custom
transformer as a Go plugin (`plugin` package `.so`), registering it under
its own identifier so it participates in the pipeline like any built-in
step. This is an end-goal driving the current interface design (transform
package lives in `pkg/`, per the plugin-extensibility rule below) —
the actual plugin loading/discovery mechanism is not implemented yet.

### Configuration of the queue

The queue must be a linked list of urls and contain a scope.

It should expose the following api :

- `Length()` : returns the length (it should be cached in the struct)
- `Pop()` : returns the top element of the queue and removes it. This operation
  is a modification, and should be protected by a mutex.
- `Push()` : pushes the element to the queue, it should be protected by a mutex.
  Before pushing an url to the queue, the queue must validate the url is actually
  in scope.

### Determining whether a url has already been pushed.

A url in this program is composed of 2 parts :

- The actual path (containing the url without the query string)
- The query string

A url can be in different states :

- `unvisited` : The url has never been visited, the actual path has never been reached
  independently from the query string
- `visited once` : The url CONTAINS a query string, has been visited once only,
  disregarding the query string.
- `visited` : The url has been visited and has no query string, OR url has been
  visited at least twice with different query strings AND has yielded the
  same hash of the content.
- `visited_parametrized` : The url has been visited and has yielded different hashes
  multiple times it has been visited.

The goal is to determine whether query strings change the content of the page.
This will allow to detect potential injections in fields.

The implementation should look something like this :

A golang map containing the following kind of struct :

A struct which implements the comparable interface based on the
hash of the url without query string. In the struct is stored a list of strings
containing all the query strings and the hash of the content for the
corresponding query string. This struct implements a `IsVisited(url) bool`
function that takes only a url and determines whether it has already been
visited (based on the previous criteria of what is a visited url). If this url
has already been fetched, return true, false otherwise, and a
`IsVisited(url, hash) bool` function that does the same, but in case the path
has already been visited, but not the query string and it is either different
or unknown.

This struct must have a `Visit(url, hash)` method that stores in the entry
the query string of the url (or empty string if not existing), alongside
with the hash of the page.

Here is the algorithm to determine whether a link should be crawled or not :

**Before fetching the page**

Check in the map if a url struct corresonding to this url is stored in the
map. If not, this url or any other parametrized url of the kind has ever been
visited.

If so, use the `IsVisited(url) bool` method to know if crawling should continue.

If validated, crawling continues, otherwise, crawling is dropped for this
link.

**After fetching the page**

If the entry was not present in the map, create a new entry in the map containing
the url without the query string as key, and call `Visit` on it.

If entry was present in the map already, call `Visit` on it.

### Configuration of the scope

The scope will be filled through a file of regexes validating a url. It will
match the url starting from the begining but the regexes will
only match prefixes. Meaning if the scope contains `https://google\.com`, any
url prefixed by `https://google.com` will be validated, but
`ahttps://google.com/test` will not be validated. In the scope file, each line
can be prefixed by a `!` to invert the regex. All of these infos should be
stored in a Scope struct that exposes a single function `IsInScope(url) bool`
that validates that a url is or not in scope.

### Crawl engine

The crawl engine should implement strategy pattern and be instantiated based
on user configuration, either a plain http client, or a google-chrome instance.
It should expose the following interface :

- `Fetch(url) (content, status)` : a function that takes an url, and returns the
  content fetched and the http status corresponding.

## Expected configuration

TODO

## Tooling

This project is built using golang.

### CI

In CI, the package should be published using goreleaser

### Nix

This project should also be packaged in a nix package in the @flake.nix
file, and a dev shell should be declared in the nix flake containing all
the required dependencies : go, goreleaser, any lib required.

## Code architecture

This project should follow a classic golang project architecture :

- `internal/` for internal lib
- `pkg/` for exposed api
- `cmd/` for binaries to be built and interacted with

It should follow clean code architecture with proper separation of concerns
to allow easier refactoring as the project advances.

Any interface that could at some point be extended by plugins (e.g. the
crawl engine strategy interface) MUST live in `pkg/`, so external code can
implement it. Concrete implementations of such interfaces live under
`internal/`, namespaced by kind, e.g. `internal/crawler/engine/<crawler-type>`
(`internal/crawler/engine/http`, `internal/crawler/engine/browser`, ...).
The interface itself lives at `pkg/crawler/engine`.

## Documentation

All usage-related documentation (how to use the crawler, its packages,
or any of its extension points such as custom transformers) **must be
stored in README.md**. Do not scatter usage docs elsewhere.

## General guidelines

Dont include claude model as co-author in **any commit**.
