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

type MetricHistogramBuilder struct {
	metric     pmetric.Metric
	datapoints map[uint64]pmetric.HistogramDataPoint
}

func (mhb *MetricHistogramBuilder) SetDescription(description string) {
	mhb.metric.SetDescription(description)
}

func (mhb *MetricHistogramBuilder) SetUnit(unit string) {
	mhb.metric.SetUnit(unit)
}

func NewMetricHistogramBuilder(metric pmetric.Metric) *MetricHistogramBuilder {
	histogram := metric.SetEmptyHistogram()
	histogram.SetAggregationTemporality(pmetric.AggregationTemporalityDelta)
	return &MetricHistogramBuilder{
		metric:     metric,
		datapoints: map[uint64]pmetric.HistogramDataPoint{},
	}
}

func (mhb *MetricHistogramBuilder) Datapoint(attr pcommon.Map, timestamp pcommon.Timestamp) pmetric.HistogramDataPoint {
	key := attrkey(attr) + uint64(timestamp)
	if item, ok := mhb.datapoints[key]; ok {
		return item
	}
	datapoint := mhb.metric.Histogram().DataPoints().AppendEmpty()
	attr.CopyTo(datapoint.Attributes())
	datapoint.SetStartTimestamp(timestamp)
	datapoint.SetTimestamp(timestamp)
	mhb.datapoints[key] = datapoint
	return datapoint
}
