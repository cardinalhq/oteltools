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
	"fmt"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type MetricDatapointBuilder interface {
	Datapoint(attr pcommon.Map, timestamp pcommon.Timestamp) (dp pmetric.NumberDataPoint, ty pmetric.MetricType, isNew bool)
}

type MetricScopeBuilder struct {
	scope   pmetric.ScopeMetrics
	metrics map[uint64]interface{}
}

func NewMetricScopeBuilder(scope pmetric.ScopeMetrics) *MetricScopeBuilder {
	return &MetricScopeBuilder{
		scope:   scope,
		metrics: make(map[uint64]interface{}),
	}
}

func (msb *MetricScopeBuilder) Metric(name string, units string, ty pmetric.MetricType) (MetricDatapointBuilder, error) {
	key := metrickey(name, units, ty)
	if item, ok := msb.metrics[key]; ok {
		if builder, ok := item.(MetricDatapointBuilder); ok {
			return builder, nil
		}
		return nil, fmt.Errorf("metric type mismatch")
	}
	metric := msb.scope.Metrics().AppendEmpty()
	metric.SetName(name)
	metric.SetUnit(units)
	var item interface{}
	switch ty {
	case pmetric.MetricTypeGauge:
		item = NewMetricGaugeBuilder(metric)
	case pmetric.MetricTypeSum:
		item = NewMetricSumBuilder(metric)
	case pmetric.MetricTypeSummary:
		item = NewMetricSummaryBuilder(metric)
	default:
		return nil, fmt.Errorf("unsupported metric type %s", ty.String())
	}
	msb.metrics[key] = item
	if builder, ok := item.(MetricDatapointBuilder); ok {
		return builder, nil
	}
	return nil, fmt.Errorf("internal error: created metric builder does not implement interface")
}

func (msb *MetricScopeBuilder) Gauge(name string) *MetricGaugeBuilder {
	key := metrickey(name, "", pmetric.MetricTypeGauge)
	if item, ok := msb.metrics[key]; ok {
		return item.(*MetricGaugeBuilder)
	}
	metric := msb.scope.Metrics().AppendEmpty()
	metric.SetName(name)
	item := NewMetricGaugeBuilder(metric)
	msb.metrics[key] = item
	return item
}

func (msb *MetricScopeBuilder) Sum(name string) *MetricSumBuilder {
	key := metrickey(name, "", pmetric.MetricTypeSum)
	if item, ok := msb.metrics[key]; ok {
		return item.(*MetricSumBuilder)
	}
	metric := msb.scope.Metrics().AppendEmpty()
	metric.SetName(name)
	item := NewMetricSumBuilder(metric)
	msb.metrics[key] = item
	return item
}

func (msb *MetricScopeBuilder) Summary(name string) *MetricSummaryBuilder {
	key := metrickey(name, "", pmetric.MetricTypeSummary)
	if item, ok := msb.metrics[key]; ok {
		return item.(*MetricSummaryBuilder)
	}
	metric := msb.scope.Metrics().AppendEmpty()
	metric.SetName(name)
	item := NewMetricSummaryBuilder(metric)
	msb.metrics[key] = item
	return item
}

func (msb *MetricScopeBuilder) Get() pmetric.ScopeMetrics {
	return msb.scope
}
