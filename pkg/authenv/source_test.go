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

package authenv

import (
	"testing"
)

func TestParseEnvironmentSource(t *testing.T) {
	tests := []struct {
		input    EnvironmentSourceString
		expected EnvironmentSource
		err      error
	}{
		{"auth", EnvironmentSourceAuth, nil},
		{"env", EnvironmentSourceEnv, nil},
		{"invalid", 0, ErrInvalidEnvironmentSource},
	}

	for _, test := range tests {
		result, err := ParseEnvironmentSource(test.input)
		if result != test.expected || err != test.err {
			t.Errorf("ParseEnvironmentSource(%q) = (%v, %v), want (%v, %v)", test.input, result, err, test.expected, test.err)
		}
	}
}
