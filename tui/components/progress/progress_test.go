package progress

import "testing"

func TestProgressClampsValue(t *testing.T) {
	m := New(10)
	m.SetPercent(2)
	if m.Percent() != 1 {
		t.Fatalf("percent = %v, want 1", m.Percent())
	}
}
