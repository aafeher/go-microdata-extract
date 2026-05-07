package extractor

import (
	"golang.org/x/net/html"
	"io"
	"strings"
)

// DublinCoreItem represents Dublin Core metadata extracted from HTML.
type DublinCoreItem struct {
	Properties map[string]any `json:"properties,omitempty"`
}

// DublinCore extracts Dublin Core metadata from HTML content.
// It supports DC.* and DCTERMS.* prefixes in <meta name> and <link rel> attributes.
func DublinCore(URL string, htmlContent string) (*DublinCoreItem, []error) {
	return parseDublinCore(URL, htmlContent)
}

func parseDublinCore(URL, htmlContent string) (*DublinCoreItem, []error) {
	return parseDublinCoreFrom(URL, strings.NewReader(htmlContent))
}

func parseDublinCoreFrom(URL string, r io.Reader) (*DublinCoreItem, []error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, []error{err}
	}
	item := &DublinCoreItem{
		Properties: make(map[string]any),
	}
	collectDublinCoreProps(doc, item, URL)
	if len(item.Properties) == 0 {
		return nil, nil
	}
	return item, nil
}

func collectDublinCoreProps(n *html.Node, item *DublinCoreItem, URL string) {
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
					item.Properties[prop] = appendValue(item.Properties[prop], resolveURL(URL, href))
				}
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		collectDublinCoreProps(c, item, URL)
	}
}

func dcProperty(name string) string {
	if strings.HasPrefix(name, "dc.") || strings.HasPrefix(name, "dcterms.") {
		return name
	}
	return ""
}
