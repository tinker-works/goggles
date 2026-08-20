package board

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/tinker-works/donsy/netomatic"
	"github.com/tinker-works/goggles/tui/components/statusline"
	"github.com/tinker-works/goggles/tui/theme"
	"github.com/tinker-works/goggles/tui/viewmodel"
)

func TestModel_MoveUpWhileZoomed_ShouldStayInTheCurrentColumn(t *testing.T) {
	m := loadedBoard().Zoom().MoveDown().MoveUp().MoveUp()
	if m.Lane != 0 || m.Index != 0 || m.Column != viewmodel.OpenColumn {
		t.Fatalf("unexpected zoomed cursor: lane=%d column=%d index=%d", m.Lane, m.Column, m.Index)
	}
}

func TestModel_ApplyFilterKey_ShouldClearOnEscape(t *testing.T) {
	m := loadedBoard().SetFilter("checkout").StartFilter()
	m = m.ApplyFilterKey(tea.KeyPressMsg{Code: tea.KeyEscape})
	if m.Filter.Value() != "" || len(m.Lanes()) != 1 || !m.Lanes()[0].Matched {
		t.Fatalf("filter was not cleared: value=%q lanes=%+v", m.Filter.Value(), m.Lanes())
	}
}

func TestModel_FilterView_ShouldStayVisibleWhileTheActiveValueIsEmpty(t *testing.T) {
	m := loadedBoard().StartFilter()

	if !m.Filter.Active {
		t.Fatal("expected the filter to be active")
	}
	if footer := m.Footer(theme.Default(), 100); !strings.Contains(footer, "enter apply") {
		t.Fatalf("expected filter controls in the footer: %q", footer)
	}
	if view := m.View(theme.Default(), nil, statusline.Model{}, 100, 20, boardNow); !strings.Contains(view, "/") || !strings.Contains(view, "▏") {
		t.Fatalf("expected the empty filter prompt in the header: %q", view)
	}

	m = m.ApplyFilterKey(tea.KeyPressMsg{Code: tea.KeyBackspace})
	if footer := m.Footer(theme.Default(), 100); !strings.Contains(footer, "filter:") {
		t.Fatalf("expected filter controls after clearing text: %q", footer)
	}
}

func TestModel_ApplyFilterKey_ShouldCloseOnEnterAndKeepTheValue(t *testing.T) {
	m := loadedBoard().StartFilter().SetFilter("checkout")
	m = m.ApplyFilterKey(tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.Filter.Active || m.Filter.Value() != "checkout" {
		t.Fatalf("expected applied filter to close, got active=%t value=%q", m.Filter.Active, m.Filter.Value())
	}
}

func TestModel_ZoomAt_ShouldIgnoreAClosedRowThatIsNotOnTheBoard(t *testing.T) {
	m := loadedBoard().ZoomAt(1)
	if m.Zoomed {
		t.Fatal("expected a click past the live lanes to be ignored")
	}
}

func TestModel_SetEpic_ShouldIgnoreAnUnknownReload(t *testing.T) {
	m := loadedBoard()
	m = m.SetEpic(netomatic.Epic{ID: "missing", Title: "not here"})
	if len(m.Epics) != len(boardEpics()) {
		t.Fatalf("unknown reload changed the board: %+v", m.Epics)
	}
}
