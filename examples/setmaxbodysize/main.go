package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/aafeher/go-microdata-extract"
)

func main() {
	// Serve a small HTML page to demonstrate the limit.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<!DOCTYPE html><html><head>
<meta property="og:title" content="Test Page">
</head><body></body></html>`)
	}))
	defer srv.Close()

	// Example 1: body fits within the limit — extraction succeeds.
	fmt.Println("=== Example 1: body within limit ===")
	em, err := extract.New().
		SetSyntaxes([]extract.Syntax{extract.SyntaxOpenGraph}).
		SetMaxBodySize(1*1024*1024). // 1 MB
		Extract(context.Background(), srv.URL, nil)
	if err != nil {
		log.Fatalf("unexpected error: %v", err)
	}
	fmt.Printf("extracted JSON:\n%s\n\n", em.GetExtractedJSON())

	// Example 2: body exceeds the limit — Extract returns a FetchError.
	fmt.Println("=== Example 2: body exceeds limit ===")
	largeContent := strings.Repeat("x", 200)
	largeSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, largeContent)
	}))
	defer largeSrv.Close()

	_, err = extract.New().
		SetMaxBodySize(100). // 100 bytes
		Extract(context.Background(), largeSrv.URL, nil)
	if err != nil {
		var fe *extract.FetchError
		if errors.As(err, &fe) {
			fmt.Printf("fetch error (as expected): %v\n", fe.Err)
		}
	}
}
