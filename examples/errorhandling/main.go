package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/aafeher/go-microdata-extract"
)

func main() {
	// Scenario 1: invalid URL — Extract returns an error immediately, no data is extracted.
	fmt.Println("=== Scenario 1: invalid URL ===")
	_, err := extract.New().Extract("not-a-valid-url", nil)
	if err != nil {
		fmt.Printf("fetch error: %v\n", err)
	}

	// Scenario 2: unreachable host — network-level error propagated from the HTTP client.
	fmt.Println("\n=== Scenario 2: unreachable host ===")
	_, err = extract.New().SetFetchTimeout(2).Extract("http://localhost:19999", nil)
	if err != nil {
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			fmt.Printf("network error (%s %s): %v\n", urlErr.Op, urlErr.URL, urlErr.Err)
		} else {
			fmt.Printf("fetch error: %v\n", err)
		}
	}

	// Scenario 3: non-200 HTTP response — Extract returns an error with the status code.
	fmt.Println("\n=== Scenario 3: non-200 HTTP response ===")
	resp, _ := http.Get("https://httpbin.org/status/404") //nolint:noctx
	if resp != nil {
		_ = resp.Body.Close()
	}
	_, err = extract.New().Extract("https://httpbin.org/status/404", nil)
	if err != nil {
		fmt.Printf("http error: %v\n", err)
	}

	// Scenario 4: successful extraction — check whether a syntax returned data before using it.
	fmt.Println("\n=== Scenario 4: check for missing syntax data ===")
	content := `<!DOCTYPE html><html><head></head><body></body></html>`
	em, err := extract.New().Extract("https://example.com", &content)
	if err != nil {
		fmt.Printf("fetch error: %v\n", err)
		return
	}

	extracted := em.GetExtracted()
	for _, syntax := range extract.SYNTAXES {
		if data, ok := extracted[syntax]; !ok || data == nil {
			fmt.Printf("syntax %q: no data found\n", syntax)
		} else {
			fmt.Printf("syntax %q: data present\n", syntax)
		}
	}
}
