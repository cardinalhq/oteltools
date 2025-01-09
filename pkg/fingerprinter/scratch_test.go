// Copyright 2024 CardinalHQ, Inc
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

//
//import (
//	"fmt"
//	"testing"
//)
//
//func TestRandom(t *testing.T) {
//	fp := NewFingerprinter()
//	input := "[c977a1b7-99d5-4419-af37-cada1836b1d0]   \\u001b[1m\\u001b[36mTicket Load (1.8ms)\\u001b[0m  \\u001b[1m\\u001b[34mSELECT `tickets`.* FROM `tickets` WHERE `tickets`.`account_id` = 11 AND `tickets`.`id` IN (106624, 106625, 106509, 106545, 106603, 106486, 106560, 106605, 106562, 107137, 107139, 107013, 107063, 107077, 107086, 107068, 107011, 107025, 107007, 107618, 106975, 106954, 107481, 107453, 107368, 107579, 107584, 106953, 107327, 107314)\\u001b[0m"
//	tokenizeInput, _, err := fp.Tokenize(input)
//	if err != nil {
//		return
//	}
//	if len(tokenizeInput) == 0 {
//		return
//	}
//	for index, token := range tokenizeInput {
//		println(fmt.Sprintf("%d.%s", index, token))
//	}
//
//}
