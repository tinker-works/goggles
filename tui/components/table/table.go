// Package table exposes a themed table primitive without domain row types.
package table

import (
	bubbletable "charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Model wraps bubbles' table model.
type Model struct {
	Table bubbletable.Model
}

// New creates a table from generic string columns and rows.
func New(columns []bubbletable.Column, rows []bubbletable.Row, width, height int) Model {
	opts := []bubbletable.Option{
		bubbletable.WithColumns(columns), bubbletable.WithRows(rows),
		bubbletable.WithWidth(width), bubbletable.WithHeight(height),
	}
	return Model{Table: bubbletable.New(opts...)}
}

// SetRows replaces table rows.
func (m *Model) SetRows(rows []bubbletable.Row) { m.Table.SetRows(rows) }

// SetColumns replaces table columns.
func (m *Model) SetColumns(columns []bubbletable.Column) { m.Table.SetColumns(columns) }

// SetSize changes the table dimensions.
func (m *Model) SetSize(width, height int) {
	m.Table.SetWidth(width)
	m.Table.SetHeight(height)
}

// SetStyles applies table styles.
func (m *Model) SetStyles(styles bubbletable.Styles) { m.Table.SetStyles(styles) }

// SelectedRow returns the currently selected row.
func (m Model) SelectedRow() bubbletable.Row { return m.Table.SelectedRow() }

// Update forwards input to the table.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.Table, cmd = m.Table.Update(msg)
	return m, cmd
}

// View renders the table.
func (m Model) View() string { return m.Table.View() }

// DefaultStyles returns styles with a neutral selected color.
func DefaultStyles() bubbletable.Styles {
	styles := bubbletable.DefaultStyles()
	styles.Selected = lipgloss.NewStyle().Bold(true)
	return styles
}
