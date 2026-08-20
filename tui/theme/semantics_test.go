package theme

import "testing"

func TestSemanticTheme_ShouldRenderPublicStates(t *testing.T) {
	theme := Default()
	for _, state := range []string{"Concept", "changes_requested", "ready", "unknown"} {
		if rendered := theme.EpicStateStyle(state).Render(state); rendered == "" {
			t.Fatalf("state %q rendered empty", state)
		}
	}
}

func TestRunDot_ShouldKeepTerminalStatesLegible(t *testing.T) {
	if RunDot("queued") != "○" || RunDot("succeeded") != "✓" || RunDot("failed") != "✗" {
		t.Fatal("unexpected run status glyphs")
	}
	if RunStatusLabel("admitted") != "building" {
		t.Fatal("admitted status should explain sandbox setup")
	}
}

func TestSwatch_ShouldBeStable(t *testing.T) {
	first := Swatch("default", true)
	second := Swatch("default", true)
	if first == "" || first != second {
		t.Fatalf("swatch was not stable: %q then %q", first, second)
	}
}
