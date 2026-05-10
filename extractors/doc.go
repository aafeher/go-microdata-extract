// Package extractor implements parsers for seven HTML structured-metadata formats:
// OpenGraph, X Cards, JSON-LD, W3C Microdata, RDFa, Dublin Core, and Microformats2.
// Each format exposes a public Parse* function that accepts a base URL and HTML content
// and returns the extracted items together with any non-fatal parse errors.
package extractor
