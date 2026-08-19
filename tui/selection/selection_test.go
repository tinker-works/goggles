package selection

import "testing"

func TestSelectionClampsAfterReplacingItems(t *testing.T) {
	m := NewString("one", "two", "three")
	m.Last()
	m.SetStrings([]string{"only"})
	if m.Index() != 0 {
		t.Fatalf("index = %d, want 0", m.Index())
	}
	item, ok := m.Selected()
	if !ok || item.Label != "only" {
		t.Fatalf("selected item = %#v, %v", item, ok)
	}
}
