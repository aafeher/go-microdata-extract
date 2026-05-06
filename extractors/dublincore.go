package extractor

import (
	"golang.org/x/net/html"
	"strings"
)

// DublinCoreItem represents Dublin Core metadata extracted from HTML.
type DublinCoreItem struct {
	Properties map[string]any `json:"properties,omitempty"`
}

// DublinCore extracts Dublin Core metadata from HTML content.
// It supports DC.* and DCTERMS.* prefixes in <meta name> and <link rel> attributes.
func DublinCore(_ string, htmlContent string) (any, []error) {
	item := parseDublinCore(htmlContent)
	var result any
	if item != nil {
		result = item
	}
	return result, nil
}

func parseDublinCore(htmlContent string) *DublinCoreItem {
	doc, _ := html.Parse(strings.NewReader(htmlContent))
	item := &DublinCoreItem{
		Properties: make(map[string]any),
	}
	collectDublinCoreProps(doc, item)
	if len(item.Properties) == 0 {
		return nil
	}
	return item
}

func collectDublinCoreProps(n *html.Node, item *DublinCoreItem) {
	if n.Type == html.ElementNode {
		switch n.Data {
		case "meta":
			name := strings.ToLower(getAttrVal(n, "name"))
			content := getAttrVal(n, "content")
			if content != "" {
				if prop := dcProperty(name); prop != "" {
					item.Properties[prop] = appendValue(item.Properties[prop], content)
				}
			}
		case "link":
			rel := strings.ToLower(getAttrVal(n, "rel"))
			href := getAttrVal(n, "href")
			if href != "" {
				if prop := dcProperty(rel); prop != "" {
					item.Properties[prop] = appendValue(item.Properties[prop], href)
				}
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		collectDublinCoreProps(c, item)
	}
}

func dcProperty(name string) string {
	if strings.HasPrefix(name, "dc.") || strings.HasPrefix(name, "dcterms.") {
		return name
	}
	return ""
}
