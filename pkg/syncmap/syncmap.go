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

import "sync"

// SyncMap is a thread-safe map that uses Load and Store semantics.
// The zero value of SyncMap is an empty map ready to use.  It must not be copied after first use.
type SyncMap[K comparable, V any] struct {
	sync.Mutex
	m map[K]V
}

// Set the size of the map.  This is a hint to the map to pre-allocate memory,
// and must be called before any write operations.
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

// Clone returns a clone of the SyncMap.
func (s *SyncMap[K, V]) Clone() SyncMap[K, V] {
	s.Lock()
	defer s.Unlock()

	s.ensure()

	clone := make(map[K]V, len(s.m))
	for k, v := range s.m {
		clone[k] = v
	}
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
// The map is unlocked while calling f.
func (s *SyncMap[K, V]) Range(f func(key K, value V) bool) {
	s.Lock()

	if len(s.m) == 0 {
		s.Unlock()
		return
	}

	for k, v := range s.m {
		s.Unlock()
		if !f(k, v) {
			return
		}
		s.Lock()
	}

	s.Unlock()
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
