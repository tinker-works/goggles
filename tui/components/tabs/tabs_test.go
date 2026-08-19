package tabs

import "testing"

func TestTabsClampSelection(t *testing.T) {
	m := New("one", "two")
	m.SetActive(4)
	if m.ActiveIndex() != 1 {
		t.Fatalf("active index = %d", m.ActiveIndex())
	}
}
