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
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type MetricScopeBuilder struct {
	scope                 pmetric.ScopeMetrics
	sums                  map[uint64]MetricSumBuilder
	gauges                map[uint64]MetricGaugeBuilder
	histograms            map[uint64]MetricHistogramBuilder
	exponentialHistograms map[uint64]MetricExponentialHistogramBuilder
	summaries             map[uint64]MetricSummaryBuilder
}

func NewMetricScopeBuilder(scope pmetric.ScopeMetrics) *MetricScopeBuilder {
	return &MetricScopeBuilder{
		scope:                 scope,
		sums:                  make(map[uint64]MetricSumBuilder),
		gauges:                make(map[uint64]MetricGaugeBuilder),
		summaries:             make(map[uint64]MetricSummaryBuilder),
		histograms:            make(map[uint64]MetricHistogramBuilder),
		exponentialHistograms: make(map[uint64]MetricExponentialHistogramBuilder),
	}
}

func (msb *MetricScopeBuilder) Gauge(name string) *MetricGaugeBuilder {
	key := metrickey(name, "", pmetric.MetricTypeGauge)
	if item, ok := msb.gauges[key]; ok {
		return &item
	}
	metric := msb.scope.Metrics().AppendEmpty()
	metric.SetName(name)
	item := NewMetricGaugeBuilder(metric)
	msb.gauges[key] = *item
	return item
}

func (msb *MetricScopeBuilder) Sum(name string) *MetricSumBuilder {
	key := metrickey(name, "", pmetric.MetricTypeSum)
	if item, ok := msb.sums[key]; ok {
		return &item
	}
	metric := msb.scope.Metrics().AppendEmpty()
	metric.SetName(name)
	item := NewMetricSumBuilder(metric)
	msb.sums[key] = *item
	return item
}

func (msb *MetricScopeBuilder) Summary(name string) *MetricSummaryBuilder {
	key := metrickey(name, "", pmetric.MetricTypeSummary)
	if item, ok := msb.summaries[key]; ok {
		return &item
	}
	metric := msb.scope.Metrics().AppendEmpty()
	metric.SetName(name)
	item := NewMetricSummaryBuilder(metric)
	msb.summaries[key] = *item
	return item
}

func (msb *MetricScopeBuilder) Histogram(name string) *MetricHistogramBuilder {
	key := metrickey(name, "", pmetric.MetricTypeHistogram)
	if item, ok := msb.histograms[key]; ok {
		return &item
	}
	metric := msb.scope.Metrics().AppendEmpty()
	metric.SetName(name)
	item := NewMetricHistogramBuilder(metric)
	msb.histograms[key] = *item
	return item
}

func (msb *MetricScopeBuilder) ExponentialHistogram(name string) *MetricExponentialHistogramBuilder {
	key := metrickey(name, "", pmetric.MetricTypeExponentialHistogram)
	if item, ok := msb.exponentialHistograms[key]; ok {
		return &item
	}
	metric := msb.scope.Metrics().AppendEmpty()
	metric.SetName(name)
	item := NewMetricExponentialHistogramBuilder(metric)
	msb.exponentialHistograms[key] = *item
	return item
}

func (msb *MetricScopeBuilder) Get() pmetric.ScopeMetrics {
	return msb.scope
}
