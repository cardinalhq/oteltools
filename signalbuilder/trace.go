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
	"encoding/hex"
	"fmt"

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

func (tb *TracesBuilder) Add(rt *ResourceTraces) error {
	resourceAttrs, err := fromRaw(rt.Resource)
	if err != nil {
		return fmt.Errorf("failed to convert resource attributes: %w", err)
	}

	resourceBuilder := tb.Resource(resourceAttrs)

	for _, scopeTrace := range rt.ScopeTraces {
		scopeAttrs, err := fromRaw(scopeTrace.Attributes)
		if err != nil {
			return fmt.Errorf("failed to convert scope attributes: %w", err)
		}

		scopeBuilder := resourceBuilder.ScopeWithInfo(
			scopeTrace.Name,
			scopeTrace.Version,
			scopeTrace.SchemaURL,
			scopeAttrs,
		)
		for _, sp := range scopeTrace.Spans {
			span := scopeBuilder.AddSpan()

			span.SetName(sp.Name)
			if sp.TraceID != "" {
				traceID, err := hex.DecodeString(sp.TraceID)
				if err != nil {
					return fmt.Errorf("invalid trace_id '%s': %w", sp.TraceID, err)
				}
				if len(traceID) == 16 {
					span.SetTraceID(pcommon.TraceID(traceID))
				} else {
					return fmt.Errorf("trace_id '%s' must be 32 hex characters (16 bytes)", sp.TraceID)
				}
			}

			if sp.SpanID != "" {
				spanID, err := hex.DecodeString(sp.SpanID)
				if err != nil {
					return fmt.Errorf("invalid span_id '%s': %w", sp.SpanID, err)
				}
				if len(spanID) == 8 {
					span.SetSpanID(pcommon.SpanID(spanID))
				} else {
					return fmt.Errorf("span_id '%s' must be 16 hex characters (8 bytes)", sp.SpanID)
				}
			}

			// Set parent span ID
			if sp.ParentSpanID != "" {
				parentSpanID, err := hex.DecodeString(sp.ParentSpanID)
				if err != nil {
					return fmt.Errorf("invalid parent_span_id '%s': %w", sp.ParentSpanID, err)
				}
				if len(parentSpanID) == 8 {
					span.SetParentSpanID(pcommon.SpanID(parentSpanID))
				} else {
					return fmt.Errorf("parent_span_id '%s' must be 16 hex characters (8 bytes)", sp.ParentSpanID)
				}
			}

			span.SetKind(ptrace.SpanKind(sp.Kind))
			span.SetStartTimestamp(pcommon.Timestamp(sp.StartTimestamp))
			span.SetEndTimestamp(pcommon.Timestamp(sp.EndTimestamp))

			if len(sp.Attributes) > 0 {
				attrs := span.Attributes()
				spanAttrs, err := fromRaw(sp.Attributes)
				if err != nil {
					return fmt.Errorf("failed to convert span attributes: %w", err)
				}
				spanAttrs.CopyTo(attrs)
			}

			span.SetDroppedAttributesCount(sp.DroppedAttributesCount)

			for _, event := range sp.Events {
				spanEvent := span.Events().AppendEmpty()
				spanEvent.SetTimestamp(pcommon.Timestamp(event.Timestamp))
				spanEvent.SetName(event.Name)

				if len(event.Attributes) > 0 {
					eventAttrs, err := fromRaw(event.Attributes)
					if err != nil {
						return fmt.Errorf("failed to convert event attributes: %w", err)
					}
					eventAttrs.CopyTo(spanEvent.Attributes())
				}

				spanEvent.SetDroppedAttributesCount(event.DroppedAttributesCount)
			}

			span.SetDroppedEventsCount(sp.DroppedEventsCount)

			for _, link := range sp.Links {
				spanLink := span.Links().AppendEmpty()

				linkTraceID, err := hex.DecodeString(link.TraceID)
				if err != nil {
					return fmt.Errorf("invalid link trace_id '%s': %w", link.TraceID, err)
				}
				if len(linkTraceID) == 16 {
					spanLink.SetTraceID(pcommon.TraceID(linkTraceID))
				} else {
					return fmt.Errorf("link trace_id '%s' must be 32 hex characters (16 bytes)", link.TraceID)
				}

				linkSpanID, err := hex.DecodeString(link.SpanID)
				if err != nil {
					return fmt.Errorf("invalid link span_id '%s': %w", link.SpanID, err)
				}
				if len(linkSpanID) == 8 {
					spanLink.SetSpanID(pcommon.SpanID(linkSpanID))
				} else {
					return fmt.Errorf("link span_id '%s' must be 16 hex characters (8 bytes)", link.SpanID)
				}

				if link.TraceState != "" {
					spanLink.TraceState().FromRaw(link.TraceState)
				}
				if len(link.Attributes) > 0 {
					linkAttrs, err := fromRaw(link.Attributes)
					if err != nil {
						return fmt.Errorf("failed to convert link attributes: %w", err)
					}
					linkAttrs.CopyTo(spanLink.Attributes())
				}

				spanLink.SetDroppedAttributesCount(link.DroppedAttributesCount)
			}

			span.SetDroppedLinksCount(sp.DroppedLinksCount)

			if sp.Status.Code != 0 || sp.Status.Message != "" {
				status := span.Status()
				if sp.Status.Code != 0 {
					status.SetCode(ptrace.StatusCode(sp.Status.Code))
				}
				if sp.Status.Message != "" {
					status.SetMessage(sp.Status.Message)
				}
			}

			span.SetFlags(sp.Flags)
		}
	}

	return nil
}

// AddFromYAML parses YAML data and adds the traces to the builder.
// Note: JSON is a subset of YAML, so this function can also accept JSON format data.
func (tb *TracesBuilder) AddFromYAML(data []byte, opts ...ParseOptions) error {
	rt, err := ParseTraces(data, opts...)
	if err != nil {
		return err
	}
	return tb.Add(rt)
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
