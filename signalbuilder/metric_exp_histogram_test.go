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

package signalbuilder

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

func TestMetricExpHistogramBuilder_Datapoint(t *testing.T) {
	metric := pmetric.NewMetric()
	builder := NewMetricExponentialHistogramBuilder(metric)

	// default temporality set on construction
	assert.Equal(t, pmetric.AggregationTemporalityDelta, metric.ExponentialHistogram().AggregationTemporality())

	attr := pcommon.NewMap()
	attr.PutStr("key", "value")
	timestamp := pcommon.NewTimestampFromTime(time.Now())

	// create new datapoint
	dp, ty, isNew := builder.Datapoint(attr, timestamp)
	assert.True(t, isNew)
	assert.Equal(t, pmetric.MetricTypeExponentialHistogram, ty)
	assert.Equal(t, timestamp, dp.Timestamp())
	assert.Equal(t, timestamp, dp.StartTimestamp())

	// default scale should be 8
	assert.Equal(t, int32(8), dp.Scale())

	// attributes copied
	v, found := dp.Attributes().Get("key")
	assert.True(t, found)
	assert.Equal(t, "value", v.AsString())

	// buckets are initialized and writable
	dp.Positive().SetOffset(5)
	dp.Positive().BucketCounts().Append(10)
	assert.Equal(t, int32(5), dp.Positive().Offset())
	assert.Equal(t, 1, dp.Positive().BucketCounts().Len())
	assert.Equal(t, uint64(10), dp.Positive().BucketCounts().At(0))

	// retrieving existing dp with same attrs + ts should reuse it
	dp2, ty2, isNew2 := builder.Datapoint(attr, timestamp)
	assert.False(t, isNew2)
	assert.Equal(t, pmetric.MetricTypeExponentialHistogram, ty2)
	assert.Equal(t, dp, dp2)

	// different attrs => new datapoint
	attr2 := pcommon.NewMap()
	attr2.PutStr("key", "value2")
	dp3, ty3, isNew3 := builder.Datapoint(attr2, timestamp)
	assert.True(t, isNew3)
	assert.Equal(t, pmetric.MetricTypeExponentialHistogram, ty3)
	assert.NotEqual(t, dp, dp3)
	v2, found2 := dp3.Attributes().Get("key")
	assert.True(t, found2)
	assert.Equal(t, "value2", v2.AsString())

	// change scale before creating another DP
	builder.SetScale(12)
	attr3 := pcommon.NewMap()
	attr3.PutStr("a", "b")
	ts2 := pcommon.NewTimestampFromTime(time.Now().Add(1 * time.Second))
	dp4, _, _ := builder.Datapoint(attr3, ts2)
	assert.Equal(t, int32(12), dp4.Scale())

	// change temporality
	builder.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
	assert.Equal(t, pmetric.AggregationTemporalityCumulative, metric.ExponentialHistogram().AggregationTemporality())
}

func BenchmarkExpHistogramDatapoint(b *testing.B) {
	metric := pmetric.NewMetric()
	builder := NewMetricExponentialHistogramBuilder(metric)

	attr := pcommon.NewMap()
	attr.PutStr("key", "value")
	timestamp := pcommon.NewTimestampFromTime(time.Now())

	b.ResetTimer()
	for b.Loop() {
		builder.Datapoint(attr, timestamp)
	}
}
