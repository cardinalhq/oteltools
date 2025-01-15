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

package stringutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitQuotedStrings(t *testing.T) {
	tests := []struct {
		input    string
		expected []StringPart
	}{
		{
			input: `This is a test "and so is this" and this is not`,
			expected: []StringPart{
				LiteralStringPart{value: "This is a test"},
				QuotedStringPart{value: "and so is this"},
				LiteralStringPart{value: "and this is not"},
			},
		},
		{
			input: "This is a test 'and so is this' and this is not",
			expected: []StringPart{
				LiteralStringPart{value: "This is a test"},
				QuotedStringPart{value: "and so is this"},
				LiteralStringPart{value: "and this is not"},
			},
		},
		{
			input: "This is a test `and so is this` and this is not",
			expected: []StringPart{
				LiteralStringPart{value: "This is a test"},
				QuotedStringPart{value: "and so is this"},
				LiteralStringPart{value: "and this is not"},
			},
		},
		{
			input: `This is a test "and so is \"this\"" and this is not`,
			expected: []StringPart{
				LiteralStringPart{value: "This is a test"},
				QuotedStringPart{value: `and so is "this"`},
				LiteralStringPart{value: "and this is not"},
			},
		},
		{
			input: `This is a test "don't" and this is not`,
			expected: []StringPart{
				LiteralStringPart{value: "This is a test"},
				QuotedStringPart{value: "don't"},
				LiteralStringPart{value: "and this is not"},
			},
		},
		{
			input: `this is a test "this is quoted double 'single' and double again"`,
			expected: []StringPart{
				LiteralStringPart{value: "this is a test"},
				QuotedStringPart{value: `this is quoted double 'single' and double again`},
			},
		},
		{
			input: `this is a don't test`,
			expected: []StringPart{
				LiteralStringPart{value: "this is a don't test"},
			},
		},
		{
			input: `this is a test "this is quoted double 'single' and double again`,
			expected: []StringPart{
				LiteralStringPart{value: "this is a test"},
				QuotedStringPart{value: `this is quoted double 'single' and double again`},
			},
		},
		{
			input: `This is a test "double", 'single', and foo="bar"`,
			expected: []StringPart{
				LiteralStringPart{value: "This is a test"},
				QuotedStringPart{value: "double"},
				LiteralStringPart{value: ","},
				QuotedStringPart{value: "single"},
				LiteralStringPart{value: ", and foo="},
				QuotedStringPart{value: "bar"},
			},
		},
	}

	for _, test := range tests {
		result := SplitQuotedStrings(test.input)
		assert.Equal(t, test.expected, result, "input: %s", test.input)
	}
}

func TestLiteralStringPartValue(t *testing.T) {
	v := LiteralStringPart{value: "foo"}
	assert.Equal(t, "foo", v.Value())
}

func TestQuotedStringPartValue(t *testing.T) {
	v := QuotedStringPart{value: "foo"}
	assert.Equal(t, "foo", v.Value())
}

func TestIsLetter(t *testing.T) {
	assert.True(t, isLetter('a'))
	assert.True(t, isLetter('A'))
	assert.True(t, isLetter('z'))
	assert.True(t, isLetter('Z'))
	assert.False(t, isLetter('1'))
	assert.False(t, isLetter(' '))
	assert.False(t, isLetter('!'))
}
