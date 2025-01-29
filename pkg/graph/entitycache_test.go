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

package graph

import (
	"go.opentelemetry.io/collector/pdata/pcommon"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKubernetesEntityRelationships(t *testing.T) {
	ec := NewResourceEntityCache()

	attributes := pcommon.NewMap()
	attributes.PutStr(string(semconv.K8SClusterNameKey), "cluster-1")
	attributes.PutStr("k8s.cluster.region", "us-east-1")
	attributes.PutStr(string(semconv.K8SNodeNameKey), "node-1")
	attributes.PutStr("k8s.node.cpu", "16")
	attributes.PutStr(string(semconv.K8SNamespaceNameKey), "default")
	attributes.PutStr("k8s.namespace.quota", "high")
	attributes.PutStr(string(semconv.K8SPodNameKey), "pod-1")
	attributes.PutStr("k8s.pod.owner", "deployment-1")

	ec.Provision(attributes)

	entities := ec.GetAllEntities()

	expectedEntities := map[string]string{
		"cluster-1": "k8s.cluster",
		"node-1":    "k8s.node",
		"default":   "k8s.namespace",
		"pod-1":     "k8s.pod",
	}

	for name, entityType := range expectedEntities {
		entity, exists := entities[toEntityId(name, entityType)]
		assert.True(t, exists, "Expected entity %s not found", name)
		assert.Equal(t, entityType, entity.Type, "Incorrect type for entity %s", name)
	}

	assert.Equal(t, "us-east-1", entities[toEntityId("cluster-1", "k8s.cluster")].Attributes["k8s.cluster.region"])
	assert.Equal(t, "16", entities[toEntityId("node-1", "k8s.node")].Attributes["k8s.node.cpu"])
	assert.Equal(t, "high", entities[toEntityId("default", "k8s.namespace")].Attributes["k8s.namespace.quota"])
	assert.Equal(t, "deployment-1", entities[toEntityId("pod-1", "k8s.pod")].Attributes["k8s.pod.owner"])

	assert.Equal(t, HasNode, entities[toEntityId("cluster-1", "k8s.cluster")].Edges["node-1:k8s.node"])
	assert.Equal(t, BelongsToCluster, entities[toEntityId("node-1", "k8s.node")].Edges["cluster-1:k8s.cluster"])
	assert.Equal(t, ContainsPod, entities[toEntityId("default", "k8s.namespace")].Edges["pod-1:k8s.pod"])
	assert.Equal(t, BelongsToNamespace, entities[toEntityId("pod-1", "k8s.pod")].Edges["default:k8s.namespace"])
}

func TestInterdependencyBetweenRelationshipMaps(t *testing.T) {
	ec := NewResourceEntityCache()

	attributes := pcommon.NewMap()
	attributes.PutStr(string(semconv.K8SClusterNameKey), "cluster-1")
	attributes.PutStr("k8s.cluster.region", "us-east-1")
	attributes.PutStr(string(semconv.K8SNodeNameKey), "node-1")
	attributes.PutStr("k8s.node.cpu", "16")
	attributes.PutStr(string(semconv.OSNameKey), "linux")
	attributes.PutStr("os.version", "5.15.0")
	attributes.PutStr(string(semconv.K8SNamespaceNameKey), "default")
	attributes.PutStr("k8s.namespace.quota", "high")
	attributes.PutStr(string(semconv.K8SPodNameKey), "pod-1")
	attributes.PutStr("k8s.pod.owner", "deployment-1")

	ec.Provision(attributes)

	entities := ec.GetAllEntities()

	expectedEntities := map[string]string{
		"cluster-1": "k8s.cluster",
		"node-1":    "k8s.node",
		"default":   "k8s.namespace",
		"pod-1":     "k8s.pod",
		"linux":     "os",
	}

	for name, entityType := range expectedEntities {
		entityId := toEntityId(name, entityType)
		entity, exists := entities[entityId]
		assert.True(t, exists, "Expected entity %s not found", entityId)
		assert.Equal(t, entityType, entity.Type, "Incorrect type for entity %s", entityId)
	}

	assert.Equal(t, "us-east-1", entities[toEntityId("cluster-1", "k8s.cluster")].Attributes["k8s.cluster.region"])
	assert.Equal(t, "16", entities[toEntityId("node-1", "k8s.node")].Attributes["k8s.node.cpu"])
	assert.Equal(t, "5.15.0", entities[toEntityId("linux", "os")].Attributes["os.version"])
	assert.Equal(t, "high", entities[toEntityId("default", "k8s.namespace")].Attributes["k8s.namespace.quota"])
	assert.Equal(t, "deployment-1", entities[toEntityId("pod-1", "k8s.pod")].Attributes["k8s.pod.owner"])

	assert.Equal(t, HasNode, entities[toEntityId("cluster-1", "k8s.cluster")].Edges[toEntityId("node-1", "k8s.node")])
	assert.Equal(t, BelongsToCluster, entities[toEntityId("node-1", "k8s.node")].Edges[toEntityId("cluster-1", "k8s.cluster")])
	assert.Equal(t, ContainsPod, entities[toEntityId("default", "k8s.namespace")].Edges[toEntityId("pod-1", "k8s.pod")])
	assert.Equal(t, BelongsToNamespace, entities[toEntityId("pod-1", "k8s.pod")].Edges[toEntityId("default", "k8s.namespace")])
	assert.Equal(t, RunsOnOperatingSystem, entities[toEntityId("node-1", "k8s.node")].Edges[toEntityId("linux", "os")])
}

func TestContainerRelationships(t *testing.T) {
	ec := NewResourceEntityCache()

	attributes := pcommon.NewMap()
	attributes.PutStr("container.name", "my-container")
	attributes.PutStr("container.id", "abc123")
	attributes.PutStr("container.runtime", "docker")
	attributes.PutStr("container.command", "/bin/bash")
	attributes.PutStr("oci.runtime", "runc")
	attributes.PutStr("oci.version", "1.0.2")
	attributes.PutStr("container.image.name", "nginx")
	attributes.PutStr("container.image.tag", "1.19.3")

	ec.Provision(attributes)

	entities := ec.GetAllEntities()

	expectedEntities := map[string]string{
		"my-container": "container",
		"nginx":        "container.image",
	}

	for name, entityType := range expectedEntities {
		entityId := toEntityId(name, entityType)
		entity, exists := entities[entityId]
		assert.True(t, exists, "Expected entity %s not found", entityId)
		assert.Equal(t, entityType, entity.Type, "Incorrect type for entity %s", entityId)
	}

	assert.Equal(t, "abc123", entities[toEntityId("my-container", "container")].Attributes["container.id"])
	assert.Equal(t, "docker", entities[toEntityId("my-container", "container")].Attributes["container.runtime"])
	assert.Equal(t, "/bin/bash", entities[toEntityId("my-container", "container")].Attributes["container.command"])
	assert.Equal(t, "runc", entities[toEntityId("my-container", "container")].Attributes["oci.runtime"])
	assert.Equal(t, "1.0.2", entities[toEntityId("my-container", "container")].Attributes["oci.version"])

	assert.Equal(t, "1.19.3", entities[toEntityId("nginx", "container.image")].Attributes["container.image.tag"])

	assert.Equal(t, UsesImage, entities[toEntityId("my-container", "container")].Edges[toEntityId("nginx", "container.image")])
	assert.Equal(t, IsUsedByContainer, entities[toEntityId("nginx", "container.image")].Edges[toEntityId("my-container", "container")])
}

func TestCloudRelationships(t *testing.T) {
	ec := NewResourceEntityCache()

	attributes := pcommon.NewMap()
	attributes.PutStr(string(semconv.CloudProviderKey), "aws")
	attributes.PutStr(string(semconv.CloudAccountIDKey), "123456789012")
	attributes.PutStr(string(semconv.CloudRegionKey), "us-west-1")
	attributes.PutStr(string(semconv.CloudAvailabilityZoneKey), "us-west-1a")
	attributes.PutStr("cloud.account.name", "my-aws-account")

	ec.Provision(attributes)

	entities := ec.GetAllEntities()

	expectedEntities := map[string]string{
		"aws":          "cloud.provider",
		"123456789012": "cloud.account",
		"us-west-1":    "cloud.region",
		"us-west-1a":   "cloud.availability_zone",
	}

	// Check that expected entities exist
	for name, entityType := range expectedEntities {
		entityId := toEntityId(name, entityType)
		entity, exists := entities[entityId]
		assert.True(t, exists, "Expected entity %s not found", entityId)
		assert.Equal(t, entityType, entity.Type, "Incorrect type for entity %s", entityId)
	}

	// Validate attributes
	assert.Equal(t, "my-aws-account", entities[toEntityId("123456789012", "cloud.account")].Attributes["cloud.account.name"])

	// Validate relationships
	assert.Equal(t, ManagesAccount, entities[toEntityId("aws", "cloud.provider")].Edges[toEntityId("123456789012", "cloud.account")])
	assert.Equal(t, ContainsRegion, entities[toEntityId("aws", "cloud.provider")].Edges[toEntityId("us-west-1", "cloud.region")])
	assert.Equal(t, ContainsAvailabilityZone, entities[toEntityId("aws", "cloud.provider")].Edges[toEntityId("us-west-1a", "cloud.availability_zone")])
	assert.Equal(t, BelongsToProvider, entities[toEntityId("123456789012", "cloud.account")].Edges[toEntityId("aws", "cloud.provider")])
	assert.Equal(t, HasResourcesInRegion, entities[toEntityId("123456789012", "cloud.account")].Edges[toEntityId("us-west-1", "cloud.region")])
	assert.Equal(t, BelongsToProvider, entities[toEntityId("us-west-1", "cloud.region")].Edges[toEntityId("aws", "cloud.provider")])
	assert.Equal(t, ContainsAvailabilityZone, entities[toEntityId("us-west-1", "cloud.region")].Edges[toEntityId("us-west-1a", "cloud.availability_zone")])
	assert.Equal(t, ContainsResourcesFromAccount, entities[toEntityId("us-west-1", "cloud.region")].Edges[toEntityId("123456789012", "cloud.account")])
	assert.Equal(t, BelongsToRegion, entities[toEntityId("us-west-1a", "cloud.availability_zone")].Edges[toEntityId("us-west-1", "cloud.region")])
}
