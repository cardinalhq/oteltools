package chqpb

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTopKByValue_AddAndSortedSlice(t *testing.T) {
	topK := NewTopKByValue(3, time.Second)

	topK.Add(1, 10)
	topK.Add(2, 20)
	topK.Add(3, 30)

	sorted := topK.SortedSlice()
	assert.Len(t, sorted, 3)
	assert.Equal(t, int64(3), sorted[0].Tid)
	assert.Equal(t, float64(30), sorted[0].Value)
	assert.Equal(t, int64(2), sorted[1].Tid)
	assert.Equal(t, int64(1), sorted[2].Tid)
}

func TestTopKByValue_Eviction(t *testing.T) {
	topK := NewTopKByValue(3, time.Second)

	topK.Add(1, 10)
	topK.Add(2, 20)
	topK.Add(3, 30)
	topK.Add(4, 25) // Should evict tid=1 (value 10)

	sorted := topK.SortedSlice()
	assert.Len(t, sorted, 3)
	assert.NotContains(t, extractTids(sorted), int64(1))
	assert.Contains(t, extractTids(sorted), int64(4))
}

func TestTopKByValue_UpdateImproved(t *testing.T) {
	topK := NewTopKByValue(3, time.Second)

	topK.Add(1, 10)
	topK.Add(2, 20)
	topK.Add(3, 30)

	topK.Add(1, 40) // Should update tid=1 and move to top

	sorted := topK.SortedSlice()
	assert.Equal(t, int64(1), sorted[0].Tid)
	assert.Equal(t, float64(40), sorted[0].Value)
}

func TestTopKByValue_UpdateWorse(t *testing.T) {
	topK := NewTopKByValue(3, time.Second)

	topK.Add(1, 10)
	topK.Add(2, 20)
	topK.Add(3, 30)

	topK.Add(1, 5) // Worse than current: should have no effect

	sorted := topK.SortedSlice()
	assert.Equal(t, float64(10), sorted[2].Value)
}

func TestTopKByValue_Eligible(t *testing.T) {
	topK := NewTopKByValue(3, time.Second)

	// Warmup: less than k items, any value is eligible
	assert.True(t, topK.Eligible(1))

	topK.Add(1, 10)
	topK.Add(2, 20)

	// Still in warmup
	assert.True(t, topK.Eligible(5))
	assert.True(t, topK.Eligible(100))

	topK.Add(3, 30) // Reaches capacity

	// Now only values greater than min (10) are eligible
	assert.False(t, topK.Eligible(5))  // Too low
	assert.False(t, topK.Eligible(10)) // Equal to min
	assert.True(t, topK.Eligible(15))  // Higher than min
	assert.True(t, topK.Eligible(100)) // Much higher
}

func extractTids(items []itemWithValue) []int64 {
	var out []int64
	for _, item := range items {
		out = append(out, item.Tid)
	}
	return out
}
