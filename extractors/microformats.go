package extractor

import (
	"golang.org/x/net/html"
	"io"
	"strings"
)

// MicroformatItem represents a single Microformats2 item extracted from HTML.
type MicroformatItem struct {
	Type       []string           `json:"type,omitempty"`
	Properties map[string]any     `json:"properties,omitempty"`
	Children   []*MicroformatItem `json:"children,omitempty"`
}

// ParseMicroformats extracts Microformats2 items from htmlContent, resolving relative u-* URL values against URL.
func ParseMicroformats(URL string, htmlContent string) ([]MicroformatItem, []error) {
	return parseMicroformats(URL, htmlContent)
}

func parseMicroformats(baseURL, htmlContent string) ([]MicroformatItem, []error) {
	return parseMicroformatsFrom(baseURL, strings.NewReader(htmlContent))
}

func parseMicroformatsFrom(baseURL string, r io.Reader) ([]MicroformatItem, []error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, []error{err}
	}
	var items []MicroformatItem
	walkMicroformats(doc, baseURL, &items)
	return items, nil
}

func walkMicroformats(n *html.Node, baseURL string, items *[]MicroformatItem) {
	if n.Type == html.ElementNode {
		if hClasses := mfRootClasses(n); len(hClasses) > 0 {
			item := buildMicroformatItem(n, hClasses, baseURL)
			*items = append(*items, *item)
			return
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		walkMicroformats(c, baseURL, items)
	}
}

func buildMicroformatItem(n *html.Node, hClasses []string, baseURL string) *MicroformatItem {
	item := &MicroformatItem{
		Type:       hClasses,
		Properties: make(map[string]any),
	}
	collectMicroformatProperties(n, item, baseURL)
	return item
}

func collectMicroformatProperties(n *html.Node, item *MicroformatItem, baseURL string) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type != html.ElementNode {
			continue
		}
		classes := mfClasses(c)
		hClasses := filterPrefix(classes, "h-")
		pClasses := filterPrefix(classes, "p-")
		uClasses := filterPrefix(classes, "u-")
		dtClasses := filterPrefix(classes, "dt-")
		eClasses := filterPrefix(classes, "e-")

		isRoot := len(hClasses) > 0
		isProp := len(pClasses)+len(uClasses)+len(dtClasses)+len(eClasses) > 0

		switch {
		case isRoot && isProp:
			// nested item acting as a property value
			nested := buildMicroformatItem(c, hClasses, baseURL)
			for _, prop := range pClasses {
				item.Properties[prop] = appendValue(item.Properties[prop], nested)
			}
			for _, prop := range uClasses {
				item.Properties[prop] = appendValue(item.Properties[prop], nested)
			}
			for _, prop := range dtClasses {
				item.Properties[prop] = appendValue(item.Properties[prop], nested)
			}
			for _, prop := range eClasses {
				item.Properties[prop] = appendValue(item.Properties[prop], nested)
			}
		case isRoot:
			// child root item → Children
			nested := buildMicroformatItem(c, hClasses, baseURL)
			item.Children = append(item.Children, nested)
		default:
			for _, prop := range pClasses {
				item.Properties[prop] = appendValue(item.Properties[prop], mfTextValue(c))
			}
			for _, prop := range uClasses {
				item.Properties[prop] = appendValue(item.Properties[prop], mfURLValue(c, baseURL))
			}
			for _, prop := range dtClasses {
				item.Properties[prop] = appendValue(item.Properties[prop], mfDatetimeValue(c))
			}
			for _, prop := range eClasses {
				item.Properties[prop] = appendValue(item.Properties[prop], mfEmbeddedValue(c))
			}
			collectMicroformatProperties(c, item, baseURL)
		}
	}
}

// mfRootClasses returns all h-* class values on an element.
func mfRootClasses(n *html.Node) []string {
	return filterPrefix(mfClasses(n), "h-")
}

// mfClasses returns all class tokens on an element.
func mfClasses(n *html.Node) []string {
	raw := getAttrVal(n, "class")
	if raw == "" {
		return nil
	}
	return strings.Fields(raw)
}

// filterPrefix returns tokens that start with the given prefix.
func filterPrefix(tokens []string, prefix string) []string {
	var out []string
	for _, t := range tokens {
		if strings.HasPrefix(t, prefix) {
			out = append(out, t)
		}
	}
	return out
}

// mfTextValue returns the plain-text value for a p-* property.
func mfTextValue(n *html.Node) string {
	if n.Data == "abbr" || n.Data == "link" {
		if v := getAttrVal(n, "title"); v != "" {
			return v
		}
	}
	if n.Data == "data" || n.Data == "input" {
		if v := getAttrVal(n, "value"); v != "" {
			return v
		}
	}
	if n.Data == "img" || n.Data == "area" {
		if v := getAttrVal(n, "alt"); v != "" {
			return v
		}
	}
	return strings.TrimSpace(mfInnerText(n))
}

// mfURLValue returns the URL value for a u-* property, resolved against baseURL.
func mfURLValue(n *html.Node, baseURL string) string {
	switch n.Data {
	case "a", "area", "link":
		if v := getAttrVal(n, "href"); v != "" {
			return resolveURL(baseURL, v)
		}
	case "img", "video", "audio", "source", "iframe", "embed":
		if v := getAttrVal(n, "src"); v != "" {
			return resolveURL(baseURL, v)
		}
	case "object":
		if v := getAttrVal(n, "data"); v != "" {
			return resolveURL(baseURL, v)
		}
	case "form":
		if v := getAttrVal(n, "action"); v != "" {
			return resolveURL(baseURL, v)
		}
	}
	return strings.TrimSpace(mfInnerText(n))
}

// mfDatetimeValue returns the datetime value for a dt-* property.
func mfDatetimeValue(n *html.Node) string {
	switch n.Data {
	case "time", "ins", "del":
		if v := getAttrVal(n, "datetime"); v != "" {
			return v
		}
	case "abbr":
		if v := getAttrVal(n, "title"); v != "" {
			return v
		}
	case "data", "input":
		if v := getAttrVal(n, "value"); v != "" {
			return v
		}
	}
	return strings.TrimSpace(mfInnerText(n))
}

// mfEmbeddedValue returns the inner HTML for an e-* property.
func mfEmbeddedValue(n *html.Node) string {
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		_ = html.Render(&sb, c)
	}
	return strings.TrimSpace(sb.String())
}

// mfInnerText returns the concatenated text content of a node and its descendants.
func mfInnerText(n *html.Node) string {
	var sb strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			sb.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return sb.String()
}
