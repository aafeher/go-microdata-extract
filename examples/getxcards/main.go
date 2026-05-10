package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aafeher/go-microdata-extract"
	extractor "github.com/aafeher/go-microdata-extract/extractors"
)

func main() {
	url := "https://github.com/aafeher/go-microdata-extract"
	content := `<!DOCTYPE html>
<html>
<head>
  <meta name="twitter:card" content="summary">
  <meta name="twitter:site" content="@example">
  <meta name="twitter:creator" content="@author">
  <meta name="twitter:title" content="Example Page">
  <meta name="twitter:description" content="An example page with X Cards metadata.">
  <meta name="twitter:image" content="https://example.com/image.jpg">
</head>
<body></body>
</html>`

	e := extract.New().SetSyntaxes([]extract.Syntax{extract.SyntaxXCards})
	em, err := e.Extract(context.Background(), url, &content)
	if err != nil {
		log.Fatalf("extraction failed: %v", err)
	}

	extracted := em.GetExtracted()

	xc, ok := extracted[extract.SyntaxXCards].(*extractor.XCards)
	if !ok || xc == nil {
		fmt.Println("no X Cards data found")
		return
	}

	fmt.Printf("Card:        %s\n", xc.Card)
	fmt.Printf("Site:        %s\n", xc.Site)
	fmt.Printf("Creator:     %s\n", xc.Creator)
	fmt.Printf("Title:       %s\n", xc.Title)
	fmt.Printf("Description: %s\n", xc.Description)
	if len(xc.XCardsImage) > 0 {
		fmt.Printf("Image URL:   %s\n", xc.XCardsImage[0].URL)
	}
}
