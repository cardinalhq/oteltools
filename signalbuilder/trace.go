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

type TracesBuilder struct {
	pm       ptrace.Traces
	builders map[uint64]*TraceResourceBuilder
}

func NewTracesBuilder() *TracesBuilder {
	return &TracesBuilder{
		pm:       ptrace.NewTraces(),
		builders: map[uint64]*TraceResourceBuilder{},
	}
}

func (mb *TracesBuilder) Resource(rattr pcommon.Map) *TraceResourceBuilder {
	key := attrkey(rattr)
	if item, ok := mb.builders[key]; ok {
		return item
	}
	resource := mb.pm.ResourceSpans().AppendEmpty()
	rattr.CopyTo(resource.Resource().Attributes())
	item := NewTraceResourceBuilder(resource)
	mb.builders[key] = item
	return item
}

func (mb *TracesBuilder) Build() ptrace.Traces {
	return removeEmptyTraces(mb.pm)
}

func removeEmptyTraces(pm ptrace.Traces) ptrace.Traces {
	pm.ResourceSpans().RemoveIf(func(rm ptrace.ResourceSpans) bool {
		rm.ScopeSpans().RemoveIf(func(sm ptrace.ScopeSpans) bool {
			return sm.Spans().Len() == 0
		})
		return rm.ScopeSpans().Len() == 0
	})
	return pm
}
