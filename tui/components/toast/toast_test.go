package toast

import "testing"

func TestToastVisibility(t *testing.T) {
	m := New()
	m.Show("saved", Success)
	if m.View() == "" {
		t.Fatal("visible toast is empty")
	}
	m.Hide()
	if m.View() != "" {
		t.Fatal("hidden toast rendered content")
	}
}
