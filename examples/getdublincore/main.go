package main

import (
	"fmt"
	"log"

	"github.com/aafeher/go-microdata-extract"
	extractor "github.com/aafeher/go-microdata-extract/extractors"
)

func main() {
	url := "https://example.com"
	content := `<!DOCTYPE html>
<html>
<head>
  <meta name="DC.title" content="Example Page">
  <meta name="DC.creator" content="Jane Doe">
  <meta name="DC.subject" content="Technology">
  <meta name="DC.subject" content="Science">
  <meta name="DC.description" content="An example page with Dublin Core metadata.">
  <meta name="DC.date" content="2026-01-01">
  <meta name="DC.language" content="en">
</head>
<body></body>
</html>`

	e := extract.New().SetSyntaxes([]extract.Syntax{extract.SyntaxDublinCore})
	em, err := e.Extract(url, &content)
	if err != nil {
		log.Fatalf("extraction failed: %v", err)
	}

	extracted := em.GetExtracted()

	item, ok := extracted[extract.SyntaxDublinCore].(*extractor.DublinCoreItem)
	if !ok || item == nil {
		fmt.Println("no Dublin Core data found")
		return
	}

	fmt.Printf("found %d Dublin Core property(ies)\n\n", len(item.Properties))

	for prop, val := range item.Properties {
		switch v := val.(type) {
		case []any:
			fmt.Printf("  %s:\n", prop)
			for _, entry := range v {
				fmt.Printf("    - %v\n", entry)
			}
		default:
			fmt.Printf("  %s: %v\n", prop, v)
		}
	}
}
