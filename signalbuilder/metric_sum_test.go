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

func TestDatapoint_NewDatapoint(t *testing.T) {
	metric := pmetric.NewMetric()
	builder := NewMetricSumBuilder(metric)

	attr := pcommon.NewMap()
	attr.PutStr("key", "value")
	timestamp := pcommon.NewTimestampFromTime(time.Now())

	dp, ty, isNew := builder.Datapoint(attr, timestamp)
	assert.Equal(t, pmetric.MetricTypeSum, ty)

	assert.True(t, isNew)
	assert.Equal(t, timestamp, dp.StartTimestamp())
	assert.Equal(t, timestamp, dp.Timestamp())
	v, found := dp.Attributes().Get("key")
	assert.True(t, found)
	assert.Equal(t, "value", v.AsString())
}

func TestDatapoint_ExistingDatapoint(t *testing.T) {
	metric := pmetric.NewMetric()
	builder := NewMetricSumBuilder(metric)

	attr := pcommon.NewMap()
	attr.PutStr("key", "value")
	timestamp := pcommon.NewTimestampFromTime(time.Now())

	dp1, ty1, isNew1 := builder.Datapoint(attr, timestamp)
	dp2, ty2, isNew2 := builder.Datapoint(attr, timestamp)
	assert.Equal(t, pmetric.MetricTypeSum, ty1)
	assert.Equal(t, pmetric.MetricTypeSum, ty2)
	assert.True(t, isNew1)
	assert.False(t, isNew2)
	assert.Equal(t, dp1, dp2)
}
