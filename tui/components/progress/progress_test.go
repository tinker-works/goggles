package progress

import (
	"math"
	"testing"
)

func TestProgressClampsValue(t *testing.T) {
	m := New(10)
	m.SetPercent(2)
	if m.Percent() != 1 {
		t.Fatalf("percent = %v, want 1", m.Percent())
	}
}

func TestProgressNormalizesNonFiniteValues(t *testing.T) {
	tests := []struct {
		value float64
		want  float64
	}{
		{value: math.NaN(), want: 0},
		{value: math.Inf(-1), want: 0},
		{value: math.Inf(1), want: 1},
	}

	for _, tt := range tests {
		m := New(10)
		m.SetPercent(tt.value)
		if got := m.Percent(); got != tt.want {
			t.Errorf("Percent(%v) = %v, want %v", tt.value, got, tt.want)
		}
	}

	m := New(10)
	m.Value = math.NaN()
	if got := m.Percent(); got != 0 {
		t.Fatalf("Percent(NaN) = %v, want 0", got)
	}
}
