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
	"math"
	"testing"

	"github.com/DataDog/sketches-go/ddsketch"
	"github.com/DataDog/sketches-go/ddsketch/mapping"
	"github.com/DataDog/sketches-go/ddsketch/store"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

func TestDDSketch_ExponentialHistogram_RoundTrip(t *testing.T) {
	// --- build a source DDSketch with mixed values (pos/neg/zero) ---
	m, err := mapping.NewLogarithmicMapping(0.01) // ~1% rel accuracy
	if err != nil {
		t.Fatalf("mapping: %v", err)
	}
	src := ddsketch.NewDDSketchFromStoreProvider(m, store.DefaultProvider)

	// zeros
	_ = src.AddWithCount(0, 7)

	// positives
	_ = src.AddWithCount(0.5, 3)
	_ = src.AddWithCount(1, 5)
	_ = src.AddWithCount(2, 8)
	_ = src.AddWithCount(4, 6)
	_ = src.AddWithCount(8, 2)
	_ = src.AddWithCount(32, 1)

	// negatives
	_ = src.AddWithCount(-0.75, 4)
	_ = src.AddWithCount(-3, 3)
	_ = src.AddWithCount(-12, 1)

	srcCount := src.GetCount()
	srcZero := src.GetZeroCount()
	srcSum := src.GetSum()
	srcMin, _ := src.GetMinValue()
	srcMax, _ := src.GetMaxValue()
	qList := []float64{0.5, 0.9, 0.99}
	srcQ := make([]float64, len(qList))
	for i, q := range qList {
		v, _ := src.GetValueAtQuantile(q)
		srcQ[i] = v
	}

	// --- convert: DDSketch -> ExponentialHistogram ---
	dp := pmetric.NewExponentialHistogramDataPoint()
	if err := ConvertDDSketchToExponentialHistogram(dp, src); err != nil {
		t.Fatalf("ConvertDDSketchToExponentialHistogram: %v", err)
	}

	// quick smoke: scale should be a sane integer, and some buckets present
	if dp.Scale() < -10 || dp.Scale() > 20 {
		t.Fatalf("unexpected scale: %d", dp.Scale())
	}
	if dp.Positive().BucketCounts().Len() == 0 && dp.Negative().BucketCounts().Len() == 0 && dp.ZeroCount() == 0 {
		t.Fatalf("histogram is empty after conversion")
	}

	// --- convert back: ExponentialHistogram -> DDSketch ---
	rt, err := ConvertExponentialHistogramToDDSketch(dp)
	if err != nil {
		t.Fatalf("ConvertExponentialHistogramToDDSketch: %v", err)
	}

	// --- compare stats with tolerances (allowing small discretization error) ---
	const (
		absTol = 1e-6
		relTol = 0.02 // 2% relative tolerance
	)

	// count & zeros are precise integers in our path
	if !approxEqual(srcCount, rt.GetCount(), 0, 0) {
		t.Fatalf("count mismatch: got %v, want %v", rt.GetCount(), srcCount)
	}
	if !approxEqual(srcZero, rt.GetZeroCount(), 0, 0) {
		t.Fatalf("zeroCount mismatch: got %v, want %v", rt.GetZeroCount(), srcZero)
	}

	// sum/min/max should be very close
	rtSum := rt.GetSum()
	rtMin, _ := rt.GetMinValue()
	rtMax, _ := rt.GetMaxValue()

	if !approxEqual(srcSum, rtSum, absTol, relTol) {
		t.Fatalf("sum mismatch: got %v, want %v", rtSum, srcSum)
	}
	if !approxEqual(srcMin, rtMin, absTol, relTol) {
		t.Fatalf("min mismatch: got %v, want %v", rtMin, srcMin)
	}
	if !approxEqual(srcMax, rtMax, absTol, relTol) {
		t.Fatalf("max mismatch: got %v, want %v", rtMax, srcMax)
	}

	// a few quantiles
	for i, q := range qList {
		want, _ := src.GetValueAtQuantile(q)
		got, _ := rt.GetValueAtQuantile(q)
		if !approxEqual(want, got, absTol, 3*relTol) { // slightly looser for quantiles
			t.Fatalf("q=%.3f mismatch: got %v, want %v (i=%d)", q, got, want, i)
		}
	}
}

func approxEqual(want, got, absTol, relTol float64) bool {
	if math.IsNaN(want) || math.IsNaN(got) {
		return false
	}
	diff := math.Abs(want - got)
	if diff <= absTol {
		return true
	}
	den := math.Max(1.0, math.Abs(want))
	return diff/den <= relTol
}
