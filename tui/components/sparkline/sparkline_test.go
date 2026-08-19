package sparkline

import "testing"

func TestSparklineRendersOneGlyphPerValue(t *testing.T) {
	m := New(4)
	m.SetValues([]float64{1, 2, 3})
	if got := m.View(); len([]rune(got)) != 3 {
		t.Fatalf("sparkline = %q", got)
	}
}
