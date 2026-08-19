package zones

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestManagerHitAndPosition(t *testing.T) {
	m := New()
	defer m.Close()
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

func TestManagerMarkScanAndHit(t *testing.T) {
	m := New()
	defer m.Close()

	marked := m.Mark("save", "save")
	if marked == "save" {
		t.Fatal("Mark did not add a zone marker")
	}
	if got := m.Scan("top\n" + marked); got != "top\nsave" {
		t.Fatalf("Scan = %q, want %q", got, "top\nsave")
	}

	zone, ok := m.Get("save")
	if !ok {
		t.Fatal("Get did not find the scanned zone")
	}
	if zone.StartX != 0 || zone.StartY != 1 || zone.EndX != 3 || zone.EndY != 1 {
		t.Fatalf("zone = %#v, want coordinates (0, 1)-(3, 1)", zone)
	}

	msg := tea.MouseClickMsg{X: 2, Y: 1, Button: tea.MouseLeft}
	hit, ok := m.Hit(msg)
	if !ok || hit.ID != "save" {
		t.Fatalf("hit = %#v, %v", hit, ok)
	}
	if x, y := hit.Pos(msg); x != 2 || y != 0 {
		t.Fatalf("position = (%d, %d), want (2, 0)", x, y)
	}
}

func TestManagerScanReplacesRenderedZones(t *testing.T) {
	m := New()
	defer m.Close()

	m.Scan(m.Mark("save", "save"))
	if _, ok := m.Get("save"); !ok {
		t.Fatal("initial scan did not record the zone")
	}

	if got := m.Scan("plain"); got != "plain" {
		t.Fatalf("Scan = %q, want %q", got, "plain")
	}
	if _, ok := m.Get("save"); ok {
		t.Fatal("zone from the previous render was not removed")
	}
}
