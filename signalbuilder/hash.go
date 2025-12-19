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

package signalbuilder

import (
	"encoding/binary"
	"math"
	"slices"
	"sync"

	"github.com/cespare/xxhash/v2"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
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

// Pre-allocated markers for type identification
var (
	mapMarker   = []byte("map:")
	keyMarker   = []byte("key:")
	valueMarker = []byte("value:")
	strMarker   = []byte("s:")
	intMarker   = []byte("i:")
	floatMarker = []byte("f:")
	boolTrue    = []byte("bt")
	boolFalse   = []byte("bf")
	sliceStart  = []byte("[")
	sliceEnd    = []byte("]")
	bytesMarker = []byte("b:")
	nullMarker  = []byte("null")
	mapStart    = []byte("{")
	mapEnd      = []byte("}")
)

func attrkey(attr pcommon.Map) uint64 {
	if attr.Len() == 0 {
		return 1
	}
	h := getHasher()
	defer putHasher(h)
	hashPcommonMap(h, attr)
	return h.Sum64()
}

// hashPcommonMap hashes a pcommon.Map directly without calling AsRaw()
func hashPcommonMap(h *xxhash.Digest, m pcommon.Map) {
	_, _ = h.Write(mapMarker)

	// Collect and sort keys for stable ordering
	keys := make([]string, 0, m.Len())
	m.Range(func(k string, _ pcommon.Value) bool {
		keys = append(keys, k)
		return true
	})
	slices.Sort(keys)

	// Hash each key-value pair
	for _, k := range keys {
		_, _ = h.Write(keyMarker)
		_, _ = h.WriteString(k)
		_, _ = h.Write(valueMarker)
		v, _ := m.Get(k)
		hashPcommonValue(h, v)
	}
}

// hashPcommonValue hashes a pcommon.Value directly
func hashPcommonValue(h *xxhash.Digest, v pcommon.Value) {
	switch v.Type() {
	case pcommon.ValueTypeStr:
		_, _ = h.Write(strMarker)
		_, _ = h.WriteString(v.Str())
	case pcommon.ValueTypeInt:
		_, _ = h.Write(intMarker)
		var buf [8]byte
		binary.LittleEndian.PutUint64(buf[:], uint64(v.Int()))
		_, _ = h.Write(buf[:])
	case pcommon.ValueTypeDouble:
		_, _ = h.Write(floatMarker)
		var buf [8]byte
		binary.LittleEndian.PutUint64(buf[:], math.Float64bits(v.Double()))
		_, _ = h.Write(buf[:])
	case pcommon.ValueTypeBool:
		if v.Bool() {
			_, _ = h.Write(boolTrue)
		} else {
			_, _ = h.Write(boolFalse)
		}
	case pcommon.ValueTypeBytes:
		_, _ = h.Write(bytesMarker)
		_, _ = h.Write(v.Bytes().AsRaw())
	case pcommon.ValueTypeSlice:
		_, _ = h.Write(sliceStart)
		slice := v.Slice()
		for i := range slice.Len() {
			hashPcommonValue(h, slice.At(i))
		}
		_, _ = h.Write(sliceEnd)
	case pcommon.ValueTypeMap:
		_, _ = h.Write(mapStart)
		hashPcommonMap(h, v.Map())
		_, _ = h.Write(mapEnd)
	case pcommon.ValueTypeEmpty:
		_, _ = h.Write(nullMarker)
	}
}

func metrickey(name string, units string, ty pmetric.MetricType) uint64 {
	h := getHasher()
	defer putHasher(h)
	_, _ = h.WriteString(name)
	_, _ = h.Write([]byte{0})
	_, _ = h.WriteString(units)
	_, _ = h.Write([]byte{0, byte(ty)})
	return h.Sum64()
}

func scopekey(name, version, schemaURL string, attr pcommon.Map) uint64 {
	if attr.Len() == 0 && name == "" && version == "" && schemaURL == "" {
		return 1
	}
	h := getHasher()
	defer putHasher(h)
	// Include name, version, and schema URL in the hash for uniqueness
	_, _ = h.WriteString(name)
	_, _ = h.Write([]byte{0})
	_, _ = h.WriteString(version)
	_, _ = h.Write([]byte{0})
	_, _ = h.WriteString(schemaURL)
	_, _ = h.Write([]byte{0})
	hashPcommonMap(h, attr)
	return h.Sum64()
}
