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

package pii

import (
	"testing"

	"github.com/cardinalhq/oteltools/pkg/pii/tokenizer"
	"github.com/stretchr/testify/assert"
)

func TestTokenizeInput(t *testing.T) {
	detector := NewDetector()

	tests := []struct {
		input    string
		expected []Token
		hasError bool
	}{
		{
			input: "test input",
			expected: []Token{
				{Type: tokenizer.TokenString, Value: "test"},
				{Type: tokenizer.TokenString, Value: "input"},
			},
			hasError: false,
		},
		{
			input: "  test input with spaces  ",
			expected: []Token{
				{Type: tokenizer.TokenString, Value: "test"},
				{Type: tokenizer.TokenString, Value: "input"},
				{Type: tokenizer.TokenString, Value: "with"},
				{Type: tokenizer.TokenString, Value: "spaces"},
			},
			hasError: false,
		},
		{
			input: "example@example.com",
			expected: []Token{
				{Type: tokenizer.TokenEmail, Value: "example@example.com"},
			},
		},
		{
			input: "test input email example@example.com",
			expected: []Token{
				{Type: tokenizer.TokenString, Value: "test"},
				{Type: tokenizer.TokenString, Value: "input"},
				{Type: tokenizer.TokenString, Value: "email"},
				{Type: tokenizer.TokenEmail, Value: "example@example.com"},
			},
			hasError: false,
		},
		{
			input: "10.42.20.1",
			expected: []Token{
				{Type: tokenizer.TokenIPv4, Value: "10.42.20.1"},
			},
		},
		{
			input: "test input ip 10.42.20.100",
			expected: []Token{
				{Type: tokenizer.TokenString, Value: "test"},
				{Type: tokenizer.TokenString, Value: "input"},
				{Type: tokenizer.TokenString, Value: "ip"},
				{Type: tokenizer.TokenIPv4, Value: "10.42.20.100"},
			},
		},
		{
			input:    "",
			expected: []Token{},
			hasError: false,
		},
	}

	for _, tt := range tests {
		tokens, err := detector.TokenizeInput(tt.input)
		if tt.hasError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, tokens)
		}
	}
}
