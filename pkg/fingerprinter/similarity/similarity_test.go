package similarity

import (
	"testing"

	"golang.org/x/exp/constraints"
)

func TestJaccardSimilarity(t *testing.T) {
	tests := []struct {
		setA, setB []string
		expected   float64
	}{
		{[]string{}, []string{}, 1.0},
		{[]string{"a"}, []string{"b"}, 0.0},
		{[]string{"a", "b"}, []string{"b", "c"}, 1.0 / 3.0},           // intersection = {"b"}, union = {"a","b","c"}
		{[]string{"a", "a", "b"}, []string{"b", "b", "c"}, 1.0 / 3.0}, // duplicates handled as sets
		{[]string{"a"}, []string{"a", "b", "c"}, 1.0 / 3.0},           // A is shorter than B
		{[]string{"a", "b", "c"}, []string{"a"}, 1.0 / 3.0},           // B is shorter than A
		{[]string{}, []string{"a"}, 0.0},                              // A is empty
		{[]string{"a"}, []string{}, 0.0},                              // B is empty
		{[]string{}, []string{}, 1.0},                                 // both empty
		{[]string{"a", "b", "c"}, []string{"b", "c", "d"}, 1.0 / 2.0}, // Test case where A pointer is incremented more often
		{[]string{"b", "c", "d"}, []string{"a", "b", "c"}, 1.0 / 2.0}, // Test case where B pointer is incremented more often
	}

	for _, tc := range tests {
		got := JaccardSimilarity(tc.setA, tc.setB)
		if got != tc.expected {
			t.Errorf("JaccardSimilarity(%v, %v) = %v, expected %v",
				tc.setA, tc.setB, got, tc.expected)
		}
	}
}

func setupTestSets(n int, offset int64) ([]int64, []int64) {
	setA := make([]int64, n)
	setB := make([]int64, n)
	for i := int64(0); i < int64(n); i++ {
		setA[i] = i
		setB[i] = i + offset
	}
	return setA, setB
}

func BenchmarkJaccardSimilarity(b *testing.B) {
	setA, setB := setupTestSets(10, 1)
	b.Run("10_items_mostly_overlap", func(b *testing.B) {
		benchmarkJaccardSimilarity(b, setA, setB)
	})

	setA, setB = setupTestSets(10, 10)
	b.Run("10_items_no_overlap", func(b *testing.B) {
		benchmarkJaccardSimilarity(b, setA, setB)
	})

	setA, setB = setupTestSets(100, 1)
	b.Run("100_items_mostly_overlap", func(b *testing.B) {
		benchmarkJaccardSimilarity(b, setA, setB)
	})

	setA, setB = setupTestSets(100, 100)
	b.Run("100_items_no_overlap", func(b *testing.B) {
		benchmarkJaccardSimilarity(b, setA, setB)
	})

	setA, setB = setupTestSets(1000, 1)
	b.Run("1000_items_mostly_overlap", func(b *testing.B) {
		benchmarkJaccardSimilarity(b, setA, setB)
	})

	setA, setB = setupTestSets(1000, 1000)
	b.Run("1000_items_no_overlap", func(b *testing.B) {
		benchmarkJaccardSimilarity(b, setA, setB)
	})
}

func benchmarkJaccardSimilarity[T constraints.Ordered](b *testing.B, setA, setB []T) {
	b.Helper()
	for i := 0; i < b.N; i++ {
		JaccardSimilarity(setA, setB)
	}
}
