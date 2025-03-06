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

	"github.com/cardinalhq/oteltools/hashutils"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

func TestAttrkey(t *testing.T) {
	tests := []struct {
		name string
		attr pcommon.Map
		want uint64
	}{
		{
			name: "empty map",
			attr: pcommon.NewMap(),
			want: 1,
		},
		{
			name: "non-empty map",
			attr: func() pcommon.Map {
				m := pcommon.NewMap()
				m.PutStr("key1", "value1")
				m.PutInt("key2", 123)
				return m
			}(),
			want: hashutils.HashAny(nil, map[string]any{
				"key2": int64(123),
				"key1": "value1",
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := attrkey(tt.attr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMetrickey(t *testing.T) {
	tests := []struct {
		name  string
		mname string
		units string
		ty    pmetric.MetricType
		want  uint64
	}{
		{
			name:  "empty metric",
			mname: "",
			units: "",
			ty:    pmetric.MetricTypeEmpty,
			want:  hashutils.HashStrings(nil, "", "", pmetric.MetricTypeEmpty.String()),
		},
		{
			name:  "non-empty metric",
			mname: "metric1",
			units: "ms",
			ty:    pmetric.MetricTypeGauge,
			want:  hashutils.HashStrings(nil, "metric1", "ms", pmetric.MetricTypeGauge.String()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := metrickey(tt.mname, tt.units, tt.ty)
			assert.Equal(t, tt.want, got)
		})
	}
}

func BenchmarkAttrkey(b *testing.B) {
	attr := func() pcommon.Map {
		m := pcommon.NewMap()
		m.PutStr("key1", "value")
		m.PutInt("key2", 123)
		return m
	}()
	b.ResetTimer()
	for b.Loop() {
		attrkey(attr)
	}
}

func BenchmarkMetrickey(b *testing.B) {
	b.ResetTimer()
	for b.Loop() {
		metrickey("metric1", "ms", pmetric.MetricTypeGauge)
	}
}
