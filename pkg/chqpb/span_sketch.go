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

// Copyright 2024-2025 CardinalHQ, Inc.
//
// CardinalHQ, Inc. proprietary and confidential.
// Unauthorized copying, distribution, or modification of this file,
// via any medium, is strictly prohibited without prior written consent.
//
// All rights reserved.

package chqpb

import (
	"fmt"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"golang.org/x/exp/slog"
	"hash/fnv"
	"sort"
	"sync"
	"time"

	"github.com/DataDog/sketches-go/ddsketch"
	"github.com/DataDog/sketches-go/ddsketch/mapping"
	"github.com/DataDog/sketches-go/ddsketch/store"
	"github.com/cardinalhq/oteltools/pkg/fingerprinter"
	"github.com/cardinalhq/oteltools/pkg/translate"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
)

// DecodeSketch reconstructs a DDSketch from its encoded bytes.
func DecodeSketch(data []byte) (*ddsketch.DDSketch, error) {
	m, err := mapping.NewLogarithmicMapping(0.01)
	if err != nil {
		return nil, err
	}
	sk, err := ddsketch.DecodeDDSketch(data, store.DefaultProvider, m)
	if err != nil {
		return nil, err
	}
	return sk, nil
}

func Merge(sketch *ddsketch.DDSketch, other *ddsketch.DDSketch) error {
	return sketch.MergeWith(other)
}

func MergeEncodedSketch(a, b []byte) ([]byte, error) {
	skA, err := DecodeSketch(a)
	if err != nil {
		return nil, fmt.Errorf("decoding sketch A: %w", err)
	}
	skB, err := DecodeSketch(b)
	if err != nil {
		return nil, fmt.Errorf("decoding sketch B: %w", err)
	}
	if err := Merge(skA, skB); err != nil {
		return nil, fmt.Errorf("merging sketches: %w", err)
	}
	return Encode(skA), nil
}

type spanSketchEntry struct {
	mu       sync.Mutex
	proto    *SpanSketchProto
	internal *ddsketch.DDSketch
}

type SpanSketchCache struct {
	buckets    sync.Map // map[int64]*sync.Map where each inner sync.Map: map[string]*spanSketchEntry
	customerId string
	interval   time.Duration
	fpr        fingerprinter.Fingerprinter
	flushFunc  func(*SpanSketchList) error
	marshaller ptrace.JSONMarshaler
}

func NewSpanSketchCache(interval time.Duration, cid string, flushFunc func(*SpanSketchList) error) *SpanSketchCache {
	c := &SpanSketchCache{
		interval:   interval,
		customerId: cid,
		fpr:        fingerprinter.NewFingerprinter(fingerprinter.NewTrieClusterManager(0.5)),
		flushFunc:  flushFunc,
		marshaller: ptrace.JSONMarshaler{},
	}
	go c.loop()
	return c
}

func (c *SpanSketchCache) loop() {
	t := time.NewTicker(c.interval)
	for range t.C {
		c.flush()
	}
}

func (c *SpanSketchCache) Update(
	metricName string,
	tagValues map[string]string,
	span ptrace.Span,
	isAggregate bool,
	resource pcommon.Resource,
) {
	interval := span.EndTimestamp().AsTime().Truncate(c.interval).Unix()
	tid := computeTID(metricName, tagValues)

	bucketIface, _ := c.buckets.LoadOrStore(interval, &sync.Map{})
	skMap := bucketIface.(*sync.Map) // inner map[string]*spanSketchEntry

	val, ok := skMap.Load(tid)
	var entry *spanSketchEntry
	if !ok {
		proto := &SpanSketchProto{
			MetricName:         metricName,
			Tid:                tid,
			Interval:           interval,
			Tags:               tagValues,
			TotalCount:         0,
			ErrorCount:         0,
			ExceptionCount:     0,
			ExceptionsMap:      make(map[int64]string),
			ExceptionCountsMap: make(map[int64]int64),
			IsAggregate:        isAggregate,
		}
		ps, _ := ddsketch.NewDefaultDDSketch(0.01)
		entry = &spanSketchEntry{
			proto:    proto,
			internal: ps,
		}
		skMap.Store(tid, entry)
	} else {
		entry = val.(*spanSketchEntry)
	}

	entry.mu.Lock()
	defer entry.mu.Unlock()

	entry.proto.TotalCount++
	if span.Status().Code() == ptrace.StatusCodeError {
		entry.proto.ErrorCount++
	}

	if v, ok := span.Attributes().Get(translate.CardinalFieldSpanDuration); ok {
		_ = entry.internal.Add(v.Double())
	}

	hasException := false
	for i := 0; i < span.Events().Len(); i++ {
		e := span.Events().At(i)
		if e.Name() != semconv.ExceptionEventName {
			continue
		}
		hasException = true
		exMsg := extractExceptionMessage(e.Attributes())
		entry.proto.ExceptionCount++
		if c.fpr != nil {
			fp, _, _, err := c.fpr.Fingerprint(exMsg)
			if err == nil {
				bytes, err := c.spanToJson(span, resource)
				if err == nil {
					entry.proto.ExceptionsMap[fp] = string(bytes)
					entry.proto.ExceptionCountsMap[fp]++
				}
			}
		}
	}

	// No exception event, but status is error: synthesize from attributes
	if !hasException && span.Status().Code() == ptrace.StatusCodeError {
		exMsg := synthesizeErrorMessageFromSpan(span)
		entry.proto.ExceptionCount++
		if c.fpr != nil {
			fp, _, _, err := c.fpr.Fingerprint(exMsg)
			if err == nil {
				bytes, err := c.spanToJson(span, resource)
				if err == nil {
					entry.proto.ExceptionsMap[fp] = string(bytes)
					entry.proto.ExceptionCountsMap[fp]++
				}
			}
		}
	}
}

func (c *SpanSketchCache) spanToJson(src ptrace.Span, resource pcommon.Resource) ([]byte, error) {
	td := ptrace.NewTraces()
	rs := td.ResourceSpans().AppendEmpty()
	resource.CopyTo(rs.Resource())

	ss := rs.ScopeSpans().AppendEmpty()
	dst := ss.Spans().AppendEmpty()
	src.CopyTo(dst)

	return c.marshaller.MarshalTraces(td)
}

func synthesizeErrorMessageFromSpan(span ptrace.Span) string {
	attr := span.Attributes()
	candidates := []string{
		toStrValue(attr, "error.message"),
		toStrValue(attr, "error"),
		toStrValue(attr, string(semconv.HTTPResponseStatusCodeKey)),
		toStrValue(attr, string(semconv.RPCGRPCStatusCodeKey)),
		toStrValue(attr, string(semconv.ExceptionMessageKey)),
	}
	msg := ""
	for _, val := range candidates {
		if val != "" {
			if msg != "" {
				msg += " | "
			}
			msg += val
		}
	}

	// Fallback: dump all string attributes sorted lexicographically
	if msg == "" {
		var kvs []string
		attr.Range(func(k string, v pcommon.Value) bool {
			if v.Type() == pcommon.ValueTypeStr {
				kvs = append(kvs, fmt.Sprintf("%s=%s", k, v.Str()))
			}
			return true
		})
		sort.Strings(kvs)
		msg = fmt.Sprintf("attrs=[%s]", joinWithSep(kvs, " | "))
	}
	return msg
}

// toStrValue returns string value of a pcommon.Map entry (if it exists and is a string).
func toStrValue(attrs pcommon.Map, key string) string {
	if v, ok := attrs.Get(key); ok && v.Type() == pcommon.ValueTypeStr {
		return v.Str()
	}
	return ""
}

func joinWithSep(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	out := parts[0]
	for _, p := range parts[1:] {
		out += sep + p
	}
	return out
}

func (c *SpanSketchCache) flush() {
	now := time.Now().Truncate(c.interval).Unix()
	list := &SpanSketchList{CustomerId: c.customerId}

	c.buckets.Range(func(intervalKey, v interface{}) bool {
		interval := intervalKey.(int64)
		if interval >= now {
			return true
		}

		skMap := v.(*sync.Map)

		skMap.Range(func(tid, entryVal interface{}) bool {
			entry := entryVal.(*spanSketchEntry)

			entry.mu.Lock()
			entry.proto.Sketch = Encode(entry.internal)
			list.Sketches = append(list.Sketches, entry.proto)
			entry.mu.Unlock()

			return true
		})
		c.buckets.Delete(intervalKey)
		return true
	})

	if len(list.Sketches) > 0 {
		if err := c.flushFunc(list); err != nil {
			slog.Error("failed to flush sketches", slog.String("customerId", c.customerId), slog.String("error", err.Error()))
		}
	}
}

func extractExceptionMessage(attributes pcommon.Map) string {
	exceptionMessage := ""
	if et, ok := attributes.Get(string(semconv.ExceptionTypeKey)); ok {
		exceptionMessage = et.AsString()
	}
	if msg, ok := attributes.Get(string(semconv.ExceptionMessageKey)); ok {
		if exceptionMessage != "" {
			exceptionMessage += ": "
		}
		exceptionMessage += msg.AsString()
	}
	if st, ok := attributes.Get(string(semconv.ExceptionStacktraceKey)); ok {
		if exceptionMessage != "" {
			exceptionMessage += ": "
		}
		exceptionMessage += "\n" + st.AsString()
	}
	return exceptionMessage
}

func Encode(sketch *ddsketch.DDSketch) []byte {
	var buf []byte
	sketch.Encode(&buf, false)
	return buf
}

func computeTID(metricName string, tags map[string]string) int64 {
	keys := make([]string, 0, len(tags))
	for k := range tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	h := fnv.New64a()
	_, _ = h.Write([]byte(metricName))
	for _, k := range keys {
		_, _ = h.Write([]byte(k + "=" + tags[k] + "|"))
	}
	return int64(h.Sum64())
}
