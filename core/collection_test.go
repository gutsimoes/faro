package core

import (
	"math"
	"testing"
)

func newTestCollection(t *testing.T) *Collection {
	t.Helper()
	c, err := NewCollection(2)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestNewCollectionRejectsBadDimension(t *testing.T) {
	for _, dim := range []int{0, -1} {
		if _, err := NewCollection(dim); err == nil {
			t.Errorf("NewCollection(%d) = nil error, want error", dim)
		}
	}
}

func TestInsertRejectsInvalidInput(t *testing.T) {
	c := newTestCollection(t)
	if err := c.Insert("a", []float32{1, 2}); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		id   string
		vec  []float32
	}{
		{"empty id", "", []float32{1, 2}},
		{"duplicate id", "a", []float32{3, 4}},
		{"wrong dimension", "b", []float32{1, 2, 3}},
		{"NaN", "c", []float32{float32(math.NaN()), 0}},
		{"Inf", "d", []float32{float32(math.Inf(1)), 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := c.Insert(tt.id, tt.vec); err == nil {
				t.Error("want error, got nil")
			}
		})
	}
	if c.Len() != 1 {
		t.Errorf("Len() = %d after rejected inserts, want 1", c.Len())
	}
}

func TestInsertCopiesVector(t *testing.T) {
	c := newTestCollection(t)
	vec := []float32{1, 1}
	if err := c.Insert("a", vec); err != nil {
		t.Fatal(err)
	}
	vec[0] = 100 // changing the caller's slice must not change the stored vector

	got, err := c.Search([]float32{1, 1}, 1)
	if err != nil {
		t.Fatal(err)
	}
	if got[0].Distance != 0 {
		t.Errorf("distance = %v, want 0", got[0].Distance)
	}
}

func TestSearchOrdersByDistance(t *testing.T) {
	c := newTestCollection(t)
	for id, vec := range map[string][]float32{
		"origin": {0, 0},
		"near":   {1, 1},
		"middle": {3, 4},
		"far":    {10, 10},
	} {
		if err := c.Insert(id, vec); err != nil {
			t.Fatal(err)
		}
	}

	got, err := c.Search([]float32{0, 0}, 3)
	if err != nil {
		t.Fatal(err)
	}
	wantIDs := []string{"origin", "near", "middle"}
	if len(got) != len(wantIDs) {
		t.Fatalf("got %d results, want %d", len(got), len(wantIDs))
	}
	for i, id := range wantIDs {
		if got[i].ID != id {
			t.Errorf("result %d = %q, want %q", i, got[i].ID, id)
		}
	}
	if got[2].Distance != 5 {
		t.Errorf("distance to middle = %v, want 5", got[2].Distance)
	}
}

func TestSearchBreaksTiesByID(t *testing.T) {
	c := newTestCollection(t)
	for _, id := range []string{"b", "c", "a"} {
		if err := c.Insert(id, []float32{1, 0}); err != nil {
			t.Fatal(err)
		}
	}
	got, err := c.Search([]float32{0, 0}, 3)
	if err != nil {
		t.Fatal(err)
	}
	for i, id := range []string{"a", "b", "c"} {
		if got[i].ID != id {
			t.Errorf("result %d = %q, want %q", i, got[i].ID, id)
		}
	}
}

func TestSearchEdgeCases(t *testing.T) {
	c := newTestCollection(t)

	got, err := c.Search([]float32{0, 0}, 5)
	if err != nil || len(got) != 0 {
		t.Errorf("empty collection: got %v, %v; want no results and no error", got, err)
	}

	if err := c.Insert("a", []float32{1, 1}); err != nil {
		t.Fatal(err)
	}
	got, err = c.Search([]float32{0, 0}, 10)
	if err != nil || len(got) != 1 {
		t.Errorf("k larger than Len: got %v, %v; want 1 result", got, err)
	}

	if _, err := c.Search([]float32{0, 0}, 0); err == nil {
		t.Error("k=0: want error")
	}
	if _, err := c.Search([]float32{0, 0, 0}, 1); err == nil {
		t.Error("wrong query dimension: want error")
	}
}

func TestL2(t *testing.T) {
	if got := L2([]float32{0, 0}, []float32{3, 4}); got != 5 {
		t.Errorf("L2 = %v, want 5", got)
	}
	if got := L2([]float32{1, 2, 3}, []float32{1, 2, 3}); got != 0 {
		t.Errorf("L2 of equal vectors = %v, want 0", got)
	}
}
