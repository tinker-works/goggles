package panel

import "testing"

func TestPanelView(t *testing.T) {
	m := New("heading")
	m.SetContent("body")
	if got := m.View(); got == "" {
		t.Fatal("panel rendered empty output")
	}
}
