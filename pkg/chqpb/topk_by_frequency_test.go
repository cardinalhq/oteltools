// Copyright 2024-2025 CardinalHQ, Inc.
//
// CardinalHQ, Inc. proprietary and confidential.
// Unauthorized copying, distribution, or modification of this file,
// via any medium, is strictly prohibited without prior written consent.
//
// All rights reserved.

package chqpb

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTopKByFrequency_EligibleWithCount(t *testing.T) {
	topK := NewTopKByFrequency(3, time.Second)

	// Test 1: heap not full → any tid is eligible
	assert.True(t, topK.EligibleWithCount(42, 1))

	topK.Add(1) // count = 1
	topK.Add(2) // count = 1
	topK.Add(3) // count = 1

	// Heap now full

	// Test 2: existing TID should always be eligible
	assert.True(t, topK.EligibleWithCount(2, 1))

	// Test 3: new tid with count ≤ min is not eligible
	assert.False(t, topK.EligibleWithCount(99, 1)) // min is 1

	// Test 4: new tid with count > min is eligible
	assert.True(t, topK.EligibleWithCount(99, 2))

	// Increase counts so that min becomes 2
	topK.Add(1)
	topK.Add(2)
	topK.Add(3)

	// Test 5: new tid with projected count = min (2) is not eligible
	assert.False(t, topK.EligibleWithCount(88, 2))

	// Test 6: new tid with projected count > min is eligible
	assert.True(t, topK.EligibleWithCount(77, 3))

	// Test 7: tid already has partial count
	topK.count[200] = 1                             // simulate prior count, but not in heap
	assert.True(t, topK.EligibleWithCount(200, 2))  // total = 3 > min
	assert.False(t, topK.EligibleWithCount(200, 1)) // total = 2 == min → not eligible
}

func TestTopKByFrequency_CleanupExpired(t *testing.T) {
	ttl := 100 * time.Millisecond
	topK := NewTopKByFrequency(5, ttl)

	// Add TID 1 and let it age out
	topK.Add(1)
	time.Sleep(2 * ttl)

	// Add TID 2 and TID 3 within TTL window
	topK.Add(2)
	topK.Add(3)

	// Perform cleanup
	topK.CleanupExpired()

	// Only TID 2 and 3 should remain
	sorted := topK.SortedSlice()
	tids := extractTidsFreq(sorted)

	assert.Len(t, sorted, 2)
	assert.Contains(t, tids, int64(2))
	assert.Contains(t, tids, int64(3))
	assert.NotContains(t, tids, int64(1))
}

func TestTopKByFrequency_AddCount(t *testing.T) {
	topK := NewTopKByFrequency(3, time.Second)

	// Add tid=100 with count=5 in one shot
	topK.AddCount(100, 5)

	// Add tid=200 and 300 with smaller counts
	topK.AddCount(200, 2)
	topK.AddCount(300, 3)

	// tid=400 should not enter with count=1
	topK.AddCount(400, 1)

	sorted := topK.SortedSlice()
	assert.Len(t, sorted, 3)

	assert.Equal(t, int64(100), sorted[0].Tid)
	assert.Equal(t, int64(5), sorted[0].Count)

	assert.Equal(t, int64(300), sorted[1].Tid)
	assert.Equal(t, int64(3), sorted[1].Count)

	assert.Equal(t, int64(200), sorted[2].Tid)
	assert.Equal(t, int64(2), sorted[2].Count)

	assert.False(t, topK.Eligible(400)) // not in top-K
	assert.True(t, topK.Eligible(100))
	assert.True(t, topK.Eligible(200))
	assert.True(t, topK.Eligible(300))
}

func TestTopKByFrequency_AddAndEligible(t *testing.T) {
	topK := NewTopKByFrequency(3, time.Second)

	// Fill to k with one hit each.
	topK.Add(1)
	topK.Add(2)
	topK.Add(3)

	assert.True(t, topK.Eligible(1))
	assert.True(t, topK.Eligible(2))
	assert.True(t, topK.Eligible(3))

	// New TID with count 1 should NOT enter top-K yet.
	topK.Add(4)
	assert.False(t, topK.Eligible(4))
}

func TestTopKByFrequency_Eviction(t *testing.T) {
	topK := NewTopKByFrequency(3, time.Second)

	// All start with count = 1
	topK.Add(1)
	topK.Add(2)
	topK.Add(3)

	// Make TID 4 reach count = 2 → should evict one of the 1-count tids.
	topK.Add(4)
	topK.Add(4)

	sorted := topK.SortedSlice()
	assert.Len(t, sorted, 3)

	assert.Equal(t, int64(4), sorted[0].Tid)   // 4 is now most-frequent
	assert.Equal(t, int64(2), sorted[0].Count) // and its count is 2

	assert.NotContains(t, extractTidsFreq(sorted), int64(1)) // one 1-count tid is gone
	assert.True(t, topK.Eligible(4))                         // 4 is definitely eligible
}

func TestTopKByFrequency_UpdateInPlace(t *testing.T) {
	topK := NewTopKByFrequency(3, time.Second)

	topK.Add(10)
	topK.Add(20)
	topK.Add(30)

	// Hit 20 twice more → count = 3
	topK.Add(20)
	topK.Add(20)

	sorted := topK.SortedSlice()
	assert.Equal(t, int64(20), sorted[0].Tid)
	assert.Equal(t, int64(3), sorted[0].Count)
}

func TestTopKByFrequency_EvictionWithTie(t *testing.T) {
	topK := NewTopKByFrequency(2, time.Second)

	// Two tids with count 1 each
	topK.Add(1)
	topK.Add(2)

	// First hit for 3 (count=1) won't evict anyone
	topK.Add(3)
	assert.False(t, topK.Eligible(3))

	// Second hit for 3 makes count=2 → should evict one of the 1-count tids
	topK.Add(3)

	assert.True(t, topK.Eligible(3))
	sorted := topK.SortedSlice()
	assert.Len(t, sorted, 2)
	assert.Equal(t, int64(3), sorted[0].Tid)
	assert.Equal(t, int64(2), sorted[0].Count)
}

func TestTopKByFrequency_SortedOrder(t *testing.T) {
	topK := NewTopKByFrequency(5, time.Second)

	// 10 seen 3×
	topK.Add(10)
	topK.Add(10)
	topK.Add(10)

	// 30 seen 2×
	topK.Add(30)
	topK.Add(30)

	// 20, 40, 50 seen 1×
	topK.Add(20)
	topK.Add(40)
	topK.Add(50)

	sorted := topK.SortedSlice()
	assert.Equal(t, int64(10), sorted[0].Tid)
	assert.Equal(t, int64(3), sorted[0].Count)
	assert.Equal(t, int64(30), sorted[1].Tid)
	assert.Equal(t, int64(2), sorted[1].Count)
	assert.Equal(t, int64(20), sorted[2].Tid)
	assert.Equal(t, int64(1), sorted[2].Count)
	assert.Equal(t, int64(40), sorted[3].Tid)
	assert.Equal(t, int64(1), sorted[3].Count)
	assert.Equal(t, int64(50), sorted[4].Tid)
	assert.Equal(t, int64(1), sorted[4].Count)
}

// Test TopKByFrequency.AddCount behavior and boolean return
func TestTopKByFrequency_AddCount_Bool(t *testing.T) {
	k := 3
	ttl := time.Minute
	topK := NewTopKByFrequency(k, ttl)

	// Add initial entries: expect true
	if !topK.AddCount(1, 1) {
		t.Errorf("AddCount(1,1) = false; want true")
	}
	if !topK.AddCount(2, 2) {
		t.Errorf("AddCount(2,2) = false; want true")
	}
	if !topK.AddCount(3, 3) {
		t.Errorf("AddCount(3,3) = false; want true")
	}

	// Heap is full: AddCount for lower count -> false
	if topK.AddCount(4, 1) {
		t.Errorf("AddCount(4,1) = true; want false (1 <= min)")
	}

	// AddCount for higher count -> true and eviction of tid=1
	if !topK.AddCount(4, 5) {
		t.Errorf("AddCount(4,5) = false; want true (5 > min)")
	}
	// After eviction, tid=1 should no longer be eligible
	if topK.Eligible(1) {
		t.Errorf("Eligible(1) = true; want false (evicted)")
	}
	// tid=4 should be eligible
	if !topK.Eligible(4) {
		t.Errorf("Eligible(4) = false; want true (added)")
	}

	// Update existing tid with higher count -> true
	if !topK.AddCount(4, 1) {
		t.Errorf("AddCount(4,1) = false; want true (existing updated)")
	}
	// And count should have increased: SortedSlice()[0] should be tid=4
	sorted := topK.SortedSlice()
	if len(sorted) == 0 || sorted[0].Tid != 4 {
		t.Errorf("SortedSlice()[0].Tid = %v; want 4", sorted[0].Tid)
	}
}

// helper
func extractTidsFreq(items []itemWithCount) []int64 {
	out := make([]int64, len(items))
	for i, itm := range items {
		out[i] = itm.Tid
	}
	return out
}
