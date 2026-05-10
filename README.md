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

(from https://w3techs.com/technologies/overview/structured_data, 2026-05-05)


| Format                                                                                 | Usage | Supported |
| -------------------------------------------------------------------------------------- |------:| :-------: |
| None                                                                                   | 21.2% |          |
| [OpenGraph](https://ogp.me/)                                                           | 70.6% |    ✔    |
| [X Cards](https://developer.x.com/en/docs/x-for-websites/cards/guides/getting-started) | 56.4% |    ✔    |
| [JSON-LD](https://www.w3.org/TR/json-ld/)                                              | 53.5% |    ✔    |
| [RDFa](https://www.w3.org/TR/rdfa-primer/)                                             | 38.9% |    ✔    |
| [Microdata](https://html.spec.whatwg.org/multipage/microdata.html)                     | 22.5% |    ✔    |
| [Dublin Core](https://www.dublincore.org/specifications/dublin-core/dc-html/)          |  0.8% |    ✔    |
| [Microformats](https://microformats.org/wiki/Main_Page)                                |  0.5% |    ✔    |

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

- syntaxes: `[]Syntax{extract.SyntaxOpenGraph, extract.SyntaxXCards, extract.SyntaxJSONLD, extract.SyntaxMicrodata, extract.SyntaxRDFa, extract.SyntaxDublinCore, extract.SyntaxMicroformats}`
- userAgent: `"go-microdata-extract (+https://github.com/aafeher/go-microdata-extract/blob/main/README.md)"`
- fetchTimeout: `3` seconds
- httpClient: `nil` (a default `http.Client` is created from `fetchTimeout` on each request)
- maxBodySize: `10485760` (10 MB — HTTP response bodies larger than this are rejected with a `FetchError`)

### Overwrite defaults

#### Syntaxes

To retrieve the full list of supported syntaxes, use `DefaultSyntaxes()`:

```go
syntaxes := extract.DefaultSyntaxes() // []Syntax — safe copy, mutation does not affect New()
```

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

#### Custom HTTP client

To inject a fully custom `*http.Client` (e.g. with a custom transport, proxy, or mutual TLS), use `SetHTTPClient()`.
When a custom client is set, `fetchTimeout` is ignored — the client's own timeout applies.

```go
e := extract.New().SetHTTPClient(&http.Client{
    Timeout: 10 * time.Second,
})
```

#### Maximum body size

To set the maximum number of bytes read from an HTTP response body, use `SetMaxBodySize()`.
Responses larger than the limit are rejected with a `FetchError`. The default is 10 MB.

```go
e := extract.New()
e = e.SetMaxBodySize(5 * 1024 * 1024) // 5 MB
```
... or ...
```go
e := extract.New().SetMaxBodySize(5 * 1024 * 1024)
```

#### Reading configuration

Use the corresponding `Get*()` methods to inspect the active configuration:

```go
e := extract.New().SetUserAgent("MyBot").SetFetchTimeout(10)

fmt.Println(e.GetUserAgent())    // "MyBot"
fmt.Println(e.GetFetchTimeout()) // 10
fmt.Println(e.GetMaxBodySize())  // 10485760 (default 10 MB)
fmt.Println(e.GetSyntaxes())     // [opengraph xcards json-ld microdata rdfa dublincore microformats]
fmt.Println(e.GetHTTPClient())   // <nil>
```

#### Chaining methods

In both cases, the functions return a pointer to the main object of the package, allowing you to chain these setting methods in a fluent interface style:

```go
e := extract.New().
    SetSyntaxes([]Syntax{extract.SyntaxOpenGraph, extract.SyntaxJSONLD}).
    SetUserAgent("YourUserAgent").
    SetFetchTimeout(10)
```

### Extract

Once you have properly initialized and configured your instance, you can extract structured data using the `Extract()` function.

The `Extract()` function takes three parameters:

- `ctx`: a `context.Context` for cancellation and timeout control of the HTTP fetch,
- `url`: the URL of the webpage,
- `urlContent`: an optional string pointer for the content of the URL

If you wish to provide the content yourself, pass it as the third parameter. If not, pass `nil` and the function will fetch the content on its own.
The `Extract()` function performs concurrent extracting and fetching optimized by the use of Go's goroutines and sync package, ensuring efficient structured data handling.

```go
e, err := e.Extract(context.Background(), "https://github.com/aafeher/go-microdata-extract", nil)
```

In this example, structured data is extracted from `"https://github.com/aafeher/go-microdata-extract"`. The function fetches the content itself, as we passed `nil` as `urlContent`.

### Error handling

`Extract()` returns a `*FetchError` when the HTTP fetch step fails (network error or non-200 status). Use `errors.As` to distinguish fetch failures from parse failures:

```go
ctx := context.Background()
e, err := extract.New().Extract(ctx, "https://example.com", nil)
if err != nil {
    var fe *extract.FetchError
    if errors.As(err, &fe) {
        fmt.Printf("fetch failed for %s: %v\n", fe.URL, fe.Err)
    }
}
```

Per-extractor parse errors are wrapped in `*ParseError` (tagged with the producing syntax) and do not cause `Extract()` to return early. Retrieve them after extraction with `GetErrors()`:

```go
e, err := extract.New().Extract(ctx, "https://example.com", nil)
if err == nil {
    for _, parseErr := range e.GetErrors() {
        var pe *extract.ParseError
        if errors.As(parseErr, &pe) {
            fmt.Printf("parse error in %s: %v\n", pe.Syntax, pe.Err)
        }
    }
}
```

### Get results as JSON

To retrieve all extracted metadata serialised as an indented JSON byte slice, use `GetExtractedJSON()`.

```go
em, err := extract.New().Extract(context.Background(), "https://example.com", nil)
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(em.GetExtractedJSON()))
```

## Performance

Benchmarks run on a 12th Gen Intel Core i7-1260P (16 threads), Go 1.25.0, measuring parse time per operation with content provided directly (no HTTP fetch overhead). Each extractor is run in isolation.

| Format        | ns/op   | B/op    | allocs/op |
|---------------|--------:|--------:|----------:|
| OpenGraph     | ~14,600 |   9,152 |       115 |
| X Cards       | ~22,000 |  16,232 |       163 |
| JSON-LD       | ~22,100 |  12,272 |       115 |
| W3C Microdata | ~19,200 |  14,916 |       170 |
| RDFa          | ~14,300 |  12,408 |       132 |
| Dublin Core   | ~10,900 |   9,184 |        87 |
| Microformats  | ~17,900 |  12,760 |       149 |
| All 7 (concurrent) | ~108,000 | 112,788 | 1,073 |

Run benchmarks with:

```bash
go test -bench=. -benchmem ./...
```

## Examples

Examples can be found in [/examples](https://github.com/aafeher/go-microdata-extract/tree/main/examples).
