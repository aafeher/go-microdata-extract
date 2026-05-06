# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/aafeher/go-microdata-extract/compare/v0.6.0...HEAD
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
