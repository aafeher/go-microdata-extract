package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aafeher/go-microdata-extract"
	extractor "github.com/aafeher/go-microdata-extract/extractors"
)

func main() {
	url := "https://example.com"
	content := `<!DOCTYPE html>
<html>
<body>
  <div vocab="https://schema.org/" typeof="Person">
    <span property="name">Jane Doe</span>
    <span property="jobTitle">Software Engineer</span>
    <div property="address" typeof="PostalAddress">
      <span property="streetAddress">123 Main St</span>
      <span property="addressLocality">Springfield</span>
    </div>
  </div>
</body>
</html>`

	e := extract.New().SetSyntaxes([]extract.Syntax{extract.SyntaxRDFa})
	em, err := e.Extract(context.Background(), url, &content)
	if err != nil {
		log.Fatalf("extraction failed: %v", err)
	}

	extracted := em.GetExtracted()

	items, ok := extracted[extract.SyntaxRDFa].([]extractor.RDFaItem)
	if !ok || len(items) == 0 {
		fmt.Println("no RDFa data found")
		return
	}

	fmt.Printf("found %d RDFa item(s)\n\n", len(items))

	for i, item := range items {
		fmt.Printf("item %d:\n", i+1)
		fmt.Printf("  vocab: %s\n", item.Vocab)
		fmt.Printf("  type:  %s\n", item.Type)
		if item.Resource != nil {
			fmt.Printf("  resource: %s\n", *item.Resource)
		}
		if item.About != nil {
			fmt.Printf("  about: %s\n", *item.About)
		}

		for prop, val := range item.Properties {
			switch v := val.(type) {
			case *extractor.RDFaItem:
				fmt.Printf("  %s (nested %s):\n", prop, v.Type)
				for subProp, subVal := range v.Properties {
					fmt.Printf("    %s: %v\n", subProp, subVal)
				}
			default:
				fmt.Printf("  %s: %v\n", prop, val)
			}
		}
		fmt.Println()
	}
}
