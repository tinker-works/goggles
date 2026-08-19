package filter

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestMatchesIsCaseInsensitive(t *testing.T) {
	m := New("search")
	m.SetValue("tea")
	if !m.Matches("Bubble Tea") || m.Matches("goggles") {
		t.Fatal("filter matching is incorrect")
	}
}

func TestNewFilterAcceptsKeypresses(t *testing.T) {
	m := New("search")
	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Text: "g"}))

	if m.Value() != "g" {
		t.Fatalf("filter value after keypress = %q, want %q", m.Value(), "g")
	}
}
