package main

import (
	"fmt"
	"os"

	"github.com/gutsimoes/faro/vector"
)

const version = "0.0-dev"

func main() {
	if err := demo(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func demo() error {
	collection, err := vector.NewCollection(2)
	if err != nil {
		return err
	}
	for _, item := range []struct {
		id     string
		values []float64
	}{
		{"origin", []float64{0, 0}},
		{"near", []float64{1, 1}},
		{"middle", []float64{3, 4}},
		{"far", []float64{10, 10}},
	} {
		if err := collection.Insert(item.id, item.values); err != nil {
			return err
		}
	}
	query := []float64{0, 0}
	results, err := collection.Search(query, 3)
	if err != nil {
		return err
	}
	fmt.Printf("Faro %s — exact L2 search\n", version)
	fmt.Printf("Vectors: %d | Query: %v | k: 3\n", collection.Len(), query)
	for i, result := range results {
		fmt.Printf("%d. %-6s distance=%.4f\n", i+1, result.ID, result.Distance)
	}
	return nil
}
