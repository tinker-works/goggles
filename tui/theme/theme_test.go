package theme

import "testing"

func TestDefaultThemeHasGeneralStyles(t *testing.T) {
	theme := New()
	if theme.Palette.Primary == nil {
		t.Fatal("primary color is nil")
	}
	if theme.RenderTitle("title") == "" || theme.RenderPanel("body", 20) == "" {
		t.Fatal("general styles rendered empty output")
	}
}

func TestWithPalettePreservesBackgroundMode(t *testing.T) {
	light := New(DefaultPalette(), false)
	if got := light.WithPalette(DefaultPalette()); got.IsDark {
		t.Fatal("custom palette should preserve the light background mode")
	}
}
