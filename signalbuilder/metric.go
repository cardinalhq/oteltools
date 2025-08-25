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

type MetricsBuilder struct {
	pm       pmetric.Metrics
	builders map[uint64]*MetricResourceBuilder
}

func NewMetricsBuilder() *MetricsBuilder {
	return &MetricsBuilder{
		pm:       pmetric.NewMetrics(),
		builders: map[uint64]*MetricResourceBuilder{},
	}
}

func (mb *MetricsBuilder) Resource(rattr pcommon.Map) *MetricResourceBuilder {
	key := attrkey(rattr)
	if item, ok := mb.builders[key]; ok {
		return item
	}
	resource := mb.pm.ResourceMetrics().AppendEmpty()
	rattr.CopyTo(resource.Resource().Attributes())
	item := NewMetricResourceBuilder(resource)
	mb.builders[key] = item
	return item
}

func (mb *MetricsBuilder) Add(rm *ResourceMetrics) error {
	resourceAttrs, err := fromRaw(rm.Resource)
	if err != nil {
		return fmt.Errorf("failed to convert resource attributes: %w", err)
	}

	resourceBuilder := mb.Resource(resourceAttrs)

	for _, scopeMetric := range rm.ScopeMetrics {
		scopeAttrs, err := fromRaw(scopeMetric.Attributes)
		if err != nil {
			return fmt.Errorf("failed to convert scope attributes: %w", err)
		}

		scopeBuilder := resourceBuilder.ScopeWithInfo(
			scopeMetric.Name,
			scopeMetric.Version,
			scopeMetric.SchemaURL,
			scopeAttrs,
		)

		for _, m := range scopeMetric.Metrics {
			switch m.Type {
			case "gauge":
				if err := mb.addGaugeMetric(scopeBuilder, &m); err != nil {
					return fmt.Errorf("failed to add gauge metric '%s': %w", m.Name, err)
				}
			case "sum":
				if err := mb.addSumMetric(scopeBuilder, &m); err != nil {
					return fmt.Errorf("failed to add sum metric '%s': %w", m.Name, err)
				}
			default:
				return fmt.Errorf("unsupported metric type '%s' for metric '%s'", m.Type, m.Name)
			}
		}
	}

	return nil
}

func (mb *MetricsBuilder) addGaugeMetric(scopeBuilder *MetricScopeBuilder, m *Metric) error {
	gaugeBuilder := scopeBuilder.Gauge(m.Name)
	gaugeBuilder.SetDescription(m.Description)
	gaugeBuilder.SetUnit(m.Unit)

	for _, dp := range m.Gauge.DataPoints {
		attrs, err := fromRaw(dp.Attributes)
		if err != nil {
			return fmt.Errorf("failed to convert data point attributes: %w", err)
		}

		datapoint, _, _ := gaugeBuilder.Datapoint(attrs, pcommon.Timestamp(dp.Timestamp))
		datapoint.SetDoubleValue(dp.Value)
		datapoint.SetStartTimestamp(pcommon.Timestamp(dp.StartTimestamp))
		datapoint.SetFlags(pmetric.DataPointFlags(dp.Flags))
	}

	return nil
}

func (mb *MetricsBuilder) addSumMetric(scopeBuilder *MetricScopeBuilder, m *Metric) error {
	sumBuilder := scopeBuilder.Sum(m.Name)
	sumBuilder.SetDescription(m.Description)
	sumBuilder.SetUnit(m.Unit)

	sum := sumBuilder.metric.Sum()
	sum.SetIsMonotonic(m.Sum.IsMonotonic)
	
	switch m.Sum.AggregationTemporality {
	case "cumulative":
		sum.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
	case "delta":
		sum.SetAggregationTemporality(pmetric.AggregationTemporalityDelta)
	case "":
		// Keep default from NewMetricSumBuilder (delta)
	default:
		return fmt.Errorf("unsupported aggregation temporality '%s'", m.Sum.AggregationTemporality)
	}

	for _, dp := range m.Sum.DataPoints {
		attrs, err := fromRaw(dp.Attributes)
		if err != nil {
			return fmt.Errorf("failed to convert data point attributes: %w", err)
		}

		datapoint, _, _ := sumBuilder.Datapoint(attrs, pcommon.Timestamp(dp.Timestamp))
		datapoint.SetDoubleValue(dp.Value)
		datapoint.SetStartTimestamp(pcommon.Timestamp(dp.StartTimestamp))
		datapoint.SetFlags(pmetric.DataPointFlags(dp.Flags))
	}

	return nil
}

// AddFromYAML parses YAML data and adds the metrics to the builder.
// Note: JSON is a subset of YAML, so this function can also accept JSON format data.
func (mb *MetricsBuilder) AddFromYAML(data []byte, opts ...ParseOptions) error {
	rm, err := ParseMetrics(data, opts...)
	if err != nil {
		return err
	}
	return mb.Add(rm)
}

func (mb *MetricsBuilder) Build() pmetric.Metrics {
	return removeEmptyMetrics(mb.pm)
}

func removeEmptyMetrics(pm pmetric.Metrics) pmetric.Metrics {
	pm.ResourceMetrics().RemoveIf(func(rm pmetric.ResourceMetrics) bool {
		rm.ScopeMetrics().RemoveIf(func(sm pmetric.ScopeMetrics) bool {
			sm.Metrics().RemoveIf(func(m pmetric.Metric) bool {
				switch m.Type() {
				case pmetric.MetricTypeEmpty:
					return true
				case pmetric.MetricTypeGauge:
					return m.Gauge().DataPoints().Len() == 0
				case pmetric.MetricTypeSum:
					return m.Sum().DataPoints().Len() == 0
				case pmetric.MetricTypeHistogram:
					return m.Histogram().DataPoints().Len() == 0
				case pmetric.MetricTypeExponentialHistogram:
					return m.ExponentialHistogram().DataPoints().Len() == 0
				case pmetric.MetricTypeSummary:
					return m.Summary().DataPoints().Len() == 0
				}
				return false
			})
			return sm.Metrics().Len() == 0
		})
		return rm.ScopeMetrics().Len() == 0
	})
	return pm
}
