package disco

import "testing"

func TestUpdateAdvancesFrame(t *testing.T) {
	m := New("a", "b")
	m, _ = m.Update(TickMsg{})
	if m.View() != "b" {
		t.Fatalf("frame = %q, want b", m.View())
	}
}
