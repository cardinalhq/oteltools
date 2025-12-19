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

func TestAttrkey(t *testing.T) {
	t.Run("empty map returns 1", func(t *testing.T) {
		got := attrkey(pcommon.NewMap())
		assert.Equal(t, uint64(1), got)
	})

	t.Run("same map produces same hash", func(t *testing.T) {
		m := pcommon.NewMap()
		m.PutStr("key1", "value1")
		m.PutInt("key2", 123)

		hash1 := attrkey(m)
		hash2 := attrkey(m)
		assert.Equal(t, hash1, hash2)
	})

	t.Run("equivalent maps produce same hash", func(t *testing.T) {
		m1 := pcommon.NewMap()
		m1.PutStr("key1", "value1")
		m1.PutInt("key2", 123)

		m2 := pcommon.NewMap()
		m2.PutInt("key2", 123)
		m2.PutStr("key1", "value1")

		assert.Equal(t, attrkey(m1), attrkey(m2))
	})

	t.Run("different maps produce different hashes", func(t *testing.T) {
		m1 := pcommon.NewMap()
		m1.PutStr("key1", "value1")

		m2 := pcommon.NewMap()
		m2.PutStr("key1", "value2")

		require.NotEqual(t, attrkey(m1), attrkey(m2))
	})

	t.Run("non-empty map produces non-1 hash", func(t *testing.T) {
		m := pcommon.NewMap()
		m.PutStr("key1", "value1")
		require.NotEqual(t, uint64(1), attrkey(m))
	})
}

func TestMetrickey(t *testing.T) {
	t.Run("same inputs produce same hash", func(t *testing.T) {
		hash1 := metrickey("metric1", "ms", pmetric.MetricTypeGauge)
		hash2 := metrickey("metric1", "ms", pmetric.MetricTypeGauge)
		assert.Equal(t, hash1, hash2)
	})

	t.Run("different names produce different hashes", func(t *testing.T) {
		hash1 := metrickey("metric1", "ms", pmetric.MetricTypeGauge)
		hash2 := metrickey("metric2", "ms", pmetric.MetricTypeGauge)
		require.NotEqual(t, hash1, hash2)
	})

	t.Run("different units produce different hashes", func(t *testing.T) {
		hash1 := metrickey("metric1", "ms", pmetric.MetricTypeGauge)
		hash2 := metrickey("metric1", "s", pmetric.MetricTypeGauge)
		require.NotEqual(t, hash1, hash2)
	})

	t.Run("different types produce different hashes", func(t *testing.T) {
		hash1 := metrickey("metric1", "ms", pmetric.MetricTypeGauge)
		hash2 := metrickey("metric1", "ms", pmetric.MetricTypeSum)
		require.NotEqual(t, hash1, hash2)
	})
}

func BenchmarkAttrkey(b *testing.B) {
	attr := func() pcommon.Map {
		m := pcommon.NewMap()
		m.PutStr("key1", "value")
		m.PutInt("key2", 123)
		return m
	}()
	for b.Loop() {
		attrkey(attr)
	}
}

func BenchmarkMetrickey(b *testing.B) {
	for b.Loop() {
		metrickey("metric1", "ms", pmetric.MetricTypeGauge)
	}
}
