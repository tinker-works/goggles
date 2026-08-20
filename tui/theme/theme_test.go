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

func TestNamedPalettesShouldUseTheirAdvertisedColors(t *testing.T) {
	defaultTheme := New(DefaultPaletteName, true)
	for _, name := range Names()[1:] {
		got := New(name, true)
		if got.Name != name {
			t.Fatalf("expected palette name %q, got %q", name, got.Name)
		}
		if got.Palette.Primary == defaultTheme.Palette.Primary {
			t.Fatalf("palette %q still uses the default primary color", name)
		}
	}
}

func TestWithBackgroundShouldKeepNamedPalette(t *testing.T) {
	got := New("nord", true).WithBackground(false)
	if got.Name != "nord" || got.Palette.Primary != New("nord", true).Palette.Primary {
		t.Fatalf("named palette was not retained: %+v", got)
	}
}
