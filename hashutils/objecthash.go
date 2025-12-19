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
	"math"
	"reflect"
	"sort"
	"sync"

	"github.com/cespare/xxhash/v2"
)

// hasherPool pools xxhash.Digest instances to reduce allocations
var hasherPool = sync.Pool{
	New: func() any {
		return xxhash.New()
	},
}

func getHasher() *xxhash.Digest {
	return hasherPool.Get().(*xxhash.Digest)
}

func putHasher(h *xxhash.Digest) {
	h.Reset()
	hasherPool.Put(h)
}

// Hasher is an interface that wraps a Sum64 method.  Any hash implementation that
// implements Sum64 can be used with HashAny.
type Hasher interface {
	Write(p []byte) (n int, err error)
	Sum64() uint64
}

// HashJSON takes a blob of JSON, unmarshals it, and returns a 64-bit hash using the provided hasher.
// If there's an error unmarshaling, it returns 0 and the error.
func HashJSON(hasher Hasher, input []byte) (uint64, error) {
	if hasher == nil {
		h := getHasher()
		defer putHasher(h)
		var v any
		if err := json.Unmarshal(input, &v); err != nil {
			return 0, err
		}
		writeHash(h, v)
		return h.Sum64(), nil
	}
	var v any
	if err := json.Unmarshal(input, &v); err != nil {
		return 0, err
	}
	writeHash(hasher, v)
	return hasher.Sum64(), nil
}

// HashAny hashes an arbitrary Go value (the result of json.Unmarshal) in a
// canonical way using the provided hasher. It returns a 64-bit non-cryptographic hash.
func HashAny(hasher Hasher, value any) uint64 {
	if hasher == nil {
		h := getHasher()
		defer putHasher(h)
		writeHash(h, value)
		return h.Sum64()
	}
	writeHash(hasher, value)
	return hasher.Sum64()
}

func HashStrings(hasher Hasher, values ...string) uint64 {
	if hasher == nil {
		h := getHasher()
		defer putHasher(h)
		for _, value := range values {
			_, _ = h.WriteString(value)
			_, _ = h.Write([]byte{0})
		}
		return h.Sum64()
	}
	for _, value := range values {
		_, _ = hasher.Write([]byte(value + "\x00"))
	}
	return hasher.Sum64()
}

// Pre-allocated byte slices for common type markers
var (
	nilMarker       = []byte("nil")
	boolTrueMarker  = []byte("bool:true")
	boolFalseMarker = []byte("bool:false")
	sliceMarker     = []byte("slice:")
	mapMarker       = []byte("map:")
	keyMarker       = []byte("key:")
	valueMarker     = []byte("value:")
	structMarker    = []byte("struct:")
	ptrMarker       = []byte("ptr:")
	colonMarker     = []byte(":")
)

// writeHash serializes and writes values into the hasher
func writeHash(h Hasher, value any) {
	if value == nil {
		_, _ = h.Write(nilMarker)
		return
	}

	v := reflect.ValueOf(value)

	switch v.Kind() {
	case reflect.Bool:
		if v.Bool() {
			_, _ = h.Write(boolTrueMarker)
		} else {
			_, _ = h.Write(boolFalseMarker)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var buf [8]byte
		binary.LittleEndian.PutUint64(buf[:], uint64(v.Int()))
		_, _ = h.Write(buf[:])
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		var buf [8]byte
		binary.LittleEndian.PutUint64(buf[:], v.Uint())
		_, _ = h.Write(buf[:])
	case reflect.Float32, reflect.Float64:
		var buf [8]byte
		binary.LittleEndian.PutUint64(buf[:], math.Float64bits(v.Float()))
		_, _ = h.Write(buf[:])
	case reflect.Complex64, reflect.Complex128:
		c := v.Complex()
		var buf [16]byte
		binary.LittleEndian.PutUint64(buf[:8], math.Float64bits(real(c)))
		binary.LittleEndian.PutUint64(buf[8:], math.Float64bits(imag(c)))
		_, _ = h.Write(buf[:])
	case reflect.String:
		_, _ = h.Write([]byte(v.String()))
	case reflect.Slice, reflect.Array:
		_, _ = h.Write(sliceMarker)
		for i := range v.Len() {
			writeHash(h, v.Index(i).Interface())
		}
	case reflect.Map:
		_, _ = h.Write(mapMarker)
		keys := make([]string, 0, v.Len())
		for _, key := range v.MapKeys() {
			keys = append(keys, fmt.Sprintf("%v", key.Interface()))
		}
		sort.Strings(keys) // Ensure stable ordering
		for _, key := range keys {
			_, _ = h.Write(keyMarker)
			_, _ = h.Write([]byte(key))
			_, _ = h.Write(valueMarker)
			writeHash(h, v.MapIndex(reflect.ValueOf(key)).Interface())
		}
	case reflect.Struct:
		_, _ = h.Write(structMarker)
		t := v.Type()
		for i := range v.NumField() {
			if !t.Field(i).IsExported() {
				continue
			}
			_, _ = h.Write([]byte(t.Field(i).Name))
			_, _ = h.Write(colonMarker)
			writeHash(h, v.Field(i).Interface())
		}
	case reflect.Ptr:
		_, _ = h.Write(ptrMarker)
		if !v.IsNil() {
			writeHash(h, v.Elem().Interface()) // Dereference pointer
		}
	default:
		_, _ = fmt.Fprintf(h, "unknown:%T", value)
	}
}
