package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aafeher/go-microdata-extract"
)

func main() {
	url := "https://github.com/aafeher/go-microdata-extract"
	content := `<!DOCTYPE html>
<html>
<head>
  <meta property="og:title" content="Example Page">
  <meta property="og:description" content="An example page.">
  <meta name="twitter:card" content="summary">
  <meta name="twitter:title" content="Example Page">
  <script type="application/ld+json">
  {"@context":"https://schema.org","@type":"WebPage","name":"Example Page"}
  </script>
</head>
<body></body>
</html>`

	em, err := extract.New().Extract(context.Background(), url, &content)
	if err != nil {
		log.Fatalf("extraction failed: %v", err)
	}

	fmt.Println(string(em.GetExtractedJSON()))
}
