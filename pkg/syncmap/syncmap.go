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

package syncmap

import (
	"maps"
	"sync"
)

// SyncMap is a thread-safe map that uses Load and Store semantics.
// The zero value of SyncMap is an empty map ready to use.  It must not be copied after first use.
type SyncMap[K comparable, V any] struct {
	sync.Mutex
	m map[K]V
}

// SetSize ets the size of the map.  This will pre-allocate the map to the specified size
// and must be called before any other operations or the call will be a no-op.
func (s *SyncMap[K, V]) SetSize(size int) {
	s.Lock()
	defer s.Unlock()
	s.ensure(size)
}

func (s *SyncMap[K, V]) ensure(size ...int) {
	if s.m == nil {
		if len(size) > 0 {
			s.m = make(map[K]V, size[0])
		} else {
			s.m = make(map[K]V)
		}
	}
}

// Clone returns a clone of the SyncMap.  The clone has a separate
// underlying map, but depending on how the values are stored, the
// values may be shared between the two maps.
// The new map has its own lock from the source.
func (s *SyncMap[K, V]) Clone() SyncMap[K, V] {
	s.Lock()
	defer s.Unlock()

	s.ensure()

	clone := make(map[K]V, len(s.m))
	maps.Copy(clone, s.m)
	return SyncMap[K, V]{m: clone}
}

// Load returns the value stored in the map for a key, or nil if no value is present.
// The ok result indicates whether value was found in the map.
func (s *SyncMap[K, V]) Load(key K) (value V, ok bool) {
	s.Lock()
	defer s.Unlock()

	s.ensure()

	value, ok = s.m[key]
	return
}

// Store sets the value for a key.
func (s *SyncMap[K, V]) Store(key K, value V) {
	s.Lock()
	defer s.Unlock()

	s.ensure()

	s.m[key] = value
}

// LoadOrStoreFunc returns the existing value for the key if present,
// otherwise it calls f and stores the result of f() in the map.
// If f returns an error, the value is not stored and the error is returned.
// The lock is held while calling f, so f must not block.
func (s *SyncMap[K, V]) LoadOrStoreFunc(key K, f func() (V, error)) (actual V, created bool, err error) {
	s.Lock()
	defer s.Unlock()

	s.ensure()

	if value, ok := s.m[key]; ok {
		return value, false, nil
	}
	value, err := f()
	if err != nil {
		return value, false, err
	}
	s.m[key] = value
	return value, true, nil
}

// ReplaceFunc calls f for a key and replaces the value with the result of f.
// If the key does not exist, the function is not called and the map is not modified.
// If the function returns an error, the value is not stored and the error is returned.
// The lock is held while calling f, so f must not block.
func (s *SyncMap[K, V]) ReplaceFunc(key K, f func(current V) (V, error)) (actual V, replaced bool, err error) {
	s.Lock()
	defer s.Unlock()

	s.ensure()

	if value, ok := s.m[key]; ok {
		actual = value
		value, err = f(value)
		if err != nil {
			return actual, false, err
		}
		s.m[key] = value
		return actual, true, nil
	}
	return actual, false, nil
}

// Replace replaces the value for a key, and returns the previous value.
func (s *SyncMap[K, V]) Replace(key K, value V) (previous V, ok bool) {
	s.Lock()
	defer s.Unlock()

	s.ensure()

	previous, ok = s.m[key]
	s.m[key] = value
	return
}

// Delete deletes the value for a key.
func (s *SyncMap[K, V]) Delete(key K) {
	s.Lock()
	defer s.Unlock()

	s.ensure()

	delete(s.m, key)
}

// Range calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration.
// The map is locked while calling f, so f must not block.
func (s *SyncMap[K, V]) Range(f func(key K, value V) bool) {
	s.Lock()
	defer s.Unlock()

	if len(s.m) == 0 {
		return
	}

	for k, v := range s.m {
		if !f(k, v) {
			return
		}
	}
}

// Keys returns a slice of all keys in the map.
func (s *SyncMap[K, V]) Keys() []K {
	s.Lock()
	defer s.Unlock()

	s.ensure()

	keys := make([]K, 0, len(s.m))
	for k := range s.m {
		keys = append(keys, k)
	}
	return keys
}

// Values returns a slice of all values in the map.
func (s *SyncMap[K, V]) Values() []V {
	s.Lock()
	defer s.Unlock()

	s.ensure()

	values := make([]V, 0, len(s.m))
	for _, v := range s.m {
		values = append(values, v)
	}
	return values
}

// Touch calls a function for a specific key.  The syncmap is
// locked while calling the function, so it must not block.
// This allows a safe way to update a value in the map without
// needing to lock the object itself.
// If the key does not exist, the funciton is not called
// and the map is not modified.
func (s *SyncMap[K, V]) Touch(key K, f func(value V) V) (found bool) {
	s.Lock()
	defer s.Unlock()

	s.ensure()

	if value, ok := s.m[key]; ok {
		s.m[key] = f(value)
		return true
	}
	return false
}

// RemoveIf calls a function for all keys in the map.  The syncmap is
// locked while calling the function, so it must not block.
// If the function returns true, the key is removed from the map.
// It is safe to "pluck" out the values from within the function.
func (s *SyncMap[K, V]) RemoveIf(f func(key K, value V) bool) {
	s.Lock()
	defer s.Unlock()

	s.ensure()

	for k, v := range s.m {
		if f(k, v) {
			delete(s.m, k)
		}
	}
}
