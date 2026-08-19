package filter

import "testing"

func TestMatchesIsCaseInsensitive(t *testing.T) {
	m := New("search")
	m.SetValue("tea")
	if !m.Matches("Bubble Tea") || m.Matches("goggles") {
		t.Fatal("filter matching is incorrect")
	}
}
