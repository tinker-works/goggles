package scroll

import "testing"

func TestScrollContent(t *testing.T) {
	m := New(20, 2)
	m.SetContent("one\ntwo")
	if m.View() == "" {
		t.Fatal("scroll view is empty")
	}
}
