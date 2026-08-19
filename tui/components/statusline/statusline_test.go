package statusline

import (
	"strings"
	"testing"
)

func TestStatuslineContainsBothSides(t *testing.T) {
	m := New(20)
	m.Set("left", "right")
	if got := m.View(); got == "" || !strings.Contains(got, "left") || !strings.Contains(got, "right") {
		t.Fatalf("statusline = %q", got)
	}
}
