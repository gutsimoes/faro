package core

import "math"

// L2 returns the Euclidean distance between a and b.

func L2(a, b []float32) float32 {
	var sum float32
	for i := range a {
		d := a[i] - b[i]
		sum += d * d
	}
	return float32(math.Sqrt(float64(sum)))
}

// CosineDistance returns one minus cosine similarity for vectors of equal length.
// A zero vector is assigned similarity zero, giving distance one.
func CosineDistance(a, b []float32) float32 {
	var dot, normA, normB float64
	for i, value := range a {
		x, y := float64(value), float64(b[i])
		dot += x * y
		normA += x * x
		normB += y * y
	}
	if normA == 0 || normB == 0 {
		return 1
	}
	cosine := dot / (math.Sqrt(normA) * math.Sqrt(normB))
	// Roundoff can push the cosine slightly outside its mathematical range.
	return float32(1 - min(1, max(-1, cosine)))
}

// DotProductDistance returns the negated dot product for vectors of equal length.
// A larger dot product ranks closer when scores are ordered ascending.
func DotProductDistance(a, b []float32) float32 {
	var dot float64
	for i, value := range a {
		dot += float64(value) * float64(b[i])
	}
	return float32(-dot)
}
