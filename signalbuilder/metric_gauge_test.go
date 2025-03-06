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

func TestMetricGaugeBuilder_Datapoint(t *testing.T) {
	metric := pmetric.NewMetric()
	builder := NewMetricGaugeBuilder(metric)

	attr := pcommon.NewMap()
	attr.PutStr("key", "value")
	timestamp := pcommon.NewTimestampFromTime(time.Now())

	// Test creating a new datapoint
	dp, ty, isNew := builder.Datapoint(attr, timestamp)
	assert.True(t, isNew)
	assert.Equal(t, pmetric.MetricTypeGauge, ty)
	assert.Equal(t, timestamp, dp.Timestamp())
	assert.Equal(t, timestamp, dp.StartTimestamp())
	v, found := dp.Attributes().Get("key")
	assert.True(t, found)
	assert.Equal(t, "value", v.AsString())

	// Test retrieving an existing datapoint
	dp2, ty2, isNew := builder.Datapoint(attr, timestamp)
	assert.False(t, isNew)
	assert.Equal(t, pmetric.MetricTypeGauge, ty2)
	assert.Equal(t, dp, dp2)

	// Test creating a new datapoint with different attributes
	attr2 := pcommon.NewMap()
	attr2.PutStr("key", "value2")
	dp3, ty2, isNew := builder.Datapoint(attr2, timestamp)
	assert.True(t, isNew)
	assert.Equal(t, pmetric.MetricTypeGauge, ty2)
	assert.NotEqual(t, dp, dp3)
	v, found = dp3.Attributes().Get("key")
	assert.True(t, found)
	assert.Equal(t, "value2", v.AsString())
}

func BenchmarkDatapoint(b *testing.B) {
	metric := pmetric.NewMetric()
	builder := NewMetricGaugeBuilder(metric)

	attr := pcommon.NewMap()
	attr.PutStr("key", "value")
	timestamp := pcommon.NewTimestampFromTime(time.Now())

	b.ResetTimer()
	for b.Loop() {
		builder.Datapoint(attr, timestamp)
	}
}
