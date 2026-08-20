// Package theme contains the daemon-independent visual language used by
// goggles. Domain-specific status colors intentionally do not live here.
package theme

import (
	"image/color"
	"strings"
	"sync"

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

const DefaultPaletteName = "default"

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

func paletteForName(name string) (Palette, string) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case DefaultPaletteName:
		return DefaultPalette(), DefaultPaletteName
	case "nord":
		return Palette{
			Background: lipgloss.Color("#2E3440"),
			Foreground: lipgloss.Color("#ECEFF4"),
			Primary:    lipgloss.Color("#88C0D0"),
			Secondary:  lipgloss.Color("#81A1C1"),
			Accent:     lipgloss.Color("#B48EAD"),
			Muted:      lipgloss.Color("#7B88A1"),
			Subtle:     lipgloss.Color("#4C566A"),
			Border:     lipgloss.Color("#434C5E"),
			Success:    lipgloss.Color("#A3BE8C"),
			Warning:    lipgloss.Color("#EBCB8B"),
			Error:      lipgloss.Color("#BF616A"),
		}, "nord"
	case "gruvbox":
		return Palette{
			Background: lipgloss.Color("#282828"),
			Foreground: lipgloss.Color("#EBDBB2"),
			Primary:    lipgloss.Color("#FABD2F"),
			Secondary:  lipgloss.Color("#83A598"),
			Accent:     lipgloss.Color("#D3869B"),
			Muted:      lipgloss.Color("#A89984"),
			Subtle:     lipgloss.Color("#665C54"),
			Border:     lipgloss.Color("#504945"),
			Success:    lipgloss.Color("#B8BB26"),
			Warning:    lipgloss.Color("#FABD2F"),
			Error:      lipgloss.Color("#FB4934"),
		}, "gruvbox"
	case "dracula":
		return Palette{
			Background: lipgloss.Color("#282A36"),
			Foreground: lipgloss.Color("#F8F8F2"),
			Primary:    lipgloss.Color("#BD93F9"),
			Secondary:  lipgloss.Color("#8BE9FD"),
			Accent:     lipgloss.Color("#FF79C6"),
			Muted:      lipgloss.Color("#6272A4"),
			Subtle:     lipgloss.Color("#44475A"),
			Border:     lipgloss.Color("#44475A"),
			Success:    lipgloss.Color("#50FA7B"),
			Warning:    lipgloss.Color("#F1FA8C"),
			Error:      lipgloss.Color("#FF5555"),
		}, "dracula"
	case "solarized":
		return Palette{
			Background: lipgloss.Color("#002B36"),
			Foreground: lipgloss.Color("#839496"),
			Primary:    lipgloss.Color("#268BD2"),
			Secondary:  lipgloss.Color("#2AA198"),
			Accent:     lipgloss.Color("#D33682"),
			Muted:      lipgloss.Color("#839496"),
			Subtle:     lipgloss.Color("#586E75"),
			Border:     lipgloss.Color("#073642"),
			Success:    lipgloss.Color("#859900"),
			Warning:    lipgloss.Color("#B58900"),
			Error:      lipgloss.Color("#DC322F"),
		}, "solarized"
	case "mono":
		return Palette{
			Background: lipgloss.Color("#18181B"),
			Foreground: lipgloss.Color("#FAFAFA"),
			Primary:    lipgloss.Color("#FFFFFF"),
			Secondary:  lipgloss.Color("#D4D4D8"),
			Accent:     lipgloss.Color("#FFFFFF"),
			Muted:      lipgloss.Color("#8A8F98"),
			Subtle:     lipgloss.Color("#52525B"),
			Border:     lipgloss.Color("#3F3F46"),
			Success:    lipgloss.Color("#FAFAFA"),
			Warning:    lipgloss.Color("#D4D4D8"),
			Error:      lipgloss.Color("#FFFFFF"),
		}, "mono"
	default:
		return DefaultPalette(), DefaultPaletteName
	}
}

// Theme is a collection of general rendering styles. It has no knowledge of
// daemon or domain states.
type Theme struct {
	Name    string
	IsDark  bool
	Palette Palette

	Base            lipgloss.Style
	Title           lipgloss.Style
	Subtitle        lipgloss.Style
	Text            lipgloss.Style
	Muted           lipgloss.Style
	Subtle          lipgloss.Style
	Accent          lipgloss.Style
	Border          lipgloss.Style
	Panel           lipgloss.Style
	Selected        lipgloss.Style
	Success         lipgloss.Style
	Warning         lipgloss.Style
	Error           lipgloss.Style
	Code            lipgloss.Style
	Help            lipgloss.Style
	Body            lipgloss.Style
	Tool            lipgloss.Style
	PanelEdge       lipgloss.Style
	PanelEdgeActive lipgloss.Style
	SubtleText      lipgloss.Style
}

// New constructs a theme using the optional palette. The string/bool form is
// accepted for semantic callers that name a palette and terminal background;
// the current goggles palette has no alternate light palette yet.
func New(args ...any) Theme {
	p := DefaultPalette()
	name := DefaultPaletteName
	isDark := true
	if len(args) > 0 {
		switch value := args[0].(type) {
		case Palette:
			p = value
		case string:
			if strings.TrimSpace(value) != "" {
				p, name = paletteForName(value)
			}
		}
	}
	if len(args) > 1 {
		if background, ok := args[1].(bool); ok {
			isDark = background
		}
	}
	return Theme{
		Name:            name,
		IsDark:          isDark,
		Palette:         p,
		Base:            lipgloss.NewStyle().Foreground(p.Foreground).Background(p.Background),
		Title:           lipgloss.NewStyle().Bold(true).Foreground(p.Primary),
		Subtitle:        lipgloss.NewStyle().Foreground(p.Secondary),
		Text:            lipgloss.NewStyle().Foreground(p.Foreground),
		Muted:           lipgloss.NewStyle().Foreground(p.Muted),
		Subtle:          lipgloss.NewStyle().Foreground(p.Subtle),
		Accent:          lipgloss.NewStyle().Bold(true).Foreground(p.Accent),
		Border:          lipgloss.NewStyle().BorderForeground(p.Border),
		Panel:           lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(p.Border).Padding(0, 1),
		PanelEdge:       lipgloss.NewStyle().Foreground(p.Border),
		PanelEdgeActive: lipgloss.NewStyle().Foreground(p.Accent),
		Selected:        lipgloss.NewStyle().Bold(true).Foreground(p.Primary),
		Success:         lipgloss.NewStyle().Foreground(p.Success),
		Warning:         lipgloss.NewStyle().Foreground(p.Warning),
		Error:           lipgloss.NewStyle().Foreground(p.Error),
		Code:            lipgloss.NewStyle().Foreground(p.Secondary),
		Help:            lipgloss.NewStyle().Foreground(p.Muted),
		Body:            lipgloss.NewStyle().Foreground(p.Foreground),
		Tool:            lipgloss.NewStyle().Foreground(p.Secondary),
		SubtleText:      lipgloss.NewStyle().Foreground(p.Subtle),
	}
}

// PaletteInfo is the presentation metadata shown by the settings screen.
type PaletteInfo struct {
	Name        string
	Description string
}

func Names() []string {
	return []string{DefaultPaletteName, "nord", "gruvbox", "dracula", "solarized", "mono"}
}

func Palettes() []PaletteInfo {
	return []PaletteInfo{
		{Name: DefaultPaletteName, Description: "goggles default palette"},
		{Name: "nord", Description: "cool blue contrast"},
		{Name: "gruvbox", Description: "warm earthy contrast"},
		{Name: "dracula", Description: "deep purple contrast"},
		{Name: "solarized", Description: "low-contrast reading"},
		{Name: "mono", Description: "terminal-native monochrome"},
	}
}

// Default returns the default goggles theme.
func Default() Theme { return New() }

// WithPalette returns a copy using palette while preserving the style layout.
func (t Theme) WithPalette(value any) Theme {
	if palette, ok := value.(Palette); ok {
		return New(palette, t.IsDark)
	}
	if name, ok := value.(string); ok {
		return New(name, t.IsDark)
	}
	return New(t.Palette, t.IsDark)
}

// FromPalette builds a theme from a caller-supplied palette. isDark is retained
// as part of the presentation API for callers that also select semantic swatches.
func FromPalette(p Palette, isDark bool) Theme { return New(p, isDark) }

// WithBackground rebuilds the theme while retaining its palette.
func (t Theme) WithBackground(isDark bool) Theme {
	next := New(t.Palette, isDark)
	next.Name = t.Name
	return next
}

// RenderPanel applies the panel style to content at the requested width.
func (t Theme) RenderPanel(content string, width int) string {
	if width > 0 {
		return t.Panel.Width(width).Render(content)
	}
	return t.Panel.Render(content)
}

// RenderTitle renders a heading using the theme title style.
func (t Theme) RenderTitle(title string) string { return t.Title.Render(title) }

func normalize(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "_", "-")
	value = strings.ReplaceAll(value, "changesrequested", "changes-requested")
	return strings.ReplaceAll(value, "hostfailed", "host-failed")
}

// semanticStyle maps a daemon state to a palette role. Unknown states are
// deliberately muted so a newly added daemon state is visible without being
// assigned a misleading meaning.
func (t Theme) semanticStyle(value string) lipgloss.Style {
	switch normalize(value) {
	case "concept", "closed", "queued", "cancelled", "absent", "stopped", "deleting":
		return t.Muted
	case "refine", "running", "open":
		return t.Tool
	case "review", "stalled", "stopping":
		return t.Warning
	case "changes-requested", "stale", "broken", "failed":
		return t.Error
	case "proposed", "coding", "admitted", "creating", "starting":
		return t.Code
	case "ready", "pr", "special":
		return t.Accent
	case "done", "merged", "succeeded":
		return t.Success
	default:
		return t.Muted
	}
}

// EpicStateStyle returns the style for an epic's wire state.
func (t Theme) EpicStateStyle(state string) lipgloss.Style { return t.semanticStyle(state) }

// IssueStateStyle returns the style for an issue's wire state.
func (t Theme) IssueStateStyle(state string) lipgloss.Style { return t.semanticStyle(state) }

// AgentRunStatusStyle returns the style for an agent run's wire status.
func (t Theme) AgentRunStatusStyle(status string) lipgloss.Style {
	if normalize(status) == "host-failed" {
		return t.Warning
	}
	return t.semanticStyle(status)
}

// SandboxStatusStyle returns the style for a sandbox's wire status.
func (t Theme) SandboxStatusStyle(status string) lipgloss.Style {
	if normalize(status) == "ready" {
		return t.Success
	}
	return t.semanticStyle(status)
}

// Badge renders a small colored status marker.
func (t Theme) Badge(style lipgloss.Style, label string) string {
	return style.Render("● " + label)
}

// RunDot returns the status glyph used in the runs rail.
func RunDot(status string) string {
	switch normalize(status) {
	case "queued", "admitted":
		return "○"
	case "succeeded", "done":
		return "✓"
	case "failed", "stalled", "cancelled":
		return "✗"
	case "host-failed":
		return "⚠"
	default:
		return "●"
	}
}

// RunStatusLabel returns the user-facing label for a run status.
func RunStatusLabel(status string) string {
	if normalize(status) == "admitted" {
		return "building"
	}
	return status
}

// RunBadge renders a run's status as a glyph and label.
func (t Theme) RunBadge(status string) string {
	return t.AgentRunStatusStyle(status).Render(RunDot(status) + " " + RunStatusLabel(status))
}

// SandboxDot renders a sandbox status as a single rail glyph.
func (t Theme) SandboxDot(status string) string {
	glyph := "●"
	switch normalize(status) {
	case "absent", "stopped", "deleting":
		glyph = "○"
	case "broken":
		glyph = "✗"
	}
	return t.SandboxStatusStyle(status).Render(glyph)
}

var (
	swatchMu    sync.Mutex
	swatchCache = map[string]string{}
)

// Swatch previews the semantic colors used by an epic-state palette list.
func Swatch(name string, isDark bool) string {
	key := name + "|" + map[bool]string{true: "dark", false: "light"}[isDark]
	swatchMu.Lock()
	defer swatchMu.Unlock()
	if value, ok := swatchCache[key]; ok {
		return value
	}
	th := New(name, isDark)
	states := []string{"Concept", "Refine", "Review", "Ready", "Done", "Failed"}
	var out strings.Builder
	for _, state := range states {
		out.WriteString(th.EpicStateStyle(state).Render("█"))
	}
	swatchCache[key] = out.String()
	return out.String()
}
