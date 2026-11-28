package core

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"testing"
	"time"
)

func TestSearchParallelMatchesSequential(t *testing.T) {
	c := newTestCollection(t)
	query := []float32{0, 0}
	got, err := c.SearchParallel(t.Context(), query, 3, 4)
	if err != nil || got == nil || len(got) != 0 {
		t.Fatalf("empty collection: got %v, %v; want an empty slice", got, err)
	}

	for _, item := range []struct {
		id  string
		vec []float32
	}{
		{"z", []float32{1, 0}},
		{"far", []float32{3, 4}},
		{"b", []float32{0, 1}},
		{"origin", []float32{0, 0}},
		{"a", []float32{-1, 0}},
	} {
		if err := c.Insert(item.id, item.vec); err != nil {
			t.Fatal(err)
		}
	}

	for _, k := range []int{1, 2, 3, 5, 7} {
		want, err := c.Search(query, k)
		if err != nil {
			t.Fatal(err)
		}
		for _, workers := range []int{1, 2, 3, 8} {
			got, err := c.SearchParallel(t.Context(), query, k, workers)
			if err != nil || !slices.Equal(got, want) {
				t.Errorf("k=%d workers=%d: got %v, %v; want %v", k, workers, got, err, want)
			}
		}
	}
}

func TestSearchParallelRejectsInvalidInput(t *testing.T) {
	c := newTestCollection(t)
	for _, tc := range []struct {
		query   []float32
		k       int
		workers int
	}{
		{[]float32{0, 0}, 1, 0},
		{[]float32{0, 0}, 1, -1},
		{[]float32{0, 0}, 0, 1},
		{[]float32{0}, 1, 1},
	} {
		if got, err := c.SearchParallel(t.Context(), tc.query, tc.k, tc.workers); err == nil || got != nil {
			t.Errorf("SearchParallel(%v, %d, %d) = %v, %v; want an error and no results", tc.query, tc.k, tc.workers, got, err)
		}
	}
}

func TestSearchParallelCanceledContext(t *testing.T) {
	c := newTestCollection(t)
	if err := c.Insert("a", []float32{1, 1}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	expired, cancelDeadline := context.WithDeadline(t.Context(), time.Now().Add(-time.Second))
	defer cancelDeadline()
	for _, tc := range []struct {
		ctx  context.Context
		want error
	}{
		{ctx, context.Canceled},
		{expired, context.DeadlineExceeded},
	} {
		got, err := c.SearchParallel(tc.ctx, []float32{0, 0}, 1, 2)
		if !errors.Is(err, tc.want) || got != nil {
			t.Errorf("got %v, %v; want nil results and %v", got, err, tc.want)
		}
	}
}

func TestSearchParallelCancelsDuringScan(t *testing.T) {
	c := newTestCollection(t)
	if err := c.Insert("a", []float32{1, 1}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	got, err := c.SearchParallel(cancelOnScanContext{ctx, cancel}, []float32{0, 0}, 1, 1)
	if !errors.Is(err, context.Canceled) || got != nil {
		t.Fatalf("got %v, %v; want nil results and context.Canceled", got, err)
	}
}

// Cancel when the worker starts observing the context, without relying on a timer.
type cancelOnScanContext struct {
	context.Context
	cancel context.CancelFunc
}

func (ctx cancelOnScanContext) Done() <-chan struct{} {
	ctx.cancel()
	return ctx.Context.Done()
}

func TestConcurrentInsertAndSearch(t *testing.T) {
	c := newTestCollection(t)
	if err := c.Insert("origin", []float32{0, 0}); err != nil {
		t.Fatal(err)
	}
	const goroutines, iterations = 4, 50
	var group sync.WaitGroup
	start := make(chan struct{})
	for worker := range goroutines {
		group.Go(func() {
			<-start
			for i := range iterations {
				id := fmt.Sprintf("%d/%d", worker, i)
				if err := c.Insert(id, []float32{float32(worker + 1), float32(i)}); err != nil {
					t.Error(err)
					return
				}
			}
		})
		group.Go(func() {
			<-start
			query := []float32{0, 0}
			want := []Result{{ID: "origin", Distance: 0}}
			for range iterations {
				vec, found := c.Get("origin")
				if !found || !slices.Equal(vec, query) {
					t.Errorf("Get(origin): got %v, %t; want %v, true", vec, found, query)
					return
				}
				got, err := c.Search(query, 1)
				if err != nil || !slices.Equal(got, want) {
					t.Errorf("Search: got %v, %v; want %v", got, err, want)
					return
				}
				got, err = c.SearchParallel(t.Context(), query, 1, 3)
				if err != nil || !slices.Equal(got, want) {
					t.Errorf("SearchParallel: got %v, %v; want %v", got, err, want)
					return
				}
			}
		})
	}
	close(start)
	group.Wait()

	if got, want := c.Len(), 1+goroutines*iterations; got != want {
		t.Errorf("Len() = %d; want %d", got, want)
	}
}
