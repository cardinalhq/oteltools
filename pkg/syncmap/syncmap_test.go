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
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncMap_SetSize(t *testing.T) {
	var m SyncMap[int, string]
	m.SetSize(10)
	assert.Len(t, m.m, 0)
}

func TestSyncMap_StoreAndLoad(t *testing.T) {
	var m SyncMap[int, string]
	m.Store(1, "one")
	value, ok := m.Load(1)
	assert.True(t, ok)
	assert.Equal(t, "one", value)
}

func TestSyncMap_Delete(t *testing.T) {
	var m SyncMap[int, string]
	m.Store(1, "one")
	m.Delete(1)
	_, ok := m.Load(1)
	assert.False(t, ok)
}

func TestSyncMap_Clone(t *testing.T) {
	var m SyncMap[int, string]
	m.Store(1, "one")
	clone := m.Clone()
	value, ok := clone.Load(1)
	assert.True(t, ok)
	assert.Equal(t, "one", value)
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
	assert.True(t, ok)
	assert.Equal(t, "one", previous)

	value, ok := m.Load(1)
	assert.True(t, ok)
	assert.Equal(t, "uno", value)

	previous, ok = m.Replace(2, "dos")
	assert.False(t, ok)

	value, ok = m.Load(2)
	assert.True(t, ok)
	assert.Equal(t, "dos", value)
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

func TestSyncMap_LoadOrStore(t *testing.T) {
	var m SyncMap[int, string]

	// Test storing a new value
	value, err := m.LoadOrStore(1, func() (string, error) {
		return "one", nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "one", value)

	// Test loading an existing value
	value, err = m.LoadOrStore(1, func() (string, error) {
		return "uno", nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "one", value)

	// Test storing another new value
	value, err = m.LoadOrStore(2, func() (string, error) {
		return "two", nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "two", value)

	wantErr := errors.New("wanted error")
	// Test erroring when storing a new value
	value, err = m.LoadOrStore(3, func() (string, error) {
		return "", wantErr
	})
	assert.Error(t, err)

	// Verify the map contents
	val1, ok1 := m.Load(1)
	assert.True(t, ok1)
	assert.Equal(t, "one", val1)

	val2, ok2 := m.Load(2)
	assert.True(t, ok2)
	assert.Equal(t, "two", val2)

	_, ok3 := m.Load(3)
	assert.False(t, ok3)
}

func TestSyncMap_RemoveIf(t *testing.T) {
	var m SyncMap[int, string]
	m.Store(1, "one")
	m.Store(2, "two")
	m.Store(3, "three")

	// Remove keys where the value contains the letter 'o'
	m.RemoveIf(func(key int, value string) bool {
		return strings.Contains(value, "o")
	})

	// Verify the remaining keys and values
	_, ok1 := m.Load(1)
	assert.False(t, ok1) // "one" should be removed

	_, ok2 := m.Load(2)
	assert.False(t, ok2) // "two" should be removed

	val3, ok3 := m.Load(3)
	assert.True(t, ok3) // "three" should remain
	assert.Equal(t, "three", val3)
}

func TestSyncMap_RemoveIf_NoMatch(t *testing.T) {
	var m SyncMap[int, string]
	m.Store(1, "one")
	m.Store(2, "two")
	m.Store(3, "three")

	// Remove keys where the value contains the letter 'z' (no match)
	m.RemoveIf(func(key int, value string) bool {
		return strings.Contains(value, "z")
	})

	// Verify all keys and values remain
	val1, ok1 := m.Load(1)
	assert.True(t, ok1)
	assert.Equal(t, "one", val1)

	val2, ok2 := m.Load(2)
	assert.True(t, ok2)
	assert.Equal(t, "two", val2)

	val3, ok3 := m.Load(3)
	assert.True(t, ok3)
	assert.Equal(t, "three", val3)
}

func TestSyncMap_RemoveIf_AllMatch(t *testing.T) {
	var m SyncMap[int, string]
	m.Store(1, "one")
	m.Store(2, "two")
	m.Store(3, "three")

	// Remove all keys
	m.RemoveIf(func(key int, value string) bool {
		return true
	})

	// Verify the map is empty
	_, ok1 := m.Load(1)
	assert.False(t, ok1)

	_, ok2 := m.Load(2)
	assert.False(t, ok2)

	_, ok3 := m.Load(3)
	assert.False(t, ok3)
}
