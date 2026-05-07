package extract

import (
	"os"
	"testing"
)

const benchAllSyntaxesHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Combined all formats benchmark</title>
    <meta property="og:title" content="OpenGraph Article Title"/>
    <meta property="og:type" content="article"/>
    <meta property="og:url" content="https://www.example.com/article"/>
    <meta name="twitter:card" content="summary_large_image"/>
    <meta name="twitter:title" content="X Cards Article Title"/>
    <meta name="DC.title" content="Dublin Core Title"/>
    <meta name="DC.creator" content="Jane Doe"/>
    <meta name="DC.subject" content="Technology"/>
    <script type="application/ld+json">
    {"@context":"https://schema.org/","@type":"Article","name":"JSON-LD Article","author":{"@type":"Person","name":"John Doe"}}
    </script>
</head>
<body>
<div itemscope itemtype="https://schema.org/Person">
    <span itemprop="name">John Doe</span>
    <span itemprop="jobTitle">Engineer</span>
</div>
<div vocab="https://schema.org/" typeof="Person">
    <span property="name">Jane Doe</span>
    <span property="jobTitle">Designer</span>
</div>
<div class="h-card">
    <span class="p-name">Test Person</span>
    <a class="u-url" href="https://example.com">example.com</a>
</div>
</body>
</html>`

func readBenchFixture(b *testing.B, path string) string {
	b.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		b.Fatalf("failed to read fixture %s: %v", path, err)
	}
	return string(data)
}

func BenchmarkExtract_OpenGraph(b *testing.B) {
	content := readBenchFixture(b, "test/test-11-opengraph-article.html")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e := New().SetSyntaxes([]Syntax{SyntaxOpenGraph})
		_, _ = e.Extract("http://example.com", &content)
	}
}

func BenchmarkExtract_XCards(b *testing.B) {
	content := readBenchFixture(b, "test/test-25-xcards-article.html")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e := New().SetSyntaxes([]Syntax{SyntaxXCards})
		_, _ = e.Extract("http://example.com", &content)
	}
}

func BenchmarkExtract_JSONLD(b *testing.B) {
	content := readBenchFixture(b, "test/test-31-ldjson-multiple-objects.html")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e := New().SetSyntaxes([]Syntax{SyntaxJSONLD})
		_, _ = e.Extract("http://example.com", &content)
	}
}

func BenchmarkExtract_Microdata(b *testing.B) {
	content := readBenchFixture(b, "test/test-34-w3cmicrodata-extended.html")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e := New().SetSyntaxes([]Syntax{SyntaxMicrodata})
		_, _ = e.Extract("http://example.com", &content)
	}
}

func BenchmarkExtract_RDFa(b *testing.B) {
	content := readBenchFixture(b, "test/test-42-rdfa-multiple.html")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e := New().SetSyntaxes([]Syntax{SyntaxRDFa})
		_, _ = e.Extract("http://example.com", &content)
	}
}

func BenchmarkExtract_DublinCore(b *testing.B) {
	content := readBenchFixture(b, "test/test-47-dublincore-multiple.html")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e := New().SetSyntaxes([]Syntax{SyntaxDublinCore})
		_, _ = e.Extract("http://example.com", &content)
	}
}

func BenchmarkExtract_Microformats(b *testing.B) {
	content := readBenchFixture(b, "test/test-53-microformats-extended.html")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e := New().SetSyntaxes([]Syntax{SyntaxMicroformats})
		_, _ = e.Extract("http://example.com", &content)
	}
}

func BenchmarkExtract_AllSyntaxes(b *testing.B) {
	content := benchAllSyntaxesHTML
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e := New()
		_, _ = e.Extract("http://example.com", &content)
	}
}
