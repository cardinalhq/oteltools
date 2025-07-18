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
	"go.opentelemetry.io/collector/pdata/pcommon"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"testing"
	"time"

	"github.com/DataDog/sketches-go/ddsketch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/ptrace"

	"github.com/cardinalhq/oteltools/pkg/translate"
)

func TestEncodeDecodeSketch(t *testing.T) {
	sk, err := ddsketch.NewDefaultDDSketch(0.01)
	require.NoError(t, err)

	require.NoError(t, sk.Add(100))
	require.NoError(t, sk.Add(200))
	require.NoError(t, sk.Add(300))

	p50, err := sk.GetValueAtQuantile(0.5)
	require.NoError(t, err)
	assert.Greater(t, p50, 100.0)
	assert.Less(t, p50, 300.0)

	p95, err := sk.GetValueAtQuantile(0.95)
	require.NoError(t, err)
	assert.InEpsilon(t, 200.0, p95, 0.05)

	data := Encode(sk)
	dsk, err := DecodeSketch(data)
	require.NoError(t, err)

	dP95, err := dsk.GetValueAtQuantile(0.95)
	require.NoError(t, err)
	assert.InEpsilon(t, p95, dP95, 0.01)
}

func TestMergeEncodedSketch(t *testing.T) {
	skA, err := ddsketch.NewDefaultDDSketch(0.01)
	require.NoError(t, err)
	skB, err := ddsketch.NewDefaultDDSketch(0.01)
	require.NoError(t, err)

	require.NoError(t, skA.Add(100))
	require.NoError(t, skA.Add(200))
	require.NoError(t, skB.Add(300))

	dataA := Encode(skA)
	dataB := Encode(skB)

	merged, err := MergeEncodedSketch(dataA, dataB)
	require.NoError(t, err)

	dsk, err := DecodeSketch(merged)
	require.NoError(t, err)

	p50, err := dsk.GetValueAtQuantile(0.5)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, p50, 100.0)
	assert.LessOrEqual(t, p50, 300.0)
}

func TestSketchCache_FlushAndGrouping(t *testing.T) {
	var flushed *SpanSketchList
	flushFn := func(lst *SpanSketchList) error {
		flushed = lst
		return nil
	}

	res := pcommon.NewResource()
	res.Attributes().PutStr(string(semconv.ServiceNameKey), "auth")
	cache := NewSpanSketchCache(time.Minute, "cust1", 2, flushFn)

	// Span A: auth service
	spanA := ptrace.NewSpan()
	spanA.Attributes().PutDouble(translate.CardinalFieldSpanDuration, 100)
	spanA.Status().SetCode(ptrace.StatusCodeOk)
	tagsA := map[string]string{"service.name": "auth"}
	cache.Update("db.calls", "count", tagsA, spanA, res, 0, 0)

	// Span B: billing service
	res = pcommon.NewResource()
	res.Attributes().PutStr(string(semconv.ServiceNameKey), "billing")
	spanB := ptrace.NewSpan()
	spanB.Attributes().PutDouble(translate.CardinalFieldSpanDuration, 200)
	spanB.Status().SetCode(ptrace.StatusCodeOk)
	tagsB := map[string]string{"service.name": "billing"}
	cache.Update("db.calls", "count", tagsB, spanB, res, 0, 0)

	// Flush manually
	cache.flush()

	require.NotNil(t, flushed)
	assert.Equal(t, "cust1", flushed.CustomerId)
	require.Len(t, flushed.Sketches, 2)

	seen := map[string]bool{}
	for _, proto := range flushed.Sketches {
		serviceTag := proto.Tags["service.name"]
		seen[serviceTag] = true
		require.Equal(t, "db.calls", proto.MetricName)
		require.NotEmpty(t, proto.Sketch)
		require.Equal(t, int64(1), proto.TotalCount)
		require.Equal(t, int64(0), proto.ErrorCount)
		require.Equal(t, int64(0), proto.ExceptionCount)
	}
	assert.True(t, seen["auth"])
	assert.True(t, seen["billing"])
}
