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
	buckets         sync.Map // map[int64]*sync.Map where each inner sync.Map: map[string]*spanSketchEntry
	customerId      string
	interval        time.Duration
	clusterManager  *fingerprinter.TrieClusterManager
	flushFunc       func(*SpanSketchList) error
	marshaller      ptrace.JSONMarshaler
	errorCountTopKs sync.Map
	latencyTopKs    sync.Map
	maxK            int
}

func NewSpanSketchCache(interval time.Duration, cid string, maxK int, flushFunc func(*SpanSketchList) error) *SpanSketchCache {
	c := &SpanSketchCache{
		interval:       interval,
		customerId:     cid,
		clusterManager: fingerprinter.NewTrieClusterManager(0.5),
		flushFunc:      flushFunc,
		maxK:           maxK,
		marshaller:     ptrace.JSONMarshaler{},
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

func getKey(metricName string, direction Direction, parentTid int64, tagFamilyId int64) string {
	return fmt.Sprintf("%s:%s:%d:%d", metricName, direction.String(), parentTid, tagFamilyId)
}

func (c *SpanSketchCache) getErrorCountTopK(metricName string, parentTid int64, tagFamilyId int64) *TopKByFrequency {
	key := getKey(metricName, Direction_UP, parentTid, tagFamilyId)
	topK, _ := c.errorCountTopKs.LoadOrStore(key, NewTopKByFrequency(c.maxK, 2*c.interval, Direction_UP))
	return topK.(*TopKByFrequency)
}

func (c *SpanSketchCache) getLatencyTopK(metricName string, parentTid int64, tagFamilyId int64) *TopKByValue {
	key := getKey(metricName, Direction_UP, parentTid, tagFamilyId)
	topK, _ := c.latencyTopKs.LoadOrStore(key, NewTopKByValue(c.maxK, 2*c.interval, Direction_UP))
	return topK.(*TopKByValue)
}

func (c *SpanSketchCache) Update(
	metricName string,
	metricType string,
	tagValues map[string]string,
	span ptrace.Span,
	resource pcommon.Resource,
	parentTID int64,
	tagFamilyId int64,
) int64 {
	interval := span.EndTimestamp().AsTime().Truncate(c.interval).Unix()
	tid := computeTID(metricName, tagValues)

	bucketIface, _ := c.buckets.LoadOrStore(interval, &sync.Map{})
	skMap := bucketIface.(*sync.Map) // inner map[string]*spanSketchEntry
	spanDurationVal, _ := span.Attributes().Get(translate.CardinalFieldSpanDuration)
	spanDuration := spanDurationVal.Double()

	var shouldEligible = false
	if span.Status().Code() == ptrace.StatusCodeError {
		errorCountTopK := c.getErrorCountTopK(metricName, parentTID, tagFamilyId)
		shouldEligible = errorCountTopK.EligibleWithCount(tid, 1)
	} else {
		latencyTopK := c.getLatencyTopK(metricName, parentTID, tagFamilyId)
		shouldEligible = latencyTopK.Eligible(spanDuration)
	}

	val, ok := skMap.Load(tid)
	var entry *spanSketchEntry
	if !ok {
		if !shouldEligible {
			return tid
		}
		proto := &SpanSketchProto{
			MetricName:         metricName,
			MetricType:         metricType,
			Tid:                tid,
			Interval:           interval,
			Tags:               tagValues,
			TotalCount:         0,
			ErrorCount:         0,
			ExceptionCount:     0,
			ExceptionsMap:      make(map[int64]string),
			ExceptionCountsMap: make(map[int64]int64),
			TagFamilyId:        tagFamilyId,
			ParentTID:          parentTID,
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

	_ = entry.internal.Add(spanDuration)

	hasException := false
	for i := 0; i < span.Events().Len(); i++ {
		e := span.Events().At(i)
		if e.Name() != semconv.ExceptionEventName {
			continue
		}
		hasException = true
		exMsg := extractExceptionMessage(e.Attributes())
		entry.proto.ExceptionCount++
		if c.clusterManager != nil {
			fp, _, _, err := fingerprinter.Fingerprint(exMsg, c.clusterManager)
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

		if exMsg == "" {
			fp, ok := span.Attributes().Get(translate.CardinalFieldFingerprint)
			if ok {
				c.putException(span, resource, entry, fp.Int())
			}
		} else if c.clusterManager != nil {
			fp, _, _, err := fingerprinter.Fingerprint(exMsg, c.clusterManager)
			if err == nil {
				c.putException(span, resource, entry, fp)
			}
		}
	}
	return tid
}

func (c *SpanSketchCache) putException(span ptrace.Span, resource pcommon.Resource, entry *spanSketchEntry, fp int64) {
	bytes, err := c.spanToJson(span, resource)
	if err == nil {
		entry.proto.ExceptionsMap[fp] = string(bytes)
		entry.proto.ExceptionCountsMap[fp]++
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

	return msg
}

// toStrValue returns string value of a pcommon.Map entry (if it exists and is a string).
func toStrValue(attrs pcommon.Map, key string) string {
	if v, ok := attrs.Get(key); ok && v.Type() == pcommon.ValueTypeStr {
		return v.Str()
	}
	return ""
}

func (c *SpanSketchCache) flush() {
	now := time.Now().Truncate(c.interval).Unix()
	out := &SpanSketchList{CustomerId: c.customerId}

	c.buckets.Range(func(intervalKey, v interface{}) bool {
		interval := intervalKey.(int64)
		if interval >= now {
			return true
		}
		skMap := v.(*sync.Map)

		// ─── Phase 1: Populate both heaps ─────────────────────────────────
		skMap.Range(func(_, val interface{}) bool {
			entry := val.(*spanSketchEntry)
			mn, tid := entry.proto.MetricName, entry.proto.Tid

			// Expire old entries
			errTK := c.getErrorCountTopK(mn, entry.proto.ParentTID, entry.proto.TagFamilyId)
			errTK.CleanupExpired()
			latTK := c.getLatencyTopK(mn, entry.proto.ParentTID, entry.proto.TagFamilyId)
			latTK.CleanupExpired()

			// Update both heaps based on this interval’s data
			errTK.AddCount(tid, int(entry.proto.ExceptionCount))
			if p50, err := entry.internal.GetValueAtQuantile(0.5); err == nil {
				latTK.Add(tid, p50)
			}
			return true
		})

		// ─── Phase 2: Emit only the final top-K winners ─────────────────────
		skMap.Range(func(_, val interface{}) bool {
			entry := val.(*spanSketchEntry)
			mn, tid := entry.proto.MetricName, entry.proto.Tid

			errTK := c.getErrorCountTopK(mn, entry.proto.ParentTID, entry.proto.TagFamilyId)
			latTK := c.getLatencyTopK(mn, entry.proto.ParentTID, entry.proto.TagFamilyId)

			// Check membership in each heap’s index
			_, inErr := errTK.h.index[tid]
			_, inVal := latTK.h.index[tid]
			if !inErr && !inVal {
				return true
			}

			// Serialize & collect
			entry.mu.Lock()
			entry.proto.Sketch = Encode(entry.internal)
			out.Sketches = append(out.Sketches, entry.proto)
			entry.mu.Unlock()
			return true
		})

		c.buckets.Delete(intervalKey)
		return true
	})

	if len(out.Sketches) > 0 {
		if err := c.flushFunc(out); err != nil {
			slog.Error("failed to flush sketches",
				slog.String("customerId", c.customerId),
				slog.String("error", err.Error()),
			)
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
