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
	"testing"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

// BenchmarkMetrics100 benchmarks building 100 unique metrics with datapoints
func BenchmarkMetrics100(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		mb := NewMetricsBuilder()

		// Create resource attributes
		rattr := pcommon.NewMap()
		rattr.PutStr("service.name", "test-service")
		rattr.PutStr("host.name", "test-host")
		r := mb.Resource(rattr)

		// Create scope attributes
		sattr := pcommon.NewMap()
		sattr.PutStr("scope.name", "test-scope")
		s := r.Scope(sattr)

		// Create 100 metrics with 1 datapoint each
		for i := 0; i < 100; i++ {
			m, err := s.Metric(fmt.Sprintf("metric_%d", i), "unit", pmetric.MetricTypeGauge)
			if err != nil {
				b.Fatal(err)
			}

			dpAttr := pcommon.NewMap()
			dpAttr.PutStr("label", fmt.Sprintf("value_%d", i))
			dp, _, _ := m.Datapoint(dpAttr, pcommon.Timestamp(1000000+i))
			dp.SetDoubleValue(float64(i))
		}

		_ = mb.Build()
	}
}

// BenchmarkMetrics100MultiDatapoints benchmarks 100 metrics with 10 datapoints each
func BenchmarkMetrics100MultiDatapoints(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		mb := NewMetricsBuilder()

		rattr := pcommon.NewMap()
		rattr.PutStr("service.name", "test-service")
		r := mb.Resource(rattr)

		sattr := pcommon.NewMap()
		s := r.Scope(sattr)

		for i := 0; i < 100; i++ {
			m, err := s.Metric(fmt.Sprintf("metric_%d", i), "unit", pmetric.MetricTypeGauge)
			if err != nil {
				b.Fatal(err)
			}

			for j := 0; j < 10; j++ {
				dpAttr := pcommon.NewMap()
				dpAttr.PutStr("instance", fmt.Sprintf("inst_%d", j))
				dp, _, _ := m.Datapoint(dpAttr, pcommon.Timestamp(1000000+j))
				dp.SetDoubleValue(float64(i*10 + j))
			}
		}

		_ = mb.Build()
	}
}

// BenchmarkAttrKey benchmarks the attrkey function specifically
func BenchmarkAttrKey(b *testing.B) {
	attr := pcommon.NewMap()
	attr.PutStr("service.name", "test-service")
	attr.PutStr("host.name", "test-host")
	attr.PutStr("label1", "value1")
	attr.PutStr("label2", "value2")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = attrkey(attr)
	}
}

// BenchmarkAttrKeyEmpty benchmarks attrkey with empty map
func BenchmarkAttrKeyEmpty(b *testing.B) {
	attr := pcommon.NewMap()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = attrkey(attr)
	}
}

// BenchmarkMetricKey benchmarks the metrickey function
func BenchmarkMetricKey(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = metrickey("test_metric_name", "seconds", pmetric.MetricTypeGauge)
	}
}

// BenchmarkResourceLookup benchmarks looking up existing resources
func BenchmarkResourceLookup(b *testing.B) {
	mb := NewMetricsBuilder()
	rattr := pcommon.NewMap()
	rattr.PutStr("service.name", "test-service")
	rattr.PutStr("host.name", "test-host")

	// Create the resource first
	_ = mb.Resource(rattr)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = mb.Resource(rattr)
	}
}

// BenchmarkBuildOnly benchmarks just the Build() call with pre-populated data
func BenchmarkBuildOnly(b *testing.B) {
	// Pre-build a populated MetricsBuilder
	mb := NewMetricsBuilder()
	rattr := pcommon.NewMap()
	rattr.PutStr("service.name", "test-service")
	r := mb.Resource(rattr)
	s := r.Scope(pcommon.NewMap())

	for i := 0; i < 100; i++ {
		m, _ := s.Metric(fmt.Sprintf("metric_%d", i), "unit", pmetric.MetricTypeGauge)
		dpAttr := pcommon.NewMap()
		dpAttr.PutStr("label", fmt.Sprintf("value_%d", i))
		dp, _, _ := m.Datapoint(dpAttr, pcommon.Timestamp(1000000+i))
		dp.SetDoubleValue(float64(i))
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = mb.Build()
	}
}
