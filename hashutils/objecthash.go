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

package hashutils

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash"
	"math"
	"sort"
)

// Hasher is an interface that wraps a Sum64 method.  Any hash implementation that
// implements Sum64 can be used with HashAny.
type Hasher interface {
	hash.Hash
	Sum64() uint64
}

// HashJSON takes a blob of JSON, unmarshals it, and returns a 64-bit hash using the provided hasher.
// If there's an error unmarshaling, it returns 0 and the error.
func HashJSON(input []byte, hasher Hasher) (uint64, error) {
	var v any
	if err := json.Unmarshal(input, &v); err != nil {
		return 0, err
	}
	return HashAny(v, hasher), nil
}

// HashAny hashes an arbitrary Go value (the result of json.Unmarshal) in a
// canonical way using the provided hasher. It returns a 64-bit non-cryptographic hash.
func HashAny(value any, hasher Hasher) uint64 {
	hashValue(hasher, value)
	return hasher.Sum64()
}

// hashValue writes a stable representation of 'value' into the hasher.
func hashValue(h hash.Hash, value any) {
	switch v := value.(type) {
	case nil:
		h.Write([]byte("null:"))
	case bool:
		if v {
			h.Write([]byte("bool:true:"))
		} else {
			h.Write([]byte("bool:false:"))
		}

	case float64:
		h.Write([]byte("num:"))
		buf := make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, math.Float64bits(v))
		h.Write(buf)
		h.Write([]byte(":"))

	case string:
		h.Write([]byte("str:"))
		h.Write([]byte(v))
		h.Write([]byte(":"))

	case []any:
		h.Write([]byte("arr["))
		for _, elem := range v {
			hashValue(h, elem)
			h.Write([]byte(","))
		}
		h.Write([]byte("]"))

	case map[string]any:
		h.Write([]byte("obj{"))
		keys := make([]string, 0, len(v))
		for key := range v {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			h.Write([]byte("key:"))
			h.Write([]byte(key))
			h.Write([]byte("="))
			hashValue(h, v[key])
			h.Write([]byte(","))
		}
		h.Write([]byte("}"))

	default:
		h.Write([]byte(fmt.Sprintf("unknown:%T", v)))
	}
}
