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

package stats

import (
	"encoding/json"
	"fmt"
	"github.com/DataDog/sketches-go/ddsketch"
	"github.com/DataDog/sketches-go/ddsketch/mapping"
	"github.com/DataDog/sketches-go/ddsketch/store"
	"github.com/cardinalhq/oteltools/pkg/fingerprinter"
	"github.com/cardinalhq/oteltools/pkg/translate"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
	"sync"
	"time"
)

var ServiceNameKey = string(semconv.ServiceNameKey)
var ClusterNameKey = string(semconv.K8SClusterNameKey)
var NamespaceNameKey = string(semconv.K8SNamespaceNameKey)
var UnknownKey = "unknown"

type PercentileSketch struct {
	sketch *ddsketch.DDSketch
}

func NewPercentileSketch() (*PercentileSketch, error) {
	sketch, err := ddsketch.NewDefaultDDSketch(0.01)
	if err != nil {
		return nil, err
	}
	return &PercentileSketch{sketch: sketch}, nil
}

func (ps *PercentileSketch) Add(value float64) error {
	return ps.sketch.Add(value)
}

func (ps *PercentileSketch) Merge(other *PercentileSketch) error {
	return ps.sketch.MergeWith(other.sketch)
}

func (ps *PercentileSketch) GetPercentile(p float64) (float64, error) {
	return ps.sketch.GetValueAtQuantile(p)
}

func (ps *PercentileSketch) Serialize() []byte {
	var serialized []byte
	ps.sketch.Encode(&serialized, false)
	return serialized
}

func DeserializePercentileSketch(data []byte) (*PercentileSketch, error) {
	indexMapping, err := mapping.NewLogarithmicMapping(0.01)
	if err != nil {
		return nil, err
	}

	sketch, err := ddsketch.DecodeDDSketch(data, store.DefaultProvider, indexMapping)
	if err != nil {
		return nil, err
	}

	return &PercentileSketch{sketch: sketch}, nil
}

type SpanSketch struct {
	timestamp                   int64
	serviceId                   string
	name                        string
	kind                        string
	fingerprint                 int64
	Attributes                  map[string]any
	latencySketch               *PercentileSketch
	totalCount                  int64
	totalErrorCount             int64
	exceptionsByFingerprint     map[int64]string
	exceptionCountByFingerprint map[int64]int64

	mu sync.Mutex
}

func (s *SpanSketch) Serialize() ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	latencyBytes := s.latencySketch.Serialize()

	serializable := struct {
		ServiceId                   string           `json:"service_id"`
		Timestamp                   int64            `json:"timestamp"`
		Name                        string           `json:"span_name"`
		Kind                        string           `json:"span_kind"`
		Fingerprint                 int64            `json:"fingerprint"`
		Attributes                  map[string]any   `json:"attributes"`
		LatencySketch               []byte           `json:"latency_sketch"`
		TotalCount                  int64            `json:"total_count"`
		ErrorCount                  int64            `json:"error_count"`
		ExceptionsByFingerprint     map[int64]string `json:"exceptions_by_fingerprint"`
		ExceptionCountByFingerprint map[int64]int64  `json:"exception_count_by_fingerprint"`
	}{
		ServiceId:                   s.serviceId,
		Timestamp:                   s.timestamp,
		Name:                        s.name,
		Kind:                        s.kind,
		Fingerprint:                 s.fingerprint,
		Attributes:                  s.Attributes,
		LatencySketch:               latencyBytes,
		TotalCount:                  s.totalCount,
		ErrorCount:                  s.totalErrorCount,
		ExceptionsByFingerprint:     s.exceptionsByFingerprint,
		ExceptionCountByFingerprint: s.exceptionCountByFingerprint,
	}

	return json.Marshal(serializable)
}

func DeserializeSpanSketch(data []byte) (*SpanSketch, error) {
	var deserialized struct {
		ServiceId                   string           `json:"service_id"`
		Name                        string           `json:"span_name"`
		Kind                        string           `json:"span_kind"`
		Fingerprint                 int64            `json:"fingerprint"`
		Attributes                  map[string]any   `json:"attributes"`
		LatencySketch               []byte           `json:"latency_sketch"`
		TotalCount                  int64            `json:"total_count"`
		ErrorCount                  int64            `json:"error_count"`
		ExceptionsByFingerprint     map[int64]string `json:"exceptions_by_fingerprint"`
		ExceptionCountByFingerprint map[int64]int64  `json:"exception_count_by_fingerprint"`
	}

	if err := json.Unmarshal(data, &deserialized); err != nil {
		return nil, err
	}

	latencySketch, err := DeserializePercentileSketch(deserialized.LatencySketch)
	if err != nil {
		return nil, err
	}

	return &SpanSketch{
		serviceId:                   deserialized.ServiceId,
		name:                        deserialized.Name,
		kind:                        deserialized.Kind,
		fingerprint:                 deserialized.Fingerprint,
		Attributes:                  deserialized.Attributes,
		latencySketch:               latencySketch,
		totalCount:                  deserialized.TotalCount,
		totalErrorCount:             deserialized.ErrorCount,
		exceptionsByFingerprint:     deserialized.ExceptionsByFingerprint,
		exceptionCountByFingerprint: deserialized.ExceptionCountByFingerprint,
		mu:                          sync.Mutex{},
	}, nil
}

func (s *SpanSketch) Update(latency float64, isError bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.totalCount++
	if isError {
		s.totalErrorCount++
	}
	return s.latencySketch.Add(latency)
}

type SketchCache struct {
	sketches      sync.Map
	flushEvery    time.Duration
	fingerprinter fingerprinter.Fingerprinter
	flushFunc     func(spans []*SpanSketch)
}

func NewSketchCache(flushEvery time.Duration, flushFunc func([]*SpanSketch)) *SketchCache {
	cache := &SketchCache{
		sketches:      sync.Map{},
		fingerprinter: fingerprinter.NewFingerprinter(fingerprinter.WithMaxTokens(100)),
		flushEvery:    flushEvery,
		flushFunc:     flushFunc,
	}
	go cache.startFlusher()
	return cache
}

func ToKey(timestamp int64, serviceId string, fingerprint int64) string {
	return fmt.Sprintf("%d:%s:%d", timestamp, serviceId, fingerprint)
}

func GetFromResource(rl pcommon.Resource, key string) string {
	resourceAttributes := rl.Attributes()
	v, vFound := resourceAttributes.Get(key)
	cluster := v.AsString()
	if !vFound {
		cluster = UnknownKey
	}
	return cluster
}

func (c *SketchCache) UpdateSpanSketch(resource pcommon.Resource, span ptrace.Span) error {
	spanAttributes := span.Attributes()
	fingerprint, found := spanAttributes.Get(translate.CardinalFieldFingerprint)
	if !found {
		return fmt.Errorf("fingerprint not found in span")
	}
	interval := c.truncatedNow()
	serviceName := GetFromResource(resource, ServiceNameKey)
	clusterName := GetFromResource(resource, ClusterNameKey)
	namespaceName := GetFromResource(resource, NamespaceNameKey)

	serviceId := fmt.Sprintf("%s:%s:%s", serviceName, clusterName, namespaceName)
	key := ToKey(interval, serviceId, fingerprint.Int())
	latency, latencyFound := spanAttributes.Get(translate.CardinalFieldSpanDuration)
	if !latencyFound {
		return fmt.Errorf("latency not found in span")
	}
	isError := span.Status().Code() == ptrace.StatusCodeError

	existing, ok := c.sketches.Load(key)
	if ok {
		existingSketch := existing.(*SpanSketch)
		err := existingSketch.Update(latency.Double(), isError)
		return err
	}
	ps, err := NewPercentileSketch()
	if err != nil {
		return err
	}

	err = ps.Add(latency.Double())
	if err != nil {
		return err
	}

	var errorCount int64 = 0
	if isError {
		errorCount = 1
	}

	newSpanSketch := &SpanSketch{
		serviceId:                   serviceId,
		timestamp:                   interval,
		name:                        span.Name(),
		kind:                        span.Kind().String(),
		fingerprint:                 fingerprint.Int(),
		Attributes:                  spanAttributes.AsRaw(),
		latencySketch:               ps,
		totalCount:                  1,
		totalErrorCount:             errorCount,
		exceptionsByFingerprint:     map[int64]string{},
		exceptionCountByFingerprint: map[int64]int64{},
	}

	for i := 0; i < span.Events().Len(); i++ {
		event := span.Events().At(i)
		if event.Name() == semconv.ExceptionEventName {
			exceptionType, exceptionTypeFound := event.Attributes().Get(string(semconv.ExceptionTypeKey))
			exceptionMessage, exceptionMessageFound := event.Attributes().Get(string(semconv.ExceptionMessageKey))
			exceptionStackTrace, exceptionStackTraceFound := event.Attributes().Get(string(semconv.ExceptionStacktraceKey))

			if !exceptionTypeFound || !exceptionMessageFound || !exceptionStackTraceFound {
				continue
			}

			toFingerprint := exceptionType.AsString() + " " + exceptionMessage.AsString() + " " + exceptionStackTrace.AsString()
			exceptionFingerprint, _, _, _, err := c.fingerprinter.Fingerprint(toFingerprint)
			if err != nil {
				continue
			}

			newSpanSketch.exceptionsByFingerprint[exceptionFingerprint] = toFingerprint
			newSpanSketch.exceptionCountByFingerprint[exceptionFingerprint]++
		}
	}

	c.sketches.Store(key, newSpanSketch)
	return nil
}

func (c *SketchCache) truncatedNow() int64 {
	return time.Now().Truncate(c.flushEvery).Unix()
}

func (c *SketchCache) flush() {
	stepTimestamp := time.Now().Truncate(c.flushEvery)

	var spans []*SpanSketch
	c.sketches.Range(func(key, value interface{}) bool {
		spanSketch := value.(*SpanSketch)
		if spanSketch.timestamp < stepTimestamp.UnixMilli() {
			spans = append(spans, spanSketch)
		}
		return true
	})

	c.flushFunc(spans)
	c.sketches = sync.Map{}
}

func (c *SketchCache) MergeSpanSketch(incoming *SpanSketch) error {
	key := ToKey(incoming.timestamp, incoming.serviceId, incoming.fingerprint)
	existingSketch, ok := c.sketches.Load(key)
	if ok {
		existing := existingSketch.(*SpanSketch)
		existing.mu.Lock()
		err := existing.latencySketch.Merge(incoming.latencySketch)
		if err != nil {
			return err
		}
		existing.totalCount += incoming.totalCount
		existing.totalErrorCount += incoming.totalErrorCount
		for fingerprint, exception := range incoming.exceptionsByFingerprint {
			existing.exceptionsByFingerprint[fingerprint] = exception
		}

		for fingerprint, count := range incoming.exceptionCountByFingerprint {
			existing.exceptionCountByFingerprint[fingerprint] += count
		}

		existing.mu.Unlock()
		return nil
	}

	c.sketches.Store(key, incoming)
	return nil
}

func (c *SketchCache) startFlusher() {
	ticker := time.NewTicker(c.flushEvery)
	defer ticker.Stop()

	for range ticker.C {
		c.flush()
	}
}
