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

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"sync"
	"unicode"

	"github.com/cespare/xxhash/v2"
	"github.com/db47h/ragel/v2"

	"github.com/cardinalhq/oteltools/pkg/fingerprinter/tokenizer"
	"github.com/cardinalhq/oteltools/stringutils"
)

// OptimizedFingerprinter reduces memory allocations through aggressive pooling and reuse
type OptimizedFingerprinter struct {
	maxTokens int
}

// ragelScanner wraps ragel scanner with a reusable reader
type ragelScanner struct {
	tk      *tokenizer.FingerprintTokenizer
	reader  *bytes.Reader
	scanner *ragel.Scanner
}

var (
	// Pool for ragel scanners with readers
	scannerPool = sync.Pool{
		New: func() any {
			tk := tokenizer.NewFingerprintTokenizer()
			reader := bytes.NewReader([]byte{})
			return &ragelScanner{
				tk:      tk,
				reader:  reader,
				scanner: ragel.New("fingerprint", reader, tk),
			}
		},
	}

	// Pool for map[string]any used in JSON parsing
	jsonMapPool = sync.Pool{
		New: func() any {
			return make(map[string]any, 8)
		},
	}

	// Pool for JSON key slices
	jsonKeySlicePool = sync.Pool{
		New: func() any {
			return make([]string, 0, 16)
		},
	}
)

func NewOptimizedFingerprinter(opts ...Option) *OptimizedFingerprinter {
	fp := &OptimizedFingerprinter{
		maxTokens: 15,
	}

	// Apply any options
	for _, opt := range opts {
		opt((*fingerprinterImpl)(fp))
	}

	return fp
}

func (fp *OptimizedFingerprinter) IsWord(word string) bool {
	if _, exists := englishWords[strings.ToLower(word)]; exists {
		return true
	}
	// If the word is entirely uppercase or entirely lowercase, it needs to fully match.
	if strings.ToUpper(word) == word || strings.ToLower(word) == word {
		return false
	}

	words := splitWordsOptimized(word)
	for _, w := range words {
		if !fp.IsWord(w) {
			return false
		}
	}
	return true
}

func (fp *OptimizedFingerprinter) Fingerprint(input string, clusterManager *TrieClusterManager) (fingerprint int64, level string, js map[string]any, err error) {
	t, level, js, err := fp.tokenizeInputOptimized(input)
	if err != nil {
		return 0, "", nil, err
	}
	defer putTokenSeq(t)

	// Optimize JSON key extraction
	if len(js) > 0 {
		jsonKeys := getJSONKeySlice()
		defer putJSONKeySlice(jsonKeys)

		collectKeysOptimized(js, "", jsonKeys)
		t.jsonKeys = jsonKeys

		if len(jsonKeys) > 0 {
			return fp.fingerprintItemsAndJSONKeysOptimized(t), level, js, nil
		}
	}

	return clusterManager.cluster(t), level, js, nil
}

func (fp *OptimizedFingerprinter) tokenizeInputOptimized(input string) (*tokenSeq, string, map[string]any, error) {
	// Light preprocessing
	input = strings.TrimSpace(input)
	input = stringutils.RemoveANSICodes(input)

	// Check for JSON content
	prefix, jsonContent, suffix := findJSONContentOptimized(input)
	if jsonContent != "" {
		data := getJSONMap()
		defer putJSONMap(data)

		// Try to parse JSON efficiently
		if err := parseJSONOptimized(jsonContent, data); err == nil {
			tokenized, level, err := fp.tokenizeJSONContentOptimized(prefix, data, suffix)
			if err != nil {
				return newTokenSeq(), "", nil, err
			}

			// Return a copy of data to avoid issues with pooled map
			result := make(map[string]any, len(data))
			for k, v := range data {
				result[k] = v
			}
			return tokenized, level, result, nil
		}
	}

	// Truncate to first newline
	if i := strings.IndexAny(input, "\n\r"); i != -1 {
		input = input[:i]
	}

	tokenized, level, err := fp.tokenizeStringOptimized(input)
	if err != nil {
		return newTokenSeq(), "", nil, err
	}
	return tokenized, level, nil, nil
}

func (fp *OptimizedFingerprinter) tokenizeStringOptimized(input string) (*tokenSeq, string, error) {
	tk := getTokenizer()
	defer putTokenizer(tk)

	quotedStrings := getStringSlice()
	defer putStringSlice(quotedStrings)

	sb := getStringBuilder()
	defer putStringBuilder(sb)

	// Extract quoted strings
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
	tokenMap, level, err := fp.tokenizeWithTokenizerOptimized(tk, targetString, quotedStrings)
	if err != nil {
		return nil, "", err
	}

	return tokenMap, strings.ToLower(level), nil
}

func (fp *OptimizedFingerprinter) tokenizeWithTokenizerOptimized(tk *tokenizer.FingerprintTokenizer, input string, quotedStrings []string) (*tokenSeq, string, error) {
	level := ""
	tokenMap := newTokenSeq()
	currentQuotedStringIndex := 0

	// Get scanner from pool and reuse it
	scanner := getScanner()
	defer putScanner(scanner)

	// Reset reader with new input
	scanner.reader.Reset([]byte(input))

	// Reset the ragel scanner to use the new reader content
	scanner.scanner.Reset()

	// Use the pooled scanner
	s := scanner.scanner

	for {
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
				currentQuotedStringIndex++
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

func (fp *OptimizedFingerprinter) tokenizeJSONContentOptimized(prefix string, data map[string]any, suffix string) (*tokenSeq, string, error) {
	message := getStringKeyOptimized(data, "message", "msg")
	level := getStringKeyOptimized(data, "level", "loglevel")
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
	s, nlevel, err := fp.tokenizeStringOptimized(body)
	if err != nil {
		return nil, "", err
	}
	if level == "" {
		level = nlevel
	}

	return s, level, nil
}

func (fp *OptimizedFingerprinter) fingerprintItemsAndJSONKeysOptimized(t *tokenSeq) int64 {
	h := xxhash.New()
	for i, item := range t.items {
		if i > 0 {
			h.Write([]byte(":"))
		}
		h.WriteString(item)
	}
	for _, key := range t.jsonKeys {
		h.Write([]byte(":"))
		h.WriteString(key)
	}
	return int64(h.Sum64())
}

// Optimized helper functions

func findJSONContentOptimized(input string) (string, string, string) {
	start := strings.IndexByte(input, '{')
	if start == -1 {
		return "", "", ""
	}

	end := strings.LastIndexByte(input, '}')
	if end == -1 || end <= start {
		return "", "", ""
	}

	return input[:start], input[start : end+1], input[end+1:]
}

func parseJSONOptimized(jsonContent string, target map[string]any) error {
	// Clear the target map
	for k := range target {
		delete(target, k)
	}

	// Try direct unmarshal first
	if err := json.Unmarshal([]byte(jsonContent), &target); err != nil {
		// Try replacing => with : for Ruby-style hashes
		jsonContent = strings.ReplaceAll(jsonContent, "=>", ":")
		return json.Unmarshal([]byte(jsonContent), &target)
	}
	return nil
}

func getStringKeyOptimized(body map[string]any, keys ...string) string {
	for _, key := range keys {
		if v, ok := body[key]; ok {
			if str, ok := v.(string); ok {
				return str
			}
		}
	}
	return ""
}

func collectKeysOptimized(m map[string]any, prefix string, result []string) {
	for k, v := range m {
		fullKey := k
		if prefix != "" {
			fullKey = prefix + "." + k
		}
		result = append(result, fullKey)

		if nested, ok := v.(map[string]any); ok {
			collectKeysOptimized(nested, fullKey, result)
		}
	}
}

func splitWordsOptimized(input string) []string {
	var result []string
	var word []byte

	for i, r := range input {
		if unicode.IsUpper(r) {
			if i != 0 && input[i-1] != '_' {
				if len(word) > 0 {
					result = append(result, string(word))
					word = word[:0]
				}
			}
			word = append(word, byte(unicode.ToLower(r)))
		} else if r == '_' {
			if len(word) > 0 {
				result = append(result, string(word))
				word = word[:0]
			}
		} else {
			word = append(word, byte(r))
		}
	}

	if len(word) > 0 {
		result = append(result, string(word))
	}

	return result
}

// Pool helper functions

func getScanner() *ragelScanner {
	return scannerPool.Get().(*ragelScanner)
}

func putScanner(s *ragelScanner) {
	// Scanner can be reused - no size limits needed for reader
	scannerPool.Put(s)
}

func getJSONMap() map[string]any {
	m := jsonMapPool.Get().(map[string]any)
	// Clear the map
	for k := range m {
		delete(m, k)
	}
	return m
}

func putJSONMap(m map[string]any) {
	if len(m) > 64 {
		return
	}
	jsonMapPool.Put(m)
}

func getJSONKeySlice() []string {
	s := jsonKeySlicePool.Get().([]string)
	return s[:0]
}

func putJSONKeySlice(s []string) {
	if cap(s) > 128 {
		return
	}
	jsonKeySlicePool.Put(s)
}
