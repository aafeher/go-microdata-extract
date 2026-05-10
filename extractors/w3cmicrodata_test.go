package extractor

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestW3CMicrodata_Basic(t *testing.T) {
	htmlStr := `<html><body>
		<div itemscope itemtype="http://schema.org/Person">
			<span itemprop="name">Alice</span>
			<span itemprop="age">30</span>
		</div>
	</body></html>`
	items, errs := ParseW3CMicrodata("", htmlStr)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	item := items[0]
	if item.Type != "http://schema.org/Person" {
		t.Errorf("Type: %q", item.Type)
	}
	if item.Properties["name"] != "Alice" {
		t.Errorf("name: %v", item.Properties["name"])
	}
	if item.Properties["age"] != "30" {
		t.Errorf("age: %v", item.Properties["age"])
	}
}

func TestW3CMicrodata_WithItemID(t *testing.T) {
	htmlStr := `<html><body>
		<div itemscope itemtype="http://schema.org/Person" itemid="http://example.com/alice">
			<span itemprop="name">Alice</span>
		</div>
	</body></html>`
	items, errs := ParseW3CMicrodata("", htmlStr)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].ID == nil {
		t.Fatal("expected non-nil ID")
	}
	if *items[0].ID != "http://example.com/alice" {
		t.Errorf("ID: %q", *items[0].ID)
	}
}

func TestW3CMicrodata_NestedItems(t *testing.T) {
	htmlStr := `<html><body>
		<div itemscope itemtype="http://schema.org/Article">
			<span itemprop="name">Article Title</span>
			<div itemprop="author" itemscope itemtype="http://schema.org/Person" itemid="http://example.com/bob">
				<span itemprop="name">Bob</span>
			</div>
		</div>
	</body></html>`
	items, errs := ParseW3CMicrodata("", htmlStr)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	article := items[0]
	if article.Type != "http://schema.org/Article" {
		t.Errorf("Article type: %q", article.Type)
	}
	authorProp := article.Properties["author"]
	if authorProp == nil {
		t.Fatal("expected author property")
	}
	// author should be a *MicrodataItem
	authorItem, ok := authorProp.(*MicrodataItem)
	if !ok {
		t.Fatalf("expected *MicrodataItem, got %T", authorProp)
	}
	if authorItem.Type != "http://schema.org/Person" {
		t.Errorf("Author type: %q", authorItem.Type)
	}
	if authorItem.ID == nil || *authorItem.ID != "http://example.com/bob" {
		t.Errorf("Author ID: %v", authorItem.ID)
	}
	if authorItem.Properties["name"] != "Bob" {
		t.Errorf("Author name: %v", authorItem.Properties["name"])
	}
}

func TestW3CMicrodata_NoItemscope(t *testing.T) {
	htmlStr := `<html><body><div><span>plain text</span></div></body></html>`
	items, errs := ParseW3CMicrodata("", htmlStr)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}

func TestParseProperties_MetaWithContent(t *testing.T) {
	htmlStr := `<html><body>
		<div itemscope>
			<meta itemprop="description" content="Meta content">
		</div>
	</body></html>`
	items, errs := ParseW3CMicrodata("", htmlStr)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Properties["description"] != "Meta content" {
		t.Errorf("description: %v", items[0].Properties["description"])
	}
}

func TestParseProperties_Datetime(t *testing.T) {
	htmlStr := `<html><body>
		<div itemscope>
			<time itemprop="published" datetime="2023-01-15">January 15</time>
		</div>
	</body></html>`
	items, errs := ParseW3CMicrodata("", htmlStr)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Properties["published"] != "2023-01-15" {
		t.Errorf("published: %v", items[0].Properties["published"])
	}
}

func TestParseProperties_URLProp(t *testing.T) {
	// itemprop="url" → resolves href
	htmlStr := `<html><body>
		<div itemscope>
			<a itemprop="url" href="/page">Link</a>
		</div>
	</body></html>`
	items, errs := ParseW3CMicrodata("http://example.com", htmlStr)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Properties["url"] != "http://example.com/page" {
		t.Errorf("url: %v", items[0].Properties["url"])
	}
}

func TestParseProperties_SuffixUrlProp(t *testing.T) {
	// itemprop ending with "Url" → resolves href
	htmlStr := `<html><body>
		<div itemscope>
			<a itemprop="imageUrl" href="/img.jpg">Image</a>
		</div>
	</body></html>`
	items, errs := ParseW3CMicrodata("http://example.com", htmlStr)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Properties["imageUrl"] != "http://example.com/img.jpg" {
		t.Errorf("imageUrl: %v", items[0].Properties["imageUrl"])
	}
}

func TestParseProperties_NonItempropChild(t *testing.T) {
	// Child without itemprop → recurse into it
	htmlStr := `<html><body>
		<div itemscope>
			<div>
				<span itemprop="name">Nested Name</span>
			</div>
		</div>
	</body></html>`
	items, errs := ParseW3CMicrodata("", htmlStr)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Properties["name"] != "Nested Name" {
		t.Errorf("name: %v", items[0].Properties["name"])
	}
}

func makeNode(htmlStr string, tag string) *html.Node {
	doc, _ := html.Parse(strings.NewReader(htmlStr))
	var find func(*html.Node) *html.Node
	find = func(n *html.Node) *html.Node {
		if n.Type == html.ElementNode && n.Data == tag {
			return n
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if found := find(c); found != nil {
				return found
			}
		}
		return nil
	}
	return find(doc)
}

func TestGetAttr(t *testing.T) {
	node := makeNode(`<div itemscope></div>`, "div")
	if node == nil {
		t.Fatal("could not find div")
	}
	if !getAttr(node, "itemscope") {
		t.Error("expected itemscope to be present")
	}
	if getAttr(node, "missing") {
		t.Error("expected missing attribute to be absent")
	}
}

func TestGetAttrVal(t *testing.T) {
	node := makeNode(`<div itemtype="http://schema.org/Thing"></div>`, "div")
	if node == nil {
		t.Fatal("could not find div")
	}
	if v := getAttrVal(node, "itemtype"); v != "http://schema.org/Thing" {
		t.Errorf("itemtype: %q", v)
	}
	if v := getAttrVal(node, "missing"); v != "" {
		t.Errorf("missing attr should return empty, got %q", v)
	}
}

func TestAppendValue(t *testing.T) {
	// nil existing → returns value
	result := appendValue(nil, "first")
	if result != "first" {
		t.Errorf("expected 'first', got %v", result)
	}

	// slice existing → append
	result = appendValue([]any{"a", "b"}, "c")
	s, ok := result.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", result)
	}
	if len(s) != 3 {
		t.Errorf("expected 3 elements, got %d", len(s))
	}
	if s[2] != "c" {
		t.Errorf("expected 'c', got %v", s[2])
	}

	// scalar existing → wrap to slice
	result = appendValue("first", "second")
	s2, ok := result.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", result)
	}
	if len(s2) != 2 {
		t.Errorf("expected 2 elements, got %d", len(s2))
	}
	if s2[0] != "first" || s2[1] != "second" {
		t.Errorf("expected [first, second], got %v", s2)
	}
}

func TestGetTextContent(t *testing.T) {
	t.Run("text node first-text-wins", func(t *testing.T) {
		node := makeNode(`<div>Hello <span>World</span></div>`, "div")
		if node == nil {
			t.Fatal("no div")
		}
		got := getTextContent(node)
		// First-text-wins: first text node is "Hello "
		if got != "Hello" {
			t.Errorf("getTextContent() = %q; want %q", got, "Hello")
		}
	})

	t.Run("element with value attr", func(t *testing.T) {
		// An element with value="" - writes val to sb immediately
		node := makeNode(`<input value="myvalue">`, "input")
		if node == nil {
			t.Fatal("no input")
		}
		got := getTextContent(node)
		if got != "myvalue" {
			t.Errorf("getTextContent() = %q; want %q", got, "myvalue")
		}
	})

	t.Run("nested elements without text", func(t *testing.T) {
		node := makeNode(`<div><span value="child-value"></span></div>`, "div")
		if node == nil {
			t.Fatal("no div")
		}
		got := getTextContent(node)
		// child element has value attr
		if got != "child-value" {
			t.Errorf("getTextContent() = %q; want %q", got, "child-value")
		}
	})
}

func TestParseW3CMicrodataFrom_ErrReader(t *testing.T) {
	items, errs := parseW3CMicrodataFrom("", errReader{})
	if len(errs) == 0 {
		t.Error("expected error from errReader")
	}
	if items != nil {
		t.Errorf("expected nil items on error, got %v", items)
	}
}

func TestW3CMicrodata_MultiplePropsForSamePropName(t *testing.T) {
	// Two elements with same itemprop → slice
	htmlStr := `<html><body>
		<div itemscope>
			<span itemprop="tag">go</span>
			<span itemprop="tag">programming</span>
		</div>
	</body></html>`
	items, errs := ParseW3CMicrodata("", htmlStr)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	tagProp := items[0].Properties["tag"]
	switch v := tagProp.(type) {
	case []any:
		if len(v) != 2 {
			t.Errorf("expected 2 tags, got %d", len(v))
		}
	default:
		t.Errorf("expected []any, got %T: %v", tagProp, tagProp)
	}
}
