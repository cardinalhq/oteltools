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
	"testing"
	"time"
)

func TestGenericSketchCache_Flush_TopKChurn(t *testing.T) {
	interval := time.Second
	maxK := 2
	metricName := "metric1"
	metricType := "gauge"
	tags1 := map[string]string{"key": "v1"}
	tags2 := map[string]string{"key": "v2"}
	tags3 := map[string]string{"key": "v3"}
	tags4 := map[string]string{"key": "v4"}

	// Capture flushed output
	var flushed *GenericSketchList
	cache := NewGenericSketchCache(interval, "cust", "telemetryType", maxK,
		func(out *GenericSketchList) error {
			flushed = out
			return nil
		},
	)

	// Use a timestamp older than two intervals so flush will pick it up
	oldTime := time.Now().Add(-2 * interval)

	// Submit three entries with increasing values
	_ = cache.Update(metricName, metricType, tags1, 0, 0, 1.0, oldTime)
	_ = cache.Update(metricName, metricType, tags2, 0, 0, 2.0, oldTime)
	tid3 := cache.Update(metricName, metricType, tags3, 0, 0, 3.0, oldTime)
	tid4 := cache.Update(metricName, metricType, tags4, 0, 0, 4.0, oldTime)

	// Manually invoke flush
	cache.flush()

	if len(flushed.Sketches) != maxK {
		t.Fatalf("expected %d sketches, got %d", maxK, len(flushed.Sketches))
	}

	// Collect and sort the returned TIDs
	got := make(map[int64]struct{})
	for _, proto := range flushed.Sketches {
		got[proto.Tid] = struct{}{}
	}

	want := map[int64]struct{}{tid4: {}, tid3: {}}

	if len(got) != len(want) {
		t.Fatalf("expected %d sketches, got %d", len(want), len(got))
	}

	for tid := range want {
		if _, ok := got[tid]; !ok {
			t.Errorf("missing expected tid %d in flushed output", tid)
		}
	}
}

func TestTopKByFrequency_AddCount_And_Eligibility(t *testing.T) {
	ttl := 2 * time.Second
	k := 2
	topK := NewTopKByFrequency(k, ttl)

	// Add counts for two tids
	if !topK.AddCount(1001, 5) {
		t.Fatal("expected AddCount to return true for tid 1001")
	}
	if !topK.AddCount(1002, 10) {
		t.Fatal("expected AddCount to return true for tid 1002")
	}

	// Now add a new TID with a count high enough to evict the current min (1001)
	if !topK.AddCount(1003, 6) {
		t.Fatal("expected AddCount to return true for tid 1003 (should evict 1001)")
	}

	// Only top 2 should remain
	sorted := topK.SortedSlice()
	if len(sorted) != 2 {
		t.Fatalf("expected 2 items in heap, got %d", len(sorted))
	}

	// Expect topK to have 1002 and 1003
	got := map[int64]struct{}{sorted[0].Tid: {}, sorted[1].Tid: {}}
	want := map[int64]struct{}{1002: {}, 1003: {}}

	for tid := range want {
		if _, ok := got[tid]; !ok {
			t.Errorf("expected tid %d to be in top K", tid)
		}
	}

	// Test that an unqualified low-count TID is not added
	if topK.AddCount(9999, 1) {
		t.Errorf("expected AddCount to return false for low-count tid 9999")
	}
}
