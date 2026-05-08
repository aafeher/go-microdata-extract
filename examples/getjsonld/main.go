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
  <script type="application/ld+json">
  {
    "@context": "https://schema.org",
    "@type": "Article",
    "headline": "Example Article",
    "author": {
      "@type": "Person",
      "name": "Jane Doe"
    },
    "datePublished": "2025-01-01"
  }
  </script>
  <script type="application/ld+json">
  {
    "@context": "https://schema.org",
    "@type": "BreadcrumbList",
    "itemListElement": [
      { "@type": "ListItem", "position": 1, "name": "Home", "item": "https://example.com" },
      { "@type": "ListItem", "position": 2, "name": "Articles", "item": "https://example.com/articles" }
    ]
  }
  </script>
</head>
<body></body>
</html>`

	e := extract.New().SetSyntaxes([]extract.Syntax{extract.SyntaxJSONLD})
	em, err := e.Extract(context.Background(), url, &content)
	if err != nil {
		log.Fatalf("extraction failed: %v", err)
	}

	extracted := em.GetExtracted()

	items, ok := extracted[extract.SyntaxJSONLD].([]map[string]any)
	if !ok || len(items) == 0 {
		fmt.Println("no JSON-LD data found")
		return
	}

	fmt.Printf("found %d JSON-LD item(s)\n\n", len(items))

	for i, item := range items {
		fmt.Printf("item %d:\n", i+1)
		fmt.Printf("  @type: %v\n", item["@type"])

		if headline, ok := item["headline"].(string); ok {
			fmt.Printf("  headline: %s\n", headline)
		}
		if author, ok := item["author"].(map[string]any); ok {
			fmt.Printf("  author: %v\n", author["name"])
		}
		if datePublished, ok := item["datePublished"].(string); ok {
			fmt.Printf("  datePublished: %s\n", datePublished)
		}
		fmt.Println()
	}
}
