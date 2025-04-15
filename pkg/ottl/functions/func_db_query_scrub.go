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

package functions

import (
	"context"
	"fmt"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl"
	"regexp"
	"strings"
)

type DBQueryScrubArguments[K any] struct {
	Query ottl.StringGetter[K]
}

func NewDbQueryScrubFactory[K any]() ottl.Factory[K] {
	return ottl.NewFactory("DBQueryScrub", &DBQueryScrubArguments[K]{}, createDbQueryScrubFunction[K])
}

func createDbQueryScrubFunction[K any](_ ottl.FunctionContext, oArgs ottl.Arguments) (ottl.ExprFunc[K], error) {
	args, ok := oArgs.(*DBQueryScrubArguments[K])
	if !ok {
		return nil, fmt.Errorf("DBQueryScrub args must be of type *DBQueryScrubArguments[K]")
	}
	return dbQueryScrub(args.Query), nil
}

func dbQueryScrub[K any](queryGetter ottl.StringGetter[K]) ottl.ExprFunc[K] {
	return func(ctx context.Context, tCtx K) (any, error) {
		query, err := queryGetter.Get(ctx, tCtx)
		if err != nil {
			return false, fmt.Errorf("failed to get query: %v", err)
		}

		scrubbedQuery := normalizeQuery(query)
		return scrubbedQuery, nil
	}
}

var sqlScrubRegex = regexp.MustCompile(`\b\d+\.\d+\b|\b\d+\b|'[^']*'|"[^"]*"`)
var doUpdateSetRegex = regexp.MustCompile(`UPDATE SET\s+(\?=\?(\.\?)?,?)*`)
var repeatingTypeRegex = regexp.MustCompile(`(?:\? \w+,?\s*)+`)

var specialChars = map[uint8]bool{
	'*': true,
	'?': true,
	'=': true,
	';': true,
	'.': true,
	',': true,
	'(': true,
	')': true,
	'_': true,
	'>': true,
	'<': true,
}

func ScrubWord(word string) string {
	var tokenList []uint8
	for j, r := range strings.ToLower(word) {
		if r >= 'a' && r <= 'z' || specialChars[uint8(r)] {
			tokenList = append(tokenList, word[j])
		} else {
			break
		}
	}
	return string(tokenList)
}

func collapsePlaceholders(s string) string {
	// If the token is enclosed in parentheses...
	if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
		// Get the inner part.
		inner := s[1 : len(s)-1]
		// If the inner part contains a comma, assume it is a batch of placeholders.
		if strings.Contains(inner, ",") {
			return "(?)"
		}
	}
	return s
}

func normalizeQuery(query string) string {
	query = sqlScrubRegex.ReplaceAllString(query, "?")
	query = doUpdateSetRegex.ReplaceAllString(query, "UPDATE SET ?=?")
	if strings.HasPrefix(query, "CREATE") {
		query = repeatingTypeRegex.ReplaceAllString(query, "")
	}

	tokens := strings.Fields(query)
	for i, token := range tokens {
		tokens[i] = ScrubWord(collapsePlaceholders(token))
	}

	return strings.Join(tokens, " ")
}
