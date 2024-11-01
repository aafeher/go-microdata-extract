package main

import (
	"fmt"
	"github.com/aafeher/go-microdata-extract"
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

	extractedJSON := em.GetExtractedJSON()
	fmt.Printf("Extracted data in JSON: %s\n", extractedJSON)
}
