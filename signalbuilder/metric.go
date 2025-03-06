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

func (mb *MetricsBuilder) Build() pmetric.Metrics {
	return mb.pm
}
