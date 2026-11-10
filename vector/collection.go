// Package vector implements an in-memory collection with exact nearest-neighbor search.
package vector

import (
	"fmt"
	"math"
	"sort"
)

// Collection stores vectors of a fixed dimension. It is not safe for concurrent use.
type Collection struct {
	dimensions int
	vectors    map[string][]float64
}

// Result identifies a neighbor and its Euclidean distance from the query.
type Result struct {
	ID       string
	Distance float64
}

func NewCollection(dimensions int) (*Collection, error) {
	if dimensions <= 0 {
		return nil, fmt.Errorf("dimensions must be positive")
	}
	return &Collection{dimensions: dimensions, vectors: make(map[string][]float64)}, nil
}

func (c *Collection) Len() int { return len(c.vectors) }

// Insert copies the vector. Duplicate IDs are rejected without changing existing data.
func (c *Collection) Insert(id string, values []float64) error {
	if id == "" {
		return fmt.Errorf("ID must not be empty")
	}
	if err := c.validate(values); err != nil {
		return err
	}
	if _, exists := c.vectors[id]; exists {
		return fmt.Errorf("ID %q already exists", id)
	}
	c.vectors[id] = append([]float64(nil), values...)
	return nil
}

// Search returns up to k neighbors, ordered by distance then ID for stable ties.
// It scans every stored vector; an empty collection returns an empty result.
func (c *Collection) Search(query []float64, k int) ([]Result, error) {
	if k <= 0 {
		return nil, fmt.Errorf("k must be positive")
	}
	if err := c.validate(query); err != nil {
		return nil, err
	}
	results := make([]Result, 0, len(c.vectors))
	for id, values := range c.vectors {
		distance, err := L2(query, values)
		if err != nil {
			return nil, fmt.Errorf("distance to %q: %w", id, err)
		}
		results = append(results, Result{ID: id, Distance: distance})
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Distance == results[j].Distance {
			return results[i].ID < results[j].ID
		}
		return results[i].Distance < results[j].Distance
	})
	if k < len(results) {
		results = results[:k]
	}
	return results, nil
}

func (c *Collection) validate(values []float64) error {
	if len(values) != c.dimensions {
		return fmt.Errorf("expected %d dimensions, got %d", c.dimensions, len(values))
	}
	return finite(values)
}

func finite(values []float64) error {
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return fmt.Errorf("coordinates must be finite")
		}
	}
	return nil
}

// L2 computes Euclidean distance between nonempty vectors of equal dimension.
func L2(a, b []float64) (float64, error) {
	if len(a) == 0 || len(a) != len(b) {
		return 0, fmt.Errorf("vectors must have the same positive dimension")
	}
	if err := finite(a); err != nil {
		return 0, err
	}
	if err := finite(b); err != nil {
		return 0, err
	}
	var distance float64
	for i := range a {
		distance = math.Hypot(distance, a[i]-b[i])
	}
	if math.IsInf(distance, 0) {
		return 0, fmt.Errorf("distance exceeds float64 range")
	}
	return distance, nil
}
