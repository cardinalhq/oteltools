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
	"go.opentelemetry.io/collector/pdata/plog"
)

type LogResourceBuilder struct {
	resource plog.ResourceLogs
	scopes   map[uint64]*LogScopeBuilder
}

func NewLogResourceBuilder(resource plog.ResourceLogs) *LogResourceBuilder {
	return &LogResourceBuilder{
		resource: resource,
		scopes:   map[uint64]*LogScopeBuilder{},
	}
}

func (lrb *LogResourceBuilder) Scope(sattr pcommon.Map) *LogScopeBuilder {
	return lrb.ScopeWithInfo("", "", "", sattr)
}

func (lrb *LogResourceBuilder) ScopeWithInfo(name, version, schemaURL string, sattr pcommon.Map) *LogScopeBuilder {
	key := scopekey(name, version, schemaURL, sattr)
	if item, ok := lrb.scopes[key]; ok {
		return item
	}
	scope := lrb.resource.ScopeLogs().AppendEmpty()
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

	item := NewLogScopeBuilder(scope)
	lrb.scopes[key] = item
	return item
}
