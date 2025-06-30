// Copyright 2024-2025 CardinalHQ, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package chqpb

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTopKByValue_AddAndSortedSlice(t *testing.T) {
	topK := NewTopKByValue(3, time.Second, Direction_UP)

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
	topK := NewTopKByValue(3, time.Second, Direction_UP)

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
	topK := NewTopKByValue(3, time.Second, Direction_UP)

	topK.Add(1, 10)
	topK.Add(2, 20)
	topK.Add(3, 30)

	topK.Add(1, 40) // Should update tid=1 and move to top

	sorted := topK.SortedSlice()
	assert.Equal(t, int64(1), sorted[0].Tid)
	assert.Equal(t, float64(40), sorted[0].Value)
}

func TestTopKByValue_UpdateWorse(t *testing.T) {
	topK := NewTopKByValue(3, time.Second, Direction_UP)

	topK.Add(1, 10)
	topK.Add(2, 20)
	topK.Add(3, 30)

	topK.Add(1, 5) // Worse than current: should have no effect

	sorted := topK.SortedSlice()
	assert.Equal(t, float64(10), sorted[2].Value)
}

func TestTopKByValue_Eligible(t *testing.T) {
	topK := NewTopKByValue(3, time.Second, Direction_UP)

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

func TestTopKByValue_Add(t *testing.T) {
	k := 3
	ttl := time.Minute
	topK := NewTopKByValue(k, ttl, Direction_UP)

	// Add initial entries: expect true
	if !topK.Add(1, 10.0) {
		t.Errorf("Add(1, 10.0) = false; want true")
	}
	if !topK.Add(2, 20.0) {
		t.Errorf("Add(2, 20.0) = false; want true")
	}
	if !topK.Add(3, 30.0) {
		t.Errorf("Add(3, 30.0) = false; want true")
	}

	// Heap is full: Add lower value -> false
	if topK.Add(4, 5.0) {
		t.Errorf("Add(4, 5.0) = true; want false (5 <= min)")
	}

	// Add higher value -> true and eviction of tid=1
	if !topK.Add(4, 40.0) {
		t.Errorf("Add(4,40.0) = false; want true (40 > min)")
	}
	// After eviction, tid=1 should no longer be eligible
	if topK.Eligible(10.0) && reflect.DeepEqual(topK.SortedSlice()[len(topK.SortedSlice())-1].Tid, int64(1)) {
		t.Errorf("tid=1 should have been evicted")
	}

	// Update existing tid lower value -> true (existing, no improvement)
	if !topK.Add(4, 35.0) {
		t.Errorf("Add(4, 35.0) = false; want true (existing)")
	}
	// Update existing tid with higher value -> true
	if !topK.Add(4, 50.0) {
		t.Errorf("Add(4, 50.0) = false; want true (improved value)")
	}

	// Final top element should be tid=4 with highest value
	sorted := topK.SortedSlice()
	if len(sorted) == 0 || sorted[0].Tid != 4 {
		t.Errorf("SortedSlice()[0].Tid = %v; want 4", sorted[0].Tid)
	}
}

func extractTids(items []itemWithValue) []int64 {
	out := make([]int64, len(items))
	for i, it := range items {
		out[i] = it.Tid
	}
	return out
}

func TestTopKByValue_DownDirection(t *testing.T) {
	const k = 3
	ttl := time.Second
	topK := NewTopKByValue(k, ttl, Direction_DOWN)

	// Before capacity: everything is eligible
	assert.True(t, topK.Eligible(100), "warmup: any value should be eligible")

	// Fill with three entries
	topK.Add(1, 100)
	topK.Add(2, 200)
	topK.Add(3, 300)

	// Now at capacity: should keep the bottom-K smallest, sorted ascending
	sorted := topK.SortedSlice()
	assert.Len(t, sorted, k)
	assert.Equal(t, []int64{1, 2, 3}, extractTids(sorted), "initial bottom-K in ascending order")

	// A new SMALLER value (50) should evict the current largest (300)
	topK.Add(4, 50)
	sorted = topK.SortedSlice()
	assert.Len(t, sorted, k)
	assert.NotContains(t, extractTids(sorted), int64(3), "300 should have been evicted")
	assert.Equal(t, []int64{4, 1, 2}, extractTids(sorted), "after eviction, should be [4,1,2]")

	// A new LARGER value should NOT be accepted
	assert.False(t, topK.Add(5, 500), "500 is too large to enter bottom-K")
	assert.NotContains(t, extractTids(topK.SortedSlice()), int64(5))

	// Eligibility after capacity:
	//  - values smaller than current max (200) should be eligible
	//  - values ≥ 200 should not
	assert.True(t, topK.Eligible(150), "150 < root(200) ⇒ eligible")
	assert.False(t, topK.Eligible(250), "250 ≥ root(200) ⇒ not eligible")

	// Updating an existing entry to a WORSE value (higher for DOWN) should keep it, but may demote its rank
	assert.True(t, topK.Add(1, 250), "existing tid=1 should still be accepted")
	assert.True(t, topK.Eligible(1), "tid=1 remains eligible after update")
}
