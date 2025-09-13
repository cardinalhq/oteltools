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

package fingerprinter

// TODO:  include JSON (or json-like) keys as an ordered list in the fingerprint
// TODO:  add the map<string,any> as a return value for when we parse json-like content

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"unicode"

	"github.com/cespare/xxhash/v2"
	"github.com/db47h/ragel/v2"

	"github.com/cardinalhq/oteltools/maputils"
	"github.com/cardinalhq/oteltools/pkg/fingerprinter/tokenizer"
	"github.com/cardinalhq/oteltools/stringutils"
)

const (
	LogLevelPlaceHolder   = "<Loglevel>"
	IdentifierPlaceHolder = "<Identifier>"
)

type Fingerprinter interface {
	IsWord(word string) bool
	Fingerprint(input string, clusterManager *TrieClusterManager) (fingerprint int64, level string, js map[string]any, err error)
}

type fingerprinterImpl struct {
	maxTokens int
}

var _ Fingerprinter = (*fingerprinterImpl)(nil)

// Use a pattern where options can be passed into the constructor as a series of functional options.

func NewFingerprinter(opts ...Option) *fingerprinterImpl {
	fp := fingerprinterImpl{
		maxTokens: 15,
	}

	for _, opt := range opts {
		opt(&fp)
	}
	return &fp
}

type Option func(*fingerprinterImpl)

func WithMaxTokens(maxlen int) Option {
	return func(fp *fingerprinterImpl) {
		fp.maxTokens = maxlen
	}
}

func findJSONContent(input string) (string, string, string) {
	var message, jsonContent, extra string
	if strings.Contains(input, "{") && strings.Contains(input, "}") {
		start := strings.Index(input, "{")
		end := strings.LastIndex(input, "}")
		if start < end {
			message = input[:start]
			jsonContent = input[start : end+1]
			extra = input[end+1:]
		}
	}
	return message, jsonContent, extra
}

func lookupKey(bodyMap map[string]any, key string) string {
	var findKey func(map[string]any, string) string
	findKey = func(currentMap map[string]any, searchKey string) string {
		for k, v := range currentMap {
			if k == searchKey {
				if str, ok := v.(string); ok {
					return str
				}
				return ""
			}
			if nestedMap, ok := v.(map[string]any); ok {
				result := findKey(nestedMap, searchKey)
				if result != "" {
					return result
				}
			}
		}
		return ""
	}

	return findKey(bodyMap, key)
}

func getStringKey(body map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := lookupKey(body, key); value != "" {
			return value
		}
	}
	return ""
}

func (fp *fingerprinterImpl) tokenizeJSONContent(prefix string, data map[string]any, suffix string) (*tokenSeq, string, error) {
	message := getStringKey(data, "message", "msg")
	level := getStringKey(data, "level", "loglevel")
	level = strings.ToLower(level)
	if !slices.Contains(tokenizer.LogLevelNames, level) {
		level = ""
	}

	sb := getStringBuilder()
	defer putStringBuilder(sb)
	sb.WriteString(prefix)
	sb.WriteString(" ")
	sb.WriteString(level)
	sb.WriteString(message)
	sb.WriteString(" ")
	sb.WriteString(suffix)
	sb.WriteString(" ")
	body := sb.String()
	s, nlevel, err := fp.tokenizeString(body)
	if err != nil {
		return nil, "", err
	}
	if level == "" {
		level = nlevel
	}

	return s, level, nil
}

func (fp *fingerprinterImpl) Fingerprint(input string, clusterManager *TrieClusterManager) (fingerprint int64, level string, js map[string]any, err error) {
	t, level, js, err := fp.tokenizeInput(input)
	if err != nil {
		return 0, "", nil, err
	}
	defer putTokenSeq(t) // Return to pool when done

	t.jsonKeys = maputils.DeepKeys(js)
	if len(t.jsonKeys) > 0 {
		return fp.fingerprintItemsAndJSONKeys(t), level, js, nil
	}
	return clusterManager.cluster(t), level, js, nil
}

func (fp *fingerprinterImpl) fingerprintItemsAndJSONKeys(t *tokenSeq) int64 {
	h := xxhash.New()
	for i, item := range t.items {
		if i > 0 {
			_, _ = h.Write([]byte(":"))
		}
		_, _ = h.WriteString(item)
	}
	for _, key := range t.jsonKeys {
		_, _ = h.Write([]byte(":"))
		_, _ = h.WriteString(key)
	}
	return int64(h.Sum64())
}

func (fp *fingerprinterImpl) tokenizeInput(input string) (*tokenSeq, string, map[string]any, error) {
	// Do some light pre-processing here to make it easier on the ragel code.
	input = strings.TrimSpace(input)
	input = stringutils.RemoveANSICodes(input)

	prefix, jsonContent, suffix := findJSONContent(input)
	if jsonContent != "" {
		var data map[string]any
		err := json.Unmarshal([]byte(jsonContent), &data)
		if err != nil {
			// Try to see if we can just replace `=>` with `:` and parse it then.
			jsonContent = strings.ReplaceAll(jsonContent, "=>", ":")
			err = json.Unmarshal([]byte(jsonContent), &data)
		}
		if err == nil {
			tokenized, level, err := fp.tokenizeJSONContent(prefix, data, suffix)
			if err != nil {
				return newTokenSeq(), "", nil, err
			}
			return tokenized, level, data, nil
		}
	}

	// Truncate the string to the first newline or CR character
	if i := strings.IndexAny(input, "\n\r"); i != -1 {
		input = input[:i]
	}
	tokenized, level, err := fp.tokenizeString(input)
	if err != nil {
		return newTokenSeq(), "", nil, err
	}
	return tokenized, level, nil, nil
}

// test helper methods - only for internal testing
func (fp *fingerprinterImpl) testTokenizeString(input string) (*tokenSeq, string, error) {
	return fp.tokenizeString(input)
}

func (fp *fingerprinterImpl) testTokenizeInput(input string) (*tokenSeq, string, map[string]any, error) {
	return fp.tokenizeInput(input)
}

func (fp *fingerprinterImpl) IsWord(word string) bool {
	if _, exists := englishWords[strings.ToLower(word)]; exists {
		return true
	}
	// If the word is entirely uppercase or entirely lowercase, it needs to fully match.
	if strings.ToUpper(word) == word || strings.ToLower(word) == word {
		return false
	}

	words := splitWords(word)
	for _, w := range words {
		if !fp.IsWord(w) {
			return false
		}
	}
	return true
}

type tokenSeq struct {
	index    int
	items    []string
	jsonKeys []string
}

func (tm *tokenSeq) add(replacement string) {
	tm.items = append(tm.items, replacement)
	tm.index += 1
}

func newTokenSeq() *tokenSeq {
	return getTokenSeq()
}

func (fp *fingerprinterImpl) tokenizeString(input string) (*tokenSeq, string, error) {
	tk := getTokenizer()
	defer putTokenizer(tk)

	quotedStrings := getStringSlice()
	defer putStringSlice(quotedStrings)

	sb := getStringBuilder()
	defer putStringBuilder(sb)

	substrings := stringutils.SplitQuotedStrings(input)
	for _, substr := range substrings {
		switch substr.(type) {
		case stringutils.LiteralStringPart:
			if sb.Len() > 0 {
				sb.WriteString(" ")
			}
			sb.WriteString(substr.Value())
		case stringutils.QuotedStringPart:
			quotedStrings = append(quotedStrings, substr.Value())
			if sb.Len() > 0 {
				sb.WriteString(" ")
			}
			sb.WriteString("quotedstringplaceholder")
		}
	}
	targetString := sb.String()

	var err error
	tokenMap, level, err := fp.tokenizeWithTokenizer(tk, targetString, quotedStrings)
	if err != nil {
		return nil, "", err
	}

	return tokenMap, strings.ToLower(level), nil
}

func (fp *fingerprinterImpl) tokenizeWithTokenizer(tk *tokenizer.FingerprintTokenizer, input string, quotedStrings []string) (*tokenSeq, string, error) {
	level := ""
	tokenMap := newTokenSeq()
	currentQuotedStringIndex := 0

	s := ragel.New("test", strings.NewReader(input), tk)
	for {
		// Check length prior to adding the next token since we use 'continue' liberally
		if tokenMap.index >= fp.maxTokens {
			return tokenMap, strings.ToLower(level), nil
		}
		_, tok, literal := s.Next()
		lowerCaseLiteral := strings.ToLower(literal)

		switch tok {
		case ragel.EOF:
			return tokenMap, strings.ToLower(level), nil
		case ragel.Error:
			return nil, "", fmt.Errorf("error: %s", literal)
		case tokenizer.TokenQuotedString:
			if currentQuotedStringIndex < len(quotedStrings) {
				tokenMap.add("<QuotedString>")
				currentQuotedStringIndex += 1
			}
		case tokenizer.TokenList:
			quotedStringCount := strings.Count(lowerCaseLiteral, "quotedstringplaceholder")
			if currentQuotedStringIndex < len(quotedStrings) && currentQuotedStringIndex+quotedStringCount <= len(quotedStrings) {
				tokenMap.add("<List>")
			}
		case tokenizer.TokenLoglevel:
			if level == "" {
				level = literal
				tokenMap.add(LogLevelPlaceHolder)
			} else {
				tokenMap.add(lowerCaseLiteral)
			}
		case tokenizer.TokenIdentifier:
			if level == "" && slices.Contains(tokenizer.LogLevelNames, strings.ToLower(literal)) {
				level = literal
				tokenMap.add(LogLevelPlaceHolder)
				continue
			}
			if fp.IsWord(literal) {
				tokenMap.add(lowerCaseLiteral)
				continue
			}
			if len(tokenMap.items) > 0 && tokenMap.items[len(tokenMap.items)-1] != IdentifierPlaceHolder {
				tokenMap.add(IdentifierPlaceHolder)
			}
		case tokenizer.TokenString:
			if fp.IsWord(literal) {
				tokenMap.add(lowerCaseLiteral)
			}
		default:
			tokenMap.add("<" + tk.TokenString(tok) + ">")
		}
	}
}

func splitWords(input string) []string {
	var result []string
	var word strings.Builder

	for i, r := range input {
		// Check if the character is uppercase
		if unicode.IsUpper(r) {
			// If it's not the first character and the previous character is not an underscore,
			// it indicates the start of a new word
			if i != 0 && input[i-1] != '_' {
				result = append(result, word.String())
				word.Reset()
			}
			word.WriteRune(unicode.ToLower(r))
		} else if r == '_' {
			// If underscore is encountered, add the current word to the result
			if word.Len() > 0 {
				result = append(result, word.String())
				word.Reset()
			}
		} else {
			// Append lowercase characters to the current word
			word.WriteRune(r)
		}
	}

	// Add the last word
	if word.Len() > 0 {
		result = append(result, word.String())
	}

	return result
}
