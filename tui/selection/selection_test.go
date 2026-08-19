package selection

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestNewSelectionAcceptsKeypresses(t *testing.T) {
	m := NewString("one", "two")
	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Text: "j"}))

	if m.Index() != 1 {
		t.Fatalf("selection index after keypress = %d, want 1", m.Index())
	}
}

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
