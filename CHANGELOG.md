# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- `CHANGELOG.md` file with full history from v0.1.0 to v0.1.7
- `codecov.yml` configuration file for Codecov

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

[Unreleased]: https://github.com/aafeher/go-microdata-extract/compare/v0.1.7...HEAD
[0.1.7]: https://github.com/aafeher/go-microdata-extract/compare/v0.1.6...v0.1.7
[0.1.6]: https://github.com/aafeher/go-microdata-extract/compare/v0.1.5...v0.1.6
[0.1.5]: https://github.com/aafeher/go-microdata-extract/compare/v0.1.4...v0.1.5
[0.1.4]: https://github.com/aafeher/go-microdata-extract/compare/v0.1.3...v0.1.4
[0.1.3]: https://github.com/aafeher/go-microdata-extract/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/aafeher/go-microdata-extract/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/aafeher/go-microdata-extract/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/aafeher/go-microdata-extract/releases/tag/v0.1.0
