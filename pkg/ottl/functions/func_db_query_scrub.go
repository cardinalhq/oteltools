package functions

import (
	"context"
	"fmt"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl"
	"regexp"
	"strings"
)

type SqlScrubArguments[K any] struct {
	Query ottl.StringGetter[K]
}

func NewDbQueryScrubFactory[K any]() ottl.Factory[K] {
	return ottl.NewFactory("SqlQueryScrub", &SqlScrubArguments[K]{}, createDbQueryScrubFunction[K])
}

func createDbQueryScrubFunction[K any](_ ottl.FunctionContext, oArgs ottl.Arguments) (ottl.ExprFunc[K], error) {
	args, ok := oArgs.(*SqlScrubArguments[K])
	if !ok {
		return nil, fmt.Errorf("SqlQueryScrub args must be of type *SqlScrubArguments[K]")
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

func normalizeQuery(query string) string {
	query = sqlScrubRegex.ReplaceAllString(query, "?")

	tokens := strings.Fields(query)
	for i, token := range tokens {
		if containsAlphaAndNumeric(token) {
			tokens[i] = "?"
		}
	}

	return strings.Join(tokens, " ")
}
