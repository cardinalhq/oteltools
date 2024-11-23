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
	wasInitialized  bool
	incrementCalled bool
	count           int64
	name            string
	key             uint64
}

func (m *MockStatsObject) Key() uint64 { return m.key }

func (m *MockStatsObject) Matches(other StatsObject) bool {
	if o, ok := other.(*MockStatsObject); ok {
		return m.name == o.name
	}
	return false
}

func (m *MockStatsObject) Increment(_ string, count int, _ int64) error {
	m.incrementCalled = true
	m.count += int64(count)
	return nil
}

func (m *MockStatsObject) Initialize() error {
	m.wasInitialized = true
	return nil
}
