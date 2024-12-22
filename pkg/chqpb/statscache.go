// Copyright 2024 CardinalHQ, Inc
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

package chqpb

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type FlushCallback[T any] func(expiredItems []T)

type Entry[T any] struct {
	key       string
	value     T
	timestamp time.Time
	mutex     sync.Mutex // Per-entry lock for fine-grained locking
}
type StatsCache[T any] struct {
	capacity      int
	cache         map[string]*Entry[T]
	keys          []string
	keyIndex      map[string]int
	cacheMutex    sync.RWMutex
	expiry        time.Duration
	flushCallback FlushCallback[T]
	initialize    func() (T, error)
	clock         Clock // Use Clock interface for time
	stopCleanup   chan struct{}
	randSource    *rand.Rand
}

func NewStatsCache[T any](capacity int, expiry time.Duration,
	flushCallback FlushCallback[T],
	initialize func() (T, error),
	clock Clock) *StatsCache[T] {
	return &StatsCache[T]{
		capacity:      capacity,
		cache:         make(map[string]*Entry[T]),
		keys:          []string{},
		keyIndex:      make(map[string]int),
		expiry:        expiry,
		flushCallback: flushCallback,
		initialize:    initialize,
		clock:         clock,
		stopCleanup:   make(chan struct{}),
		randSource:    rand.New(rand.NewSource(clock.Now().UnixNano())),
	}
}

func (b *StatsCache[T]) cleanupExpiredEntries() {
	b.cacheMutex.Lock()
	defer b.cacheMutex.Unlock()

	now := b.clock.Now()
	var expiredItems []T

	for key, entry := range b.cache {
		entry.mutex.Lock()
		if now.Sub(entry.timestamp) > b.expiry {
			expiredItems = append(expiredItems, entry.value)
			b.removeKey(key)
		}
		entry.mutex.Unlock()
	}

	if len(expiredItems) > 0 && b.flushCallback != nil {
		b.flushCallback(expiredItems)
	}
}

func (b *StatsCache[T]) startCleanup() {
	ticker := time.NewTicker(b.expiry / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.cleanupExpiredEntries()
		case <-b.stopCleanup:
			return
		}
	}
}

func (b *StatsCache[T]) Compute(key string, entryUpdater func(existing T) error) error {
	b.cacheMutex.RLock()
	entry, found := b.cache[key]
	b.cacheMutex.RUnlock()

	if found {
		entry.mutex.Lock()
		defer entry.mutex.Unlock()

		if time.Since(entry.timestamp) > b.expiry {
			b.cacheMutex.Lock()
			defer b.cacheMutex.Unlock()
			b.removeKey(key)
			return fmt.Errorf("entry expired")
		}

		return entryUpdater(entry.value)
	}

	value, err := b.initialize()
	if err != nil {
		return err
	}
	err = entryUpdater(value)
	if err != nil {
		return err
	}

	b.cacheMutex.Lock()
	defer b.cacheMutex.Unlock()

	if len(b.cache) >= b.capacity {
		b.evictRandom()
	}

	b.addKey(key, value)
	return nil
}

func (b *StatsCache[T]) addKey(key string, value T) {
	entry := &Entry[T]{
		key:       key,
		value:     value,
		timestamp: time.Now(),
	}
	b.cache[key] = entry
	b.keys = append(b.keys, key)
	b.keyIndex[key] = len(b.keys) - 1
}

func (b *StatsCache[T]) removeKey(key string) {
	index, found := b.keyIndex[key]
	if !found {
		return
	}

	delete(b.cache, key)

	// Remove key from keys slice by swapping it with the last element
	lastKey := b.keys[len(b.keys)-1]
	b.keys[index] = lastKey
	b.keyIndex[lastKey] = index
	b.keys = b.keys[:len(b.keys)-1]

	// Remove from keyIndex map
	delete(b.keyIndex, key)
}

func (b *StatsCache[T]) evictRandom() {
	if len(b.keys) == 0 {
		return
	}

	randomIndex := b.randSource.Intn(len(b.keys))
	randomKey := b.keys[randomIndex]

	entry := b.cache[randomKey]
	if b.flushCallback != nil {
		b.flushCallback([]T{entry.value})
	}

	b.removeKey(randomKey)
}

func (b *StatsCache[T]) Close() {
	close(b.stopCleanup)
}
