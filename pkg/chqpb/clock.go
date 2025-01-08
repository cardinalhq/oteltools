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

package chqpb

import "time"

type Clock interface {
	Now() time.Time
	After(d time.Duration) <-chan time.Time
}

type RealClock struct{}

func (RealClock) Now() time.Time {
	return time.Now()
}

func (RealClock) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}

type MockClock struct {
	currentTime time.Time
	channels    []chan time.Time
}

func NewMockClock(start time.Time) *MockClock {
	return &MockClock{
		currentTime: start,
		channels:    []chan time.Time{},
	}
}

func (m *MockClock) Now() time.Time {
	return m.currentTime
}

func (m *MockClock) After(d time.Duration) <-chan time.Time {
	ch := make(chan time.Time, 1)
	m.channels = append(m.channels, ch)
	return ch
}

func (m *MockClock) Advance(d time.Duration) {
	m.currentTime = m.currentTime.Add(d)
	for _, ch := range m.channels {
		ch <- m.currentTime
	}
	m.channels = nil // Clear channels after firing
}
