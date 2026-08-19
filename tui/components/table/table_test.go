package table

import (
	"testing"

	bubbletable "charm.land/bubbles/v2/table"
)

func TestTableSelectedRow(t *testing.T) {
	m := New([]bubbletable.Column{{Title: "Name", Width: 8}}, []bubbletable.Row{{"one"}}, 20, 3)
	if got := m.SelectedRow(); len(got) != 1 || got[0] != "one" {
		t.Fatalf("selected row = %v", got)
	}
}
