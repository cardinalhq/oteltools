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

// nolint

package tokenizer

import (
    "fmt"

    "github.com/db47h/ragel/v2"
)

// Token types
const (
    TokenString ragel.Token = iota
    TokenUrl
    TokenIPv4
    TokenIPv6
    TokenEmail
    TokenFQDN
)

var TokenNames = map[ragel.Token]string{
    TokenString:      "String",
    TokenUrl:         "Url",
    TokenIPv4:        "IPv4",
    TokenIPv6:        "IPv6",
    TokenEmail:       "Email",
    TokenFQDN:        "FQDN",
}

// make golangci-lint happy
var (
    _ = tokenizer_en_main
    _ = tokenizer_error
)

type PIITokenizer struct {}

func NewPIITokenizer() *PIITokenizer {
	return &PIITokenizer{}
}

func (*PIITokenizer) TokenString(t ragel.Token) string {
    if t < 0 || t >= ragel.Token(len(TokenNames)) {
        return "Token(" + fmt.Sprintf("%d", t) + ")"
    }
    return TokenNames[t]
}

// ragel state machine definition.
%%{
    machine tokenizer;

    # utf-8 support
    include UTF8 "utf8.rl";

    newline = '\n' @{ s.Newline(p) };

    main := |*
        alpha_u = uletter | '_';
        alnum_u = alpha_u | digit;

        number = digit+ ('.' digit+)? | '.' digit+ | digit + '.';

        ansicode = digit+ (';' digit+)*  uletter;

        dnslabel = alnum_u+ ('-' alnum_u+)*;
        fqdn = dnslabel ('.' dnslabel)+;
        email = alnum_u+ (('.' | '-' | '+' | '_') alnum_u+)* '@' fqdn;

        ipv4 = digit{1,3} '.' digit{1,3} '.' digit{1,3} '.' digit{1,3};

        protocol = alnum_u+;
        url_creds = (alnum_u+)? ':' (alnum_u+)? '@';
        url_path = ('/' alnum_u+)*;
        url_host = fqdn | ipv4;
        url_port = ':' digit{1,5};
        url = protocol '://' (url_creds)? url_host? url_port? url_path;
        httpmethod = [Gg][Ee][Tt] | [Pp][Oo][Ss][Tt] | [Pp][Uu][Tt] | [Dd][Ee][Ll][Ee][Tt][Ee] | [Hh][Ee][Aa][Dd] | [Pp][Aa][Tt][Cc][Hh];

        brackets = '(' | ')' | '[' | ']' | '{' | '}' | '<' | '>';
        punctuation = '.' | ',' | ';' | ':' | '!' | '?' | '"' | '\'' | '*' | '-' | '_' | '@' | '#' | '$' | '%' | '&' | '^' | '|' | '~' | '`' | '+' | '=' | '\\' | '|';
        skipcharacters = space | newline | cntrl | 0x7f | brackets |  punctuation;

        wordEndOfSentence = [a-zA-Z]+ '. ';

        # pre-filtering
        ansicode;

        ipv4 {
            s.Emit(ts, TokenIPv4, string(data[ts:te]))
        };

        url {
            s.Emit(ts, TokenUrl, string(data[ts:te]))
        };

        email {
            s.Emit(ts, TokenEmail, string(data[ts:te]))
        };

        fqdn {
            s.Emit(ts, TokenFQDN, string(data[ts:te]))
        };

        wordEndOfSentence {
            s.Emit(ts, TokenString, string(data[ts:te]))
        };

        alpha_u alnum_u* {
            s.Emit(ts, TokenString, string(data[ts:te]))
        };

        skipcharacters+;

        '/';
    *|;
}%%

%%write data nofinal;


func (PIITokenizer) Init(s *ragel.State) (int, int) {
    var cs, ts, te, act int
    %%write init;
    s.SaveVars(cs, ts, te, act)
    return %%{ write start; }%%, %%{ write error; }%%
}

func (PIITokenizer) Run(s *ragel.State, p, pe, eof int) (int, int) {
    cs, ts, te, act, data := s.GetVars()
    %%write exec;
    s.SaveVars(cs, ts, te, act)
    return p, pe
}
