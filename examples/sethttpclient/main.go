package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aafeher/go-microdata-extract"
)

func main() {
	url := "https://github.com/aafeher/go-microdata-extract"

	customClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	e := extract.New().SetHTTPClient(customClient)
	em, err := e.Extract(context.Background(), url, nil)
	if err != nil {
		log.Fatalf("extraction failed: %v", err)
	}

	fmt.Println("extraction successful with custom HTTP client")
	fmt.Printf("extracted JSON:\n%s\n", em.GetExtractedJSON())
}
