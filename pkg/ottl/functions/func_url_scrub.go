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
	"sort"
	"strings"

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
		urlValue, err := urlGetter.Get(ctx, tCtx)
		if err != nil {
			return false, fmt.Errorf("failed to get url: %v", err)
		}

		parsedURL, err := url.Parse(urlValue)
		if err != nil {
			return urlValue, nil
		}

		if isIPv4Address(parsedURL.Host) {
			parsedURL.Host = "<ip>"
		}

		parsedURL.RawQuery = scrubQueryString(parsedURL.RawQuery)
		parsedURL.Path = scrubPath(parsedURL.Path)

		unescaped, err := url.PathUnescape(parsedURL.String())
		if err != nil {
			return urlValue, nil
		}
		return unescaped, nil
	}
}

func scrubQueryString(queryString string) string {
	if queryString == "" {
		return queryString
	}

	paramMap := make(map[string]string)
	var keys []string

	params := strings.Split(queryString, "&")
	for _, param := range params {
		parts := strings.SplitN(param, "=", 2)
		key := parts[0]
		paramMap[key] = "_"
		keys = append(keys, key)
	}

	sort.Strings(keys)

	var sortedParams []string
	for _, key := range keys {
		sortedParams = append(sortedParams, key+"="+paramMap[key])
	}

	return strings.Join(sortedParams, "&")
}

func scrubPath(path string) string {
	if path == "/" {
		return path
	}

	segments := strings.Split(path, "/")

	for i, segment := range segments {
		if isVersion(segment) {
			continue
		} else if containsAlphaAndNumeric(segment) {
			segments[i] = "<value>"
		} else if isNumber(segment) {
			segments[i] = "<number>"
		}
	}

	return strings.Join(segments, "/")
}

var (
	letterRegex         = regexp.MustCompile(`[a-zA-Z]`)
	numberRegex         = regexp.MustCompile(`\d`)
	multipleNumberRegex = regexp.MustCompile(`\d+$`)
	versionRegex        = regexp.MustCompile(`^v\d+$`)
	ipv4Regex           = regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`)
)

func isNumber(segment string) bool {
	return multipleNumberRegex.MatchString(segment)
}

func isVersion(segment string) bool {
	return versionRegex.MatchString(segment)
}

func isIPv4Address(segment string) bool {
	return ipv4Regex.MatchString(segment)
}

func containsAlphaAndNumeric(segment string) bool {
	hasLetter := letterRegex.MatchString(segment)
	hasNumber := numberRegex.MatchString(segment)
	return hasLetter && hasNumber
}
