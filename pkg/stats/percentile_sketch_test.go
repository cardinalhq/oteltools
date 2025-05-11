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

package stats

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/ptrace"

	"github.com/cardinalhq/oteltools/pkg/translate"
)

func TestPercentileSketchBasic(t *testing.T) {
	ps, err := NewPercentileSketch()
	require.NoError(t, err)

	require.NoError(t, ps.Add(100))
	require.NoError(t, ps.Add(200))
	require.NoError(t, ps.Add(300))

	p50, err := ps.Quantile(0.50)
	require.NoError(t, err)
	assert.Greater(t, p50, 100.0)
	assert.Less(t, p50, 300.0)

	p95, err := ps.Quantile(0.95)
	require.NoError(t, err)
	assert.InEpsilon(t, 200.0, p95, 0.05)

	data := ps.Encode()
	dps, err := DecodeSketch(data)
	require.NoError(t, err)

	dP95, err := dps.Quantile(0.95)
	require.NoError(t, err)
	assert.InEpsilon(t, p95, dP95, 0.01)
}

func TestMergeEncodedSketch(t *testing.T) {
	ps1, err := NewPercentileSketch()
	require.NoError(t, err)
	ps2, err := NewPercentileSketch()
	require.NoError(t, err)

	require.NoError(t, ps1.Add(100))
	require.NoError(t, ps1.Add(200))
	require.NoError(t, ps2.Add(300))

	data1 := ps1.Encode()
	data2 := ps2.Encode()

	merged, err := MergeEncodedSketch(data1, data2)
	require.NoError(t, err)

	dps, err := DecodeSketch(merged)
	require.NoError(t, err)

	p50, err := dps.Quantile(0.50)
	require.NoError(t, err)
	// p50 should lie between the inputs
	assert.GreaterOrEqual(t, p50, 100.0)
	assert.LessOrEqual(t, p50, 300.0)
}

func TestSketchCache_FlushAndGrouping(t *testing.T) {
	var flushed *SpanSketchList
	flushFunc := func(lst *SpanSketchList) {
		flushed = lst
	}

	cache := NewSketchCache(time.Minute, flushFunc)

	// Span A: auth service
	spanA := ptrace.NewSpan()
	spanA.Attributes().PutDouble(translate.CardinalFieldSpanDuration, 100)
	spanA.Status().SetCode(ptrace.StatusCodeOk)
	tagsA := map[string]string{"service.name": "auth"}
	cache.Update("db.calls", tagsA, spanA)

	// Span B: billing service
	spanB := ptrace.NewSpan()
	spanB.Attributes().PutDouble(translate.CardinalFieldSpanDuration, 200)
	spanB.Status().SetCode(ptrace.StatusCodeOk)
	tagsB := map[string]string{"service.name": "billing"}
	cache.Update("db.calls", tagsB, spanB)

	// Flush manually
	cache.flush()

	// Expect two sketches
	require.Len(t, flushed.Sketches, 2)

	seen := map[string]bool{}
	for _, proto := range flushed.Sketches {
		serviceTag := proto.Tags["service.name"]
		seen[serviceTag] = true
		require.Equal(t, "db.calls", proto.MetricName)
		require.NotEmpty(t, proto.Sketch)
		require.Equal(t, int64(1), proto.TotalCount)
	}
	assert.True(t, seen["auth"])
	assert.True(t, seen["billing"])
}
