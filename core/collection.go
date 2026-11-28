package core

import (
	"fmt"
	"math"
	"slices"
	"sort"
	"sync"
)

// Collection stores vectors of a fixed dimension in one contiguous block
// Vector i occupies data[i*dim : (i+1)*dim], and ids[i] is the external ID.
// safe for concurrent use

type Collection struct {
	mu   sync.RWMutex
	dim  int
	data []float32
	ids  []string
	pos  map[string]int // external ID -> position

}

type Result struct {
	ID       string
	Distance float32
}

func NewCollection(dim int) (*Collection, error) {
	if dim <= 0 {
		return nil, fmt.Errorf("dimension must be greater than 0, got %d", dim)
	}
	return &Collection{dim: dim, pos: make(map[string]int)}, nil
}

// Dim returns the dimensions of the vectors stored in a  collection
func (c *Collection) Dim() int {
	return c.dim
}

// Len returns the number of stored vectores in a collection
func (c *Collection) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.ids)
}

// Insert stores a copy of the vector in the collection under "id"
// It rejects empty/duplicate ids and vectors with NaN or infinite values
// Rejections leave the collection unchganed
func (c *Collection) Insert(id string, vec []float32) error {
	if id == "" {
		return fmt.Errorf("id cannot be blank")
	}

	if err := c.validate(vec); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock() // unlocks when the function returns

	if _, exists := c.pos[id]; exists {
		return fmt.Errorf("id %q already exists", id)
	}

	c.pos[id] = len(c.ids)
	c.ids = append(c.ids, id)
	c.data = append(c.data, vec...) //append copies the value into the block
	return nil
}

// Get returns a copy of the vector stored under id, or nil and false if absent.
func (c *Collection) Get(id string) ([]float32, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	index, exists := c.pos[id]
	if !exists {
		return nil, false
	}
	start := index * c.dim
	return slices.Clone(c.data[start : start+c.dim]), true
}

// Search returns up to k nearest neighbors of query by L2 distance
// ties are decided by ID so results are always stable. has to scan every vector
func (c *Collection) Search(query []float32, k int) ([]Result, error) {
	if err := c.validateSearch(query, k); err != nil {
		return nil, err
	}
	c.mu.RLock()
	results := make([]Result, len(c.ids))
	for i, id := range c.ids {
		vec := c.data[i*c.dim : (i+1)*c.dim]
		results[i] = Result{ID: id, Distance: L2(query, vec)}
	}
	c.mu.RUnlock()

	return rankResults(results, k), nil
}

func rankResults(results []Result, k int) []Result {
	sort.Slice(results, func(i, j int) bool {
		if results[i].Distance == results[j].Distance {
			return results[i].ID < results[j].ID
		}
		return results[i].Distance < results[j].Distance
	})

	return results[:min(k, len(results))]
}

func (c *Collection) validateSearch(query []float32, k int) error {
	if k <= 0 {
		return fmt.Errorf("k must be greater than 0, got %d", k)
	}
	return c.validate(query)
}

func (c *Collection) validate(vec []float32) error {
	if len(vec) != c.dim {
		return fmt.Errorf("expected %d dimensions, got %d", c.dim, len(vec))
	}

	for _, v := range vec {
		f := float64(v)
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return fmt.Errorf("values must be finite")
		}
	}
	return nil
}
