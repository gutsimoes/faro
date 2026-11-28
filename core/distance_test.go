package core

import (
	"math"
	"testing"
)

func TestCosineDistance(t *testing.T) {
	for _, tc := range []struct {
		name string
		a, b []float32
		want float32
	}{
		{"same direction", []float32{1, 2}, []float32{2, 4}, 0},
		{"perpendicular", []float32{1, 0}, []float32{0, 1}, 1},
		{"opposite directions", []float32{1, 2}, []float32{-1, -2}, 2},
		{"diagonal", []float32{1, 0}, []float32{1, 1}, float32(1 - math.Sqrt(0.5))},
		{"zero first", []float32{0, 0}, []float32{1, 2}, 1},
		{"zero second", []float32{1, 2}, []float32{0, 0}, 1},
		{"both zero", []float32{0, 0}, []float32{0, 0}, 1},
		{"extreme magnitudes", []float32{math.MaxFloat32}, []float32{math.SmallestNonzeroFloat32}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := CosineDistance(tc.a, tc.b)
			if math.IsNaN(float64(got)) || math.Abs(float64(got-tc.want)) > 1e-6 {
				t.Errorf("CosineDistance(%v, %v) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestDotProductDistance(t *testing.T) {
	for _, tc := range []struct {
		name string
		a, b []float32
		want float32
	}{
		{"positive dot product", []float32{1, 2}, []float32{3, 4}, -11},
		{"negative dot product", []float32{1, 2}, []float32{-3, -4}, 11},
		{"zero dot product", []float32{0, 0}, []float32{1, 2}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := DotProductDistance(tc.a, tc.b); got != tc.want {
				t.Errorf("DotProductDistance(%v, %v) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
		})
	}
}
