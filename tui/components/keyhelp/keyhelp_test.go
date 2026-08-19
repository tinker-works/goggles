package keyhelp

import (
	"testing"

	"github.com/tinker-works/goggles/tui/keys"
)

func TestViewUsesKeyMap(t *testing.T) {
	m := New()
	m.SetWidth(80)
	if got := m.View(keys.Default()); got == "" {
		t.Fatal("help view is empty")
	}
}
