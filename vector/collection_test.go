package vector

import (
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"testing"
)

func TestL2(t *testing.T) {
	for _, tc := range []struct {
		name string
		a, b []float64
		want float64
	}{
		{"three-four-five", []float64{0, 0}, []float64{3, 4}, 5},
		{"identical", []float64{2, -3}, []float64{2, -3}, 0},
		{"negative", []float64{-3, -4}, []float64{0, 0}, 5},
		{"large-finite", []float64{0}, []float64{1e200}, 1e200},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := L2(tc.a, tc.b)
			if err != nil || math.Abs(got-tc.want) > 1e-10 {
				t.Fatalf("L2 = %g, %v; want %g", got, err, tc.want)
			}
		})
	}
}

func TestL2RejectsInvalidVectors(t *testing.T) {
	for _, pair := range [][2][]float64{
		{nil, nil},
		{{1}, {1, 2}},
		{{math.NaN()}, {0}},
		{{0}, {math.Inf(1)}},
		{{math.MaxFloat64}, {-math.MaxFloat64}},
	} {
		if _, err := L2(pair[0], pair[1]); err == nil {
			t.Errorf("expected error for %v", pair)
		}
	}
}

func TestSearch(t *testing.T) {
	c, err := NewCollection(2)
	if err != nil {
		t.Fatal(err)
	}
	for id, values := range map[string][]float64{
		"far": {3, 4}, "b": {1, 0}, "a": {-1, 0}, "origin": {0, 0},
	} {
		if err := c.Insert(id, values); err != nil {
			t.Fatal(err)
		}
	}
	want := []Result{{"origin", 0}, {"a", 1}, {"b", 1}, {"far", 5}}
	for _, k := range []int{1, 3, 4, 10} {
		got, err := c.Search([]float64{0, 0}, k)
		end := min(k, len(want))
		if err != nil || !reflect.DeepEqual(got, want[:end]) {
			t.Errorf("Search(k=%d) = %v, %v; want %v", k, got, err, want[:end])
		}
	}
}

func TestEmptyCollection(t *testing.T) {
	c, _ := NewCollection(2)
	got, err := c.Search([]float64{0, 0}, 3)
	if err != nil || len(got) != 0 {
		t.Fatalf("empty search = %v, %v", got, err)
	}
}

func TestValidationAndOwnership(t *testing.T) {
	for _, dimensions := range []int{0, -1} {
		if _, err := NewCollection(dimensions); err == nil {
			t.Fatal("expected invalid dimensions error")
		}
	}
	c, _ := NewCollection(2)
	values := []float64{3, 4}
	if err := c.Insert("point", values); err != nil {
		t.Fatal(err)
	}
	values[0] = 100
	for _, tc := range []struct {
		id     string
		values []float64
	}{
		{"", []float64{0, 0}},
		{"wrong", []float64{0}},
		{"nan", []float64{math.NaN(), 0}},
		{"inf", []float64{0, math.Inf(-1)}},
		{"point", []float64{0, 0}},
	} {
		if err := c.Insert(tc.id, tc.values); err == nil {
			t.Errorf("expected insert error for %q", tc.id)
		}
	}
	got, err := c.Search([]float64{0, 0}, 1)
	if err != nil || len(got) != 1 || got[0].Distance != 5 || c.Len() != 1 {
		t.Fatalf("insert changed stored data: %v, %v, len=%d", got, err, c.Len())
	}
	for _, query := range [][]float64{nil, {1}, {math.NaN(), 0}, {0, math.Inf(1)}} {
		if _, err := c.Search(query, 1); err == nil {
			t.Errorf("expected invalid query error for %v", query)
		}
	}
	for _, k := range []int{0, -1} {
		if _, err := c.Search([]float64{0, 0}, k); err == nil {
			t.Errorf("expected invalid k error for %d", k)
		}
	}
}

// BenchmarkSearch excludes data generation and insertion from the timed region.
func BenchmarkSearch(b *testing.B) {
	for _, n := range []int{1000, 10000} {
		for _, dimensions := range []int{16, 128} {
			b.Run(fmt.Sprintf("n=%d/d=%d", n, dimensions), func(b *testing.B) {
				c, _ := NewCollection(dimensions)
				rng := rand.New(rand.NewSource(42))
				for i := 0; i < n; i++ {
					values := make([]float64, dimensions)
					for j := range values {
						values[j] = rng.Float64()
					}
					if err := c.Insert(fmt.Sprint(i), values); err != nil {
						b.Fatal(err)
					}
				}
				query := make([]float64, dimensions)
				for i := range query {
					query[i] = rng.Float64()
				}
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					if _, err := c.Search(query, 10); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}
