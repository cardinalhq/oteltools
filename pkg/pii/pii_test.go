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

	"github.com/stretchr/testify/assert"
)

// TestSanitize tests both the tokenizer and the sanitizer in one go, so we can
// avoid repeating the input strings in both tokenizer and sanitizer tests.
func TestSanitize(t *testing.T) {
	tests := []struct {
		input          string
		types          []PIIType
		expectedTokens []Token
		expected       string
	}{
		{
			input:          "test input",
			expectedTokens: []Token{},
			expected:       "test input",
		},
		{
			input: "example@example.com",
			expectedTokens: []Token{
				{Type: PIITypeEmail, Value: "example@example.com"},
			},
			expected: "REDACTED",
		},
		{
			input: "test input email example@example.com",
			expectedTokens: []Token{
				{Type: PIITypeEmail, Value: "example@example.com"},
			},
			expected: "test input email REDACTED",
		},
		{
			input: "10.42.20.1",
			expectedTokens: []Token{
				{Type: PIITypeIPv4, Value: "10.42.20.1"},
			},
			expected: "10.42.20.1",
		},
		{
			input: "10.42.20.1",
			types: []PIIType{PIITypeIPv4},
			expectedTokens: []Token{
				{Type: PIITypeIPv4, Value: "10.42.20.1"},
			},
			expected: "REDACTED",
		},
		{
			input: "test input ip 10.42.20.100",
			types: []PIIType{PIITypeIPv4},
			expectedTokens: []Token{
				{Type: PIITypeIPv4, Value: "10.42.20.100"},
			},
			expected: "test input ip REDACTED",
		},
		{
			input: `test input "10.162.8.237,10.25.31.44"`,
			types: []PIIType{PIITypeIPv4},
			expectedTokens: []Token{
				{Type: PIITypeIPv4, Value: "10.162.8.237"},
				{Type: PIITypeIPv4, Value: "10.25.31.44"},
			},
			expected: `test input "REDACTED,REDACTED"`,
		},
		{
			input: "test input 8.8.8.8, sender=example@example.com",
			types: []PIIType{PIITypeIPv4},
			expectedTokens: []Token{
				{Type: PIITypeIPv4, Value: "8.8.8.8"},
				{Type: PIITypeEmail, Value: "example@example.com"},
			},
			expected: "test input REDACTED, sender=example@example.com",
		},
		{
			input: "test input 8.8.8.8, sender=example@example.com",
			types: []PIIType{PIITypeEmail},
			expectedTokens: []Token{
				{Type: PIITypeIPv4, Value: "8.8.8.8"},
				{Type: PIITypeEmail, Value: "example@example.com"},
			},
			expected: "test input 8.8.8.8, sender=REDACTED",
		},
		{
			input: "test input 8.8.8.8, sender=example@example.com",
			types: []PIIType{PIITypeIPv4, PIITypeEmail},
			expectedTokens: []Token{
				{Type: PIITypeIPv4, Value: "8.8.8.8"},
				{Type: PIITypeEmail, Value: "example@example.com"},
			},
			expected: "test input REDACTED, sender=REDACTED",
		},
		{
			input: "test input with SSN 123-45-6789",
			types: []PIIType{PIITypeSSN},
			expectedTokens: []Token{
				{Type: PIITypeSSN, Value: "123-45-6789"},
			},
			expected: "test input with SSN REDACTED",
		},
		{
			input: "test input with SSN 123-45-6789",
			types: []PIIType{PIITypeSSN, PIITypeEmail},
			expectedTokens: []Token{
				{Type: PIITypeSSN, Value: "123-45-6789"},
			},
			expected: "test input with SSN REDACTED",
		},
		{
			input:          "test input with CCN 4111-1111-1111-1112 which is not a real number",
			types:          []PIIType{PIITypeCCN},
			expectedTokens: []Token{},
			expected:       "test input with CCN 4111-1111-1111-1112 which is not a real number",
		},
		{
			input: "test input with CCN 4111 1111 1111 1111 which is a valid number",
			types: []PIIType{PIITypeCCN},
			expectedTokens: []Token{
				{Type: PIITypeCCN, Value: "4111 1111 1111 1111"},
			},
			expected: "test input with CCN REDACTED which is a valid number",
		},
		{
			input: "amex 3782 822463 10005",
			types: []PIIType{PIITypeCCN},
			expectedTokens: []Token{
				{Type: PIITypeCCN, Value: "3782 822463 10005"},
			},
			expected: "amex REDACTED",
		},
		{
			input:          "",
			expectedTokens: []Token{},
			expected:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			opts := []Option{}
			if len(tt.types) > 0 {
				opts = append(opts, WithPIITypes(tt.types...))
			}
			detector := NewDetector(opts...)
			tokens, err := detector.Tokenize(tt.input)
			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.expectedTokens, tokens)
			output := detector.Sanitize(tt.input, tokens)
			assert.Equal(t, tt.expected, output)
		})
	}
}

func TestValidCCN(t *testing.T) {
	tests := []struct {
		ccn      string
		expected bool
	}{
		{"1234567812345670", true},    // Valid CCN
		{"1234-5678-1234-5670", true}, // Valid CCN with dashes
		{"1234 5678 1234 5670", true}, // Valid CCN with spaces
		{"1234567812345678", false},   // Invalid CCN
		{"4111111111111111", true},    // Valid Visa
		{"5500000000000004", true},    // Valid MasterCard
		{"340000000000009", true},     // Valid American Express
		{"3782 822463 10005", true},   // Valid American Express with spaces
		{"30000000000004", true},      // Valid Diners Club
		{"6011000000000004", true},    // Valid Discover
		{"201400000000009", true},     // Valid enRoute
		{"3088000000000009", true},    // Valid JCB
		{"", false},                   // Empty string
		{"1234", false},               // Too short
		{"abcdabcdabcdabcd", false},   // Non-numeric
	}

	for _, tt := range tests {
		t.Run(tt.ccn, func(t *testing.T) {
			result := validCCN(tt.ccn)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func BenchmarkTokenize(b *testing.B) {
	detector := NewDetector()
	var t []Token
	var err error
	for b.Loop() {
		t, err = detector.Tokenize("test input with SSN 123-45-6789")
		if err != nil {
			b.Fail()
		}
	}
	if len(t) == 0 {
		b.Fail()
	}
}

func BenchmarkSanitize(b *testing.B) {
	detector := NewDetector()
	s := "test input with SSN 123-45-6789, phone 123-456-7890, email example@example.com, and CCN 4111-1111-1111-1111"
	tokens, _ := detector.Tokenize(s)
	for b.Loop() {
		detector.Sanitize(s, tokens)
	}
}
