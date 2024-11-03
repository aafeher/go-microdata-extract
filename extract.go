package extract

import (
	"bytes"
	"encoding/json"
	"fmt"
	extractor "github.com/aafeher/go-microdata-extract/extractors"
	"io"
	"net/http"
	"sync"
	"time"
)

type (
	// Extractor is a struct used for extracting metadata from web content or a provided URL. It utilizes various processors.
	Extractor struct {
		cfg       config
		url       string
		content   string
		extracted map[Syntax]interface{}
		errs      []error
	}

	// config represents configuration settings for an Extractor, including syntax options, user agent, and fetch timeout.
	config struct {
		syntaxes     []Syntax
		userAgent    string
		fetchTimeout uint8
	}

	// Processor represents a data structure to hold a processor's name and function for extracting metadata.
	Processor struct {
		Name Syntax
		Func func() (interface{}, []error)
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
)

// SYNTAXES defines an array of metadata syntax identifiers supported for parsing.
var SYNTAXES = []Syntax{SyntaxOpenGraph, SyntaxXCards, SyntaxJSONLD, SyntaxMicrodata}

// New creates a new instance of Extractor with default configurations and an empty map for extracted data.
func New() *Extractor {
	e := &Extractor{
		extracted: make(map[Syntax]interface{}),
	}

	e.setConfigDefaults()

	return e
}

// setConfigDefaults initializes the Extractor with default configuration settings.
func (e *Extractor) setConfigDefaults() {
	e.cfg = config{
		syntaxes:     SYNTAXES,
		userAgent:    "go-microdata-extract (+https://github.com/aafeher/go-microdata-extract/blob/main/README.md)",
		fetchTimeout: 3,
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
		if contains(SYNTAXES, syntax) {
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

// Extract retrieves metadata from the specified URL or provided content and processes it using various parsers.
// url: The URL to extract metadata from.
// urlContent: Optional pointer to a string containing HTML content. If nil, the content at the URL will be fetched.
func (e *Extractor) Extract(url string, urlContent *string) (*Extractor, error) {
	var err error
	var mu sync.Mutex
	var wg sync.WaitGroup

	e.url = url
	e.content, err = e.setContent(urlContent)
	if err != nil {
		e.errs = append(e.errs, err)
		return e, err
	}

	var processors []Processor

	if contains(e.cfg.syntaxes, SyntaxOpenGraph) {
		processors = append(processors, Processor{
			Name: SyntaxOpenGraph,
			Func: func() (interface{}, []error) {
				return extractor.ParseOpenGraph(e.url, e.content)
			},
		})
	}
	if contains(e.cfg.syntaxes, SyntaxXCards) {
		processors = append(processors, Processor{
			Name: SyntaxXCards,
			Func: func() (interface{}, []error) {
				return extractor.ParseXCards(e.url, e.content)
			},
		})
	}
	if contains(e.cfg.syntaxes, SyntaxJSONLD) {
		processors = append(processors, Processor{
			Name: SyntaxJSONLD,
			Func: func() (interface{}, []error) {
				return extractor.JSONLD(e.url, e.content)
			},
		})
	}
	if contains(e.cfg.syntaxes, SyntaxMicrodata) {
		processors = append(processors, Processor{
			Name: SyntaxMicrodata,
			Func: func() (interface{}, []error) {
				return extractor.W3CMicrodata(e.url, e.content)
			},
		})
	}

	for _, processor := range processors {
		wg.Add(1)
		proc := processor
		go func(proc Processor) {
			defer wg.Done()
			extracted, errorsExtracted := proc.Func()

			mu.Lock()
			defer mu.Unlock()
			e.errs = append(e.errs, errorsExtracted...)
			e.extracted[proc.Name] = extracted
		}(proc)
	}

	wg.Wait()

	return e, nil
}

// setContent sets the content for the Extractor, fetching from URL if necessary. Returns the content or an error.
func (e *Extractor) setContent(urlContent *string) (string, error) {
	if urlContent != nil {
		return *urlContent, nil
	}
	mainURLContent, err := e.fetch(e.url)

	if err != nil {
		return "", err
	}
	return string(mainURLContent), nil
}

// fetch retrieves the content from the specified URL. Returns the fetched content as a byte slice or an error if failed.
func (e *Extractor) fetch(url string) ([]byte, error) {
	var body bytes.Buffer

	client := &http.Client{
		Timeout: time.Duration(e.cfg.fetchTimeout) * time.Second,
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", e.cfg.userAgent)

	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received HTTP status %d", response.StatusCode)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	_, err = io.Copy(&body, response.Body)
	if err != nil {
		return nil, err
	}

	return body.Bytes(), nil
}

// GetExtracted returns the extracted metadata as a map by processor name from the Extractor instance.
func (e *Extractor) GetExtracted() map[Syntax]interface{} {
	return e.extracted
}

// GetExtractedJSON returns the extracted metadata as a JSON-formatted byte array with indentation.
func (e *Extractor) GetExtractedJSON() json.RawMessage {
	extractedJSON, errJSON := json.MarshalIndent(e.extracted, "", "  ")
	if errJSON != nil {
		e.errs = append(e.errs, errJSON)
	}

	return extractedJSON
}

// index returns the index of the first occurrence of v in s,
// or -1 if not present.
func index[S ~[]E, E comparable](s S, v E) int {
	for i := range s {
		if v == s[i] {
			return i
		}
	}
	return -1
}

// contains reports whether v is present in s.
func contains[S ~[]E, E comparable](s S, v E) bool {
	return index(s, v) >= 0
}
