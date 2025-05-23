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

// Package hashutils provides utilities for hashing objects in a stable
// way.  The hash returned should be considered stable when
// a hasher is provided.  If a hasher is not provided, the default
// hasher is used, which is currently xxhash, but may change resulting
// in different hashes.  It is OK to use the default hasher for
// temporary hashes, but for long-term storage, a specific hasher
// should be provided.  The hash returned is a 64-bit non-cryptographic
// hash.
package hashutils
