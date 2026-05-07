package extractor

import (
	"testing"
)

func TestDublinCore_EmptyHTML(t *testing.T) {
	item, errs := DublinCore("", `<html><head></head></html>`)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if item != nil {
		t.Errorf("expected nil item, got %+v", item)
	}
}

func TestDublinCore_WithMetaTags(t *testing.T) {
	html := `<html><head>
		<meta name="DC.title" content="My Title">
		<meta name="DC.creator" content="Alice">
		<meta name="DCterms.subject" content="Science">
	</head></html>`
	item, errs := DublinCore("", html)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if item == nil {
		t.Fatal("expected non-nil item")
	}
	if item.Properties["dc.title"] != "My Title" {
		t.Errorf("dc.title: %v", item.Properties["dc.title"])
	}
	if item.Properties["dc.creator"] != "Alice" {
		t.Errorf("dc.creator: %v", item.Properties["dc.creator"])
	}
	if item.Properties["dcterms.subject"] != "Science" {
		t.Errorf("dcterms.subject: %v", item.Properties["dcterms.subject"])
	}
}

func TestDublinCore_WithLinkTags(t *testing.T) {
	html := `<html><head>
		<link rel="DC.identifier" href="/resource/123">
	</head></html>`
	item, errs := DublinCore("http://example.com", html)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if item == nil {
		t.Fatal("expected non-nil item")
	}
	val, ok := item.Properties["dc.identifier"]
	if !ok {
		t.Fatal("dc.identifier not found")
	}
	if val != "http://example.com/resource/123" {
		t.Errorf("dc.identifier: %v", val)
	}
}

func TestDublinCore_MultipleValuesForSameProp(t *testing.T) {
	html := `<html><head>
		<meta name="DC.subject" content="Topic A">
		<meta name="DC.subject" content="Topic B">
	</head></html>`
	item, errs := DublinCore("", html)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if item == nil {
		t.Fatal("expected non-nil item")
	}
	val := item.Properties["dc.subject"]
	switch v := val.(type) {
	case []any:
		if len(v) != 2 {
			t.Errorf("expected 2 subjects, got %d", len(v))
		}
	default:
		t.Errorf("expected []any, got %T: %v", val, val)
	}
}

func TestDublinCore_MetaWithNoContent(t *testing.T) {
	// meta tag without content attribute should be ignored
	html := `<html><head>
		<meta name="DC.title">
	</head></html>`
	item, errs := DublinCore("", html)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if item != nil {
		t.Errorf("expected nil item (no content), got %+v", item)
	}
}

func TestDublinCore_LinkWithNoHref(t *testing.T) {
	// link tag without href attribute should be ignored
	html := `<html><head>
		<link rel="DC.identifier">
	</head></html>`
	item, errs := DublinCore("", html)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if item != nil {
		t.Errorf("expected nil item (no href), got %+v", item)
	}
}

func TestDcProperty(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"dc.title", "dc.title"},
		{"dc.creator", "dc.creator"},
		{"dcterms.subject", "dcterms.subject"},
		{"dcterms.description", "dcterms.description"},
		{"unknown", ""},
		{"", ""},
		{"other.prop", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dcProperty(tt.name)
			if got != tt.want {
				t.Errorf("dcProperty(%q) = %q; want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestParseDublinCoreFrom_ErrReader(t *testing.T) {
	item, errs := parseDublinCoreFrom("http://example.com", errReader{})
	if len(errs) == 0 {
		t.Error("expected error from errReader")
	}
	if item != nil {
		t.Errorf("expected nil item on error, got %+v", item)
	}
}
