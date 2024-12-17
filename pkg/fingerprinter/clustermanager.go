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

import (
	"crypto/sha1"
	"encoding/hex"
	"regexp"
	"sort"
	"strings"
	"sync"
)

type Message struct {
	Text        string
	Fingerprint string
}

type MessageCluster struct {
	Fingerprint string
	Messages    []*Message
}

type ClusterManager struct {
	Clusters []MessageCluster
	mu       sync.Mutex
}

func computeClusterFingerprint(messages []*Message) string {
	if len(messages) == 0 {
		return ""
	}

	commonWords := tokenizeMessage(messages[0].Text)

	for _, message := range messages[1:] {
		messageWords := tokenizeMessage(message.Text)
		for word := range commonWords {
			if _, exists := messageWords[word]; !exists {
				delete(commonWords, word)
			}
		}
	}

	commonWordList := make([]string, 0, len(commonWords))
	for word := range commonWords {
		commonWordList = append(commonWordList, word)
	}
	sort.Strings(commonWordList)

	joinedWords := strings.Join(commonWordList, " ")
	hash := sha1.Sum([]byte(joinedWords))
	return hex.EncodeToString(hash[:])
}

func tokenizeMessage(message string) map[string]struct{} {
	message = strings.ToLower(message)
	message = regexp.MustCompile(`[^\w\s]`).ReplaceAllString(message, "")
	tokens := strings.Fields(message)

	tokenSet := make(map[string]struct{})
	for _, token := range tokens {
		tokenSet[token] = struct{}{}
	}
	return tokenSet
}

func computeJaccardIndex(setA, setB map[string]struct{}) float64 {
	intersectionSize := 0
	for token := range setA {
		if _, exists := setB[token]; exists {
			intersectionSize++
		}
	}

	unionSize := len(setA) + len(setB) - intersectionSize
	if unionSize == 0 {
		return 0.0
	}

	return float64(intersectionSize) / float64(unionSize)
}

func (cm *ClusterManager) Cluster(messages []*Message, threshold float64) (map[string]string, map[string]string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	messageToCluster := make(map[string]string)
	clusterChanges := make(map[string]string)

	for _, message := range messages {
		tokenSet := tokenizeMessage(message.Text)
		found := false

		for i, cluster := range cm.Clusters {
			clusterTokenSet := tokenizeMessage(cluster.Messages[0].Text)
			if computeJaccardIndex(tokenSet, clusterTokenSet) >= threshold {
				cm.Clusters[i].Messages = append(cm.Clusters[i].Messages, message)
				found = true

				oldFingerprint := cluster.Fingerprint
				newFingerprint := computeClusterFingerprint(cm.Clusters[i].Messages)
				cm.Clusters[i].Fingerprint = newFingerprint

				if oldFingerprint != newFingerprint {
					clusterChanges[oldFingerprint] = newFingerprint
					messageToCluster[message.Fingerprint] = newFingerprint
				}
				messageToCluster[message.Fingerprint] = cluster.Fingerprint
				break
			}
		}

		if !found {
			newCluster := MessageCluster{
				Messages:    []*Message{message},
				Fingerprint: computeClusterFingerprint([]*Message{message}),
			}
			cm.Clusters = append(cm.Clusters, newCluster)
			messageToCluster[message.Fingerprint] = newCluster.Fingerprint
		}
	}
	return messageToCluster, clusterChanges
}
