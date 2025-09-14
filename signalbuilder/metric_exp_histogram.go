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
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type MetricExpHistogramBuilder struct {
	metric      pmetric.Metric
	datapoints  map[uint64]pmetric.ExponentialHistogramDataPoint
	scale       int32
	temporality pmetric.AggregationTemporality
}

var _ MetricExponentialHistogramBuilder = (*MetricExpHistogramBuilder)(nil)

// NewMetricExponentialHistogramBuilder creates an ExponentialHistogram metric with sane defaults.
func NewMetricExponentialHistogramBuilder(metric pmetric.Metric) *MetricExpHistogramBuilder {
	eh := metric.SetEmptyExponentialHistogram()
	eh.SetAggregationTemporality(pmetric.AggregationTemporalityDelta)

	return &MetricExpHistogramBuilder{
		metric:      metric,
		datapoints:  map[uint64]pmetric.ExponentialHistogramDataPoint{},
		scale:       8,
		temporality: pmetric.AggregationTemporalityDelta,
	}
}

func (b *MetricExpHistogramBuilder) SetScale(scale int32) *MetricExpHistogramBuilder {
	b.scale = scale
	return b
}

func (b *MetricExpHistogramBuilder) SetAggregationTemporality(t pmetric.AggregationTemporality) *MetricExpHistogramBuilder {
	b.temporality = t
	b.metric.ExponentialHistogram().SetAggregationTemporality(t)
	return b
}

// Datapoint returns/creates a datapoint for (attrs,timestamp).
func (b *MetricExpHistogramBuilder) Datapoint(attr pcommon.Map, timestamp pcommon.Timestamp) (dp pmetric.ExponentialHistogramDataPoint, ty pmetric.MetricType, isNew bool) {
	key := attrkey(attr) + uint64(timestamp)
	if item, ok := b.datapoints[key]; ok {
		return item, pmetric.MetricTypeExponentialHistogram, false
	}

	datapoint := b.metric.ExponentialHistogram().DataPoints().AppendEmpty()
	attr.CopyTo(datapoint.Attributes())
	datapoint.SetStartTimestamp(timestamp)
	datapoint.SetTimestamp(timestamp)
	datapoint.SetScale(b.scale)

	_ = datapoint.Positive().BucketCounts() // touch to init
	_ = datapoint.Negative().BucketCounts()

	b.datapoints[key] = datapoint
	return datapoint, pmetric.MetricTypeExponentialHistogram, true
}
