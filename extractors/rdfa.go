package extractor

import (
	"golang.org/x/net/html"
	"io"
	"strings"
)

// RDFaItem represents a single RDFa item extracted from HTML.
type RDFaItem struct {
	Vocab      string         `json:"vocab,omitempty"`
	Type       string         `json:"type,omitempty"`
	Resource   *string        `json:"resource,omitempty"`
	About      *string        `json:"about,omitempty"`
	Properties map[string]any `json:"properties,omitempty"`
}

// RDFa extracts RDFa structured data from HTML content.
func RDFa(_ string, htmlContent string) ([]RDFaItem, []error) {
	return parseRDFa(htmlContent)
}

func parseRDFa(htmlContent string) ([]RDFaItem, []error) {
	return parseRDFaFrom(strings.NewReader(htmlContent))
}

func parseRDFaFrom(r io.Reader) ([]RDFaItem, []error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, []error{err}
	}
	prefixes := collectRDFaPrefixes(doc)

	var items []RDFaItem
	walkRDFa(doc, "", prefixes, &items)
	return items, nil
}

func walkRDFa(n *html.Node, vocab string, prefixes map[string]string, items *[]RDFaItem) {
	if n.Type == html.ElementNode {
		if v := getAttrVal(n, "vocab"); v != "" {
			vocab = v
		}
		typeof := getAttrVal(n, "typeof")
		prop := getAttrVal(n, "property")
		if typeof != "" && prop == "" {
			var extraItems []RDFaItem
			item := buildRDFaItem(n, typeof, vocab, prefixes, &extraItems)
			*items = append(*items, item)
			*items = append(*items, extraItems...)
			return
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		walkRDFa(c, vocab, prefixes, items)
	}
}

func buildRDFaItem(n *html.Node, typeof, vocab string, prefixes map[string]string, extraItems *[]RDFaItem) RDFaItem {
	item := RDFaItem{
		Properties: make(map[string]any),
	}
	if vocab != "" {
		item.Vocab = vocab
	}
	item.Type = resolveRDFaTerm(typeof, vocab, prefixes)
	if resource := getAttrVal(n, "resource"); resource != "" {
		item.Resource = &resource
	}
	if about := getAttrVal(n, "about"); about != "" {
		item.About = &about
	}
	collectRDFaProperties(n, &item, vocab, prefixes, extraItems)
	return item
}

func collectRDFaProperties(n *html.Node, item *RDFaItem, vocab string, prefixes map[string]string, extraItems *[]RDFaItem) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type != html.ElementNode {
			continue
		}
		if v := getAttrVal(c, "vocab"); v != "" {
			vocab = v
		}
		prop := getAttrVal(c, "property")
		typeof := getAttrVal(c, "typeof")
		switch {
		case prop != "" && typeof != "":
			subItem := buildRDFaItem(c, typeof, vocab, prefixes, extraItems)
			propKey := resolveRDFaTerm(prop, vocab, prefixes)
			item.Properties[propKey] = appendValue(item.Properties[propKey], &subItem)
		case prop != "":
			propKey := resolveRDFaTerm(prop, vocab, prefixes)
			val := getRDFaPropertyValue(c)
			item.Properties[propKey] = appendValue(item.Properties[propKey], val)
			collectRDFaProperties(c, item, vocab, prefixes, extraItems)
		case typeof != "":
			sibling := buildRDFaItem(c, typeof, vocab, prefixes, extraItems)
			*extraItems = append(*extraItems, sibling)
		default:
			collectRDFaProperties(c, item, vocab, prefixes, extraItems)
		}
	}
}

func getRDFaPropertyValue(n *html.Node) string {
	if content := getAttrVal(n, "content"); content != "" {
		return content
	}
	if datetime := getAttrVal(n, "datetime"); datetime != "" {
		return datetime
	}
	if href := getAttrVal(n, "href"); href != "" {
		return href
	}
	if src := getAttrVal(n, "src"); src != "" {
		return src
	}
	if resource := getAttrVal(n, "resource"); resource != "" {
		return resource
	}
	return getTextContent(n)
}

func resolveRDFaTerm(term, vocab string, prefixes map[string]string) string {
	if strings.HasPrefix(term, "http://") || strings.HasPrefix(term, "https://") {
		return term
	}
	if idx := strings.Index(term, ":"); idx > 0 {
		prefix := term[:idx]
		local := term[idx+1:]
		if uri, ok := prefixes[prefix]; ok {
			return uri + local
		}
	}
	if vocab != "" {
		return vocab + term
	}
	return term
}

func collectRDFaPrefixes(doc *html.Node) map[string]string {
	prefixes := make(map[string]string)
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if prefix := getAttrVal(n, "prefix"); prefix != "" {
				parseRDFaPrefixes(prefix, prefixes)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return prefixes
}

func parseRDFaPrefixes(prefixStr string, prefixes map[string]string) {
	parts := strings.Fields(prefixStr)
	for i := 0; i+1 < len(parts); i += 2 {
		key := strings.TrimSuffix(parts[i], ":")
		prefixes[key] = parts[i+1]
	}
}
