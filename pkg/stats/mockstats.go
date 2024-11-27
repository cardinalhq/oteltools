// Copyright 2024 CardinalHQ, Inc
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

package stats

type MockStatsObject struct {
	Id              uint64 `json:"key"`
	Count           int    `json:"count"`
	Size            int64  `json:"size"`
	wasInitialized  bool
	incrementCalled bool
}

func (m *MockStatsObject) Key() uint64 {
	return m.Id
}

func (m *MockStatsObject) Matches(other StatsObject) bool {
	if o, ok := other.(*MockStatsObject); ok {
		return o.Id == m.Id
	}
	return false
}

func (m *MockStatsObject) Increment(key string, count int, size int64) error {
	m.Count += count
	m.Size += size
	m.incrementCalled = true
	return nil
}

func (m *MockStatsObject) Initialize() error {
	m.wasInitialized = true
	return nil
}
