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

type LogBuilder struct {
	pm       plog.Logs
	builders map[uint64]*LogResourceBuilder
}

func NewLogBuilder() *LogBuilder {
	return &LogBuilder{
		pm:       plog.NewLogs(),
		builders: map[uint64]*LogResourceBuilder{},
	}
}

func (lb *LogBuilder) Resource(rattr pcommon.Map) *LogResourceBuilder {
	key := attrkey(rattr)
	if item, ok := lb.builders[key]; ok {
		return item
	}
	resource := lb.pm.ResourceLogs().AppendEmpty()
	rattr.CopyTo(resource.Resource().Attributes())
	item := NewLogResourceBuilder(resource)
	lb.builders[key] = item
	return item
}

func (lb *LogBuilder) Build() plog.Logs {
	return removeEmptyLogs(lb.pm)
}

func removeEmptyLogs(pm plog.Logs) plog.Logs {
	pm.ResourceLogs().RemoveIf(func(rl plog.ResourceLogs) bool {
		rl.ScopeLogs().RemoveIf(func(sl plog.ScopeLogs) bool {
			return sl.LogRecords().Len() == 0
		})
		return rl.ScopeLogs().Len() == 0
	})
	return pm
}
