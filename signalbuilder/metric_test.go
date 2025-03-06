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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

func TestBuild(t *testing.T) {
	mb := NewMetricsBuilder()
	r := mb.Resource(pcommon.NewMap())
	s := r.Scope(pcommon.NewMap())
	m, err := s.Metric("alice", "s", pmetric.MetricTypeGauge)
	require.NoError(t, err)
	dp, ty, isNew := m.Datapoint(pcommon.NewMap(), 123)
	require.Equal(t, pmetric.MetricTypeGauge, ty)
	require.True(t, isNew)
	dp.SetDoubleValue(42)
	pm := mb.Build()
	require.Equal(t, 1, pm.ResourceMetrics().Len())
	require.Equal(t, 1, pm.ResourceMetrics().At(0).ScopeMetrics().Len())
	require.Equal(t, 1, pm.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().Len())
	require.Equal(t, 1, pm.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Gauge().DataPoints().Len())
	require.Equal(t, float64(42), pm.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Gauge().DataPoints().At(0).DoubleValue())

	// Add another resource with otherwise identical scopes and metrics.
	attr2 := pcommon.NewMap()
	attr2.PutStr("bob", "uncle")
	r2 := mb.Resource(attr2)
	s2 := r2.Scope(pcommon.NewMap())
	m2, err := s2.Metric("alice", "s", pmetric.MetricTypeGauge)
	require.NoError(t, err)
	dp2, ty2, isNew2 := m2.Datapoint(pcommon.NewMap(), 123)
	require.Equal(t, pmetric.MetricTypeGauge, ty2)
	require.True(t, isNew2)
	dp2.SetDoubleValue(43)
	pm2 := mb.Build()
	require.Equal(t, 2, pm2.ResourceMetrics().Len())
	require.Equal(t, 1, pm2.ResourceMetrics().At(0).ScopeMetrics().Len())
	require.Equal(t, 1, pm2.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().Len())
	require.Equal(t, 1, pm2.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Gauge().DataPoints().Len())
	require.Equal(t, float64(42), pm2.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Gauge().DataPoints().At(0).DoubleValue())
	require.Equal(t, 1, pm2.ResourceMetrics().At(1).ScopeMetrics().Len())
	require.Equal(t, 1, pm2.ResourceMetrics().At(1).ScopeMetrics().At(0).Metrics().Len())
	require.Equal(t, 1, pm2.ResourceMetrics().At(1).ScopeMetrics().At(0).Metrics().At(0).Gauge().DataPoints().Len())
	require.Equal(t, float64(43), pm2.ResourceMetrics().At(1).ScopeMetrics().At(0).Metrics().At(0).Gauge().DataPoints().At(0).DoubleValue())

	// Add another datapoint to the second resource, starting at the lookup of the resource.
	r2a := mb.Resource(attr2)
	assert.Equal(t, r2, r2a)
	s2a := r2a.Scope(pcommon.NewMap())
	assert.Equal(t, s2, s2a)
	m2a, err := s2a.Metric("alice", "s", pmetric.MetricTypeGauge)
	require.NoError(t, err)
	assert.Equal(t, m2, m2a)
	dp2a, ty2a, isNew2a := m2a.Datapoint(pcommon.NewMap(), 321)
	require.Equal(t, pmetric.MetricTypeGauge, ty2a)
	require.True(t, isNew2a)
	dp2a.SetDoubleValue(44)
	pm2a := mb.Build()
	require.Equal(t, 2, pm2a.ResourceMetrics().Len())
	require.Equal(t, 1, pm2a.ResourceMetrics().At(0).ScopeMetrics().Len())
	require.Equal(t, 1, pm2a.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().Len())
	require.Equal(t, 1, pm2a.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Gauge().DataPoints().Len())
	require.Equal(t, float64(42), pm2a.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Gauge().DataPoints().At(0).DoubleValue())
	require.Equal(t, 1, pm2a.ResourceMetrics().At(1).ScopeMetrics().Len())
	require.Equal(t, 1, pm2a.ResourceMetrics().At(1).ScopeMetrics().At(0).Metrics().Len())
	require.Equal(t, 2, pm2a.ResourceMetrics().At(1).ScopeMetrics().At(0).Metrics().At(0).Gauge().DataPoints().Len())
	require.Equal(t, float64(43), pm2a.ResourceMetrics().At(1).ScopeMetrics().At(0).Metrics().At(0).Gauge().DataPoints().At(0).DoubleValue())
	require.Equal(t, float64(44), pm2a.ResourceMetrics().At(1).ScopeMetrics().At(0).Metrics().At(0).Gauge().DataPoints().At(1).DoubleValue())
}

func BenchmarkBuilding(b *testing.B) {
	for b.Loop() {
		mb := NewMetricsBuilder()
		r := mb.Resource(pcommon.NewMap())
		s := r.Scope(pcommon.NewMap())
		m, err := s.Metric("alice", "s", pmetric.MetricTypeGauge)
		if err != nil {
			b.Fail()
		}
		dp, ty, isNew := m.Datapoint(pcommon.NewMap(), 123)
		if ty != pmetric.MetricTypeGauge || !isNew {
			b.Fail()
		}
		dp.SetDoubleValue(42)

		// Add another resource with otherwise identical scopes and metrics.
		attr2 := pcommon.NewMap()
		attr2.PutStr("bob", "uncle")
		r2 := mb.Resource(attr2)
		s2 := r2.Scope(pcommon.NewMap())
		m2, err := s2.Metric("alice", "s", pmetric.MetricTypeGauge)
		if err != nil {
			b.Fail()
		}
		dp2, ty2, isNew2 := m2.Datapoint(pcommon.NewMap(), 123)
		if ty2 != pmetric.MetricTypeGauge || !isNew2 {
			b.Fail()
		}
		dp2.SetDoubleValue(43)

		// Add another datapoint to the second resource, starting at the lookup of the resource.
		r2a := mb.Resource(attr2)
		if r2 != r2a {
			b.Fail()
		}
		s2a := r2a.Scope(pcommon.NewMap())
		if s2 != s2a {
			b.Fail()
		}
		m2a, err := s2a.Metric("alice", "s", pmetric.MetricTypeGauge)
		if err != nil {
			b.Fail()
		}
		dp2a, ty2a, isNew2a := m2a.Datapoint(pcommon.NewMap(), 321)
		if ty2a != pmetric.MetricTypeGauge || !isNew2a {
			b.Fail()
		}
		dp2a.SetDoubleValue(44)

		mb.Build()
	}
}
