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

type MetricGaugeBuilder struct {
	metric     pmetric.Metric
	datapoints map[uint64]pmetric.NumberDataPoint
}

func (mgb *MetricGaugeBuilder) SetDescription(description string) {
	mgb.metric.SetDescription(description)
}

func (mgb *MetricGaugeBuilder) SetUnit(unit string) {
	mgb.metric.SetUnit(unit)
}

func NewMetricGaugeBuilder(metric pmetric.Metric) *MetricGaugeBuilder {
	metric.SetEmptyGauge()
	return &MetricGaugeBuilder{
		metric:     metric,
		datapoints: map[uint64]pmetric.NumberDataPoint{},
	}
}

func (mb *MetricGaugeBuilder) Datapoint(attr pcommon.Map, timestamp pcommon.Timestamp) (dp pmetric.NumberDataPoint, ty pmetric.MetricType, isNew bool) {
	key := attrkey(attr) + uint64(timestamp)
	if item, ok := mb.datapoints[key]; ok {
		return item, pmetric.MetricTypeGauge, false
	}
	datapoint := mb.metric.Gauge().DataPoints().AppendEmpty()
	attr.CopyTo(datapoint.Attributes())
	datapoint.SetStartTimestamp(timestamp)
	datapoint.SetTimestamp(timestamp)
	mb.datapoints[key] = datapoint
	return datapoint, pmetric.MetricTypeGauge, true
}
