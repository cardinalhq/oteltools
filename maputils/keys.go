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

package maputils

import (
	"slices"
	"strconv"
)

// DeepKeys walks the map and return all keys with . separator, sorted
func DeepKeys(m map[string]any) []string {
	k := collectKeys(m)
	slices.Sort(k)
	return k
}

func collectKeys(data map[string]any) []string {
	var result []string
	for key, value := range data {
		result = append(result, traverseKey(key, value)...)
	}
	return result
}

func traverseKey(prefix string, value any) []string {
	switch v := value.(type) {
	case map[string]any:
		return traverseMap(prefix, v)
	case []any:
		if isSliceOfMaps(v) {
			return traverseSlice(prefix, v) // Only traverse the maps in the slice
		}
		return []string{prefix} // Include the slice key itself
	default:
		return []string{prefix}
	}
}

func traverseMap(prefix string, m map[string]any) []string {
	var keys []string
	for key, value := range m {
		fullKey := prefix + "." + key
		keys = append(keys, traverseKey(fullKey, value)...)
	}
	return keys
}

func traverseSlice(prefix string, s []any) []string {
	var keys []string
	for i, value := range s {
		// If the element is a map, traverse it. Otherwise, ignore.
		if mapVal, ok := value.(map[string]any); ok {
			fullKey := prefix + "." + strconv.Itoa(i)
			keys = append(keys, traverseMap(fullKey, mapVal)...)
		}
	}
	return keys
}

func isSliceOfMaps(s []any) bool {
	for _, value := range s {
		if _, ok := value.(map[string]any); ok {
			return true
		}
	}
	return false
}
