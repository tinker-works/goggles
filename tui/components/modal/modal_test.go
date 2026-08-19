package modal

import "testing"

func TestModalVisibility(t *testing.T) {
	m := New("title", "body")
	if m.View() != "" {
		t.Fatal("new modal is visible")
	}
	m.Show("body")
	if m.View() == "" {
		t.Fatal("shown modal is empty")
	}
	m.Hide()
	if m.View() != "" {
		t.Fatal("hidden modal rendered content")
	}
}
