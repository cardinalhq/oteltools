// Copyright 2024 CardinalHQ, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package chqpb

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type FlushCallback[T any] func(expiredItems []T)

type InitializeCallback[T any] func(tsHour int64, key string) (T, error)

type Entry[T any] struct {
	key       string
	value     T
	timestamp time.Time
	mutex     sync.Mutex // Per-entry lock for fine-grained locking
}

type StatsCache[T any] struct {
	capacity           int
	numBins            uint16
	cache              map[int64]map[int]map[string]*Entry[T]
	bucketLocks        map[int64]*sync.RWMutex
	binLocks           map[int64]map[int]*sync.RWMutex
	cacheMutex         sync.RWMutex
	expiry             time.Duration
	flushCallback      FlushCallback[T]
	initializeCallback InitializeCallback[T]
	clock              Clock
	stopCleanup        chan struct{}
	randSource         *rand.Rand
}

func NewStatsCache[T any](capacity int,
	numBins uint16,
	expiry time.Duration,
	flushCallback FlushCallback[T],
	initializeCallback InitializeCallback[T],
	clock Clock) *StatsCache[T] {
	statsCache := &StatsCache[T]{
		capacity:           capacity,
		numBins:            numBins,
		cache:              make(map[int64]map[int]map[string]*Entry[T]),
		bucketLocks:        make(map[int64]*sync.RWMutex),
		binLocks:           make(map[int64]map[int]*sync.RWMutex),
		expiry:             expiry,
		flushCallback:      flushCallback,
		initializeCallback: initializeCallback,
		clock:              clock,
		stopCleanup:        make(chan struct{}),
		randSource:         rand.New(rand.NewSource(clock.Now().UnixNano())),
	}
	return statsCache
}

func (b *StatsCache[T]) Start() {
	go func() {
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
	}()
}

func (b *StatsCache[T]) cleanupExpiredEntries() {
	now := b.clock.Now()
	expiredBucket := now.Truncate(b.expiry).Add(-b.expiry).Unix()

	b.cacheMutex.Lock()
	expiredBucketLock, bucketExists := b.bucketLocks[expiredBucket]
	if !bucketExists {
		b.cacheMutex.Unlock()
		return
	}

	expiredBinLocks := b.binLocks[expiredBucket]
	delete(b.bucketLocks, expiredBucket)

	keyBuckets, found := b.cache[expiredBucket]
	if found {
		delete(b.cache, expiredBucket)
		delete(b.binLocks, expiredBucket)
	}
	b.cacheMutex.Unlock()

	if !found {
		return
	}

	expiredBucketLock.Lock()
	var expiredItems []T
	for binIndex, entryMap := range keyBuckets {
		binLock, binExists := expiredBinLocks[binIndex]
		if binExists {
			binLock.Lock()
			fmt.Printf("cleanupExpiredEntries: expiredBucket:%d, numEntries:%d", expiredBucket, len(entryMap))
			for _, entry := range entryMap {
				expiredItems = append(expiredItems, entry.value)
			}

			binLock.Unlock()
		}
	}
	expiredBucketLock.Unlock()

	if len(expiredItems) > 0 && b.flushCallback != nil {
		go b.flushCallback(expiredItems)
	}
}

func (b *StatsCache[T]) Compute(tsHour int64, key string, entryUpdater func(existing T) error) error {
	now := b.clock.Now()
	truncatedTimestamp := now.Truncate(b.expiry).Unix()

	hash := sha256.Sum256([]byte(key))
	binIndex := int(binary.BigEndian.Uint16(hash[:]) % b.numBins)

	b.cacheMutex.RLock()
	binMap, bucketExists := b.cache[truncatedTimestamp]
	bucketLock, lockExists := b.bucketLocks[truncatedTimestamp]
	b.cacheMutex.RUnlock()

	if !bucketExists {
		b.cacheMutex.Lock()
		// Double check that bucket wasn't added during lock upgrade
		if binMap, bucketExists = b.cache[truncatedTimestamp]; !bucketExists {
			binMap = make(map[int]map[string]*Entry[T])
			b.cache[truncatedTimestamp] = binMap
			b.bucketLocks[truncatedTimestamp] = &sync.RWMutex{}
			b.binLocks[truncatedTimestamp] = make(map[int]*sync.RWMutex)
		}
		if _, lockExists = b.bucketLocks[truncatedTimestamp]; !lockExists {
			b.bucketLocks[truncatedTimestamp] = &sync.RWMutex{}
			b.binLocks[truncatedTimestamp] = make(map[int]*sync.RWMutex)
		}
		b.cacheMutex.Unlock()
	}

	bucketLock = b.bucketLocks[truncatedTimestamp]
	bucketLock.RLock()
	entryMap, binExists := binMap[binIndex]
	bucketLock.RUnlock()

	if !binExists {
		bucketLock.Lock()
		if entryMap, binExists = binMap[binIndex]; !binExists {
			entryMap = make(map[string]*Entry[T])
			binMap[binIndex] = entryMap
			if b.binLocks[truncatedTimestamp][binIndex] == nil {
				b.binLocks[truncatedTimestamp][binIndex] = &sync.RWMutex{}
			}
		}
		bucketLock.Unlock()
	}

	binLock := b.binLocks[truncatedTimestamp][binIndex]

	binLock.RLock()
	entry, entryExists := entryMap[key]
	binLock.RUnlock()

	if entryExists {
		entry.mutex.Lock()
		defer entry.mutex.Unlock()
		return entryUpdater(entry.value)
	}

	value, err := b.initializeCallback(tsHour, key)
	if err != nil {
		return err
	}
	if err = entryUpdater(value); err != nil {
		return err
	}

	binLock.Lock()
	defer binLock.Unlock()

	// Check if entry was added during lock upgrade to avoid duplicate entries
	if _, entryExists = entryMap[key]; !entryExists {
		if len(entryMap) >= b.capacity {
			b.evictRandom(entryMap)
		}
		entryMap[key] = &Entry[T]{
			key:       key,
			value:     value,
			timestamp: now,
		}
	}

	return nil
}

func (b *StatsCache[T]) evictRandom(entryMap map[string]*Entry[T]) {
	if len(entryMap) == 0 {
		return
	}

	keys := make([]string, 0, len(entryMap))
	for key := range entryMap {
		keys = append(keys, key)
	}
	randomKey := keys[b.randSource.Intn(len(keys))]
	if b.flushCallback != nil {
		entry := entryMap[randomKey]
		b.flushCallback([]T{entry.value})
	}

	delete(entryMap, randomKey)
}

// Close stops the cleanup goroutine
func (b *StatsCache[T]) Close() {
	close(b.stopCleanup)
}
