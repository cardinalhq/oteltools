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

package accumulator

// All accumulators are treated as a histogram,  This might not be
// very memory efficient, but it lets us treat them almost identically
// in the rest of the code.
type Accumulator[T int64 | float64] interface {
	Add(value []T) error
	Buckets() []T
	Count() uint64
	Sum() []T
	Avg() []T
	Max() []T
	Min() []T
}
