package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/aafeher/go-microdata-extract"
)

func main() {
	// Default configuration — inspect with Get* methods.
	e := extract.New()

	fmt.Println("=== Default configuration ===")
	fmt.Printf("UserAgent:   %s\n", e.GetUserAgent())
	fmt.Printf("FetchTimeout: %d s\n", e.GetFetchTimeout())
	fmt.Printf("MaxBodySize: %d bytes\n", e.GetMaxBodySize())
	fmt.Printf("Syntaxes:    %v\n", e.GetSyntaxes())
	fmt.Printf("HTTPClient:  %v\n", e.GetHTTPClient())

	// Override a few settings and verify them.
	customClient := &http.Client{Timeout: 10 * time.Second}
	e.SetUserAgent("MyBot/1.0").
		SetFetchTimeout(5).
		SetMaxBodySize(5 * 1024 * 1024).
		SetSyntaxes([]extract.Syntax{extract.SyntaxOpenGraph, extract.SyntaxJSONLD}).
		SetHTTPClient(customClient)

	fmt.Println("\n=== After configuration ===")
	fmt.Printf("UserAgent:    %s\n", e.GetUserAgent())
	fmt.Printf("FetchTimeout: %d s\n", e.GetFetchTimeout())
	fmt.Printf("MaxBodySize:  %d bytes\n", e.GetMaxBodySize())
	fmt.Printf("Syntaxes:     %v\n", e.GetSyntaxes())
	fmt.Printf("HTTPClient:   %v\n", e.GetHTTPClient())
}
