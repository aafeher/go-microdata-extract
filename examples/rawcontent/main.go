package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aafeher/go-microdata-extract"
)

func main() {
	url := "https://example.com"
	content := `<!DOCTYPE html>
<html>
<head>
  <meta property="og:title" content="Example Page" />
  <meta property="og:description" content="This is an example page." />
  <meta property="og:url" content="https://example.com" />
  <meta property="og:image" content="https://example.com/image.jpg" />
</head>
<body></body>
</html>`

	e := extract.New()
	em, err := e.Extract(context.Background(), url, &content)
	if err != nil {
		log.Fatalf("extraction failed: %v", err)
	}

	extractedJSON := em.GetExtractedJSON()
	fmt.Printf("Extracted data in JSON: %s\n", extractedJSON)
}
