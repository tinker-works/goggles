package sparkline

import (
	"math"
	"testing"
)

func TestSparklineRendersOneGlyphPerValue(t *testing.T) {
	m := New(4)
	m.SetValues([]float64{1, 2, 3})
	if got := m.View(); len([]rune(got)) != 3 {
		t.Fatalf("sparkline = %q", got)
	}
}

func TestSparklineIgnoresNonFiniteValues(t *testing.T) {
	m := New(4)
	m.SetValues([]float64{math.Inf(-1), 1, math.NaN(), 3, math.Inf(1)})

	if got := m.View(); got != "▁█" {
		t.Fatalf("sparkline = %q, want %q", got, "▁█")
	}
}

func TestSparklineHandlesFiniteExtremeValues(t *testing.T) {
	m := New(4)
	m.SetValues([]float64{-math.MaxFloat64, math.MaxFloat64})

	if got := m.View(); got != "▁█" {
		t.Fatalf("sparkline = %q, want %q", got, "▁█")
	}
}

func TestSparklineRendersFlatValues(t *testing.T) {
	m := New(4)
	m.SetValues([]float64{0, 0})

	if got := m.View(); got != "▁▁" {
		t.Fatalf("sparkline = %q, want %q", got, "▁▁")
	}
}

func TestSparklineRendersEmptyForOnlyNonFiniteValues(t *testing.T) {
	m := New(4)
	m.Values = []float64{math.Inf(-1), math.NaN(), math.Inf(1)}

	if got := m.View(); got != "" {
		t.Fatalf("sparkline = %q, want empty output", got)
	}
}
