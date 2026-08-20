// Package statusline renders a one-line left/right status bar.
package statusline

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/tinker-works/goggles/tui/theme"
)

// Model stores statusline content.
type Model struct {
	Left       string
	Right      string
	Width      int
	LeftStyle  lipgloss.Style
	RightStyle lipgloss.Style
	Branch     string
	Dirty      bool
	Offline    bool
	Loading    string
	SyncedAt   time.Time
}

// New creates a statusline of width.
func New(width int) Model { return Model{Width: width} }

// Set updates the two sides.
func (m *Model) Set(left, right string) { m.Left, m.Right = left, right }

// View renders both sides without allowing the right side to overlap the left.
func (m Model) View() string {
	left, right := m.LeftStyle.Render(m.Left), m.RightStyle.Render(m.Right)
	if m.Width <= 0 {
		return strings.TrimSpace(left + " " + right)
	}
	available := m.Width - lipgloss.Width(left) - lipgloss.Width(right)
	if available < 0 {
		return lipgloss.NewStyle().Width(m.Width).MaxWidth(m.Width).Render(left)
	}
	return left + strings.Repeat(" ", available) + right
}

// Sync records a successful remote refresh.
func (m Model) Sync(branch string, dirty bool, at time.Time) Model {
	m.Branch, m.Dirty, m.Offline, m.Loading, m.SyncedAt = branch, dirty, false, "", at
	return m
}

// Ready records a completed local read without changing the sync age.
func (m Model) Ready(branch string, dirty bool) Model {
	m.Branch, m.Dirty, m.Offline, m.Loading = branch, dirty, false, ""
	return m
}

// Fail marks the last synchronized data as offline.
func (m Model) Fail() Model {
	m.Offline, m.Loading = true, ""
	return m
}

func (m Model) TitleSuffix() string {
	if m.Offline {
		return " · ⚠ offline"
	}
	return ""
}

// Render produces the themed screen-header status used by moved screens.
func Render(th theme.Theme, m Model, now time.Time) string {
	branch := m.Branch
	if branch == "" {
		branch = "main"
	}
	if m.Offline {
		return th.Error.Render("⚠ offline") + th.Muted.Render("  showing "+Age(m.SyncedAt, now)+" data")
	}
	mark, style := "✓", th.Success
	if m.Dirty {
		mark, style = "±", th.Warning
	}
	line := style.Render(branch + " " + mark)
	if m.Loading != "" {
		return line + th.Muted.Render("  "+m.Loading)
	}
	if m.SyncedAt.IsZero() {
		return line
	}
	return line + th.Muted.Render("  synced "+Age(m.SyncedAt, now)+" ago")
}

func Age(at, now time.Time) string {
	if at.IsZero() {
		return "never"
	}
	elapsed := now.Sub(at)
	if elapsed < 0 {
		elapsed = 0
	}
	switch {
	case elapsed < time.Minute:
		return fmt.Sprintf("%ds", int(elapsed.Seconds()))
	case elapsed < time.Hour:
		return fmt.Sprintf("%dm%02ds", int(elapsed.Minutes()), int(elapsed.Seconds())%60)
	default:
		return fmt.Sprintf("%dh%02dm", int(elapsed.Hours()), int(elapsed.Minutes())%60)
	}
}

func Elapsed(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	if d < time.Hour {
		return fmt.Sprintf("%02d:%02d", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh%02dm", int(d.Hours()), int(d.Minutes())%60)
}
