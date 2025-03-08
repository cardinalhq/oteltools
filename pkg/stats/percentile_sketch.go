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
	"github.com/DataDog/sketches-go/ddsketch"
	"github.com/DataDog/sketches-go/ddsketch/mapping"
	"github.com/DataDog/sketches-go/ddsketch/store"
	"sync"
	"time"
)

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
	Key           string
	Tags          map[string]string
	latencySketch *PercentileSketch
	totalCount    int64
	errorCount    int64

	mu sync.Mutex
}

func (s *SpanSketch) Serialize() ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	latencyBytes := s.latencySketch.Serialize()

	serializable := struct {
		Key           string            `json:"key"`
		Tags          map[string]string `json:"tags"`
		LatencySketch []byte            `json:"latency_sketch"`
		TotalCount    int64             `json:"total_count"`
		ErrorCount    int64             `json:"error_count"`
	}{
		Key:           s.Key,
		Tags:          s.Tags,
		LatencySketch: latencyBytes,
		TotalCount:    s.totalCount,
		ErrorCount:    s.errorCount,
	}

	return json.Marshal(serializable)
}

func DeserializeSpanSketch(data []byte) (*SpanSketch, error) {
	var deserialized struct {
		Key           string            `json:"key"`
		Tags          map[string]string `json:"tags"`
		LatencySketch []byte            `json:"latency_sketch"`
		TotalCount    int64             `json:"total_count"`
		ErrorCount    int64             `json:"error_count"`
	}

	if err := json.Unmarshal(data, &deserialized); err != nil {
		return nil, err
	}

	latencySketch, err := DeserializePercentileSketch(deserialized.LatencySketch)
	if err != nil {
		return nil, err
	}

	return &SpanSketch{
		Key:           deserialized.Key,
		Tags:          deserialized.Tags,
		latencySketch: latencySketch,
		totalCount:    deserialized.TotalCount,
		errorCount:    deserialized.ErrorCount,
		mu:            sync.Mutex{}, // Initialize mutex
	}, nil
}

func (s *SpanSketch) UpdateSpan(latency float64, isError bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.totalCount++
	if isError {
		s.errorCount++
	}
	return s.latencySketch.Add(latency)
}

type SketchCache struct {
	sketches   sync.Map
	flushEvery time.Duration
	flushFunc  func(stepTimestamp time.Time, spans []*SpanSketch)
}

func NewSketchCache(flushEvery time.Duration, flushFunc func(time.Time, []*SpanSketch)) *SketchCache {
	cache := &SketchCache{
		sketches:   sync.Map{},
		flushEvery: flushEvery,
		flushFunc:  flushFunc,
	}
	go cache.startFlusher() // Start background flusher
	return cache
}

func (c *SketchCache) UpdateSpanSketch(key string, tags map[string]string, latency float64, isError bool) error {
	existingSketch, ok := c.sketches.Load(key)
	if ok {
		span := existingSketch.(*SpanSketch)
		err := span.UpdateSpan(latency, isError)
		return err
	}
	ps, err := NewPercentileSketch()
	if err != nil {
		return err
	}

	err = ps.Add(latency)
	if err != nil {
		return err
	}
	c.sketches.Store(key, &SpanSketch{
		Key:           key,
		Tags:          tags,
		latencySketch: ps,
		totalCount:    1,
		errorCount:    0,
	})
	return nil
}

func (c *SketchCache) flush() {
	stepTimestamp := time.Now().Truncate(c.flushEvery)

	var spans []*SpanSketch
	c.sketches.Range(func(key, value interface{}) bool {
		spans = append(spans, value.(*SpanSketch))
		return true
	})

	c.flushFunc(stepTimestamp, spans)
	c.sketches = sync.Map{}
}

func (c *SketchCache) MergeSpanSketch(incoming *SpanSketch) error {
	existingSketch, ok := c.sketches.Load(incoming.Key)
	if ok {
		existing := existingSketch.(*SpanSketch)
		existing.mu.Lock()
		err := existing.latencySketch.Merge(incoming.latencySketch)
		if err != nil {
			return err
		}
		existing.totalCount += incoming.totalCount
		existing.errorCount += incoming.errorCount

		existing.mu.Unlock()
		return nil
	}

	c.sketches.Store(incoming.Key, incoming)
	return nil
}

func (c *SketchCache) startFlusher() {
	ticker := time.NewTicker(c.flushEvery)
	defer ticker.Stop()

	for range ticker.C {
		c.flush()
	}
}
