package commentbox

import (
	"strings"
	"testing"
)

func TestViewIncludesHeaderAndBody(t *testing.T) {
	m := New()
	m.SetContent("Ada", "today", "hello")
	view := m.View()
	for _, want := range []string{"Ada", "today", "hello"} {
		if !strings.Contains(view, want) {
			t.Fatalf("view %q does not contain %q", view, want)
		}
	}
}
