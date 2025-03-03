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

func normalizeQuery(query string) string {
	query = sqlScrubRegex.ReplaceAllString(query, "?")

	tokens := strings.Fields(query)
	for i, token := range tokens {
		tokens[i] = ScrubWord(token)
	}

	return strings.Join(tokens, " ")
}
