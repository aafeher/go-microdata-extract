package extractor

import "testing"

func TestResolveURL(t *testing.T) {
	tests := []struct {
		name string
		base string
		ref  string
		want string
	}{
		{
			name: "empty ref returns ref",
			base: "http://example.com",
			ref:  "",
			want: "",
		},
		{
			name: "empty base returns ref",
			base: "",
			ref:  "http://example.com/page",
			want: "http://example.com/page",
		},
		{
			name: "path-relative ref",
			base: "http://example.com/a/b",
			ref:  "c",
			want: "http://example.com/a/c",
		},
		{
			name: "root-relative ref",
			base: "http://example.com/a/b",
			ref:  "/c",
			want: "http://example.com/c",
		},
		{
			name: "absolute ref unchanged",
			base: "http://example.com",
			ref:  "http://other.com/page",
			want: "http://other.com/page",
		},
		{
			name: "invalid base URL returns ref",
			base: "http://%zz",
			ref:  "http://example.com/page",
			want: "http://example.com/page",
		},
		{
			name: "invalid ref URL returns ref",
			base: "http://example.com",
			ref:  "%zz",
			want: "%zz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveURL(tt.base, tt.ref)
			if got != tt.want {
				t.Errorf("resolveURL(%q, %q) = %q; want %q", tt.base, tt.ref, got, tt.want)
			}
		})
	}
}
