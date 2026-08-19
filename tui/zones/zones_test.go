package zones

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestManagerHitAndPosition(t *testing.T) {
	m := New()
	m.Set(Zone{ID: "save", StartX: 2, StartY: 3, EndX: 5, EndY: 4})
	msg := tea.MouseClickMsg{X: 4, Y: 3, Button: tea.MouseLeft}
	zone, ok := m.Hit(msg)
	if !ok || zone.ID != "save" {
		t.Fatalf("hit = %#v, %v", zone, ok)
	}
	x, y := zone.Pos(msg)
	if x != 2 || y != 0 {
		t.Fatalf("position = (%d, %d), want (2, 0)", x, y)
	}
}
