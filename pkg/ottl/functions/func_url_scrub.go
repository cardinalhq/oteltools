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
	"net/url"
	"regexp"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl"
)

type UrlScrubArguments[K any] struct {
	Url ottl.StringGetter[K]
}

func NewUrlScrubFactory[K any]() ottl.Factory[K] {
	return ottl.NewFactory("UrlScrub", &UrlScrubArguments[K]{}, createUrlScrubFunction[K])
}

func createUrlScrubFunction[K any](_ ottl.FunctionContext, oArgs ottl.Arguments) (ottl.ExprFunc[K], error) {
	args, ok := oArgs.(*UrlScrubArguments[K])
	if !ok {
		return nil, fmt.Errorf("urlScrub args must be of type *UrlScrubArguments[K]")
	}
	return urlScrub(args.Url), nil
}

func urlScrub[K any](urlGetter ottl.StringGetter[K]) ottl.ExprFunc[K] {
	return func(ctx context.Context, tCtx K) (any, error) {
		url, err := urlGetter.Get(ctx, tCtx)
		if err != nil {
			return false, fmt.Errorf("failed to get url: %v", err)
		}

		scrubbedUrl := scrubQueryString(url)
		scrubbedUrl = scrubNumbersInPath(scrubbedUrl)
		return scrubbedUrl, nil
	}
}

func scrubQueryString(inputURL string) string {
	parsedURL, err := url.Parse(inputURL)
	if err != nil {
		return inputURL
	}

	query := parsedURL.Query()
	for key := range query {
		query.Set(key, "_")
	}

	parsedURL.RawQuery = query.Encode()
	return parsedURL.String()
}

func scrubNumbersInPath(inputURL string) string {
	re := regexp.MustCompile(`/\d+`)
	return re.ReplaceAllStringFunc(inputURL, func(match string) string {
		return re.ReplaceAllString(match, "/<number>")
	})
}
