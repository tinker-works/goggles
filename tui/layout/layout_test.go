package layout

import "testing"

func TestColumnsUseAllWidth(t *testing.T) {
	columns := Columns(10, 3)
	if len(columns) != 3 || columns[0]+columns[1]+columns[2] != 10 {
		t.Fatalf("columns = %v", columns)
	}
}
