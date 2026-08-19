package text

import "testing"

func TestWrapPreservesUnicode(t *testing.T) {
	if got := Wrap("αβγδ", 2); got != "αβ\nγδ" {
		t.Fatalf("wrapped text = %q", got)
	}
}
