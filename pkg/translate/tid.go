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

package translate

import (
	"cmp"
	"slices"
	"sync"

	"github.com/cespare/xxhash/v2"
	"go.opentelemetry.io/collector/pdata/pcommon"

	"github.com/cardinalhq/oteltools/pkg/authenv"
)

// Pool for xxhash digesters to avoid allocation per call
var hashPool = sync.Pool{
	New: func() any {
		return xxhash.New()
	},
}

// Pre-allocated separator byte slice
var colonBytes = []byte(":")

// kv holds a key-value pair for sorting
type kv struct {
	k, v string
}

func CalculateTID(extra map[string]string, rattr, sattr, iattr pcommon.Map, prefix string, environment authenv.Environment) int64 {
	// Get pooled hasher
	xx := hashPool.Get().(*xxhash.Digest)
	xx.Reset()
	defer hashPool.Put(xx)

	// Estimate capacity to avoid slice growth
	estimatedSize := len(extra) + rattr.Len() + sattr.Len() + iattr.Len()
	if environment != nil {
		estimatedSize += len(environment.Tags())
	}

	pairs := make([]kv, 0, estimatedSize)

	// Add extra map entries
	for k, v := range extra {
		pairs = append(pairs, kv{k, v})
	}

	// Add attributes with prefixes
	addKeysDirect(rattr, "resource.", &pairs)
	addKeysDirect(sattr, "scope.", &pairs)
	addKeysDirect(iattr, prefix+".", &pairs)

	// Add environment tags
	if environment != nil {
		for k, v := range environment.Tags() {
			pairs = append(pairs, kv{"env." + k, v})
		}
	}

	// Sort by key
	slices.SortFunc(pairs, func(a, b kv) int {
		return cmp.Compare(a.k, b.k)
	})

	// Write to hash
	first := true
	for _, p := range pairs {
		if p.v != "" {
			if !first {
				_, _ = xx.Write(colonBytes)
			}
			first = false
			_, _ = xx.WriteString(p.k)
			_, _ = xx.Write(colonBytes)
			_, _ = xx.WriteString(p.v)
		}
	}
	return int64(xx.Sum64())
}

func addKeysDirect(attr pcommon.Map, prefix string, pairs *[]kv) {
	attr.Range(func(k string, v pcommon.Value) bool {
		if len(k) > 0 && k[0] != '_' {
			*pairs = append(*pairs, kv{prefix + k, v.AsString()})
		}
		return true
	})
}

// addKeys is kept for backward compatibility with tests
func addKeys(attr pcommon.Map, prefix string, tags map[string]string) {
	attr.Range(func(k string, v pcommon.Value) bool {
		if len(k) > 0 && k[0] != '_' {
			tags[prefix+"."+k] = v.AsString()
		}
		return true
	})
}

// calculateTID is kept for backward compatibility with benchmarks
func calculateTID(tags map[string]string) int64 {
	xx := hashPool.Get().(*xxhash.Digest)
	xx.Reset()
	defer hashPool.Put(xx)

	keys := make([]string, 0, len(tags))
	for k := range tags {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	first := true
	for _, k := range keys {
		v := tags[k]
		if v != "" {
			if !first {
				_, _ = xx.Write(colonBytes)
			}
			first = false
			_, _ = xx.WriteString(k)
			_, _ = xx.Write(colonBytes)
			_, _ = xx.WriteString(v)
		}
	}
	return int64(xx.Sum64())
}
