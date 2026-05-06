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
<body>
  <div class="h-card">
    <span class="p-name">Jane Doe</span>
    <span class="p-job-title">Software Engineer</span>
    <a class="u-url" href="https://example.com/janedoe">Homepage</a>
    <div class="p-address h-adr">
      <span class="p-locality">Springfield</span>
      <span class="p-country-name">United States</span>
    </div>
  </div>
</body>
</html>`

	e := extract.New().SetSyntaxes([]extract.Syntax{extract.SyntaxMicroformats})
	em, err := e.Extract(url, &content)
	if err != nil {
		log.Fatalf("extraction failed: %v", err)
	}

	extracted := em.GetExtracted()

	items, ok := extracted[extract.SyntaxMicroformats].([]extractor.MicroformatItem)
	if !ok || len(items) == 0 {
		fmt.Println("no Microformats data found")
		return
	}

	fmt.Printf("found %d Microformats item(s)\n\n", len(items))
	printItem(items[0], 0)
}

func printItem(item extractor.MicroformatItem, indent int) {
	pad := func() string {
		s := ""
		for i := 0; i < indent; i++ {
			s += "  "
		}
		return s
	}
	fmt.Printf("%stype: %v\n", pad(), item.Type)
	for prop, val := range item.Properties {
		switch v := val.(type) {
		case *extractor.MicroformatItem:
			fmt.Printf("%s%s (nested %v):\n", pad(), prop, v.Type)
			printItem(*v, indent+1)
		default:
			fmt.Printf("%s%s: %v\n", pad(), prop, val)
		}
	}
}
