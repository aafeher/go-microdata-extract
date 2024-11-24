package extractor

import (
	"fmt"
	"golang.org/x/net/html"
	"net/url"
	"strings"
)

type MicrodataItem struct {
	Type       string         `json:"type,omitempty"`
	ID         *string        `json:"id,omitempty"`
	Properties map[string]any `json:"properties,omitempty"`
}

func W3CMicrodata(URL string, htmlContent string) ([]MicrodataItem, []error) {
	items, errors := parseW3CMicrodata(URL, htmlContent)

	var results []MicrodataItem
	for _, item := range items {
		result := MicrodataItem{
			Type:       item.Type,
			Properties: item.Properties,
		}
		if item.ID != nil {
			result.ID = item.ID
		}
		results = append(results, result)

	}

	return results, errors
}

// parseW3CMicrodata parses an HTML input string to extract W3C microdata items and returns them along with any errors.
func parseW3CMicrodata(URL string, input string) ([]*MicrodataItem, []error) {
	var errors []error

	// strings.NewReader() always provides a valid reader for html.Parse()
	doc, _ := html.Parse(strings.NewReader(input))

	var items []*MicrodataItem
	var parseNode func(*html.Node)
	parseNode = func(n *html.Node) {
		if n.Type == html.ElementNode && getAttr(n, "itemscope") {
			item := &MicrodataItem{
				Properties: make(map[string]any),
			}
			itemType := getAttrVal(n, "itemtype")
			if itemType != "" {
				item.Type = itemType
			}
			itemID := getAttrVal(n, "itemid")
			if itemID != "" {
				item.ID = &itemID
			}
			parseProperties(n, item, URL)

			items = append(items, item)
		} else {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				parseNode(c)
			}
		}
	}
	parseNode(doc)

	return items, errors
}

func parseProperties(n *html.Node, item *MicrodataItem, URL string) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode {
			if prop := getAttrVal(c, "itemprop"); prop != "" {
				if getAttr(c, "itemscope") {
					subItem := &MicrodataItem{
						Properties: make(map[string]any),
					}
					subItemType := getAttrVal(c, "itemtype")
					if subItemType != "" {
						subItem.Type = subItemType
					}
					subItemID := getAttrVal(c, "itemid")
					if subItemID != "" {
						subItem.ID = &subItemID
					}
					parseProperties(c, subItem, URL)
					item.Properties[prop] = appendValue(item.Properties[prop], subItem)
				} else {
					value := getTextContent(c)
					if datetime := getAttrVal(c, "datetime"); datetime != "" {
						value = datetime
					} else if prop == "url" || strings.HasSuffix(prop, "Url") {
						href := getAttrVal(c, "href")
						if strings.HasPrefix(href, "//") || strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
							value = href
						} else {
							baseURL := ""
							parsedURL, err := url.Parse(URL)
							if err == nil {
								baseURL = fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
							}
							value = baseURL + href
						}
					}
					item.Properties[prop] = appendValue(item.Properties[prop], value)
				}
			} else {
				parseProperties(c, item, URL)
			}
		}
	}
}

func getAttr(n *html.Node, key string) bool {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return true
		}
	}
	return false
}

func getAttrVal(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return strings.TrimSpace(attr.Val)
		}
	}
	return ""
}

func appendValue(existing any, value any) any {
	if existing == nil {
		return value
	}
	switch v := existing.(type) {
	case []any:
		return append(v, value)
	default:
		return []any{existing, value}
	}
}

func getTextContent(n *html.Node) string {
	var sb strings.Builder
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			if sb.String() == "" {
				sb.WriteString(n.Data)
			}
		} else if n.Type == html.ElementNode {
			val := ""
			for _, attr := range n.Attr {
				if attr.Key == "value" {
					val = attr.Val
					break
				}
			}
			sb.WriteString(val)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	return strings.TrimSpace(sb.String())
}
