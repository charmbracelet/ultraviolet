package uv

import (
	"testing"
)

func TestLinkParser(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want Link
	}{
		// Basic hyperlinks
		{
			name: "simple URL",
			data: []byte("8;;https://example.com"),
			want: Link{
				URL: "https://example.com",
			},
		},
		{
			name: "URL with path",
			data: []byte("8;;https://example.com/path/to/page"),
			want: Link{
				URL: "https://example.com/path/to/page",
			},
		},
		{
			name: "URL with query string",
			data: []byte("8;;https://example.com?foo=bar&baz=qux"),
			want: Link{
				URL: "https://example.com?foo=bar&baz=qux",
			},
		},
		{
			name: "URL with fragment",
			data: []byte("8;;https://example.com/page#section"),
			want: Link{
				URL: "https://example.com/page#section",
			},
		},
		{
			name: "URL with port",
			data: []byte("8;;https://example.com:8080/path"),
			want: Link{
				URL: "https://example.com:8080/path",
			},
		},
		// URLs with parameters
		{
			name: "URL with id parameter",
			data: []byte("8;id=12345;https://example.com"),
			want: Link{
				URL:    "https://example.com",
				Params: "id=12345",
			},
		},
		{
			name: "URL with multiple parameters",
			data: []byte("8;id=12345:key=value;https://example.com"),
			want: Link{
				URL:    "https://example.com",
				Params: "id=12345:key=value",
			},
		},
		{
			name: "URL with parameter containing special chars",
			data: []byte("8;id=foo:bar=baz;https://example.com"),
			want: Link{
				URL:    "https://example.com",
				Params: "id=foo:bar=baz",
			},
		},
		// Different protocols
		{
			name: "http URL",
			data: []byte("8;;http://example.com"),
			want: Link{
				URL: "http://example.com",
			},
		},
		{
			name: "ftp URL",
			data: []byte("8;;ftp://ftp.example.com/file.txt"),
			want: Link{
				URL: "ftp://ftp.example.com/file.txt",
			},
		},
		{
			name: "file URL",
			data: []byte("8;;file:///home/user/document.pdf"),
			want: Link{
				URL: "file:///home/user/document.pdf",
			},
		},
		{
			name: "mailto URL",
			data: []byte("8;;mailto:user@example.com"),
			want: Link{
				URL: "mailto:user@example.com",
			},
		},
		// Edge cases
		{
			name: "empty URL",
			data: []byte("8;;"),
			want: Link{
				URL: "",
			},
		},
		{
			name: "empty URL with parameters",
			data: []byte("8;id=12345;"),
			want: Link{
				URL:    "",
				Params: "id=12345",
			},
		},
		{
			name: "URL with encoded characters",
			data: []byte("8;;https://example.com/path%20with%20spaces"),
			want: Link{
				URL: "https://example.com/path%20with%20spaces",
			},
		},
		// Unsupported commands (not cmd 8)
		{
			name: "command 0 (ignored)",
			data: []byte("0;https://example.com"),
			want: Link{},
		},
		{
			name: "command 1 (ignored)",
			data: []byte("1;;https://example.com"),
			want: Link{},
		},
		{
			name: "command 9 (ignored)",
			data: []byte("9;;https://example.com"),
			want: Link{},
		},
		// Missing command number defaults to 0
		{
			name: "no command number",
			data: []byte(";;https://example.com"),
			want: Link{},
		},
		// Real-world examples
		{
			name: "GitHub URL",
			data: []byte("8;;https://github.com/charmbracelet/bubbletea"),
			want: Link{
				URL: "https://github.com/charmbracelet/bubbletea",
			},
		},
		{
			name: "GitHub PR URL",
			data: []byte("8;;https://github.com/charmbracelet/bubbletea/pull/123"),
			want: Link{
				URL: "https://github.com/charmbracelet/bubbletea/pull/123",
			},
		},
		{
			name: "documentation with anchor",
			data: []byte("8;;https://pkg.go.dev/github.com/charmbracelet/bubbletea#Model"),
			want: Link{
				URL: "https://pkg.go.dev/github.com/charmbracelet/bubbletea#Model",
			},
		},
		{
			name: "local file path",
			data: []byte("8;;file:///Users/username/Documents/file.txt"),
			want: Link{
				URL: "file:///Users/username/Documents/file.txt",
			},
		},
		// Complex parameter examples
		{
			name: "parameter with equals and colons",
			data: []byte("8;foo=bar:baz=qux;https://example.com"),
			want: Link{
				URL:    "https://example.com",
				Params: "foo=bar:baz=qux",
			},
		},
		{
			name: "parameter with empty value",
			data: []byte("8;id=;https://example.com"),
			want: Link{
				URL:    "https://example.com",
				Params: "id=",
			},
		},
		{
			name: "parameter without value",
			data: []byte("8;standalone;https://example.com"),
			want: Link{
				URL:    "https://example.com",
				Params: "standalone",
			},
		},
		// Very long URLs
		{
			name: "long URL",
			data: []byte("8;;https://example.com/very/long/path/with/many/segments/that/goes/on/and/on/and/on/page.html?param1=value1&param2=value2&param3=value3"),
			want: Link{
				URL: "https://example.com/very/long/path/with/many/segments/that/goes/on/and/on/and/on/page.html?param1=value1&param2=value2&param3=value3",
			},
		},
		// URLs with authentication
		{
			name: "URL with username and password",
			data: []byte("8;;https://user:pass@example.com/path"),
			want: Link{
				URL: "https://user:pass@example.com/path",
			},
		},
		{
			name: "URL with username only",
			data: []byte("8;;https://user@example.com/path"),
			want: Link{
				URL: "https://user@example.com/path",
			},
		},
		// Special characters in URL
		{
			name: "URL with parentheses",
			data: []byte("8;;https://en.wikipedia.org/wiki/Terminal_(computing)"),
			want: Link{
				URL: "https://en.wikipedia.org/wiki/Terminal_(computing)",
			},
		},
		{
			name: "URL with brackets",
			data: []byte("8;;https://example.com/path[1]"),
			want: Link{
				URL: "https://example.com/path[1]",
			},
		},
		{
			name: "URL with curly braces",
			data: []byte("8;;https://example.com/path{foo}"),
			want: Link{
				URL: "https://example.com/path{foo}",
			},
		},
		// IPv4 and IPv6
		{
			name: "IPv4 address",
			data: []byte("8;;http://192.168.1.1:8080/path"),
			want: Link{
				URL: "http://192.168.1.1:8080/path",
			},
		},
		{
			name: "IPv6 address",
			data: []byte("8;;http://[2001:db8::1]:8080/path"),
			want: Link{
				URL: "http://[2001:db8::1]:8080/path",
			},
		},
		// Localhost
		{
			name: "localhost",
			data: []byte("8;;http://localhost:3000/api/endpoint"),
			want: Link{
				URL: "http://localhost:3000/api/endpoint",
			},
		},
		// Data URLs
		{
			name: "data URL",
			data: []byte("8;;data:text/plain;base64,SGVsbG8gV29ybGQ="),
			want: Link{
				URL: "data:text/plain;base64,SGVsbG8gV29ybGQ=",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lp := &LinkParser{}
			lp.Reset()
			lp.Advance(tt.data)
			got := lp.Build()

			if got.URL != tt.want.URL {
				t.Errorf("URL = %q, want %q", got.URL, tt.want.URL)
			}
			if got.Params != tt.want.Params {
				t.Errorf("Params = %q, want %q", got.Params, tt.want.Params)
			}
		})
	}
}

func TestLinkParser_Incremental(t *testing.T) {
	tests := []struct {
		name   string
		chunks [][]byte
		want   Link
	}{
		{
			name: "URL in multiple chunks",
			chunks: [][]byte{
				[]byte("8;;https://"),
				[]byte("example.com"),
				[]byte("/path"),
			},
			want: Link{
				URL: "https://example.com/path",
			},
		},
		{
			name: "URL and params in chunks",
			chunks: [][]byte{
				[]byte("8;id="),
				[]byte("12345"),
				[]byte(";https://example.com"),
			},
			want: Link{
				URL:    "https://example.com",
				Params: "id=12345",
			},
		},
		{
			name: "command in chunks",
			chunks: [][]byte{
				[]byte("8"),
				[]byte(";;"),
				[]byte("https://example.com"),
			},
			want: Link{
				URL: "https://example.com",
			},
		},
		{
			name: "byte by byte",
			chunks: [][]byte{
				[]byte("8"),
				[]byte(";"),
				[]byte(";"),
				[]byte("h"),
				[]byte("t"),
				[]byte("t"),
				[]byte("p"),
				[]byte("s"),
				[]byte(":"),
				[]byte("/"),
				[]byte("/"),
				[]byte("e"),
				[]byte("x"),
				[]byte("a"),
				[]byte("m"),
				[]byte("p"),
				[]byte("l"),
				[]byte("e"),
				[]byte("."),
				[]byte("c"),
				[]byte("o"),
				[]byte("m"),
			},
			want: Link{
				URL: "https://example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lp := &LinkParser{}
			lp.Reset()
			for _, chunk := range tt.chunks {
				lp.Advance(chunk)
			}
			got := lp.Build()

			if got.URL != tt.want.URL {
				t.Errorf("URL = %q, want %q", got.URL, tt.want.URL)
			}
			if got.Params != tt.want.Params {
				t.Errorf("Params = %q, want %q", got.Params, tt.want.Params)
			}
		})
	}
}

func TestLinkParser_Apply(t *testing.T) {
	lp := &LinkParser{}
	lp.Reset()
	lp.Advance([]byte("8;id=123;https://example.com"))

	var link Link
	lp.Apply(&link)

	if link.URL != "https://example.com" {
		t.Errorf("URL = %q, want %q", link.URL, "https://example.com")
	}
	if link.Params != "id=123" {
		t.Errorf("Params = %q, want %q", link.Params, "id=123")
	}
}

func TestLinkParser_ApplyNil(t *testing.T) {
	lp := &LinkParser{}
	lp.Reset()
	lp.Advance([]byte("8;;https://example.com"))

	// Should not panic
	lp.Apply(nil)
}

func TestLinkParser_Reset(t *testing.T) {
	lp := &LinkParser{}
	lp.Advance([]byte("8;;https://example.com;id=123"))

	// Reset should clear everything
	lp.Reset()
	got := lp.Build()

	if got.URL != "" {
		t.Errorf("URL after reset = %q, want empty", got.URL)
	}
	if got.Params != "" {
		t.Errorf("Params after reset = %q, want empty", got.Params)
	}

	// After reset, should be able to parse new link
	lp.Advance([]byte("8;;https://new-example.com"))
	got = lp.Build()

	if got.URL != "https://new-example.com" {
		t.Errorf("URL after reset and new parse = %q, want %q", got.URL, "https://new-example.com")
	}
}

func TestLinkParser_MultipleResets(t *testing.T) {
	lp := &LinkParser{}

	// Parse first link
	lp.Reset()
	lp.Advance([]byte("8;;https://first.com"))
	first := lp.Build()

	// Parse second link
	lp.Reset()
	lp.Advance([]byte("8;id=2;https://second.com"))
	second := lp.Build()

	// Parse third link
	lp.Reset()
	lp.Advance([]byte("8;;https://third.com"))
	third := lp.Build()

	if first.URL != "https://first.com" {
		t.Errorf("first URL = %q, want %q", first.URL, "https://first.com")
	}
	if second.URL != "https://second.com" {
		t.Errorf("second URL = %q, want %q", second.URL, "https://second.com")
	}
	if second.Params != "id=2" {
		t.Errorf("second Params = %q, want %q", second.Params, "id=2")
	}
	if third.URL != "https://third.com" {
		t.Errorf("third URL = %q, want %q", third.URL, "https://third.com")
	}
}
