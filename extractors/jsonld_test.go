package extractor

import (
	"testing"
)

func TestJSONLD_PublicWrapper(t *testing.T) {
	html := `<html><head><script type="application/ld+json">{"@type":"Thing","name":"test"}</script></head></html>`
	items, errs := JSONLD("http://example.com", html)
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

func TestExtractJSONLD(t *testing.T) {
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
