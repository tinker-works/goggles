package text

import "testing"

func TestWrapPreservesUnicode(t *testing.T) {
	if got := Wrap("αβγδ", 2); got != "αβ\nγδ" {
		t.Fatalf("wrapped text = %q", got)
	}
}

func TestWrapUsesDisplayWidth(t *testing.T) {
	if got := Wrap("界界", 3); got != "界\n界" {
		t.Fatalf("wrapped wide text = %q, want %q", got, "界\n界")
	}
	if got := Wrap("界界", 1); got != "界\n界" {
		t.Fatalf("wrapped narrow text = %q, want %q", got, "界\n界")
	}
}
