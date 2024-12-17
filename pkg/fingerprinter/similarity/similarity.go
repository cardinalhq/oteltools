package similarity

import (
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"
)

// deduplicateSorted removes duplicates from a sorted slice in-place, returning a new slice with unique elements.
// The returned slice shares the same underlying array but may have a shorter length.
func deduplicateSorted[T comparable](data []T) []T {
	if len(data) == 0 {
		return data
	}

	j := 0
	for i := 1; i < len(data); i++ {
		if data[i] != data[j] {
			j++
			data[j] = data[i]
		}
	}
	return data[:j+1]
}

// JaccardSimilarity calculates the Jaccard similarity by sorting the input slices,
// deduplicating them, and then using a two-pointer approach to find the intersection.
// This function uses generics to handle any comparable type.
func JaccardSimilarity[T constraints.Ordered](a, b []T) float64 {
	slices.Sort(a)
	slices.Sort(b)

	// Deduplicate both slices
	a = deduplicateSorted(a)
	b = deduplicateSorted(b)

	lenA := len(a)
	lenB := len(b)

	// Handle edge cases
	if lenA == 0 && lenB == 0 {
		return 1.0
	}
	if lenA == 0 || lenB == 0 {
		return 0.0
	}

	// Two-pointer intersection count
	i, j := 0, 0
	intersection := 0
	for i < lenA && j < lenB {
		if a[i] == b[j] {
			intersection++
			i++
			j++
		} else if a[i] < b[j] {
			i++
		} else {
			j++
		}
	}

	union := lenA + lenB - intersection
	return float64(intersection) / float64(union)
}
