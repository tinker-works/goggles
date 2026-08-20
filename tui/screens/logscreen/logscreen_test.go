package logscreen

import (
	"errors"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/tinker-works/goggles/tui/components/statusline"
	"github.com/tinker-works/goggles/tui/theme"
)

func TestModel_Scroll_ShouldParkAndReattachAtTheEnd(t *testing.T) {
	m := New().Append([]string{"one", "two", "three", "four", "five"}, 1, nil)
	parked := m.Scroll(-2, 5, 3)
	reattached := parked.Scroll(2, 5, 3)
	if parked.Following() || parked.Top() != 0 || !reattached.Following() {
		t.Fatalf("unexpected scroll state: parked=%+v reattached=%+v", parked, reattached)
	}
}

func TestModel_View_ShouldRenderEmptyAndErrorStates(t *testing.T) {
	empty := New().View(theme.Default(), statusline.Model{}, 80, 10)
	errView := New().Append(nil, 0, errors.New("permission denied")).View(theme.Default(), statusline.Model{}, 80, 10)
	if !strings.Contains(empty, "No log entries yet.") || !strings.Contains(errView, "permission denied") {
		t.Fatalf("unexpected log states: empty=%q error=%q", empty, errView)
	}
	if lipgloss.Width(empty) != 80 || lipgloss.Height(empty) != 10 {
		t.Fatalf("expected fixed-size panel, got %dx%d", lipgloss.Width(empty), lipgloss.Height(empty))
	}
}

func TestModel_Reload_ShouldClearRowsAndResetScroll(t *testing.T) {
	m := New().Append([]string{"one", "two", "three"}, 12, nil).Scroll(-1, 3, 2)
	reloaded := m.Reload()
	if len(reloaded.Lines) != 0 || reloaded.Offset != 0 || reloaded.Top() != 0 || !reloaded.Following() {
		t.Fatalf("expected reset log, got %+v", reloaded)
	}
}

func TestModel_Append_ShouldKeepBoundedDaemonPages(t *testing.T) {
	lines := make([]string, maxLogLines+1)
	lines[0], lines[maxLogLines] = "old", "new"
	m := New().Append(lines, 1, nil)
	if len(m.Lines) != maxLogLines || m.Lines[0] != lines[1] || m.Lines[maxLogLines-1] != "new" {
		t.Fatalf("unexpected bounded log: first=%q last=%q count=%d", m.Lines[0], m.Lines[maxLogLines-1], len(m.Lines))
	}
}
