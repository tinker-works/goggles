// Package statusline renders a one-line left/right status bar.
package statusline

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
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
	if m.Left == "" && m.Right == "" && (m.Branch != "" || m.Offline || m.Loading != "" || !m.SyncedAt.IsZero()) {
		branch := m.Branch
		if branch == "" {
			branch = "main"
		}
		if m.Offline {
			return "offline  showing " + Age(m.SyncedAt, time.Now()) + " data"
		}
		line := branch + " "
		if m.Dirty {
			line += "±"
		} else {
			line += "✓"
		}
		if m.Loading != "" {
			return line + "  " + m.Loading
		}
		if !m.SyncedAt.IsZero() {
			return line + "  synced " + Age(m.SyncedAt, time.Now()) + " ago"
		}
		return line
	}
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

func (m Model) Sync(branch string, dirty bool, at time.Time) Model {
	m = m.Ready(branch, dirty)
	m.SyncedAt = at
	return m
}

func (m Model) Ready(branch string, dirty bool) Model {
	m.Branch, m.Dirty, m.Offline, m.Loading = branch, dirty, false, ""
	return m
}

func (m Model) Online() Model {
	m.Offline = false
	return m
}

func (m Model) Fail() Model {
	m.Offline, m.Loading = true, ""
	return m
}

func (m Model) Load(what string) Model {
	m.Loading = what
	return m
}

func (m Model) IsLoading() bool { return m.Loading != "" }

// Age formats elapsed time for screen rows.
func Age(at, now time.Time) string {
	if at.IsZero() {
		return "never"
	}
	elapsed := maxDuration(0, now.Sub(at))
	switch {
	case elapsed < time.Minute:
		return fmt.Sprintf("%ds", int(elapsed.Seconds()))
	case elapsed < time.Hour:
		return fmt.Sprintf("%dm%02ds", int(elapsed.Minutes()), int(elapsed.Seconds())%60)
	default:
		return fmt.Sprintf("%dh%02dm", int(elapsed.Hours()), int(elapsed.Minutes())%60)
	}
}

// Elapsed formats a run duration for compact rail rows.
func Elapsed(duration time.Duration) string {
	if duration < 0 {
		duration = 0
	}
	if duration < time.Hour {
		return fmt.Sprintf("%02d:%02d", int(duration.Minutes()), int(duration.Seconds())%60)
	}
	return fmt.Sprintf("%dh%02dm", int(duration.Hours()), int(duration.Minutes())%60)
}

// TitleSuffix provides a compact offline marker for panel titles.
func (m Model) TitleSuffix() string {
	if m.Offline {
		return " · offline"
	}
	return ""
}

func maxDuration(minimum time.Duration, value time.Duration) time.Duration {
	if value < minimum {
		return minimum
	}
	return value
}
