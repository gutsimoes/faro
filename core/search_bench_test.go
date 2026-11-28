package core

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
)

func BenchmarkSearch(b *testing.B) {
	const k = 10
	ctx := context.Background()
	for _, size := range []int{100, 1000, 10000} {
		for _, dim := range []int{16, 128} {
			b.Run(fmt.Sprintf("n=%d/d=%d", size, dim), func(b *testing.B) {
				c, query := benchmarkCollection(b, size, dim)

				b.Run("sequential", func(b *testing.B) {
					b.ReportAllocs()
					for b.Loop() {
						if _, err := c.Search(query, k); err != nil {
							b.Fatal(err)
						}
					}
				})
				for _, workers := range []int{1, 2, 4} {
					b.Run(fmt.Sprintf("parallel/workers=%d", workers), func(b *testing.B) {
						b.ReportAllocs()
						for b.Loop() {
							if _, err := c.SearchParallel(ctx, query, k, workers); err != nil {
								b.Fatal(err)
							}
						}
					})
				}
			})
		}
	}
}

func benchmarkCollection(b *testing.B, size, dim int) (*Collection, []float32) {
	b.Helper()
	c, err := NewCollection(dim)
	if err != nil {
		b.Fatal(err)
	}

	rng := rand.New(rand.NewSource(42))
	vec := make([]float32, dim)
	for i := range size {
		for j := range vec {
			vec[j] = rng.Float32()
		}
		if err := c.Insert(strconv.Itoa(i), vec); err != nil {
			b.Fatal(err)
		}
	}

	query := make([]float32, dim)
	for i := range query {
		query[i] = rng.Float32()
	}
	return c, query
}
