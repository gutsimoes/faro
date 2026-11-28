package core

import (
	"context"
	"fmt"
)

// SearchParallel returns the same L2 neighbors as Search using up to workers
// goroutines. It stops scanning when ctx is canceled and waits for its workers
// to finish before returning. It never returns partial results.
func (c *Collection) SearchParallel(ctx context.Context, query []float32, k, workers int) ([]Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if workers <= 0 {
		return nil, fmt.Errorf("workers must be greater than 0, got %d", workers)
	}
	if err := c.validateSearch(query, k); err != nil {
		return nil, err
	}

	candidates := c.parallelCandidates(ctx, query, k, workers)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	results := rankResults(candidates, k)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func (c *Collection) parallelCandidates(ctx context.Context, query []float32, k, workers int) []Result {
	c.mu.RLock()
	defer c.mu.RUnlock()

	count := len(c.ids)
	workers = min(workers, count)
	if workers == 0 || ctx.Err() != nil {
		return []Result{}
	}

	candidates := make(chan []Result, workers)
	start := 0
	for i := range workers {
		end := start + (count-start)/(workers-i)
		go func(start, end int) {
			candidates <- c.scanCandidates(ctx, query, k, start, end)
		}(start, end)
		start = end
	}

	// Drain all workers before unlocking the collection, even on cancellation.
	var combined []Result
	for range workers {
		combined = append(combined, (<-candidates)...)
	}
	return combined
}

func (c *Collection) scanCandidates(ctx context.Context, query []float32, k, start, end int) []Result {
	done := ctx.Done()
	results := make([]Result, end-start)
	for i := range results {
		select {
		case <-done:
			return nil
		default:
		}
		pos := start + i
		vec := c.data[pos*c.dim : (pos+1)*c.dim]
		results[i] = Result{ID: c.ids[pos], Distance: L2(query, vec)}
	}
	if ctx.Err() != nil {
		return nil
	}
	// A vector outside its partition's top k cannot be in the global top k.
	if k < len(results) {
		return rankResults(results, k)
	}
	return results
}
