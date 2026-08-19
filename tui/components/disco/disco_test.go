package disco

import "testing"

func TestUpdateAdvancesFrame(t *testing.T) {
	m := New("a", "b")
	m, _ = m.Update(TickMsg{})
	if m.View() != "b" {
		t.Fatalf("frame = %q, want b", m.View())
	}
}

func TestViewNormalizesNegativeFrame(t *testing.T) {
	m := Model{Frames: []string{"a", "b", "c"}, Frame: -1}
	if m.View() != "c" {
		t.Fatalf("frame = %q, want c", m.View())
	}
}
