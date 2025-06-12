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
	var got []int64
	for _, proto := range flushed.Sketches {
		got = append(got, proto.Tid)
	}
	//slices.Sort(got)

	want := []int64{tid4, tid3}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("at index %d, got %d, want %d", i, got[i], w)
		}
	}
}
