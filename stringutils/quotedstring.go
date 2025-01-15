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
	"strings"
)

type LiteralStringPart struct {
	value string
}

func (l LiteralStringPart) Value() string {
	return l.value
}

var _ StringPart = QuotedStringPart{}

type QuotedStringPart struct {
	value string
}

func (q QuotedStringPart) Value() string {
	return q.value
}

var _ StringPart = LiteralStringPart{}

type StringPart interface {
	Value() string
}

// SplitQuotedStrings will take a string and return a list of parts.
// Double quotes, single quotes, and backticks are supported.
// The backslash character is used to escape the quote character.
// Special cases where English contractions are used is supported, so "don't" will not be treated as
// the start or end of a quoted string.
func SplitQuotedStrings(input string) []StringPart {
	var tokens []StringPart
	var current strings.Builder

	var quoteChar rune
	var isInQuote bool
	var isEscaped bool
	var previousCharWasNotLetter = true

	runes := []rune(input)
	for _, r := range runes {
		if isEscaped {
			// Append the escaped character
			current.WriteRune(r)
			isEscaped = false
			continue
		}

		if r == '\\' {
			// Escape the next character
			isEscaped = true
			continue
		}

		if isInQuote {
			if r == quoteChar {
				// Close the quoted string
				tokens = append(tokens, QuotedStringPart{value: current.String()})
				current.Reset()
				isInQuote = false
				quoteChar = 0
			} else {
				current.WriteRune(r)
			}
		} else {
			if previousCharWasNotLetter && (r == '"' || r == '\'' || r == '`') {
				// Start a quoted string
				if current.Len() > 0 {
					tokens = append(tokens, LiteralStringPart{value: strings.TrimSpace(current.String())})
					current.Reset()
				}
				isInQuote = true
				quoteChar = r
			} else {
				current.WriteRune(r)
				previousCharWasNotLetter = !isLetter(r)
			}
		}
	}

	// Add any remaining content
	if current.Len() > 0 {
		if isInQuote {
			tokens = append(tokens, QuotedStringPart{value: current.String()})
		} else {
			tokens = append(tokens, LiteralStringPart{value: strings.TrimSpace(current.String())})
		}
	}

	return tokens
}

func isLetter(r rune) bool {
	return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z'
}
