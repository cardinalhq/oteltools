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

type MetricSummaryBuilder struct {
	metric     pmetric.Metric
	datapoints map[uint64]pmetric.SummaryDataPoint
}

func (msb *MetricSummaryBuilder) SetDescription(description string) {
	msb.metric.SetDescription(description)
}

func (msb *MetricSummaryBuilder) SetUnit(unit string) {
	msb.metric.SetUnit(unit)
}

func NewMetricSummaryBuilder(metric pmetric.Metric) *MetricSummaryBuilder {
	metric.SetEmptySummary()
	return &MetricSummaryBuilder{
		metric:     metric,
		datapoints: map[uint64]pmetric.SummaryDataPoint{},
	}
}

func (msb *MetricSummaryBuilder) Datapoint(attr pcommon.Map, timestamp pcommon.Timestamp) pmetric.SummaryDataPoint {
	key := attrkey(attr) + uint64(timestamp)
	if item, ok := msb.datapoints[key]; ok {
		return item
	}
	datapoint := msb.metric.Summary().DataPoints().AppendEmpty()
	attr.CopyTo(datapoint.Attributes())
	datapoint.SetStartTimestamp(timestamp)
	datapoint.SetTimestamp(timestamp)
	msb.datapoints[key] = datapoint
	return datapoint
}
