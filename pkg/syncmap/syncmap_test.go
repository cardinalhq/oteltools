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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncMap_SetSize(t *testing.T) {
	var m SyncMap[int, string]
	m.SetSize(10)
	if len(m.m) != 0 {
		t.Errorf("expected map to be empty, got %d elements", len(m.m))
	}
}

func TestSyncMap_StoreAndLoad(t *testing.T) {
	var m SyncMap[int, string]
	m.Store(1, "one")
	value, ok := m.Load(1)
	if !ok || value != "one" {
		t.Errorf("expected to load 'one', got '%v'", value)
	}
}

func TestSyncMap_Delete(t *testing.T) {
	var m SyncMap[int, string]
	m.Store(1, "one")
	m.Delete(1)
	_, ok := m.Load(1)
	if ok {
		t.Errorf("expected key 1 to be deleted")
	}
}

func TestSyncMap_Clone(t *testing.T) {
	var m SyncMap[int, string]
	m.Store(1, "one")
	clone := m.Clone()
	value, ok := clone.Load(1)
	if !ok || value != "one" {
		t.Errorf("expected to load 'one' from clone, got '%v'", value)
	}
}

func TestSyncMap_Range(t *testing.T) {
	var m SyncMap[int, string]
	m.Store(1, "one")
	m.Store(2, "two")

	keys := make(map[int]bool)
	m.Range(func(key int, _ string) bool {
		keys[key] = true
		return true
	})

	assert.Equal(t, 2, len(keys))
	assert.True(t, keys[1])
	assert.True(t, keys[2])
}

func TestSyncMap_Range_Empty(t *testing.T) {
	var m SyncMap[int, string]

	keys := make(map[int]bool)
	m.Range(func(key int, _ string) bool {
		keys[key] = true
		return true
	})

	if len(keys) != 0 {
		t.Errorf("expected no keys to be present, got %v", keys)
	}
}

func TestSyncMap_Range_Break(t *testing.T) {
	var m SyncMap[int, string]
	m.Store(1, "one")
	m.Store(2, "two")

	keys := make(map[int]bool)
	m.Range(func(key int, _ string) bool {
		keys[key] = true
		return false
	})

	val, ok := keys[1]
	assert.True(t, ok)
	assert.True(t, val)
}

func TestSyncMap_Keys(t *testing.T) {
	var m SyncMap[int, string]
	m.Store(1, "one")
	m.Store(2, "two")

	keys := m.Keys()
	assert.ElementsMatch(t, keys, []int{1, 2})
}

func TestSyncMap_Values(t *testing.T) {
	var m SyncMap[int, string]
	m.Store(1, "one")
	m.Store(2, "two")

	values := m.Values()
	assert.ElementsMatch(t, values, []string{"one", "two"})
}

func TestSyncMap_Replace(t *testing.T) {
	var m SyncMap[int, string]
	m.Store(1, "one")

	previous, ok := m.Replace(1, "uno")
	if !ok || previous != "one" {
		t.Errorf("expected to replace 'one' with 'uno', got previous value '%v'", previous)
	}

	value, ok := m.Load(1)
	if !ok || value != "uno" {
		t.Errorf("expected to load 'uno', got '%v'", value)
	}

	previous, ok = m.Replace(2, "dos")
	if ok || previous != "" {
		t.Errorf("expected to replace non-existent key, got previous value '%v'", previous)
	}

	value, ok = m.Load(2)
	if !ok || value != "dos" {
		t.Errorf("expected to load 'dos', got '%v'", value)
	}
}

func TestSyncMap_Touch(t *testing.T) {
	var m SyncMap[int, string]
	m.Store(1, "one")

	ok := m.Touch(1, func(v string) string {
		return v + "!"
	})
	require.True(t, ok)

	value, ok := m.Load(1)
	require.True(t, ok)
	require.Equal(t, "one!", value)

	ok = m.Touch(2, func(v string) string {
		return v + "!"
	})
	require.False(t, ok)
}
