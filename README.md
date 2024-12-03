# go-microdata-extract

[![codecov](https://codecov.io/gh/aafeher/go-microdata-extract/graph/badge.svg?token=BD1QYCZESR)](https://codecov.io/gh/aafeher/go-microdata-extract)
[![Go](https://github.com/aafeher/go-microdata-extract/actions/workflows/go.yml/badge.svg)](https://github.com/aafeher/go-microdata-extract/actions/workflows/go.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/aafeher/go-microdata-extract.svg)](https://pkg.go.dev/github.com/aafeher/go-microdata-extract)
[![Go Report Card](https://goreportcard.com/badge/github.com/aafeher/go-microdata-extract)](https://goreportcard.com/report/github.com/aafeher/go-microdata-extract)

A Go package for extracting structured data from HTML.

## Formats supported

For currently supported formats, see [Statistics](#statistics)

## Statistics

Usage statistics of structured data formats for websites

(from https://w3techs.com/technologies/overview/structured_data, 2024-12-03)


| Format                                                                                 | Usage | Supported |
| -------------------------------------------------------------------------------------- |-------| :-------: |
| None                                                                                   | 23.4% |          |
| [OpenGraph](https://ogp.me/)                                                           | 67.5% |    ✔    |
| [X Cards](https://developer.x.com/en/docs/x-for-websites/cards/guides/getting-started) | 52.2% |    ✔    |
| [JSON-LD](https://www.w3.org/TR/json-ld/)                                              | 49.6% |    ✔    |
| [RDFa](https://www.w3.org/TR/rdfa-primer/)                                             | 39.4% |     -     |
| [Microdata](https://html.spec.whatwg.org/multipage/microdata.html)                     | 24.1% |    ✔    |
| [Dublin Core](https://www.dublincore.org/specifications/dublin-core/dc-html/)          | 0.9%  |     -     |
| [Microformats](https://microformats.org/wiki/Main_Page)                                | 0.4%  |     -     |

## Installation

```bash
go get github.com/aafeher/go-microdata-extract
```

```go
import "github.com/aafeher/go-microdata-extract"
```

## Usage

### Create instance

To create a new instance with default settings, you can simply call the `New()` function.

```go
e := extract.New()
```

### Configuration defaults

- syntaxes: `[]Syntax{extract.SyntaxOpenGraph, extract.SyntaxXCards, extract.SyntaxJSONLD, extract.SyntaxMicrodata}`
- userAgent: `"go-microdata-extract (+https://github.com/aafeher/go-microdata-extract/blob/main/README.md)"`
- fetchTimeout: `3` seconds

### Overwrite defaults

#### Syntaxes

To set the syntaxes whose results you want to retrieve after processing, use the `SetSyntaxes()` function.

```go
e := extract.New()
e = e.SetSyntaxes([]Syntax{extract.SyntaxOpenGraph, extract.SyntaxJSONLD})
```
... or ...
```go
e := extract.New().SetSyntaxes([]Syntax{extract.SyntaxOpenGraph, extract.SyntaxJSONLD})
```

#### User Agent

To set the user agent, use the `SetUserAgent()` function.

```go
e := extract.New()
e = e.SetUserAgent("YourUserAgent")
```
... or ...
```go
e := extract.New().SetUserAgent("YourUserAgent")
```

#### Fetch timeout

To set the fetch timeout, use the `SetFetchTimeout()` function. It should be specified in seconds as an **uint8** value.

```go
e := extract.New()
e = e.SetFetchTimeout(10)
```
... or ...

```go
e := extract.New().SetFetchTimeout(10)
```

#### Chaining methods

In both cases, the functions return a pointer to the main object of the package, allowing you to chain these setting methods in a fluent interface style:

```go
e := extract.New()
     .SetSyntaxes([]Syntax{extract.SyntaxOpenGraph, extract.SyntaxJSONLD})
     .SetUserAgent("YourUserAgent")
     .SetFetchTimeout(10)
```

### Extract

Once you have properly initialized and configured your instance, you can extract structured data using the `Extract()` function.

The `Extract()` function takes in two parameters:

- `url`: the URL of the webpage,
- `urlContent`: an optional string pointer for the content of the URL

If you wish to provide the content yourself, pass the content as the second parameter. If not, simply pass nil and the function will fetch the content on its own.
The `Extract()` function performs concurrent extracting and fetching optimized by the use of Go's goroutines and sync package, ensuring efficient structured data handling.

```go
e, err := e.Extract("https://github.com/aafeher/go-microdata-extract", nil)
```

In this example, structured data is extracted from "https://github.com/aafeher/go-microdata-extract". The function fetches the content itself, as we passed nil as the urlContent.

## Examples

Examples can be found in [/examples](https://github.com/aafeher/go-microdata-extract/tree/main/examples).
