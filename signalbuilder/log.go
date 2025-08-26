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

func (lb *LogBuilder) Add(rl *ResourceLogs) error {
	resourceAttrs, err := fromRaw(rl.Resource)
	if err != nil {
		return fmt.Errorf("failed to convert resource attributes: %w", err)
	}

	resourceBuilder := lb.Resource(resourceAttrs)

	for _, scopeLog := range rl.ScopeLogs {
		scopeAttrs, err := fromRaw(scopeLog.Attributes)
		if err != nil {
			return fmt.Errorf("failed to convert scope attributes: %w", err)
		}

		scopeBuilder := resourceBuilder.ScopeWithInfo(
			scopeLog.Name,
			scopeLog.Version,
			scopeLog.SchemaURL,
			scopeAttrs,
		)
		for _, lr := range scopeLog.LogRecords {
			record := scopeBuilder.AddRecord()

			record.SetTimestamp(pcommon.Timestamp(lr.Timestamp))
			record.SetObservedTimestamp(pcommon.Timestamp(lr.ObservedTimestamp))
			record.SetSeverityText(lr.SeverityText)
			record.SetSeverityNumber(plog.SeverityNumber(lr.SeverityNumber))

			if lr.Body != nil {
				body := record.Body()
				if err := body.FromRaw(lr.Body); err != nil {
					return fmt.Errorf("failed to convert log record body: %w", err)
				}
			}

			if len(lr.Attributes) > 0 {
				attrs := record.Attributes()
				logAttrs, err := fromRaw(lr.Attributes)
				if err != nil {
					return fmt.Errorf("failed to convert log record attributes: %w", err)
				}
				logAttrs.CopyTo(attrs)
			}

			record.SetFlags(plog.LogRecordFlags(lr.Flags))
			record.SetDroppedAttributesCount(lr.DroppedAttributesCount)
			record.SetEventName(lr.EventName)

			if lr.TraceID != "" {
				traceID, err := hex.DecodeString(lr.TraceID)
				if err != nil {
					return fmt.Errorf("invalid trace_id '%s': %w", lr.TraceID, err)
				}
				if len(traceID) == 16 {
					record.SetTraceID(pcommon.TraceID(traceID))
				} else {
					return fmt.Errorf("trace_id '%s' must be 32 hex characters (16 bytes)", lr.TraceID)
				}
			}

			if lr.SpanID != "" {
				spanID, err := hex.DecodeString(lr.SpanID)
				if err != nil {
					return fmt.Errorf("invalid span_id '%s': %w", lr.SpanID, err)
				}
				if len(spanID) == 8 {
					record.SetSpanID(pcommon.SpanID(spanID))
				} else {
					return fmt.Errorf("span_id '%s' must be 16 hex characters (8 bytes)", lr.SpanID)
				}
			}
		}
	}

	return nil
}

// AddFromYAML parses YAML data and adds the logs to the builder.
// Note: JSON is a subset of YAML, so this function can also accept JSON format data.
func (lb *LogBuilder) AddFromYAML(data []byte, opts ...ParseOptions) error {
	rl, err := ParseLogs(data, opts...)
	if err != nil {
		return err
	}
	return lb.Add(rl)
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
