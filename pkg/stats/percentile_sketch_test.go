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
)

func TestPercentileSketch(t *testing.T) {
	ps, err := NewPercentileSketch()
	require.NoError(t, err)

	require.NoError(t, ps.Add(100))
	require.NoError(t, ps.Add(200))
	require.NoError(t, ps.Add(300))

	p50, err := ps.GetPercentile(0.50)
	require.NoError(t, err)
	assert.Greater(t, p50, 100.0)
	assert.Less(t, p50, 300.0)

	p95, err := ps.GetPercentile(0.95)
	require.NoError(t, err)
	assert.InEpsilon(t, 200.0, p95, 0.05)

	serialized := ps.Serialize()

	deserializedPs, err := DeserializePercentileSketch(serialized)
	require.NoError(t, err)

	deserializedP95, err := deserializedPs.GetPercentile(0.95)
	require.NoError(t, err)
	assert.InEpsilon(t, p95, deserializedP95, 0.01)
}

func TestSpanSketch(t *testing.T) {
	ps, err := NewPercentileSketch()
	require.NoError(t, err)

	span := &SpanSketch{
		Tags:          map[string]string{"service": "auth"},
		latencySketch: ps,
		totalCount:    0,
		errorCount:    0,
	}

	require.NoError(t, span.UpdateSpan(100, false))
	require.NoError(t, span.UpdateSpan(200, false))
	require.NoError(t, span.UpdateSpan(300, true))

	assert.Equal(t, int64(3), span.totalCount, "Total count should be 3")
	assert.Equal(t, int64(1), span.errorCount, "Error count should be 1")

	p95, err := span.latencySketch.GetPercentile(0.95)
	require.NoError(t, err)
	assert.InEpsilon(t, 200.0, p95, 0.05)
	assert.Less(t, p95, 300.0)
}

func TestSketchCacheFlush(t *testing.T) {
	var flushedSpans []*SpanSketch

	flushFunc := func(ts time.Time, spans []*SpanSketch) {
		flushedSpans = append(flushedSpans, spans...)
	}

	cache := NewSketchCache(999*time.Hour, flushFunc)

	require.NoError(t, cache.UpdateSpanSketch("span1", map[string]string{"service": "auth"}, 150, false))
	require.NoError(t, cache.UpdateSpanSketch("span1", map[string]string{"service": "auth"}, 250, false))
	require.NoError(t, cache.UpdateSpanSketch("span2", map[string]string{"service": "billing"}, 50, false))
	require.NoError(t, cache.UpdateSpanSketch("span2", map[string]string{"service": "billing"}, 300, true))

	count := 0
	cache.sketches.Range(func(key, value interface{}) bool {
		count += 1
		return true
	})

	assert.Equal(t, count, 2, "Expected 2 sketches in cache")
	cache.flush()
	assert.Len(t, flushedSpans, 2, "Expected 2 flushed SpanSketches")
}
