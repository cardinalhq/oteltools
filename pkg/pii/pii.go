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
	"fmt"
	"slices"
	"strings"

	"github.com/db47h/ragel/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/cardinalhq/oteltools/pkg/pii/tokenizer"
	"github.com/cardinalhq/oteltools/pkg/telemetry"
)

type Detector interface {
	Tokenize(input string) ([]Token, error)
	Sanitize(input string, tokens []Token) string
}

type PIIType int

const (
	PIITypeEmail = PIIType(tokenizer.TokenEmail)
	PIITypeIPv4  = PIIType(tokenizer.TokenIPv4)
	PIITypeURL   = PIIType(tokenizer.TokenURL)
	PIITypeFQDN  = PIIType(tokenizer.TokenFQDN)
	PIITypePhone = PIIType(tokenizer.TokenPhone)
	PIITypeSSN   = PIIType(tokenizer.TokenSSN)
	PIITypeCCN   = PIIType(tokenizer.TokenCCN)
)

func (t PIIType) String() string {
	switch t {
	case PIITypeEmail:
		return "email"
	case PIITypeIPv4:
		return "ipv4"
	case PIITypeURL:
		return "url"
	case PIITypeFQDN:
		return "fqdn"
	case PIITypePhone:
		return "phone"
	case PIITypeSSN:
		return "ssn"
	case PIITypeCCN:
		return "ccn"
	default:
		return "unknown"
	}
}

type DetectorImpl struct {
	piiTypes            []PIIType
	redactedReplacement string
	detectionStats      telemetry.DeferrableCounter[int64]
	redactionStats      telemetry.DeferrableCounter[int64]
}

var _ Detector = (*DetectorImpl)(nil)

type Token struct {
	Type  PIIType
	Value string
}

func NewDetector(opts ...Option) *DetectorImpl {
	fp := DetectorImpl{
		piiTypes: []PIIType{
			PIITypeEmail,
			PIITypePhone,
			PIITypeSSN,
			PIITypeCCN,
		},
		redactedReplacement: "REDACTED",
	}

	for _, opt := range opts {
		opt(&fp)
	}
	return &fp
}

type Option func(*DetectorImpl)

func WithPIITypes(types ...PIIType) Option {
	return func(fp *DetectorImpl) {
		fp.piiTypes = types
	}
}

func WithRedactedReplacement(replacement string) Option {
	return func(fp *DetectorImpl) {
		fp.redactedReplacement = replacement
	}
}

func WithRedactionStats(stats telemetry.DeferrableCounter[int64]) Option {
	return func(fp *DetectorImpl) {
		fp.redactionStats = stats
	}
}

func WithDetectionStats(stats telemetry.DeferrableCounter[int64]) Option {
	return func(fp *DetectorImpl) {
		fp.detectionStats = stats
	}
}

func (fp *DetectorImpl) Tokenize(input string) ([]Token, error) {
	tk := tokenizer.NewPIITokenizer()
	s := ragel.New("test", strings.NewReader(input), tk)
	tokens := []Token{}
	for {
		_, tok, literal := s.Next()
		switch tok {
		case ragel.EOF:
			return tokens, nil
		case ragel.Error:
			return nil, fmt.Errorf("tokenization: %s", literal)
		case tokenizer.TokenString:
			continue
		case tokenizer.TokenCCN:
			if validCCN(literal) {
				tokens = append(tokens, Token{PIIType(tok), literal})
			}
		default:
			tokens = append(tokens, Token{PIIType(tok), literal})
		}
	}
}

func (fp *DetectorImpl) Sanitize(input string, tokens []Token) string {
	for _, t := range tokens {
		attrset := attribute.NewSet(
			attribute.String("pii.type", t.Type.String()),
		)
		telemetry.CounterAdd(fp.detectionStats, 1, metric.WithAttributeSet(attrset))
		if len(fp.piiTypes) == 0 || slices.Contains(fp.piiTypes, t.Type) {
			telemetry.CounterAdd(fp.redactionStats, 1, metric.WithAttributeSet(attrset))
			input = strings.ReplaceAll(input, t.Value, fp.redactedReplacement)
		}
	}
	return input
}

func validCCN(ccn string) bool {
	// Luhn algorithm
	// https://en.wikipedia.org/wiki/Luhn_algorithm
	// https://rosettacode.org/wiki/Luhn_test_of_credit_card_numbers
	ccn = strings.ReplaceAll(ccn, "-", "")
	ccn = strings.ReplaceAll(ccn, " ", "")
	if len(ccn) < 13 || len(ccn) > 19 {
		return false
	}
	sum := 0
	parity := len(ccn) % 2
	for i, c := range ccn {
		digit := int(c - '0')
		if digit < 0 || digit > 9 {
			return false
		}
		if i%2 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}

	return sum%10 == 0
}
