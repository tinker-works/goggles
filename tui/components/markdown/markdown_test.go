package markdown

import (
	"strings"
	"testing"

	xansi "github.com/charmbracelet/x/ansi"
)

func TestRenderWrapsLongLines(t *testing.T) {
	m := New("one two three four")
	m.SetWidth(14)
	got := m.Render()
	plain := xansi.Strip(got)
	if !strings.Contains(plain, "one two") || !strings.Contains(plain, "three four") {
		t.Fatalf("render = %q", plain)
	}
	if strings.Contains(plain, "one two three four") {
		t.Fatalf("render did not wrap: %q", plain)
	}
}

func TestRenderFormatsMarkdown(t *testing.T) {
	m := New("# Title\n\nThis is **bold** and *italic*.\n\n[link](https://example.com)\n\n```go\nfmt.Println(\"hello\")\n```")
	got := m.Render()
	plain := xansi.Strip(got)
	if !strings.Contains(got, "\x1b[") {
		t.Fatalf("render is not ANSI-styled: %q", got)
	}
	for _, source := range []string{"# Title", "**bold**", "*italic*", "[link](https"} {
		if strings.Contains(plain, source) {
			t.Fatalf("render contains unrendered Markdown %q: %q", source, plain)
		}
	}
	for _, content := range []string{"Title", "bold", "italic", "link", "fmt.Println", "hello"} {
		if !strings.Contains(plain, content) {
			t.Fatalf("render does not contain %q: %q", content, plain)
		}
	}
}
