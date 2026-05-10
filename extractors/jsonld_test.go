package extractor

import (
	"testing"
)

func TestJSONLD_PublicWrapper(t *testing.T) {
	html := `<html><head><script type="application/ld+json">{"@type":"Thing","name":"test"}</script></head></html>`
	items, errs := ParseJSONLD("http://example.com", html)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0]["name"] != "test" {
		t.Errorf("expected name=test, got %v", items[0]["name"])
	}
}

func TestJSONLD_RelativeURLResolution(t *testing.T) {
	const base = "http://example.com"

	t.Run("resolves @id and url in top-level object", func(t *testing.T) {
		html := `<script type="application/ld+json">{"@id":"/page","url":"/page","name":"Test"}</script>`
		items, errs := ParseJSONLD(base, html)
		if len(errs) != 0 {
			t.Fatalf("unexpected errors: %v", errs)
		}
		if items[0]["@id"] != "http://example.com/page" {
			t.Errorf("@id: %q", items[0]["@id"])
		}
		if items[0]["url"] != "http://example.com/page" {
			t.Errorf("url: %q", items[0]["url"])
		}
		if items[0]["name"] != "Test" {
			t.Errorf("name should be unchanged: %q", items[0]["name"])
		}
	})

	t.Run("resolves @id in nested object", func(t *testing.T) {
		html := `<script type="application/ld+json">{"@type":"Article","author":{"@id":"/author","name":"Alice"}}</script>`
		items, errs := ParseJSONLD(base, html)
		if len(errs) != 0 {
			t.Fatalf("unexpected errors: %v", errs)
		}
		author, _ := items[0]["author"].(map[string]any)
		if author == nil {
			t.Fatal("author not found")
		}
		if author["@id"] != "http://example.com/author" {
			t.Errorf("nested @id: %q", author["@id"])
		}
	})

	t.Run("resolves @id in array of objects", func(t *testing.T) {
		html := `<script type="application/ld+json">[{"@id":"/a"},{"@id":"/b"}]</script>`
		items, errs := ParseJSONLD(base, html)
		if len(errs) != 0 {
			t.Fatalf("unexpected errors: %v", errs)
		}
		if items[0]["@id"] != "http://example.com/a" {
			t.Errorf("@id[0]: %q", items[0]["@id"])
		}
		if items[1]["@id"] != "http://example.com/b" {
			t.Errorf("@id[1]: %q", items[1]["@id"])
		}
	})

	t.Run("absolute URLs are unchanged", func(t *testing.T) {
		html := `<script type="application/ld+json">{"@id":"https://other.com/page","url":"https://other.com/page"}</script>`
		items, errs := ParseJSONLD(base, html)
		if len(errs) != 0 {
			t.Fatalf("unexpected errors: %v", errs)
		}
		if items[0]["@id"] != "https://other.com/page" {
			t.Errorf("@id should be unchanged: %q", items[0]["@id"])
		}
	})

	t.Run("resolves @id inside array-valued field", func(t *testing.T) {
		html := `<script type="application/ld+json">{"@type":"Collection","hasPart":[{"@id":"/part1"},{"@id":"/part2"}]}</script>`
		items, errs := ParseJSONLD(base, html)
		if len(errs) != 0 {
			t.Fatalf("unexpected errors: %v", errs)
		}
		parts, _ := items[0]["hasPart"].([]any)
		if len(parts) != 2 {
			t.Fatalf("expected 2 parts, got %d", len(parts))
		}
		p0, _ := parts[0].(map[string]any)
		if p0["@id"] != "http://example.com/part1" {
			t.Errorf("hasPart[0][@id]: %q", p0["@id"])
		}
		p1, _ := parts[1].(map[string]any)
		if p1["@id"] != "http://example.com/part2" {
			t.Errorf("hasPart[1][@id]: %q", p1["@id"])
		}
	})

	t.Run("empty base URL skips resolution", func(t *testing.T) {
		html := `<script type="application/ld+json">{"@id":"/page"}</script>`
		items, errs := ParseJSONLD("", html)
		if len(errs) != 0 {
			t.Fatalf("unexpected errors: %v", errs)
		}
		if items[0]["@id"] != "/page" {
			t.Errorf("@id should be unchanged when no base: %q", items[0]["@id"])
		}
	})
}

func TestExtractParseJSONLD(t *testing.T) {
	t.Run("no script tags returns nil", func(t *testing.T) {
		items, errs := extractJSONLD(`<html><head></head></html>`)
		if len(errs) != 0 {
			t.Errorf("unexpected errors: %v", errs)
		}
		if items != nil {
			t.Errorf("expected nil items, got %v", items)
		}
	})

	t.Run("single JSON object", func(t *testing.T) {
		html := `<script type="application/ld+json">{"@type":"Person","name":"Alice"}</script>`
		items, errs := extractJSONLD(html)
		if len(errs) != 0 {
			t.Errorf("unexpected errors: %v", errs)
		}
		if len(items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(items))
		}
		if items[0]["name"] != "Alice" {
			t.Errorf("expected name=Alice, got %v", items[0]["name"])
		}
	})

	t.Run("JSON array with multiple items", func(t *testing.T) {
		html := `<script type="application/ld+json">[{"@type":"A"},{"@type":"B"}]</script>`
		items, errs := extractJSONLD(html)
		if len(errs) != 0 {
			t.Errorf("unexpected errors: %v", errs)
		}
		if len(items) != 2 {
			t.Fatalf("expected 2 items, got %d", len(items))
		}
	})

	t.Run("invalid object JSON returns error", func(t *testing.T) {
		html := `<script type="application/ld+json">{invalid json}</script>`
		items, errs := extractJSONLD(html)
		if len(errs) == 0 {
			t.Error("expected an error for invalid JSON object")
		}
		if len(items) != 0 {
			t.Errorf("expected no items, got %v", items)
		}
	})

	t.Run("invalid array JSON returns error", func(t *testing.T) {
		html := `<script type="application/ld+json">[invalid json]</script>`
		items, errs := extractJSONLD(html)
		if len(errs) == 0 {
			t.Error("expected an error for invalid JSON array")
		}
		if len(items) != 0 {
			t.Errorf("expected no items, got %v", items)
		}
	})

	t.Run("empty script content is skipped", func(t *testing.T) {
		html := `<script type="application/ld+json">   </script>`
		items, errs := extractJSONLD(html)
		if len(errs) != 0 {
			t.Errorf("unexpected errors: %v", errs)
		}
		if items != nil {
			t.Errorf("expected nil items, got %v", items)
		}
	})

	t.Run("content starting with neither [ nor { is skipped silently", func(t *testing.T) {
		html := `<script type="application/ld+json">"just a string"</script>`
		items, errs := extractJSONLD(html)
		if len(errs) != 0 {
			t.Errorf("unexpected errors: %v", errs)
		}
		if items != nil {
			t.Errorf("expected nil items, got %v", items)
		}
	})

	t.Run("multiple scripts mixed valid and invalid", func(t *testing.T) {
		html := `<script type="application/ld+json">{"@type":"A"}</script>` +
			`<script type="application/ld+json">{bad}</script>` +
			`<script type="application/ld+json">[{"@type":"B"}]</script>`
		items, errs := extractJSONLD(html)
		if len(errs) != 1 {
			t.Errorf("expected 1 error, got %d", len(errs))
		}
		if len(items) != 2 {
			t.Errorf("expected 2 items, got %d", len(items))
		}
	})
}
