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
	"strings"

	"github.com/db47h/ragel/v2"

	"github.com/cardinalhq/oteltools/pkg/pii/tokenizer"
)

type Detector interface {
	TokenizeInput(input string) ([]Token, error)
	Tokenize(input string) ([]Token, error)
}

type DetectorImpl struct {
}

var _ Detector = (*DetectorImpl)(nil)

type Token struct {
	Type  ragel.Token
	Value string
}

func NewDetector(opts ...Option) *DetectorImpl {
	fp := DetectorImpl{}

	for _, opt := range opts {
		opt(&fp)
	}
	return &fp
}

type Option func(*DetectorImpl)

func (fp *DetectorImpl) TokenizeInput(input string) ([]Token, error) {
	message := strings.TrimSpace(input)

	return fp.Tokenize(message)
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
			return nil, fmt.Errorf("error: %s", literal)
		default:
			tokens = append(tokens, Token{tok, literal})
		}
	}
}
