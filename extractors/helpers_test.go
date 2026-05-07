package extractor

import "fmt"

// errReader is an io.Reader that always returns an error.
// Used to trigger the html.Parse error branch in parseDublinCoreFrom,
// parseRDFaFrom, parseMicroformatsFrom, and parseW3CMicrodataFrom.
type errReader struct{}

func (errReader) Read(_ []byte) (int, error) { return 0, fmt.Errorf("read error") }

// pointerOfString returns a pointer to the given string value.
func pointerOfString(s string) *string {
	return &s
}
