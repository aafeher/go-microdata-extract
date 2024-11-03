package main

import (
	"fmt"
	"github.com/aafeher/go-microdata-extract"
	extractor "github.com/aafeher/go-microdata-extract/extractors"
	"log"
)

func main() {
	url := "https://github.com/aafeher/go-microdata-extract"

	e := extract.New()
	em, err := e.Extract(url, nil)
	if err != nil {
		log.Printf("%v", err)
	}

	extracted := em.GetExtracted()
	fmt.Printf("Extracted data: %v\n", extracted)

	extractedOG := extracted[extract.SyntaxOpenGraph].(*extractor.OpenGraph)
	fmt.Printf("Extracted OG data: %v\n", extractedOG)

	extractedOGTitle := extracted[extract.SyntaxOpenGraph].(*extractor.OpenGraph).Title
	fmt.Printf("Extracted OG Title: %v\n", extractedOGTitle)
}
