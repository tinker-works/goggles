package markdown

import "testing"

func TestRenderWrapsLongLines(t *testing.T) {
	m := New("abcdefgh")
	m.SetWidth(3)
	if got := m.Render(); got != "abc\ndef\ngh" {
		t.Fatalf("render = %q", got)
	}
}
