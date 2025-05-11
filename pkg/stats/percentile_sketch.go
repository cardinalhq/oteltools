package stats

import (
	"fmt"
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
	"go.opentelemetry.io/collector/pdata/ptrace"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
)

// DecodeSketch reconstructs from bytes.
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
	// 2) Merge B into A
	if err := Merge(skA, skB); err != nil {
		return nil, fmt.Errorf("merging sketches: %w", err)
	}
	return Encode(skA), nil
}

// sketchEntry couples a proto and its internal sketch state.
type sketchEntry struct {
	proto    *SpanSketchProto
	internal *ddsketch.DDSketch
}

// SketchCache holds sketches for multiple metrics and emits a SpanSketchList on flush.
type SketchCache struct {
	sketches   sync.Map // map[string]*sketchEntry
	customerId string
	interval   time.Duration
	fpr        fingerprinter.Fingerprinter
	flushFunc  func(*SpanSketchList) error
}

// NewSketchCache creates a cache with flush interval and callback.
func NewSketchCache(interval time.Duration, cid string, flushFunc func(*SpanSketchList) error) *SketchCache {
	c := &SketchCache{
		interval:   interval,
		customerId: cid,
		fpr:        fingerprinter.NewFingerprinter(fingerprinter.NewTrieClusterManager(0.5)),
		flushFunc:  flushFunc,
	}
	go c.loop()
	return c
}

func (c *SketchCache) loop() {
	t := time.NewTicker(c.interval)
	for range t.C {
		c.flush()
	}
}

// Update ingests a span under a metricName and tagValues.
func (c *SketchCache) Update(metricName string, tagValues map[string]string, span ptrace.Span) {
	// Determine bucket interval
	interval := span.EndTimestamp().AsTime().Truncate(c.interval).Unix()
	// Compute tid from metricName and tags
	tid := computeTID(metricName, tagValues)
	key := fmt.Sprintf("%d:%s", interval, tid)

	val, ok := c.sketches.Load(key)
	var entry *sketchEntry
	if !ok {
		// Initialize proto
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
		}
		// Initialize internal sketch
		ps, _ := ddsketch.NewDefaultDDSketch(0.01)
		entry = &sketchEntry{proto: proto, internal: ps}
		c.sketches.Store(key, entry)
	} else {
		entry = val.(*sketchEntry)
	}

	// Update counts
	entry.proto.TotalCount++
	if span.Status().Code() == ptrace.StatusCodeError {
		entry.proto.ErrorCount++
	}

	// Update latency
	if v, ok := span.Attributes().Get(translate.CardinalFieldSpanDuration); ok {
		err := entry.internal.Add(v.Double())
		if err != nil {
			return
		}
	}

	// Update exceptions
	for i := 0; i < span.Events().Len(); i++ {
		e := span.Events().At(i)
		if e.Name() != semconv.ExceptionEventName {
			continue
		}
		exceptionMessage := ""
		exceptionTypeKey, eok := e.Attributes().Get(string(semconv.ExceptionTypeKey))
		if eok {
			exceptionMessage = exceptionTypeKey.AsString()
		}

		msg, mok := e.Attributes().Get(string(semconv.ExceptionMessageKey))
		if mok {
			if exceptionMessage != "" {
				exceptionMessage += ": "
			}
			exceptionMessage += msg.AsString()
		}

		stackTrace, sok := e.Attributes().Get(string(semconv.ExceptionStacktraceKey))
		if sok {
			if exceptionMessage != "" {
				exceptionMessage += ": "
			}
			exceptionMessage += "\n" + stackTrace.AsString()
		}

		entry.proto.ExceptionCount++
		if c.fpr != nil {
			fp, _, _, err := c.fpr.Fingerprint(msg.AsString())
			if err == nil {
				entry.proto.ExceptionsMap[fp] = msg.AsString()
				entry.proto.ExceptionCountsMap[fp]++
			}
		}
	}
}

func Encode(sketch *ddsketch.DDSketch) []byte {
	var buf []byte
	sketch.Encode(&buf, false)
	return buf
}

// flush emits a SpanSketchList proto and removes old entries.
func (c *SketchCache) flush() {
	now := time.Now().Truncate(c.interval).Unix()
	list := &SpanSketchList{}
	list.CustomerId = c.customerId
	c.sketches.Range(func(key, value interface{}) bool {
		entry := value.(*sketchEntry)
		if entry.proto.Interval < now {
			// Encode sketch bytes
			entry.proto.Sketch = Encode(entry.internal)
			list.Sketches = append(list.Sketches, entry.proto)
			c.sketches.Delete(key)
		}
		return true
	})
	err := c.flushFunc(list)
	if err != nil {
		slog.Error("failed to flush sketches", slog.String("customerId", c.customerId), slog.String("error", err.Error()))
	}
}

// computeTID hashes metricName and tagValues into a stable ID.
func computeTID(metricName string, tags map[string]string) string {
	keys := make([]string, 0, len(tags))
	for k := range tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	h := fnv.New64a()
	_, _ = h.Write([]byte(metricName))
	for _, k := range keys {
		_, _ = h.Write([]byte(k))
		_, _ = h.Write([]byte("="))
		_, _ = h.Write([]byte(tags[k]))
		_, _ = h.Write([]byte("|"))
	}
	return fmt.Sprintf("%x", h.Sum64())
}
