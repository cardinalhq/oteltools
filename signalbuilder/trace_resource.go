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
	"go.opentelemetry.io/collector/pdata/ptrace"
)

type TraceResourceBuilder struct {
	resource ptrace.ResourceSpans
	scopes   map[uint64]*TraceScopeBuilder
}

func NewTraceResourceBuilder(resource ptrace.ResourceSpans) *TraceResourceBuilder {
	return &TraceResourceBuilder{
		resource: resource,
		scopes:   map[uint64]*TraceScopeBuilder{},
	}
}

func (mrb *TraceResourceBuilder) Scope(sattr pcommon.Map) *TraceScopeBuilder {
	return mrb.ScopeWithInfo("", "", "", sattr)
}

func (mrb *TraceResourceBuilder) ScopeWithInfo(name, version, schemaURL string, sattr pcommon.Map) *TraceScopeBuilder {
	key := scopekey(name, version, schemaURL, sattr)
	if item, ok := mrb.scopes[key]; ok {
		return item
	}
	scope := mrb.resource.ScopeSpans().AppendEmpty()
	sattr.CopyTo(scope.Scope().Attributes())

	// Set the scope name, version, and schema URL
	instrScope := scope.Scope()
	if name != "" {
		instrScope.SetName(name)
	}
	if version != "" {
		instrScope.SetVersion(version)
	}
	if schemaURL != "" {
		scope.SetSchemaUrl(schemaURL)
	}

	item := NewTraceScopeBuilder(scope)
	mrb.scopes[key] = item
	return item
}
