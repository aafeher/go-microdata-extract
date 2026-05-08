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
  <div itemscope itemtype="https://schema.org/Person" itemid="https://example.com/person/1">
    <span itemprop="name">Jane Doe</span>
    <span itemprop="jobTitle">Software Engineer</span>
    <div itemprop="address" itemscope itemtype="https://schema.org/PostalAddress">
      <span itemprop="streetAddress">123 Main St</span>
      <span itemprop="addressLocality">Springfield</span>
      <span itemprop="addressCountry">US</span>
    </div>
  </div>
  <div itemscope itemtype="https://schema.org/Product">
    <span itemprop="name">Awesome Widget</span>
    <span itemprop="description">A very useful widget.</span>
    <meta itemprop="price" content="9.99" />
  </div>
</body>
</html>`

	e := extract.New().SetSyntaxes([]extract.Syntax{extract.SyntaxMicrodata})
	em, err := e.Extract(context.Background(), url, &content)
	if err != nil {
		log.Fatalf("extraction failed: %v", err)
	}

	extracted := em.GetExtracted()

	items, ok := extracted[extract.SyntaxMicrodata].([]extractor.MicrodataItem)
	if !ok || len(items) == 0 {
		fmt.Println("no Microdata found")
		return
	}

	fmt.Printf("found %d Microdata item(s)\n\n", len(items))

	for i, item := range items {
		fmt.Printf("item %d:\n", i+1)
		fmt.Printf("  type: %s\n", item.Type)
		if item.ID != nil {
			fmt.Printf("  id:   %s\n", *item.ID)
		}

		for prop, val := range item.Properties {
			switch v := val.(type) {
			case *extractor.MicrodataItem:
				// nested item (single)
				fmt.Printf("  %s (nested %s):\n", prop, v.Type)
				for subProp, subVal := range v.Properties {
					fmt.Printf("    %s: %v\n", subProp, subVal)
				}
			case []any:
				// multiple values for the same property
				fmt.Printf("  %s: %v\n", prop, v)
			default:
				fmt.Printf("  %s: %v\n", prop, val)
			}
		}
		fmt.Println()
	}
}
