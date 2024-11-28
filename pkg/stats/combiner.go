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

package stats

import (
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.etcd.io/bbolt"
)

type StatsObject interface {
	Key() uint64
	Matches(other StatsObject) bool
	Increment(key string, count int, size int64) error
	Initialize() error
}

type StatsCombiner[T StatsObject] struct {
	sync.Mutex
	interval     time.Duration
	cutoff       time.Time
	db           *bbolt.DB
	bucketBase   string
	serializer   func(T) ([]byte, error)
	deserializer func([]byte) (T, error)
}

func (sc *StatsCombiner[T]) Record(now time.Time, item T, incKey string, count int, logSize int64) (map[uint64][]T, error) {
	sc.Lock()
	defer sc.Unlock()

	currentBucket := sc.currentBucketName(now)
	fmt.Printf("Writing to bucket: %s\n", currentBucket)

	err := sc.db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(currentBucket))
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create bucket: %w", err)
	}

	key := item.Key()
	var flushedData map[uint64][]T

	err = sc.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(currentBucket))
		if bucket == nil {
			return errors.New("bucket not found")
		}

		keyBytes := uint64ToBytes(key)
		data := bucket.Get(keyBytes)

		var existing T
		if data != nil {
			var err error
			existing, err = sc.deserializer(data)
			if err != nil {
				return fmt.Errorf("failed to deserialize existing object: %w", err)
			}
			if err := existing.Increment(incKey, count, logSize); err != nil {
				return fmt.Errorf("failed to increment object: %w", err)
			}
		} else {
			if err := item.Initialize(); err != nil {
				return fmt.Errorf("failed to initialize object: %w", err)
			}
			existing = item
		}

		serialized, err := sc.serializer(existing)
		if err != nil {
			return fmt.Errorf("failed to serialize object: %w", err)
		}
		return bucket.Put(keyBytes, serialized)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update record: %w", err)
	}

	// Flush if the time interval has passed
	if now.After(sc.cutoff) {
		flushedData, err = sc.flush(now)
		if err != nil {
			return nil, fmt.Errorf("failed to flush data: %w", err)
		}
	}

	return flushedData, nil
}

func (sc *StatsCombiner[T]) flush(now time.Time) (map[uint64][]T, error) {
	flushedData := make(map[uint64][]T)
	currentBucket := sc.getBucketName(sc.cutoff.Unix() - int64(sc.interval.Seconds()))
	println("Flushing from bucket", currentBucket)

	err := sc.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(currentBucket))
		if bucket == nil {
			return nil
		}

		err := bucket.ForEach(func(k, v []byte) error {
			key := bytesToUint64(k)
			item, err := sc.deserializer(v)
			if err != nil {
				return fmt.Errorf("failed to deserialize object: %w", err)
			}
			fmt.Printf("Flushing item: %v\n", string(v))
			flushedData[key] = append(flushedData[key], item)
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to collect data: %w", err)
		}

		return tx.DeleteBucket([]byte(currentBucket))
	})
	if err != nil {
		return nil, fmt.Errorf("flush failed: %w", err)
	}

	sc.cutoff = now.Truncate(sc.interval).Add(sc.interval)

	return flushedData, nil
}

func NewStatsCombiner[T StatsObject](
	db *bbolt.DB,
	bucketBase string,
	now time.Time,
	interval time.Duration,
	serializer func(T) ([]byte, error),
	deserializer func([]byte) (T, error),
) *StatsCombiner[T] {
	return &StatsCombiner[T]{
		interval:     interval,
		cutoff:       now.Truncate(interval).Add(interval),
		db:           db,
		bucketBase:   bucketBase,
		serializer:   serializer,
		deserializer: deserializer,
	}
}

func (sc *StatsCombiner[T]) currentBucketName(t time.Time) string {
	return fmt.Sprintf("%s_%d", sc.bucketBase, t.Truncate(sc.interval).Unix())
}

func (sc *StatsCombiner[T]) getBucketName(t int64) string {
	return fmt.Sprintf("%s_%d", sc.bucketBase, t)
}

func uint64ToBytes(i uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, i)
	return b
}

func bytesToUint64(b []byte) uint64 {
	if len(b) < 8 {
		return 0
	}
	return binary.BigEndian.Uint64(b)
}
