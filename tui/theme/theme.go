// Package theme contains the daemon-independent visual language used by
// goggles. Domain-specific status colors intentionally do not live here.
package theme

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Palette contains colors used by general presentation styles.
type Palette struct {
	Background color.Color
	Foreground color.Color
	Primary    color.Color
	Secondary  color.Color
	Accent     color.Color
	Muted      color.Color
	Subtle     color.Color
	Border     color.Color
	Success    color.Color
	Warning    color.Color
	Error      color.Color
}

// DefaultPalette is the standard dark/light palette.
func DefaultPalette() Palette {
	return Palette{
		Background: lipgloss.Color("#111111"),
		Foreground: lipgloss.Color("#F8FAFC"),
		Primary:    lipgloss.Color("#C084FC"),
		Secondary:  lipgloss.Color("#60A5FA"),
		Accent:     lipgloss.Color("#F472B6"),
		Muted:      lipgloss.Color("#94A3B8"),
		Subtle:     lipgloss.Color("#64748B"),
		Border:     lipgloss.Color("#334155"),
		Success:    lipgloss.Color("#4ADE80"),
		Warning:    lipgloss.Color("#FACC15"),
		Error:      lipgloss.Color("#F87171"),
	}
}

// Theme is a collection of general rendering styles. It has no knowledge of
// daemon or domain states.
type Theme struct {
	Palette Palette

	Base     lipgloss.Style
	Title    lipgloss.Style
	Subtitle lipgloss.Style
	Text     lipgloss.Style
	Muted    lipgloss.Style
	Subtle   lipgloss.Style
	Accent   lipgloss.Style
	Border   lipgloss.Style
	Panel    lipgloss.Style
	Selected lipgloss.Style
	Success  lipgloss.Style
	Warning  lipgloss.Style
	Error    lipgloss.Style
	Code     lipgloss.Style
	Help     lipgloss.Style
}

// New constructs a theme using palette, or DefaultPalette when omitted.
func New(palettes ...Palette) Theme {
	p := DefaultPalette()
	if len(palettes) > 0 {
		p = palettes[0]
	}
	return Theme{
		Palette:  p,
		Base:     lipgloss.NewStyle().Foreground(p.Foreground).Background(p.Background),
		Title:    lipgloss.NewStyle().Bold(true).Foreground(p.Primary),
		Subtitle: lipgloss.NewStyle().Foreground(p.Secondary),
		Text:     lipgloss.NewStyle().Foreground(p.Foreground),
		Muted:    lipgloss.NewStyle().Foreground(p.Muted),
		Subtle:   lipgloss.NewStyle().Foreground(p.Subtle),
		Accent:   lipgloss.NewStyle().Bold(true).Foreground(p.Accent),
		Border:   lipgloss.NewStyle().BorderForeground(p.Border),
		Panel:    lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(p.Border).Padding(0, 1),
		Selected: lipgloss.NewStyle().Bold(true).Foreground(p.Primary),
		Success:  lipgloss.NewStyle().Foreground(p.Success),
		Warning:  lipgloss.NewStyle().Foreground(p.Warning),
		Error:    lipgloss.NewStyle().Foreground(p.Error),
		Code:     lipgloss.NewStyle().Foreground(p.Secondary),
		Help:     lipgloss.NewStyle().Foreground(p.Muted),
	}
}

// Default is the default goggles theme.
var Default = New()

// WithPalette returns a copy using palette while preserving the style layout.
func (t Theme) WithPalette(p Palette) Theme { return New(p) }

// RenderPanel applies the panel style to content at the requested width.
func (t Theme) RenderPanel(content string, width int) string {
	if width > 0 {
		return t.Panel.Width(width).Render(content)
	}
	return t.Panel.Render(content)
}

// RenderTitle renders a heading using the theme title style.
func (t Theme) RenderTitle(title string) string { return t.Title.Render(title) }
