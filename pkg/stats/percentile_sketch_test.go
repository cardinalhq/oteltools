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
	"github.com/cardinalhq/oteltools/pkg/translate"
	"go.opentelemetry.io/collector/pdata/ptrace"
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
		Attributes:      map[string]any{"service": "auth"},
		latencySketch:   ps,
		totalCount:      0,
		totalErrorCount: 0,
	}

	require.NoError(t, span.Update(100, false))
	require.NoError(t, span.Update(200, false))
	require.NoError(t, span.Update(300, true))

	assert.Equal(t, int64(3), span.totalCount, "Total count should be 3")
	assert.Equal(t, int64(1), span.totalErrorCount, "Error count should be 1")

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

	span1 := ptrace.NewSpan()
	span1.SetName("span1")
	span1.Attributes().PutStr("service", "auth")
	span1.Attributes().PutDouble(translate.CardinalFieldSpanDuration, 100)
	span1.Attributes().PutInt(translate.CardinalFieldFingerprint, 1)
	require.NoError(t, cache.UpdateSpanSketch(span1))

	span2 := ptrace.NewSpan()
	span2.SetName("span2")
	span2.Attributes().PutStr("service", "auth")
	span2.Attributes().PutInt(translate.CardinalFieldFingerprint, 1)
	span2.Attributes().PutDouble(translate.CardinalFieldSpanDuration, 100)
	require.NoError(t, cache.UpdateSpanSketch(span2))

	span3 := ptrace.NewSpan()
	span3.SetName("span3")
	span3.Attributes().PutStr("service", "billing")
	span3.Attributes().PutDouble(translate.CardinalFieldSpanDuration, 100)
	span3.Attributes().PutInt(translate.CardinalFieldFingerprint, 2)
	require.NoError(t, cache.UpdateSpanSketch(span3))

	span4 := ptrace.NewSpan()
	span4.SetName("span4")
	span4.Attributes().PutStr("service", "billing")
	span4.Attributes().PutInt(translate.CardinalFieldFingerprint, 2)
	span4.Attributes().PutDouble(translate.CardinalFieldSpanDuration, 100)
	require.NoError(t, cache.UpdateSpanSketch(span4))

	count := 0
	cache.sketches.Range(func(key, value interface{}) bool {
		count += 1
		return true
	})

	assert.Equal(t, 2, count, "Expected 2 sketches in cache")
	cache.flush()
	assert.Len(t, flushedSpans, 2, "Expected 2 flushed SpanSketches")
}

func TestSpanSketchMergeAndSerialization(t *testing.T) {
	spanSketch1, err := NewPercentileSketch()
	assert.NoError(t, err)

	span1 := &SpanSketch{
		fingerprint:                 12345,
		Attributes:                  map[string]any{"service": "auth", "method": "POST"},
		latencySketch:               spanSketch1,
		totalCount:                  100,
		totalErrorCount:             5,
		exceptionsByFingerprint:     map[int64]string{101: "NullPointerException"},
		exceptionCountByFingerprint: map[int64]int64{101: 3},
	}

	err = span1.latencySketch.Add(120.0)
	assert.NoError(t, err)

	spanSketch2, err := NewPercentileSketch()
	assert.NoError(t, err)

	span2 := &SpanSketch{
		fingerprint:                 12345,
		Attributes:                  map[string]any{"service": "auth", "method": "POST"},
		latencySketch:               spanSketch2,
		totalCount:                  50,
		totalErrorCount:             3,
		exceptionsByFingerprint:     map[int64]string{102: "TimeoutException"},
		exceptionCountByFingerprint: map[int64]int64{102: 2},
	}

	err = span2.latencySketch.Add(200.0)
	assert.NoError(t, err)

	c := NewSketchCache(999*time.Hour, nil)
	err = c.MergeSpanSketch(span1)
	assert.NoError(t, err)
	err = c.MergeSpanSketch(span2)
	assert.NoError(t, err)

	assert.Equal(t, int64(150), span1.totalCount, "Total count should be merged correctly")
	assert.Equal(t, int64(8), span1.totalErrorCount, "Total error count should be merged correctly")

	assert.Equal(t, "NullPointerException", span1.exceptionsByFingerprint[101])
	assert.Equal(t, "TimeoutException", span1.exceptionsByFingerprint[102])
	assert.Equal(t, int64(3), span1.exceptionCountByFingerprint[101])
	assert.Equal(t, int64(2), span1.exceptionCountByFingerprint[102])

	p50, err := span1.latencySketch.GetPercentile(0.5)
	assert.NoError(t, err)
	assert.Greater(t, p50, 100.0, "P50 latency should reflect merged latencies")

	serialized, err := span1.Serialize()
	assert.NoError(t, err)

	deserializedSpan, err := DeserializeSpanSketch(serialized)
	assert.NoError(t, err)

	assert.Equal(t, span1.fingerprint, deserializedSpan.fingerprint)
	assert.Equal(t, span1.totalCount, deserializedSpan.totalCount)
	assert.Equal(t, span1.totalErrorCount, deserializedSpan.totalErrorCount)
	assert.Equal(t, span1.exceptionsByFingerprint, deserializedSpan.exceptionsByFingerprint)
	assert.Equal(t, span1.exceptionCountByFingerprint, deserializedSpan.exceptionCountByFingerprint)

	p50AfterDeserialize, err := deserializedSpan.latencySketch.GetPercentile(0.5)
	assert.NoError(t, err)
	assert.Greater(t, p50AfterDeserialize, 100.0, "P50 latency should persist after deserialization")
}
