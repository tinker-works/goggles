package loader

import "testing"

func TestHiddenLoaderRendersNothing(t *testing.T) {
	m := New("loading")
	m.Visible = false
	if m.View() != "" {
		t.Fatalf("hidden loader = %q", m.View())
	}
}
