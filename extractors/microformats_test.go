package extractor

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestMicroformats_Basic_HCard(t *testing.T) {
	htmlStr := `<html><body>
		<div class="h-card">
			<span class="p-name">Alice Smith</span>
			<a class="u-url" href="http://example.com/alice">Profile</a>
		</div>
	</body></html>`
	items, errs := Microformats("", htmlStr)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(items) == 0 {
		t.Fatal("expected at least 1 item")
	}
	item := items[0]
	if len(item.Type) == 0 || item.Type[0] != "h-card" {
		t.Errorf("Type: %v", item.Type)
	}
	if item.Properties["p-name"] != "Alice Smith" {
		t.Errorf("p-name: %v", item.Properties["p-name"])
	}
	if item.Properties["u-url"] != "http://example.com/alice" {
		t.Errorf("u-url: %v", item.Properties["u-url"])
	}
}

func TestMicroformats_HEntry_AllPropertyTypes(t *testing.T) {
	htmlStr := `<html><body>
		<article class="h-entry">
			<h1 class="p-name">My Post</h1>
			<a class="u-url" href="/post/1">Permalink</a>
			<time class="dt-published" datetime="2023-01-15">January 15</time>
			<div class="e-content"><p>Hello world</p></div>
		</article>
	</body></html>`
	items, errs := Microformats("", htmlStr)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(items) == 0 {
		t.Fatal("expected at least 1 item")
	}
	item := items[0]
	if item.Properties["p-name"] != "My Post" {
		t.Errorf("p-name: %v", item.Properties["p-name"])
	}
	if item.Properties["u-url"] != "/post/1" {
		t.Errorf("u-url: %v", item.Properties["u-url"])
	}
	if item.Properties["dt-published"] != "2023-01-15" {
		t.Errorf("dt-published: %v", item.Properties["dt-published"])
	}
	if item.Properties["e-content"] == nil {
		t.Errorf("e-content: %v", item.Properties["e-content"])
	}
}

func TestMicroformats_NestedRootAsPropValue(t *testing.T) {
	// h-* that also has p-* → nested item acting as property value (isRoot && isProp)
	htmlStr := `<html><body>
		<div class="h-entry">
			<div class="h-card p-author">
				<span class="p-name">Bob</span>
			</div>
		</div>
	</body></html>`
	items, errs := Microformats("", htmlStr)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(items) == 0 {
		t.Fatal("expected at least 1 item")
	}
	// p-author should be a nested MicroformatItem
	pAuthor := items[0].Properties["p-author"]
	if pAuthor == nil {
		t.Errorf("p-author property not found: %v", items[0].Properties)
	}
}

func TestMicroformats_NestedRootAsChild(t *testing.T) {
	// h-* without property classes → child item
	htmlStr := `<html><body>
		<div class="h-feed">
			<div class="h-entry">
				<span class="p-name">Entry 1</span>
			</div>
		</div>
	</body></html>`
	items, errs := Microformats("", htmlStr)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(items) == 0 {
		t.Fatal("expected at least 1 item")
	}
	feed := items[0]
	if len(feed.Children) == 0 {
		t.Errorf("expected children, got %v", feed.Children)
	}
	if feed.Children[0].Type[0] != "h-entry" {
		t.Errorf("Child type: %v", feed.Children[0].Type)
	}
}

func TestMicroformats_NestedRootAsPropWithMultiplePropTypes(t *testing.T) {
	// h-card with u-*, dt-*, e-* classes as well (all prop types added as nested item)
	htmlStr := `<html><body>
		<div class="h-entry">
			<div class="h-card u-author dt-author e-author">
				<span class="p-name">Carol</span>
			</div>
		</div>
	</body></html>`
	items, errs := Microformats("", htmlStr)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(items) == 0 {
		t.Fatal("expected 1 item")
	}
	// u-author, dt-author, e-author all should have the nested item
	if items[0].Properties["u-author"] == nil {
		t.Errorf("u-author not set: %v", items[0].Properties)
	}
	if items[0].Properties["dt-author"] == nil {
		t.Errorf("dt-author not set: %v", items[0].Properties)
	}
	if items[0].Properties["e-author"] == nil {
		t.Errorf("e-author not set: %v", items[0].Properties)
	}
}

func TestMicroformats_MultipleValues(t *testing.T) {
	// Same property name twice → slice
	htmlStr := `<html><body>
		<div class="h-card">
			<a class="u-url" href="/1">Link 1</a>
			<a class="u-url" href="/2">Link 2</a>
		</div>
	</body></html>`
	items, errs := Microformats("", htmlStr)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(items) == 0 {
		t.Fatal("expected 1 item")
	}
	switch v := items[0].Properties["u-url"].(type) {
	case []any:
		if len(v) != 2 {
			t.Errorf("expected 2 url values, got %d", len(v))
		}
	default:
		t.Errorf("expected []any, got %T: %v", v, v)
	}
}

func findFirstElement(doc *html.Node, tag string) *html.Node {
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

func TestMfTextValue_AllBranches(t *testing.T) {
	tests := []struct {
		name string
		html string
		tag  string
		want string
	}{
		{
			name: "abbr with title",
			html: `<abbr title="Abbreviation">Abbr</abbr>`,
			tag:  "abbr",
			want: "Abbreviation",
		},
		{
			name: "link with title",
			html: `<link title="Link Title">`,
			tag:  "link",
			want: "Link Title",
		},
		{
			name: "data with value",
			html: `<data value="data-value">text</data>`,
			tag:  "data",
			want: "data-value",
		},
		{
			name: "input with value",
			html: `<input value="input-value">`,
			tag:  "input",
			want: "input-value",
		},
		{
			name: "img with alt",
			html: `<img alt="image alt" src="/img.jpg">`,
			tag:  "img",
			want: "image alt",
		},
		{
			name: "area with alt",
			html: `<area alt="area alt" href="/link">`,
			tag:  "area",
			want: "area alt",
		},
		{
			name: "fallback inner text",
			html: `<span>plain text</span>`,
			tag:  "span",
			want: "plain text",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, _ := html.Parse(strings.NewReader(tt.html))
			node := findFirstElement(doc, tt.tag)
			if node == nil {
				t.Fatalf("could not find <%s> element", tt.tag)
			}
			got := mfTextValue(node)
			if got != tt.want {
				t.Errorf("mfTextValue() = %q; want %q", got, tt.want)
			}
		})
	}
}

func TestMfURLValue_AllBranches(t *testing.T) {
	tests := []struct {
		name string
		html string
		tag  string
		want string
	}{
		{
			name: "a with href",
			html: `<a href="/link">text</a>`,
			tag:  "a",
			want: "/link",
		},
		{
			name: "area with href",
			html: `<area href="/area">`,
			tag:  "area",
			want: "/area",
		},
		{
			name: "link with href",
			html: `<link href="/link">`,
			tag:  "link",
			want: "/link",
		},
		{
			name: "img with src",
			html: `<img src="/img.jpg" alt="img">`,
			tag:  "img",
			want: "/img.jpg",
		},
		{
			name: "video with src",
			html: `<video src="/vid.mp4"></video>`,
			tag:  "video",
			want: "/vid.mp4",
		},
		{
			name: "audio with src",
			html: `<audio src="/audio.mp3"></audio>`,
			tag:  "audio",
			want: "/audio.mp3",
		},
		{
			name: "source with src",
			html: `<source src="/source.mp4">`,
			tag:  "source",
			want: "/source.mp4",
		},
		{
			name: "iframe with src",
			html: `<iframe src="/frame.html"></iframe>`,
			tag:  "iframe",
			want: "/frame.html",
		},
		{
			name: "embed with src",
			html: `<embed src="/embed.swf">`,
			tag:  "embed",
			want: "/embed.swf",
		},
		{
			name: "object with data",
			html: `<object data="/obj.pdf"></object>`,
			tag:  "object",
			want: "/obj.pdf",
		},
		{
			name: "form with action",
			html: `<form action="/submit"></form>`,
			tag:  "form",
			want: "/submit",
		},
		{
			name: "fallback inner text",
			html: `<span>http://example.com</span>`,
			tag:  "span",
			want: "http://example.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, _ := html.Parse(strings.NewReader(tt.html))
			node := findFirstElement(doc, tt.tag)
			if node == nil {
				t.Fatalf("could not find <%s> element", tt.tag)
			}
			got := mfURLValue(node, "")
			if got != tt.want {
				t.Errorf("mfURLValue() = %q; want %q", got, tt.want)
			}
		})
	}
}

func TestMfDatetimeValue_AllBranches(t *testing.T) {
	tests := []struct {
		name string
		html string
		tag  string
		want string
	}{
		{
			name: "time with datetime",
			html: `<time datetime="2023-01-15">January</time>`,
			tag:  "time",
			want: "2023-01-15",
		},
		{
			name: "ins with datetime",
			html: `<ins datetime="2023-02-01">inserted</ins>`,
			tag:  "ins",
			want: "2023-02-01",
		},
		{
			name: "del with datetime",
			html: `<del datetime="2023-03-01">deleted</del>`,
			tag:  "del",
			want: "2023-03-01",
		},
		{
			name: "abbr with title",
			html: `<abbr title="2023-04-01">April</abbr>`,
			tag:  "abbr",
			want: "2023-04-01",
		},
		{
			name: "data with value",
			html: `<data value="2023-05-01">May</data>`,
			tag:  "data",
			want: "2023-05-01",
		},
		{
			name: "input with value",
			html: `<input value="2023-06-01">`,
			tag:  "input",
			want: "2023-06-01",
		},
		{
			name: "fallback inner text",
			html: `<span>2023-07-01</span>`,
			tag:  "span",
			want: "2023-07-01",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, _ := html.Parse(strings.NewReader(tt.html))
			node := findFirstElement(doc, tt.tag)
			if node == nil {
				t.Fatalf("could not find <%s> element", tt.tag)
			}
			got := mfDatetimeValue(node)
			if got != tt.want {
				t.Errorf("mfDatetimeValue() = %q; want %q", got, tt.want)
			}
		})
	}
}

func TestMfEmbeddedValue(t *testing.T) {
	htmlStr := `<div class="e-content"><p>Hello <strong>world</strong></p></div>`
	doc, _ := html.Parse(strings.NewReader(htmlStr))
	node := findFirstElement(doc, "div")
	if node == nil {
		t.Fatal("could not find div element")
	}
	got := mfEmbeddedValue(node)
	if got == "" {
		t.Error("expected non-empty embedded value")
	}
	if !strings.Contains(got, "Hello") {
		t.Errorf("expected inner HTML to contain 'Hello', got %q", got)
	}
}

func TestMfInnerText(t *testing.T) {
	htmlStr := `<div>Hello <span>World</span></div>`
	doc, _ := html.Parse(strings.NewReader(htmlStr))
	node := findFirstElement(doc, "div")
	if node == nil {
		t.Fatal("could not find div element")
	}
	got := mfInnerText(node)
	if got != "Hello World" {
		t.Errorf("mfInnerText() = %q; want %q", got, "Hello World")
	}
}

func TestParseMicroformatsFrom_ErrReader(t *testing.T) {
	items, errs := parseMicroformatsFrom("", errReader{})
	if len(errs) == 0 {
		t.Error("expected error from errReader")
	}
	if items != nil {
		t.Errorf("expected nil items on error, got %v", items)
	}
}

func TestMicroformats_NoMfClasses(t *testing.T) {
	// No h-* classes → empty
	htmlStr := `<html><body><div class="not-mf"><span>text</span></div></body></html>`
	items, errs := Microformats("", htmlStr)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}

func TestMfClasses_EmptyClass(t *testing.T) {
	htmlStr := `<div></div>`
	doc, _ := html.Parse(strings.NewReader(htmlStr))
	node := findFirstElement(doc, "div")
	if node == nil {
		t.Fatal("could not find div element")
	}
	classes := mfClasses(node)
	if classes != nil {
		t.Errorf("expected nil for no class attr, got %v", classes)
	}
}

func TestFilterPrefix(t *testing.T) {
	tokens := []string{"h-card", "p-name", "u-url", "h-entry", "other"}
	hClasses := filterPrefix(tokens, "h-")
	if len(hClasses) != 2 {
		t.Errorf("expected 2 h- classes, got %v", hClasses)
	}
	pClasses := filterPrefix(tokens, "p-")
	if len(pClasses) != 1 {
		t.Errorf("expected 1 p- class, got %v", pClasses)
	}
	none := filterPrefix(tokens, "dt-")
	if len(none) != 0 {
		t.Errorf("expected 0 dt- classes, got %v", none)
	}
}

func TestMicroformats_RelativeURLResolution(t *testing.T) {
	const base = "http://example.com"
	htmlStr := `<html><body>
		<div class="h-card">
			<a class="u-url" href="/profile">Profile</a>
			<img class="u-photo" src="/photo.jpg">
		</div>
	</body></html>`

	items, errs := Microformats(base, htmlStr)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(items) == 0 {
		t.Fatal("expected at least 1 item")
	}
	item := items[0]

	if v, _ := item.Properties["u-url"].(string); v != "http://example.com/profile" {
		t.Errorf("u-url: %q", v)
	}
	if v, _ := item.Properties["u-photo"].(string); v != "http://example.com/photo.jpg" {
		t.Errorf("u-photo: %q", v)
	}
}
