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

type MetricExponentialHistogramBuilder struct {
	metric     pmetric.Metric
	datapoints map[uint64]pmetric.ExponentialHistogramDataPoint
}

func (mehb *MetricExponentialHistogramBuilder) SetDescription(description string) {
	mehb.metric.SetDescription(description)
}

func (mehb *MetricExponentialHistogramBuilder) SetUnit(unit string) {
	mehb.metric.SetUnit(unit)
}

func NewMetricExponentialHistogramBuilder(metric pmetric.Metric) *MetricExponentialHistogramBuilder {
	histogram := metric.SetEmptyExponentialHistogram()
	histogram.SetAggregationTemporality(pmetric.AggregationTemporalityDelta)
	return &MetricExponentialHistogramBuilder{
		metric:     metric,
		datapoints: map[uint64]pmetric.ExponentialHistogramDataPoint{},
	}
}

func (mehb *MetricExponentialHistogramBuilder) Datapoint(attr pcommon.Map, timestamp pcommon.Timestamp) pmetric.ExponentialHistogramDataPoint {
	key := attrkey(attr) + uint64(timestamp)
	if item, ok := mehb.datapoints[key]; ok {
		return item
	}
	datapoint := mehb.metric.ExponentialHistogram().DataPoints().AppendEmpty()
	attr.CopyTo(datapoint.Attributes())
	datapoint.SetStartTimestamp(timestamp)
	datapoint.SetTimestamp(timestamp)
	mehb.datapoints[key] = datapoint
	return datapoint
}