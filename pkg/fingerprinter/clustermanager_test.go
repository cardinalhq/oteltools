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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClusterAssignments(t *testing.T) {
	cm := &ClusterManager{}

	messages := []*Message{
		{Text: "Error reading file from path /var/logs/app.log", Fingerprint: "fp1"},
		{Text: "Error reading file from path /usr/logs/app.log", Fingerprint: "fp2"},
		{Text: "Failed to connect to database at db-server", Fingerprint: "fp3"},
		{Text: "Connection error to database at db-server", Fingerprint: "fp4"},
	}

	threshold := 0.5

	messageToCluster, clusterChanges := cm.Cluster(messages, threshold)

	assert.Equal(t, len(cm.Clusters), 2, "Should create two clusters")
	assert.Equal(t, clusterChanges[messageToCluster["fp1"]], cm.Clusters[0].Fingerprint, "fp1 should belong to the first cluster")
	assert.Equal(t, clusterChanges[messageToCluster["fp2"]], cm.Clusters[0].Fingerprint, "fp2 should belong to the first cluster")

	assert.Equal(t, clusterChanges[messageToCluster["fp3"]], cm.Clusters[1].Fingerprint, "fp3 should belong to the first cluster")
	assert.Equal(t, clusterChanges[messageToCluster["fp4"]], cm.Clusters[1].Fingerprint, "fp4 should belong to the first cluster")
}

func TestStreamingClusterAssignments(t *testing.T) {
	cm := &ClusterManager{}

	streamedMessages := []*Message{
		{Text: "Error reading file from path /var/logs/app.log", Fingerprint: "fp1"},
		{Text: "Error reading file from path /usr/logs/app.log", Fingerprint: "fp2"},
		{Text: "Failed to connect to database at db-server", Fingerprint: "fp3"},
		{Text: "Connection error to database at db-server", Fingerprint: "fp4"},
	}

	threshold := 0.5

	messageToCluster := make(map[string]string)

	for i, message := range streamedMessages {
		newMessageToCluster, clusterChanges := cm.Cluster([]*Message{message}, threshold)

		for msgFingerprint, clusterFingerprint := range newMessageToCluster {
			messageToCluster[msgFingerprint] = clusterFingerprint
		}

		if len(clusterChanges) > 0 {
			for oldFp, newFp := range clusterChanges {
				for msgFp, clusterFp := range messageToCluster {
					if clusterFp == oldFp {
						messageToCluster[msgFp] = newFp
					}
				}
			}
		}

		switch i {
		case 0: // After first message, 1 cluster
			assert.Equal(t, 1, len(cm.Clusters), "Should have 1 cluster after first message")
		case 1: // After second message, same cluster as first
			assert.Equal(t, 1, len(cm.Clusters), "Should still have 1 cluster after second message")
		case 2: // After third message, 2 clusters now
			assert.Equal(t, 2, len(cm.Clusters), "Should have 2 clusters after third message")
		case 3: // After fourth message, still 2 clusters
			assert.Equal(t, 2, len(cm.Clusters), "Should still have 2 clusters after fourth message")
		}
	}

	assert.Equal(t, messageToCluster["fp1"], messageToCluster["fp2"], "fp1 and fp2 should be in the same cluster")
	assert.Equal(t, messageToCluster["fp3"], messageToCluster["fp4"], "fp3 and fp4 should be in the same cluster")
	assert.NotEqual(t, messageToCluster["fp1"], messageToCluster["fp3"], "Clusters for fp1/fp2 and fp3/fp4 should be different")
}
