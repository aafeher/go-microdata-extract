package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/aafeher/go-microdata-extract"
)

func main() {
	ctx := context.Background()

	// Scenario 1: misconfigured extractor — Extract returns a ConfigError before any network call.
	fmt.Println("=== Scenario 1: invalid configuration ===")
	_, err := extract.New().SetFetchTimeout(0).Extract(ctx, "https://example.com", nil)
	if err != nil {
		var ce *extract.ConfigError
		if errors.As(err, &ce) {
			fmt.Printf("config error — field %q: %v\n", ce.Field, ce.Err)
		}
	}

	// Scenario 2: invalid URL — Extract returns a FetchError immediately, no data is extracted.
	fmt.Println("\n=== Scenario 2: invalid URL ===")
	_, err = extract.New().Extract(ctx, "not-a-valid-url", nil)
	if err != nil {
		var fe *extract.FetchError
		if errors.As(err, &fe) {
			fmt.Printf("fetch error for %s: %v\n", fe.URL, fe.Err)
		} else {
			fmt.Printf("fetch error: %v\n", err)
		}
	}

	// Scenario 3: unreachable host — network-level error propagated from the HTTP client.
	fmt.Println("\n=== Scenario 3: unreachable host ===")
	_, err = extract.New().SetFetchTimeout(2).Extract(ctx, "http://localhost:19999", nil)
	if err != nil {
		var fe *extract.FetchError
		if errors.As(err, &fe) {
			fmt.Printf("fetch error for %s: %v\n", fe.URL, fe.Err)
		} else {
			fmt.Printf("fetch error: %v\n", err)
		}
	}

	// Scenario 4: non-200 HTTP response — Extract returns a FetchError with the status code.
	fmt.Println("\n=== Scenario 4: non-200 HTTP response ===")
	resp, _ := http.Get("https://httpbin.org/status/404") //nolint:noctx
	if resp != nil {
		_ = resp.Body.Close()
	}
	_, err = extract.New().Extract(ctx, "https://httpbin.org/status/404", nil)
	if err != nil {
		var fe *extract.FetchError
		if errors.As(err, &fe) {
			fmt.Printf("http error for %s: %v\n", fe.URL, fe.Err)
		} else {
			fmt.Printf("http error: %v\n", err)
		}
	}

	// Scenario 5: successful extraction — check whether a syntax returned data before using it.
	fmt.Println("\n=== Scenario 5: check for missing syntax data ===")
	content := `<!DOCTYPE html><html><head></head><body></body></html>`
	em, err := extract.New().Extract(ctx, "https://example.com", &content)
	if err != nil {
		fmt.Printf("fetch error: %v\n", err)
		return
	}

	extracted := em.GetExtracted()
	for _, syntax := range extract.DefaultSyntaxes() {
		if data, ok := extracted[syntax]; !ok || data == nil {
			fmt.Printf("syntax %q: no data found\n", syntax)
		} else {
			fmt.Printf("syntax %q: data present\n", syntax)
		}
	}
}
