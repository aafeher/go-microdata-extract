package extract

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	extractor "github.com/aafeher/go-microdata-extract/extractors"
	"io"
	"net/http"
	"slices"
	"sync"
	"time"
)

// FetchError is returned when the HTTP fetch step fails (network error or non-200 status).
type FetchError struct {
	URL string
	Err error
}

func (e *FetchError) Error() string { return fmt.Sprintf("fetch %q: %s", e.URL, e.Err) }
func (e *FetchError) Unwrap() error { return e.Err }

// ParseError wraps a parser error together with the syntax that produced it.
type ParseError struct {
	Syntax Syntax
	Err    error
}

func (e *ParseError) Error() string { return fmt.Sprintf("parse %q: %s", e.Syntax, e.Err) }
func (e *ParseError) Unwrap() error { return e.Err }

type (
	// Extractor is a struct used for extracting metadata from web content or a provided URL. It utilizes various processors.
	Extractor struct {
		cfg       config
		url       string
		content   string
		extracted map[Syntax]any
		errs      []error
	}

	// config represents configuration settings for an Extractor, including syntax options, user agent, and fetch timeout.
	config struct {
		syntaxes     []Syntax
		userAgent    string
		fetchTimeout uint8
		httpClient   *http.Client
		maxBodySize  int64
	}

	processor struct {
		Name Syntax
		Func func() (any, []error)
	}

	Syntax string
)

const (
	// SyntaxOpenGraph is the identifier used for the Open Graph metadata syntax.
	SyntaxOpenGraph Syntax = "opengraph"

	// SyntaxXCards is the identifier used for the X Cards metadata syntax.
	SyntaxXCards Syntax = "xcards"

	// SyntaxJSONLD is the identifier used for the JSON-LD metadata syntax.
	SyntaxJSONLD Syntax = "json-ld"

	// SyntaxMicrodata is the identifier used for the W3C Microdata metadata syntax.
	SyntaxMicrodata Syntax = "microdata"

	// SyntaxRDFa is the identifier used for the RDFa metadata syntax.
	SyntaxRDFa Syntax = "rdfa"

	// SyntaxDublinCore is the identifier used for the Dublin Core metadata syntax.
	SyntaxDublinCore Syntax = "dublincore"

	// SyntaxMicroformats is the identifier used for the Microformats2 metadata syntax.
	SyntaxMicroformats Syntax = "microformats"
)

var defaultSyntaxes = []Syntax{SyntaxOpenGraph, SyntaxXCards, SyntaxJSONLD, SyntaxMicrodata, SyntaxRDFa, SyntaxDublinCore, SyntaxMicroformats}

// DefaultSyntaxes returns a copy of the default syntax list used by New().
// Callers may iterate over it or pass a subset to SetSyntaxes.
func DefaultSyntaxes() []Syntax {
	cp := make([]Syntax, len(defaultSyntaxes))
	copy(cp, defaultSyntaxes)
	return cp
}

// New creates a new instance of Extractor with default configurations and an empty map for extracted data.
func New() *Extractor {
	e := &Extractor{
		extracted: make(map[Syntax]any),
	}

	e.setConfigDefaults()

	return e
}

const defaultMaxBodySize int64 = 10 * 1024 * 1024 // 10 MB

// setConfigDefaults initializes the Extractor with default configuration settings.
func (e *Extractor) setConfigDefaults() {
	e.cfg = config{
		syntaxes:     defaultSyntaxes,
		userAgent:    "go-microdata-extract (+https://github.com/aafeher/go-microdata-extract/blob/main/README.md)",
		fetchTimeout: 3,
		maxBodySize:  defaultMaxBodySize,
	}
}

// SetSyntaxes sets the syntaxes that the Extractor will use for parsing metadata. Filters out unsupported syntaxes.
// syntaxes: A slice of Syntax representing the desired syntaxes.
// Returns the updated Extractor instance.
func (e *Extractor) SetSyntaxes(syntaxes []Syntax) *Extractor {
	if len(syntaxes) == 0 {
		return e
	}

	syntaxesToSet := make([]Syntax, 0)
	for _, syntax := range syntaxes {
		if slices.Contains(defaultSyntaxes, syntax) {
			syntaxesToSet = append(syntaxesToSet, syntax)
		}
	}
	if len(syntaxesToSet) == 0 {
		return e
	}

	e.cfg.syntaxes = syntaxesToSet

	return e
}

// SetUserAgent sets the User-Agent header for the HTTP client used by the Extractor.
// userAgent: A string representing the User-Agent to set for HTTP requests.
// Returns the updated Extractor instance.
func (e *Extractor) SetUserAgent(userAgent string) *Extractor {
	e.cfg.userAgent = userAgent

	return e
}

// SetFetchTimeout sets the HTTP client's fetch timeout value in seconds.
// fetchTimeout: A uint8 value representing the timeout duration in seconds.
// Returns the updated Extractor instance.
func (e *Extractor) SetFetchTimeout(fetchTimeout uint8) *Extractor {
	e.cfg.fetchTimeout = fetchTimeout

	return e
}

// SetHTTPClient injects a custom HTTP client to use for all fetch operations.
// When set, the client's own timeout and transport are used instead of the fetchTimeout setting.
// Returns the updated Extractor instance.
func (e *Extractor) SetHTTPClient(client *http.Client) *Extractor {
	e.cfg.httpClient = client

	return e
}

// SetMaxBodySize sets the maximum number of bytes read from an HTTP response body.
// Responses larger than this limit are rejected with a FetchError. Default is 10 MB.
// Returns the updated Extractor instance.
func (e *Extractor) SetMaxBodySize(size int64) *Extractor {
	e.cfg.maxBodySize = size

	return e
}

// Extract retrieves metadata from the specified URL or provided content and processes it using various parsers.
// ctx: A context for cancellation and timeout control of the HTTP fetch.
// url: The URL to extract metadata from.
// urlContent: Optional pointer to a string containing HTML content. If nil, the content at the URL will be fetched.
func (e *Extractor) Extract(ctx context.Context, url string, urlContent *string) (*Extractor, error) {
	var err error
	var mu sync.Mutex
	var wg sync.WaitGroup

	e.url = url
	e.content, err = e.setContent(ctx, urlContent)
	if err != nil {
		e.errs = append(e.errs, err)
		return e, err
	}

	var processors []processor

	if slices.Contains(e.cfg.syntaxes, SyntaxOpenGraph) {
		processors = append(processors, processor{
			Name: SyntaxOpenGraph,
			Func: func() (any, []error) {
				result, errs := extractor.ParseOpenGraph(e.url, e.content)
				if result == nil {
					return nil, errs
				}
				return result, errs
			},
		})
	}
	if slices.Contains(e.cfg.syntaxes, SyntaxXCards) {
		processors = append(processors, processor{
			Name: SyntaxXCards,
			Func: func() (any, []error) {
				result, errs := extractor.ParseXCards(e.url, e.content)
				if result == nil {
					return nil, errs
				}
				return result, errs
			},
		})
	}
	if slices.Contains(e.cfg.syntaxes, SyntaxJSONLD) {
		processors = append(processors, processor{
			Name: SyntaxJSONLD,
			Func: func() (any, []error) {
				return extractor.JSONLD(e.url, e.content)
			},
		})
	}
	if slices.Contains(e.cfg.syntaxes, SyntaxMicrodata) {
		processors = append(processors, processor{
			Name: SyntaxMicrodata,
			Func: func() (any, []error) {
				return extractor.W3CMicrodata(e.url, e.content)
			},
		})
	}
	if slices.Contains(e.cfg.syntaxes, SyntaxRDFa) {
		processors = append(processors, processor{
			Name: SyntaxRDFa,
			Func: func() (any, []error) {
				return extractor.RDFa(e.url, e.content)
			},
		})
	}
	if slices.Contains(e.cfg.syntaxes, SyntaxDublinCore) {
		processors = append(processors, processor{
			Name: SyntaxDublinCore,
			Func: func() (any, []error) {
				result, errs := extractor.DublinCore(e.url, e.content)
				if result == nil {
					return nil, errs
				}
				return result, errs
			},
		})
	}
	if slices.Contains(e.cfg.syntaxes, SyntaxMicroformats) {
		processors = append(processors, processor{
			Name: SyntaxMicroformats,
			Func: func() (any, []error) {
				return extractor.Microformats(e.url, e.content)
			},
		})
	}

	for _, proc := range processors {
		wg.Add(1)
		go func(proc processor) {
			defer wg.Done()
			extracted, errorsExtracted := proc.Func()

			mu.Lock()
			defer mu.Unlock()
			e.errs = append(e.errs, wrapParseErrors(proc.Name, errorsExtracted)...)
			e.extracted[proc.Name] = extracted
		}(proc)
	}

	wg.Wait()

	return e, nil
}

// setContent sets the content for the Extractor, fetching from URL if necessary. Returns the content or an error.
func (e *Extractor) setContent(ctx context.Context, urlContent *string) (string, error) {
	if urlContent != nil {
		return *urlContent, nil
	}
	mainURLContent, err := e.fetch(ctx, e.url)

	if err != nil {
		return "", err
	}
	return string(mainURLContent), nil
}

// fetch retrieves the content from the specified URL using the provided context. Returns the fetched content as a byte slice or an error if failed.
func (e *Extractor) fetch(ctx context.Context, url string) ([]byte, error) {
	var body bytes.Buffer

	client := e.cfg.httpClient
	if client == nil {
		client = &http.Client{
			Timeout: time.Duration(e.cfg.fetchTimeout) * time.Second,
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, &FetchError{URL: url, Err: err}
	}

	req.Header.Set("User-Agent", e.cfg.userAgent)

	response, err := client.Do(req)
	if err != nil {
		return nil, &FetchError{URL: url, Err: err}
	}

	if response.StatusCode != http.StatusOK {
		return nil, &FetchError{URL: url, Err: fmt.Errorf("received HTTP status %d", response.StatusCode)}
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	_, err = io.Copy(&body, io.LimitReader(response.Body, e.cfg.maxBodySize+1))
	if err != nil {
		return nil, &FetchError{URL: url, Err: err}
	}
	if int64(body.Len()) > e.cfg.maxBodySize {
		return nil, &FetchError{URL: url, Err: fmt.Errorf("response body exceeds limit of %d bytes", e.cfg.maxBodySize)}
	}

	return body.Bytes(), nil
}

// wrapParseErrors wraps each error in a ParseError tagged with the given syntax.
// Returns nil when errs is empty, preserving nil e.errs for callers that check len.
func wrapParseErrors(syntax Syntax, errs []error) []error {
	if len(errs) == 0 {
		return nil
	}
	wrapped := make([]error, len(errs))
	for i, err := range errs {
		wrapped[i] = &ParseError{Syntax: syntax, Err: err}
	}
	return wrapped
}

// GetExtracted returns the extracted metadata as a map by processor name from the Extractor instance.
func (e *Extractor) GetExtracted() map[Syntax]any {
	return e.extracted
}

// GetErrors returns the accumulated parse errors from the last Extract() call.
func (e *Extractor) GetErrors() []error {
	return e.errs
}

// GetExtractedJSON returns the extracted metadata as a JSON-formatted byte array with indentation.
func (e *Extractor) GetExtractedJSON() json.RawMessage {
	extractedJSON, errJSON := json.MarshalIndent(e.extracted, "", "  ")
	if errJSON != nil {
		e.errs = append(e.errs, errJSON)
	}

	return extractedJSON
}
