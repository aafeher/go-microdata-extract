package extractor

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestRDFa_Basic(t *testing.T) {
	// typeof without property → top-level item captured by walkRDFa
	htmlStr := `<html><body>
		<div typeof="Person">
			<span property="name">Alice</span>
		</div>
	</body></html>`
	items, errs := ParseRDFa("", htmlStr)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(items) == 0 {
		t.Fatal("expected at least 1 item")
	}
	if items[0].Type != "Person" {
		t.Errorf("Type: %q", items[0].Type)
	}
	if items[0].Properties["name"] == nil {
		t.Errorf("name property not found: %v", items[0].Properties)
	}
}

func TestRDFa_WithVocab(t *testing.T) {
	htmlStr := `<html><body vocab="http://schema.org/">
		<div typeof="Article">
			<span property="name">My Article</span>
		</div>
	</body></html>`
	items, errs := ParseRDFa("", htmlStr)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(items) == 0 {
		t.Fatal("expected at least 1 item")
	}
	if items[0].Vocab != "http://schema.org/" {
		t.Errorf("Vocab: %q", items[0].Vocab)
	}
	if items[0].Type != "http://schema.org/Article" {
		t.Errorf("Type: %q", items[0].Type)
	}
}

func TestRDFa_WithPrefixAttribute(t *testing.T) {
	htmlStr := `<html><body prefix="schema: http://schema.org/">
		<div typeof="schema:Book">
			<span property="schema:name">A Book</span>
		</div>
	</body></html>`
	items, errs := ParseRDFa("", htmlStr)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(items) == 0 {
		t.Fatal("expected at least 1 item")
	}
	if items[0].Type != "http://schema.org/Book" {
		t.Errorf("Type: %q", items[0].Type)
	}
	if items[0].Properties["http://schema.org/name"] == nil {
		t.Errorf("name property not found: %v", items[0].Properties)
	}
}

func TestRDFa_WithResourceAndAbout(t *testing.T) {
	htmlStr := `<html><body vocab="http://schema.org/">
		<div typeof="Person" resource="http://example.com/alice" about="http://example.com/about">
			<span property="name">Alice</span>
		</div>
	</body></html>`
	items, errs := ParseRDFa("", htmlStr)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(items) == 0 {
		t.Fatal("expected at least 1 item")
	}
	if items[0].Resource == nil || *items[0].Resource != "http://example.com/alice" {
		t.Errorf("Resource: %v", items[0].Resource)
	}
	if items[0].About == nil || *items[0].About != "http://example.com/about" {
		t.Errorf("About: %v", items[0].About)
	}
}

func TestRDFa_NestedItemsWithPropAndTypeof(t *testing.T) {
	// property + typeof: nested item as property value
	htmlStr := `<html><body vocab="http://schema.org/">
		<div typeof="Article">
			<div property="author" typeof="Person">
				<span property="name">Bob</span>
			</div>
		</div>
	</body></html>`
	items, errs := ParseRDFa("", htmlStr)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(items) == 0 {
		t.Fatal("expected at least 1 item")
	}
	authorProp := items[0].Properties["http://schema.org/author"]
	if authorProp == nil {
		t.Errorf("expected author property, got %v", items[0].Properties)
	}
}

func TestRDFa_NestedTypeofWithoutProp(t *testing.T) {
	// typeof without property → sibling extraItem
	htmlStr := `<html><body vocab="http://schema.org/">
		<div typeof="Article">
			<span property="name">Article Name</span>
			<div typeof="Person">
				<span property="name">Side Person</span>
			</div>
		</div>
	</body></html>`
	items, errs := ParseRDFa("", htmlStr)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	// Should have the Article + the nested Person as extra item
	if len(items) < 2 {
		t.Errorf("expected at least 2 items (article + nested person), got %d", len(items))
	}
}

func TestRDFa_DefaultBranchContinuesWalking(t *testing.T) {
	// A child with no property and no typeof → recurse deeper
	htmlStr := `<html><body vocab="http://schema.org/">
		<div typeof="Article">
			<div>
				<span property="name">Nested Name</span>
			</div>
		</div>
	</body></html>`
	items, errs := ParseRDFa("", htmlStr)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(items) == 0 {
		t.Fatal("expected at least 1 item")
	}
	if items[0].Properties["http://schema.org/name"] == nil {
		t.Errorf("name property not found: %v", items[0].Properties)
	}
}

func TestGetRDFaPropertyValue_AllBranches(t *testing.T) {
	tests := []struct {
		name string
		html string
		want string
	}{
		{
			name: "content attribute",
			html: `<span content="cv">text</span>`,
			want: "cv",
		},
		{
			name: "datetime attribute",
			html: `<time datetime="2023-01-01">January 2023</time>`,
			want: "2023-01-01",
		},
		{
			name: "href attribute",
			html: `<a href="/link">text</a>`,
			want: "/link",
		},
		{
			name: "src attribute",
			html: `<img src="/img.jpg" alt="img">`,
			want: "/img.jpg",
		},
		{
			name: "resource attribute",
			html: `<div resource="/resource">text</div>`,
			want: "/resource",
		},
		{
			name: "fallback to text content",
			html: `<span>plain text</span>`,
			want: "plain text",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, _ := html.Parse(strings.NewReader(tt.html))
			// Walk to find first element node
			var find func(*html.Node) *html.Node
			find = func(n *html.Node) *html.Node {
				if n.Type == html.ElementNode && n.Data != "html" && n.Data != "head" && n.Data != "body" {
					return n
				}
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					if found := find(c); found != nil {
						return found
					}
				}
				return nil
			}
			node := find(doc)
			if node == nil {
				t.Fatal("could not find element node")
			}
			got := getRDFaPropertyValue(node, "")
			if got != tt.want {
				t.Errorf("getRDFaPropertyValue() = %q; want %q", got, tt.want)
			}
		})
	}
}

func TestResolveRDFaTerm(t *testing.T) {
	prefixes := map[string]string{
		"schema": "http://schema.org/",
		"foaf":   "http://xmlns.com/foaf/0.1/",
	}

	tests := []struct {
		name  string
		term  string
		vocab string
		want  string
	}{
		{
			name:  "absolute http URL unchanged",
			term:  "http://example.com/type",
			vocab: "http://schema.org/",
			want:  "http://example.com/type",
		},
		{
			name:  "absolute https URL unchanged",
			term:  "https://example.com/type",
			vocab: "http://schema.org/",
			want:  "https://example.com/type",
		},
		{
			name:  "CURIE with known prefix",
			term:  "schema:Person",
			vocab: "",
			want:  "http://schema.org/Person",
		},
		{
			name:  "CURIE with unknown prefix falls back to vocab",
			term:  "unknown:Thing",
			vocab: "http://schema.org/",
			want:  "http://schema.org/unknown:Thing",
		},
		{
			name:  "no prefix with vocab",
			term:  "Article",
			vocab: "http://schema.org/",
			want:  "http://schema.org/Article",
		},
		{
			name:  "no prefix no vocab returns term as-is",
			term:  "Article",
			vocab: "",
			want:  "Article",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveRDFaTerm(tt.term, tt.vocab, prefixes)
			if got != tt.want {
				t.Errorf("resolveRDFaTerm(%q, %q) = %q; want %q", tt.term, tt.vocab, got, tt.want)
			}
		})
	}
}

func TestParseRDFaPrefixes(t *testing.T) {
	prefixes := make(map[string]string)

	// Normal case
	parseRDFaPrefixes("schema: http://schema.org/ foaf: http://xmlns.com/foaf/0.1/", prefixes)
	if prefixes["schema"] != "http://schema.org/" {
		t.Errorf("schema: %q", prefixes["schema"])
	}
	if prefixes["foaf"] != "http://xmlns.com/foaf/0.1/" {
		t.Errorf("foaf: %q", prefixes["foaf"])
	}

	// Odd-count fields edge case (last field without pair is ignored)
	prefixes2 := make(map[string]string)
	parseRDFaPrefixes("schema: http://schema.org/ orphan", prefixes2)
	if prefixes2["schema"] != "http://schema.org/" {
		t.Errorf("schema: %q", prefixes2["schema"])
	}
	if _, ok := prefixes2["orphan"]; ok {
		t.Error("orphan should not be added (no value)")
	}
}

func TestParseRDFaFrom_ErrReader(t *testing.T) {
	items, errs := parseRDFaFrom("", errReader{})
	if len(errs) == 0 {
		t.Error("expected error from errReader")
	}
	if items != nil {
		t.Errorf("expected nil items on error, got %v", items)
	}
}

func TestRDFa_VocabChangesInsideItem(t *testing.T) {
	// vocab can change mid-walk within collectRDFaProperties
	htmlStr := `<html><body>
		<div typeof="Person" vocab="http://schema.org/">
			<div vocab="http://other.org/">
				<span property="name">Name</span>
			</div>
		</div>
	</body></html>`
	items, errs := ParseRDFa("", htmlStr)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(items) == 0 {
		t.Fatal("expected at least 1 item")
	}
}

func TestRDFa_RelativeURLResolution(t *testing.T) {
	const base = "http://example.com"
	htmlStr := `<html><body vocab="http://schema.org/">
		<div typeof="Article" resource="/article" about="/about">
			<a property="url" href="/page">Page</a>
			<img property="image" src="/img.jpg">
			<div property="sameAs" resource="/same"></div>
		</div>
	</body></html>`

	items, errs := ParseRDFa(base, htmlStr)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(items) == 0 {
		t.Fatal("expected at least 1 item")
	}
	item := items[0]

	if item.Resource == nil || *item.Resource != "http://example.com/article" {
		t.Errorf("Resource: %v", item.Resource)
	}
	if item.About == nil || *item.About != "http://example.com/about" {
		t.Errorf("About: %v", item.About)
	}
	if v, _ := item.Properties["http://schema.org/url"].(string); v != "http://example.com/page" {
		t.Errorf("url property: %q", v)
	}
	if v, _ := item.Properties["http://schema.org/image"].(string); v != "http://example.com/img.jpg" {
		t.Errorf("image property: %q", v)
	}
	if v, _ := item.Properties["http://schema.org/sameAs"].(string); v != "http://example.com/same" {
		t.Errorf("sameAs property: %q", v)
	}
}
