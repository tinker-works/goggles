package keys

import (
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

func TestDefaultNavigation(t *testing.T) {
	bindings := Default()
	if !key.Matches(tea.KeyPressMsg(tea.Key{Text: "j"}), bindings.Down) {
		t.Fatal("down binding does not match j")
	}
	if len(bindings.FullHelp()) == 0 {
		t.Fatal("default key map has no full help")
	}
}
