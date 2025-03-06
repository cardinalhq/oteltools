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

type MetricResourceBuilder struct {
	resource pmetric.ResourceMetrics
	scopes   map[uint64]*MetricScopeBuilder
}

func NewMetricResourceBuilder(resource pmetric.ResourceMetrics) *MetricResourceBuilder {
	return &MetricResourceBuilder{
		resource: resource,
		scopes:   map[uint64]*MetricScopeBuilder{},
	}
}

func (mrb *MetricResourceBuilder) Scope(sattr pcommon.Map) *MetricScopeBuilder {
	key := attrkey(sattr)
	if item, ok := mrb.scopes[key]; ok {
		return item
	}
	scope := mrb.resource.ScopeMetrics().AppendEmpty()
	sattr.CopyTo(scope.Scope().Attributes())
	item := NewMetricScopeBuilder(scope)
	mrb.scopes[key] = item
	return item
}
