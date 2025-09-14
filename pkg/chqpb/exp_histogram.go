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

// Copyright 2024-2025 CardinalHQ, Inc.
//
// CardinalHQ, Inc. proprietary and confidential.
// Unauthorized copying, distribution, or modification of this file,
// via any medium, is strictly prohibited without prior written consent.
//
// All rights reserved.

package chqpb

import (
	"errors"
	"github.com/DataDog/sketches-go/ddsketch/mapping"
	"math"

	"github.com/DataDog/sketches-go/ddsketch"
	"github.com/DataDog/sketches-go/ddsketch/store"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

// ConvertDDSketchToExponentialHistogram maps a DDSketch into an OTEL ExponentialHistogramDataPoint.
func ConvertDDSketchToExponentialHistogram(dp pmetric.ExponentialHistogramDataPoint, sk *ddsketch.DDSketch) error {
	if sk == nil {
		return errors.New("nil ddsketch")
	}

	// Use exported getter; on older ddsketch, replace with sk.GetKeyMapping().
	m := sk.IndexMapping
	if m == nil {
		return errors.New("ddsketch has no mapping")
	}

	// Choose an OTEL scale whose base ~= DDSketch gamma ≈ (1+α)/(1-α).
	alpha := m.RelativeAccuracy()
	if alpha <= 0 || alpha >= 1 {
		alpha = 0.01
	}
	gamma := (1 + alpha) / (1 - alpha)
	log2gamma := math.Log2(gamma)
	if log2gamma <= 0 {
		log2gamma = 1.0 / (1 << 10)
	}
	scale := int32(math.Round(math.Log2(1.0 / log2gamma)))
	if scale < -10 {
		scale = -10
	}
	if scale > 20 {
		scale = 20
	}
	dp.SetScale(scale)

	type buckAgg struct {
		minIdx int
		maxIdx int
		counts map[int]uint64
	}
	initAgg := func() buckAgg {
		return buckAgg{
			minIdx: math.MaxInt32,
			maxIdx: math.MinInt32,
			counts: make(map[int]uint64),
		}
	}
	pos := initAgg()
	neg := initAgg()

	// log_base(v) = log2(v) / log2(base) = log2(v) * 2^scale.
	mult := math.Pow(2, float64(scale))
	toExpoIndex := func(v float64) (int, bool) {
		if v <= 0 || math.IsNaN(v) || math.IsInf(v, 0) {
			return 0, false
		}
		return int(math.Floor(math.Log2(v) * mult)), true
	}

	// Iterate a store via the stable Bins() iterator.
	iterate := func(st store.Store, into *buckAgg, isNeg bool) {
		if st == nil {
			return
		}
		for b := range st.Bins() {
			c := b.Count()
			if c <= 0 {
				continue
			}
			idx := b.Index()  // DDSketch key (integer)
			v := m.Value(idx) // representative value for that key
			if isNeg {
				v = -v
			}
			i, ok := toExpoIndex(math.Abs(v))
			if !ok {
				continue
			}
			into.counts[i] += uint64(math.Round(c))
			if i < into.minIdx {
				into.minIdx = i
			}
			if i > into.maxIdx {
				into.maxIdx = i
			}
		}
	}

	iterate(sk.GetPositiveValueStore(), &pos, false)
	iterate(sk.GetNegativeValueStore(), &neg, true)

	// Zero bucket (if present).
	if z := sk.GetZeroCount(); z > 0 {
		dp.SetZeroCount(uint64(math.Round(z)))
	} else {
		dp.SetZeroCount(0)
	}

	// Materialize positive buckets.
	if len(pos.counts) > 0 {
		if pos.minIdx == math.MaxInt32 {
			pos.minIdx = 0
		}
		if pos.maxIdx == math.MinInt32 {
			pos.maxIdx = -1
		}
		plen := pos.maxIdx - pos.minIdx + 1
		dp.Positive().SetOffset(int32(pos.minIdx))
		bc := dp.Positive().BucketCounts()
		bc.EnsureCapacity(plen)
		for i := 0; i < plen; i++ {
			bc.Append(pos.counts[pos.minIdx+i])
		}
	}

	// Materialize negative buckets.
	if len(neg.counts) > 0 {
		if neg.minIdx == math.MaxInt32 {
			neg.minIdx = 0
		}
		if neg.maxIdx == math.MinInt32 {
			neg.maxIdx = -1
		}
		nlen := neg.maxIdx - neg.minIdx + 1
		dp.Negative().SetOffset(int32(neg.minIdx))
		bc := dp.Negative().BucketCounts()
		bc.EnsureCapacity(nlen)
		for i := 0; i < nlen; i++ {
			bc.Append(neg.counts[neg.minIdx+i])
		}
	}

	dp.SetCount(uint64(math.Round(sk.GetCount())))
	dp.SetSum(sk.GetSum())
	if minValue, err := sk.GetMinValue(); err == nil {
		dp.SetMin(minValue)
	}
	if maxValue, err := sk.GetMaxValue(); err == nil {
		dp.SetMax(maxValue)
	}

	return nil
}

// ConvertExponentialHistogramToDDSketch reconstructs a DDSketch from an
// OTEL ExponentialHistogramDataPoint. It chooses a DDSketch mapping whose
// growth factor matches the OTEL base implied by dp.Scale, then replays
// bucket masses into the sketch using representative values m.Value(i).
func ConvertExponentialHistogramToDDSketch(dp pmetric.ExponentialHistogramDataPoint) (*ddsketch.DDSketch, error) {
	// If nothing recorded, return an empty-but-valid sketch with a reasonable α.
	if dp.Count() == 0 {
		m, _ := mapping.NewLogarithmicMapping(0.01)
		return ddsketch.NewDDSketchFromStoreProvider(m, store.DefaultProvider), nil
	}

	// OTEL base = 2^(1 / 2^scale)
	scale := dp.Scale()
	base := math.Exp2(1.0 / math.Exp2(float64(scale))) // 2^(1/2^scale)

	// Map OTEL base -> DDSketch relative accuracy α such that gamma ≈ base.
	// gamma = (1+α)/(1-α)  =>  α = (gamma-1)/(gamma+1) with gamma ≈ base
	alpha := (base - 1) / (base + 1)
	// Clamp α to sane bounds (avoid degenerate mappings).
	if !isFinite(alpha) || alpha <= 1e-9 {
		alpha = 1e-3
	}
	if alpha >= 0.5 { // very coarse; keep reasonable
		alpha = 0.5 - 1e-6
	}

	m, err := mapping.NewLogarithmicMapping(alpha)
	if err != nil {
		return nil, err
	}
	sk := ddsketch.NewDDSketchFromStoreProvider(m, store.DefaultProvider)
	if sk == nil {
		return nil, errors.New("failed to construct ddsketch")
	}

	addSide := func(side pmetric.ExponentialHistogramDataPointBuckets, negative bool) {
		if side.BucketCounts().Len() != 0 {
			offset := int(side.Offset())
			bc := side.BucketCounts()
			for i := 0; i < bc.Len(); i++ {
				c := bc.At(i)
				if c == 0 {
					continue
				}
				key := offset + i
				v := m.Value(key)
				if negative {
					v = -v
				}
				_ = sk.AddWithCount(v, float64(c))
			}
		}
	}

	// Positive and negative
	addSide(dp.Positive(), false)
	addSide(dp.Negative(), true)

	// Zeros (if any)
	if z := dp.ZeroCount(); z > 0 {
		_ = sk.AddWithCount(0, float64(z))
	}

	return sk, nil
}

func isFinite(f float64) bool {
	return !math.IsNaN(f) && !math.IsInf(f, 0)
}
