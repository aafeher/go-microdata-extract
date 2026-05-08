# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- `GetErrors() []error` method on `Extractor` — exposes the accumulated parse errors from the last `Extract()` call; previously `errs` was a private field with no public getter

### Added
- `DefaultSyntaxes() []Syntax` function — returns a fresh copy of the default syntax list; safe to iterate or pass to `SetSyntaxes` without risk of mutating the package-level default

### Changed
- `Processor` struct unexported to `processor` — it was only used internally by `Extract()` and was never intended as part of the public API
- `SYNTAXES` exported variable replaced by unexported `defaultSyntaxes`; use `DefaultSyntaxes()` instead — the old `var` allowed callers to mutate the global default, causing subtle cross-call bugs
- `FillMissingFieldsFromOpenGraph` in `extractors/xcards.go` unexported to `fillMissingFieldsFromOpenGraph` — it is an internal implementation detail of the XCards extractor; corresponding tests moved from `extract_test.go` to `extractors/xcards_test.go`

## [0.10.0] - 2026-05-08

### Added
- `context.Context` parameter as the first argument of `Extract()` — HTTP fetches now respect context cancellation and deadline; pass `context.Background()` for the previous behaviour
- `SetHTTPClient(client *http.Client) *Extractor` — injects a fully custom HTTP client (enables proxy, mutual TLS, custom transport, auth headers); when set, the client's own timeout is used instead of `fetchTimeout`
- `FetchError` structured error type (`URL string`, `Err error`) returned by `Extract()` on network or non-200 failures; supports `errors.As` and `errors.Unwrap`
- `ParseError` structured error type (`Syntax Syntax`, `Err error`) wrapping every per-extractor error accumulated in `Extractor.errs`; supports `errors.As` and `errors.Unwrap`
- `httpClient *http.Client` field added to internal `config` struct (zero value = use default client built from `fetchTimeout`)

### Changed
- **API: `Extract(url, content)` → `Extract(ctx, url, content)`** — `context.Context` is now the first parameter; all callers (examples, benchmarks, tests) updated
- `fetch()` now uses `http.NewRequestWithContext` instead of `http.NewRequest`, propagating the caller's context to the HTTP layer
- All errors returned from `fetch()` are now wrapped in `*FetchError`; callers can use `errors.As(err, &FetchError{})` to check for fetch vs parse failures
- Parse errors from extractor goroutines are now wrapped in `*ParseError` (tagged with the producing `Syntax`) before being appended to `Extractor.errs`
- `errorhandling` example updated to demonstrate `errors.As` with `*FetchError`

## [0.9.0] - 2026-05-08

### Changed
- Extracted shared `handleMediaSlice[T any]` generic helper to `extractors/shared.go` — consolidates the slice-management boilerplate (append-or-reuse-last pattern) that was duplicated across all six media property handlers in `opengraph.go` and `xcards.go`
- Extracted `parseMusicProperty`, `parseVideoObjectProperty`, `parseArticleProperty`, `parseBookProperty`, `parseProfileProperty` to `extractors/shared.go` — the five namespace-property switch blocks were identical between `parseOpenGraphMetaTag` and `parseXCardsMetaTag`
- Moved `handleMusicSongProperty`, `handleVideoActorProperty`, `parseIntSafely`, `parseTimeSafely` to `extractors/shared.go` — these helpers are shared across multiple files and now live in a single authoritative location
- `handleOpenGraphImageProperty`, `handleOpenGraphVideoProperty`, `handleOpenGraphAudioProperty` and their XCards counterparts are now thin wrappers over `handleMediaSlice`, reducing each handler to a format-specific field-assignment closure

## [0.8.0] - 2026-05-07

### Added
- RDFa error/edge-case test fixture (`test/test-55-rdfa-errors.html`) — HTML with no RDFa items and a dangling `property` attribute without a `typeof` parent; corresponding `test-55-rdfa-errors` test case added
- Relative URL resolution test fixture (`test/test-56-opengraph-relative-urls.html`) with corresponding `test-56-opengraph-relative-urls` test case verifying path-relative URLs are resolved against the page base URL
- `resolveURL(base, ref string) string` shared helper in `extractors/w3cmicrodata.go` using `url.URL.ResolveReference` (handles absolute, protocol-relative, and path-relative refs correctly)
- `resolveOpenGraphURLs` in `extractors/opengraph.go` — resolves `og:url`, image/video/audio URL and SecureURL fields
- `resolveXCardsURLs` in `extractors/xcards.go` — resolves twitter/og image/video/audio URL and SecureURL fields

### Changed
- **API: `ParseOpenGraph` return type** changed from `(any, []error)` to `(*OpenGraph, []error)`; `extract.go` wrapper converts typed nil to untyped nil to preserve map semantics
- **API: `ParseXCards` return type** changed from `(any, []error)` to `(*XCards, []error)`; same nil handling in wrapper
- **API: `DublinCore` return type** changed from `(any, []error)` to `(*DublinCoreItem, []error)`; same nil handling in wrapper
- `JSONLD` signature: unused named parameter `URL string` replaced with blank identifier `_ string`
- `DublinCore` link `href` values are now resolved against the page URL (previously stored verbatim)
- W3CMicrodata inline URL resolution replaced with shared `resolveURL` — protocol-relative URLs (`//host/path`) are now properly resolved to `http://host/path`; existing test updated (`test-35-w3cmicrodata-book`: `discussionUrl` value updated from `//www.example.com/…` to `http://www.example.com/…`)
- `parseDublinCore` now accepts a URL parameter and handles `html.Parse` errors
- `parseMicroformats`, `parseRDFa`, `parseW3CMicrodata` now handle `html.Parse` errors instead of discarding them

### Fixed
- Remove always-true dead condition `if len(items) >= 0` in `extractors/jsonld.go` — simplified `JSONLD()` to return `extractJSONLD()` directly
- Remove no-op nil→empty-slice init blocks in `extractors/opengraph.go` — `handleOpenGraphImageProperty`, `handleOpenGraphVideoProperty`, `handleOpenGraphAudioProperty` each had a redundant `if len == 0 { slice = []T{} }` block (Go's `append` works identically on nil and empty slices)

## [0.7.0] - 2026-05-07

### Added
- Benchmark tests (`extract_bench_test.go`) for all 7 formats: OpenGraph, X Cards, JSON-LD, W3C Microdata, RDFa, Dublin Core, and Microformats, plus a combined all-syntaxes benchmark
- Performance section in README with benchmark results table and instructions for running benchmarks

## [0.6.0] - 2026-05-06

### Added
- Microformats2 extractor (`extractors/microformats.go`) supporting `h-*` root items, `p-*` (text), `u-*` (URL), `dt-*` (datetime), and `e-*` (embedded HTML) property prefixes
- `SyntaxMicroformats` constant (`"microformats"`) added to `SYNTAXES` and enabled by default
- Nested item support: `h-*` child nodes used as both property values and `Children` entries
- Add example: Microformats extraction with recursive nested item printing (`examples/getmicroformats`)

### Changed
- Default syntaxes list extended with `SyntaxMicroformats`
- README: Microformats marked as supported in the formats table, default syntaxes updated

## [0.5.0] - 2026-05-06

### Added
- Dublin Core extractor (`extractors/dublincore.go`) supporting `DC.*` and `DCTERMS.*` prefixes in `<meta name>` and `<link rel>` attributes
- `SyntaxDublinCore` constant (`"dublincore"`) added to `SYNTAXES` and enabled by default
- Add example: Dublin Core extraction with type assertion and multiple-value handling (`examples/getdublincore`)

### Changed
- Default syntaxes list extended with `SyntaxDublinCore`
- README: Dublin Core marked as supported in the formats table, default syntaxes updated

### Fixed
- XCards extractor no longer falsely returns an empty struct when non-Twitter `<meta name>` tags (e.g. Dublin Core, keywords) are present; `xcHasValue` is now only set for recognized `twitter:`, `og:`, `music:`, `video:`, `article:`, `book:`, or `profile:` prefixes

## [0.4.0] - 2026-05-05

### Added
- RDFa extractor (`extractors/rdfa.go`) supporting `vocab`, `typeof`, `property`, `prefix`, `resource`, `about` attributes and CURIE resolution
- `SyntaxRDFa` constant (`"rdfa"`) added to `SYNTAXES` and enabled by default
- Add example: RDFa extraction with type assertion and nested items (`examples/getrdfa`)

### Changed
- Default syntaxes list extended with `SyntaxRDFa`
- README: RDFa marked as supported in the formats table, default syntaxes updated

## [0.3.0] - 2026-05-05

### Added
- Add example: raw HTML content provided directly without HTTP fetch (`examples/rawcontent`)
- Add example: error handling for fetch failures and missing syntax data (`examples/errorhandling`)
- Add example: JSON-LD extraction with type assertion and nested field access (`examples/getjsonld`)
- Add example: W3C Microdata extraction with type assertion and nested items (`examples/getmicrodata`)

## [0.2.0] - 2026-05-05

### Added
- `CHANGELOG.md` file with full history from v0.1.0 to v0.1.7
- `ROADMAP.md` progress tracker with planned versions up to v1.0.0
- `codecov.yml` configuration file for Codecov
- `.golangci.yml` linter configuration (errcheck, govet, staticcheck, ineffassign, misspell, unused, gosimple, revive)
- Add 100% coverage threshold to `codecov.yml` for both project and patch checks

### Changed
- Bump minimum Go version from 1.18 to 1.25.0
- Update dependency `golang.org/x/net` from v0.31.0 to v0.53.0
- Update GitHub Actions workflow Go version from 1.18 to 1.25.0
- Add golangci-lint job to GitHub Actions workflow, build job depends on it passing
- Add matrix strategy to CI: both jobs run on Go `1.25.0` and `stable`
- Update `actions/setup-go` from `@v4` to `@v5`
- Add `go vet` step to build job
- Add `-race` flag to test run for race condition detection
- Add vulnerability scan step (`govulncheck`) on `stable` Go version
- Update `codecov/codecov-action` from `@v4.0.1` to `@v5`, upload only on `stable`

### Fixed
- Replace unreachable `len(errs) < 0` condition with correct `len(errs) != 0` check in `extract_test.go`
- Rename unused `r *http.Request` parameter to `_` in test HTTP handler (`extract_test.go`)

## [0.1.7] - 2025-07-01

### Changed
- Refactored `FillMissingFieldsFromOpenGraph` function with comprehensive unit tests

### Fixed
- Updated README with latest format usage statistics (w3techs.com, 2025-07-01)

## [0.1.6] - 2025-02-04

### Fixed
- Handle empty `XCardsImage` cases to prevent panics on missing image data

## [0.1.5] - 2025-02-02

### Fixed
- Handle whitespace around the `type` attribute in JSON-LD `<script>` tags during extraction
- Updated README with latest format usage statistics (w3techs.com, 2025-02-02)

## [0.1.4] - 2024-12-03

### Fixed
- Initialize OpenGraph media slices (images, videos, audio) before property handling to avoid nil slice issues
- Updated README with latest format usage statistics (w3techs.com, 2024-12-03)

## [0.1.3] - 2024-12-03

### Fixed
- Improve `<meta>` tag handling in W3C Microdata extraction by prioritizing the `content` attribute

## [0.1.2] - 2024-11-24

### Changed
- Replace all `interface{}` usages with the `any` type alias (Go 1.18+)
- Update dependency `golang.org/x/net` to v0.31.0
- Updated README with latest format usage statistics (w3techs.com, 2024-11-24)

## [0.1.1] - 2024-11-03

### Changed
- Rename internal package name from `extractors` to `extractor` in the extractors directory
- Add GitHub Actions CI workflow (`go.yml`) for automated build and test on push/PR to `main`

## [0.1.0] - 2024-11-01

### Added
- Initial release
- Extract structured metadata from HTML for four formats: OpenGraph, X Cards (Twitter Cards), JSON-LD, W3C Microdata
- Concurrent extraction using goroutines with `sync.WaitGroup` and `sync.Mutex`
- Configurable HTTP client: custom User-Agent and fetch timeout
- Fluent/builder API: `New()`, `SetSyntaxes()`, `SetUserAgent()`, `SetFetchTimeout()`, `Extract()`, `GetExtracted()`, `GetExtractedJSON()`
- Support for providing raw HTML content directly (bypassing HTTP fetch)
- Examples: simple extraction, OpenGraph-only, configuring specific syntaxes

[Unreleased]: https://github.com/aafeher/go-microdata-extract/compare/v0.10.0...HEAD
[0.10.0]: https://github.com/aafeher/go-microdata-extract/compare/v0.9.0...v0.10.0
[0.9.0]: https://github.com/aafeher/go-microdata-extract/compare/v0.8.0...v0.9.0
[0.8.0]: https://github.com/aafeher/go-microdata-extract/compare/v0.7.0...v0.8.0
[0.7.0]: https://github.com/aafeher/go-microdata-extract/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/aafeher/go-microdata-extract/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/aafeher/go-microdata-extract/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/aafeher/go-microdata-extract/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/aafeher/go-microdata-extract/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/aafeher/go-microdata-extract/compare/v0.1.7...v0.2.0
[0.1.7]: https://github.com/aafeher/go-microdata-extract/compare/v0.1.6...v0.1.7
[0.1.6]: https://github.com/aafeher/go-microdata-extract/compare/v0.1.5...v0.1.6
[0.1.5]: https://github.com/aafeher/go-microdata-extract/compare/v0.1.4...v0.1.5
[0.1.4]: https://github.com/aafeher/go-microdata-extract/compare/v0.1.3...v0.1.4
[0.1.3]: https://github.com/aafeher/go-microdata-extract/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/aafeher/go-microdata-extract/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/aafeher/go-microdata-extract/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/aafeher/go-microdata-extract/releases/tag/v0.1.0
