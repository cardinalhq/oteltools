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
	"fmt"
	"testing"

	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"

	"go.opentelemetry.io/collector/pdata/pcommon"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (ec *ResourceEntityCache) allEntities() map[ResourceEntityEdgeKey]*ResourceEntity {
	entities := make(map[ResourceEntityEdgeKey]*ResourceEntity)

	ec.entityMap.Range(func(key ResourceEntityEdgeKey, value *ResourceEntity) bool {
		entities[key] = value
		return true
	})

	return entities
}

func TestKubernetesEntityRelationships(t *testing.T) {
	ec := NewResourceEntityCache()

	attributes := pcommon.NewMap()
	attributes.PutStr(string(semconv.K8SClusterNameKey), "cluster-1")
	attributes.PutStr(string(semconv.K8SClusterUIDKey), "us-east-1")
	attributes.PutStr(string(semconv.K8SNodeNameKey), "node-1")
	attributes.PutStr(string(semconv.K8SNodeUIDKey), "16")
	attributes.PutStr(string(semconv.K8SNamespaceNameKey), "default")
	attributes.PutStr(string(semconv.K8SPodNameKey), "pod-1")
	attributes.PutStr(string(semconv.K8SPodUIDKey), "pod-uid-1")
	attributes.PutStr("k8s.pod.ip", "127.0.0.1")
	attributes.PutStr("k8s.pod.label.company-name", "cardinal")
	attributes.PutStr(string(semconv.ServiceNameKey), "service-1")
	attributes.PutStr(string(semconv.K8SDeploymentNameKey), "deployment-1")
	attributes.PutStr(string(semconv.K8SReplicaSetNameKey), "replicaset-1")

	ec.ProvisionResourceAttributes(attributes)

	entities := ec.allEntities()

	expectedEdges := map[string]string{
		"cluster-1":    "k8s.cluster",
		"node-1":       "k8s.node",
		"default":      "k8s.namespace",
		"pod-1":        "k8s.pod",
		"service-1":    "service",
		"deployment-1": "k8s.deployment",
		"replicaset-1": "k8s.replicaset",
	}

	for name, entityType := range expectedEdges {
		entity, exists := entities[toEntityKey(name, entityType, "")]
		assert.True(t, exists, "Expected entity %s not found", name)
		assert.Equal(t, entityType, entity.Type, "Incorrect type for entity %s", name)
	}

	assert.Equal(t, "us-east-1", entities[toEntityKey("cluster-1", "k8s.cluster", "")].Attributes[string(semconv.K8SClusterUIDKey)])
	assert.Equal(t, "16", entities[toEntityKey("node-1", "k8s.node", "")].Attributes[string(semconv.K8SNodeUIDKey)])
	assert.Equal(t, "pod-uid-1", entities[toEntityKey("pod-1", "k8s.pod", "")].Attributes[string(semconv.K8SPodUIDKey)])
	assert.Equal(t, "127.0.0.1", entities[toEntityKey("pod-1", "k8s.pod", "")].Attributes["k8s.pod.ip"])
	assert.Equal(t, "cardinal", entities[toEntityKey("pod-1", "k8s.pod", "")].Attributes["k8s.pod.label.company-name"])

	assert.Equal(t, HasNode, entities[toEntityKey("cluster-1", "k8s.cluster", "")].Edges[toEntityKey("node-1", "k8s.node", "")].Relationship)
	assert.Equal(t, BelongsToCluster, entities[toEntityKey("node-1", "k8s.node", "")].Edges[toEntityKey("cluster-1", "k8s.cluster", "")].Relationship)
	assert.Equal(t, ContainsService, entities[toEntityKey("default", "k8s.namespace", "")].Edges[toEntityKey("service-1", "service", "")].Relationship)
	assert.Equal(t, BelongsToNamespace, entities[toEntityKey("service-1", "service", "")].Edges[toEntityKey("default", "k8s.namespace", "")].Relationship)
	assert.Equal(t, IsManagedByDeployment, entities[toEntityKey("service-1", "service", "")].Edges[toEntityKey("deployment-1", "k8s.deployment", "")].Relationship)
	assert.Equal(t, ManagesReplicaset, entities[toEntityKey("deployment-1", "k8s.deployment", "")].Edges[toEntityKey("replicaset-1", "k8s.replicaset", "")].Relationship)
	assert.Equal(t, ContainsPod, entities[toEntityKey("replicaset-1", "k8s.replicaset", "")].Edges[toEntityKey("pod-1", "k8s.pod", "")].Relationship)
}

func TestInterdependencyBetweenRelationshipMaps(t *testing.T) {
	ec := NewResourceEntityCache()

	attributes := pcommon.NewMap()
	attributes.PutStr(string(semconv.K8SClusterNameKey), "cluster-1")
	attributes.PutStr(string(semconv.K8SNodeNameKey), "node-1")
	attributes.PutStr(string(semconv.OSNameKey), "linux")
	attributes.PutStr(string(semconv.K8SNamespaceNameKey), "default")
	attributes.PutStr(string(semconv.K8SPodNameKey), "pod-1")
	attributes.PutStr(string(semconv.ServiceNameKey), "service1")

	ec.ProvisionResourceAttributes(attributes)

	entities := ec.allEntities()

	expectedEntities := map[string]string{
		"cluster-1": "k8s.cluster",
		"node-1":    "k8s.node",
		"default":   "k8s.namespace",
		"pod-1":     "k8s.pod",
		"linux":     "os",
		"service1":  "service",
	}

	for name, entityType := range expectedEntities {
		entityId := toEntityKey(name, entityType, "")
		entity, exists := entities[entityId]
		assert.True(t, exists, "Expected entity %s not found", entityId)
		assert.Equal(t, entityType, entity.Type, "Incorrect type for entity %s", entityId)
	}

	assert.Equal(t, HasNode, entities[toEntityKey("cluster-1", "k8s.cluster", "")].Edges[toEntityKey("node-1", "k8s.node", "")].Relationship)
	assert.Equal(t, BelongsToCluster, entities[toEntityKey("node-1", "k8s.node", "")].Edges[toEntityKey("cluster-1", "k8s.cluster", "")].Relationship)
	assert.Equal(t, ContainsService, entities[toEntityKey("default", "k8s.namespace", "")].Edges[toEntityKey("service1", "service", "")].Relationship)
	assert.Equal(t, BelongsToNamespace, entities[toEntityKey("service1", "service", "")].Edges[toEntityKey("default", "k8s.namespace", "")].Relationship)
	assert.Equal(t, RunsOnOperatingSystem, entities[toEntityKey("node-1", "k8s.node", "")].Edges[toEntityKey("linux", "os", "")].Relationship)
}

func TestContainerRelationships(t *testing.T) {
	ec := NewResourceEntityCache()

	attributes := pcommon.NewMap()

	// Container attributes
	attributes.PutStr(string(semconv.ContainerNameKey), "my-container")
	attributes.PutStr(string(semconv.ContainerIDKey), "container-123")
	attributes.PutStr(string(semconv.ContainerCommandKey), "/bin/bash")
	attributes.PutStr(string(semconv.ContainerCommandArgsKey), "-c echo hello")
	attributes.PutStr(string(semconv.ContainerRuntimeKey), "docker")
	attributes.PutStr(string(semconv.ContainerCommandLineKey), "/bin/bash -c echo hello")
	attributes.PutStr(string(semconv.ContainerImageIDKey), "sha256:abcdef123456")
	attributes.PutStr(string(semconv.ContainerImageTagsKey), "latest")
	attributes.PutStr(string(semconv.ContainerImageRepoDigestsKey), "nginx@sha256:abcdef123456")
	attributes.PutStr("container.label.owner", "team-a")

	// Process attributes
	attributes.PutStr(string(semconv.ProcessCommandKey), "java")
	attributes.PutStr(string(semconv.ProcessExecutableNameKey), "java")
	attributes.PutStr(string(semconv.ProcessExecutablePathKey), "/usr/bin/java")
	attributes.PutStr(string(semconv.ProcessCommandArgsKey), "-jar myapp.jar")
	attributes.PutStr(string(semconv.ProcessCommandLineKey), "java -jar myapp.jar")
	attributes.PutStr(string(semconv.ProcessOwnerKey), "root")
	attributes.PutStr(string(semconv.ProcessCreationTimeKey), "1700000000")
	attributes.PutStr(string(semconv.ProcessContextSwitchTypeKey), "voluntary")
	attributes.PutStr(string(semconv.ProcessGroupLeaderPIDKey), "1001")
	attributes.PutStr(string(semconv.ProcessParentPIDKey), "1000")
	attributes.PutStr(string(semconv.ProcessPIDKey), "2000")

	ec.ProvisionResourceAttributes(attributes)

	entities := ec.allEntities()

	expectedEntities := map[string]string{
		"my-container": "container",
		"java":         "process",
	}

	for name, entityType := range expectedEntities {
		entityID := toEntityKey(name, entityType, "")
		entity, exists := entities[entityID]
		assert.True(t, exists, "Expected entity %s not found", entityID)
		assert.Equal(t, entityType, entity.Type, "Incorrect type for entity %s", entityID)
	}

	assert.Equal(t, "container-123", entities[toEntityKey("my-container", "container", "")].Attributes[string(semconv.ContainerIDKey)])
	assert.Equal(t, "/bin/bash", entities[toEntityKey("my-container", "container", "")].Attributes[string(semconv.ContainerCommandKey)])
	assert.Equal(t, "-c echo hello", entities[toEntityKey("my-container", "container", "")].Attributes[string(semconv.ContainerCommandArgsKey)])
	assert.Equal(t, "docker", entities[toEntityKey("my-container", "container", "")].Attributes[string(semconv.ContainerRuntimeKey)])
	assert.Equal(t, "/bin/bash -c echo hello", entities[toEntityKey("my-container", "container", "")].Attributes[string(semconv.ContainerCommandLineKey)])
	assert.Equal(t, "team-a", entities[toEntityKey("my-container", "container", "")].Attributes["container.label.owner"])

	assert.Equal(t, "java", entities[toEntityKey("java", "process", "")].Attributes[string(semconv.ProcessExecutableNameKey)])
	assert.Equal(t, "/usr/bin/java", entities[toEntityKey("java", "process", "")].Attributes[string(semconv.ProcessExecutablePathKey)])
	assert.Equal(t, "-jar myapp.jar", entities[toEntityKey("java", "process", "")].Attributes[string(semconv.ProcessCommandArgsKey)])
	assert.Equal(t, "java -jar myapp.jar", entities[toEntityKey("java", "process", "")].Attributes[string(semconv.ProcessCommandLineKey)])
	assert.Equal(t, "root", entities[toEntityKey("java", "process", "")].Attributes[string(semconv.ProcessOwnerKey)])
	assert.Equal(t, "1700000000", entities[toEntityKey("java", "process", "")].Attributes[string(semconv.ProcessCreationTimeKey)])
	assert.Equal(t, "voluntary", entities[toEntityKey("java", "process", "")].Attributes[string(semconv.ProcessContextSwitchTypeKey)])
	assert.Equal(t, "1001", entities[toEntityKey("java", "process", "")].Attributes[string(semconv.ProcessGroupLeaderPIDKey)])
	assert.Equal(t, "1000", entities[toEntityKey("java", "process", "")].Attributes[string(semconv.ProcessParentPIDKey)])
	assert.Equal(t, "2000", entities[toEntityKey("java", "process", "")].Attributes[string(semconv.ProcessPIDKey)])

}

func TestDBRelationships(t *testing.T) {
	ec := NewResourceEntityCache()
	attributes := pcommon.NewMap()
	attributes.PutStr(string(semconv.ServiceNameKey), "service-1")
	globalEntityMap := ec.ProvisionResourceAttributes(attributes)

	recordAttributes := pcommon.NewMap()
	recordAttributes.PutStr(string(semconv.DBSystemKey), "mysql")
	recordAttributes.PutStr(string(semconv.DBCollectionNameKey), "glacier.tbl_17665234232")
	ec.ProvisionRecordAttributes(globalEntityMap, recordAttributes)
	dbCollectionEntityId := toEntityKey("glacier.tbl_", "database.collection", "")

	entities := ec.allEntities()
	_, exists := entities[dbCollectionEntityId]
	assert.True(t, exists, "Expected entity %s not found", dbCollectionEntityId)
	key := toEntityKey("service-1", "service", "")
	entity, found := entities[key]
	require.True(t, found)
	require.NotNil(t, entity)
	fmt.Printf("entity: %#v\n", entity)
	edge, found := entity.Edges[dbCollectionEntityId]
	require.True(t, found)
	require.NotNil(t, edge)
	assert.Equal(t, UsesDatabaseCollection, edge.Relationship)
}

func TestMessagingConsumesFromRelationship(t *testing.T) {
	ec := NewResourceEntityCache()
	attributes := pcommon.NewMap()
	attributes.PutStr(string(semconv.ServiceNameKey), "service-1")
	globalEntityMap := ec.ProvisionResourceAttributes(attributes)

	recordAttributes := pcommon.NewMap()
	recordAttributes.PutStr(string(semconv.MessagingSystemKey), "kafka")
	recordAttributes.PutStr(string(semconv.MessagingDestinationNameKey), "topic-1")
	recordAttributes.PutStr(string(semconv.MessagingOperationNameKey), "process")
	recordAttributes.PutStr(string(semconv.MessagingConsumerGroupNameKey), "consumer-group-1")

	ec.ProvisionRecordAttributes(globalEntityMap, recordAttributes)
	entities := ec.allEntities()
	messagingDestinationEntityId := toEntityKey("topic-1", MessagingDestination, "")
	messagingConsumerGroupEntityId := toEntityKey("consumer-group-1", MessagingConsumerGroup, "")

	_, messagingDestinationExists := entities[messagingDestinationEntityId]
	assert.True(t, messagingDestinationExists, "Expected entity %s not found", messagingDestinationEntityId)

	_, messagingConsumerGroupEntityExists := entities[messagingConsumerGroupEntityId]
	assert.True(t, messagingConsumerGroupEntityExists, "Expected entity %s not found", messagingConsumerGroupEntityId)

	// assert edge between service and messaging destination, and assert the edge relationship
	assert.Equal(t, ConsumesFrom, entities[toEntityKey("service-1", "service", "")].Edges[messagingDestinationEntityId].Relationship)
	assert.Equal(t, ConsumesFrom, entities[toEntityKey("service-1", "service", "")].Edges[messagingConsumerGroupEntityId].Relationship)
}

func TestMessagingProducesToRelationship(t *testing.T) {
	ec := NewResourceEntityCache()
	attributes := pcommon.NewMap()
	attributes.PutStr(string(semconv.ServiceNameKey), "service-1")
	globalEntityMap := ec.ProvisionResourceAttributes(attributes)

	recordAttributes := pcommon.NewMap()
	recordAttributes.PutStr(string(semconv.MessagingSystemKey), "kafka")
	recordAttributes.PutStr(string(semconv.MessagingDestinationNameKey), "topic-1")
	recordAttributes.PutStr(string(semconv.MessagingOperationTypeKey), "publish")

	ec.ProvisionRecordAttributes(globalEntityMap, recordAttributes)
	entities := ec.allEntities()
	messagingDestinationEntityId := toEntityKey("topic-1", MessagingDestination, "")

	_, messagingDestinationExists := entities[messagingDestinationEntityId]
	assert.True(t, messagingDestinationExists, "Expected entity %s not found", messagingDestinationEntityId)

	// assert edge between service and messaging destination, and assert the edge relationship
	assert.Equal(t, ProducesTo, entities[toEntityKey("service-1", "service", "")].Edges[messagingDestinationEntityId].Relationship)
}

func TestCloudRelationships(t *testing.T) {
	ec := NewResourceEntityCache()

	attributes := pcommon.NewMap()
	attributes.PutStr(string(semconv.CloudProviderKey), "aws")
	attributes.PutStr(string(semconv.CloudAccountIDKey), "123456789012")
	attributes.PutStr(string(semconv.CloudRegionKey), "us-west-1")
	attributes.PutStr(string(semconv.CloudAvailabilityZoneKey), "us-west-1a")

	ec.ProvisionResourceAttributes(attributes)

	entities := ec.allEntities()

	expectedEntities := map[string]string{
		"aws":          "cloud.provider",
		"123456789012": "cloud.account",
		"us-west-1":    "cloud.region",
		"us-west-1a":   "cloud.availability_zone",
	}

	// Check that expected entities exist
	for name, entityType := range expectedEntities {
		entityId := toEntityKey(name, entityType, "")
		entity, exists := entities[entityId]
		assert.True(t, exists, "Expected entity %s not found", entityId)
		assert.Equal(t, entityType, entity.Type, "Incorrect type for entity %s", entityId)
	}

	// Validate relationships
	key := toEntityKey("aws", "cloud.provider", "")
	lookup := entities[key]
	require.NotNil(t, lookup)
	edge, found := lookup.Edges[toEntityKey("123456789012", "cloud.account", "")]
	require.True(t, found)
	require.NotNil(t, edge)
	assert.Equal(t, ManagesAccount, edge.Relationship)
	edge, found = lookup.Edges[toEntityKey("us-west-1", "cloud.region", "")]
	require.True(t, found)
	require.NotNil(t, edge)
	assert.Equal(t, ContainsRegion, edge.Relationship)

	// assert.Equal(t, ContainsAvailabilityZone, entities[toEntityKey("aws", "cloud.provider", "")].Edges[toEntityKey("us-west-1a", "cloud.availability_zone", "")].Relationship)
	// assert.Equal(t, BelongsToProvider, entities[toEntityKey("123456789012", "cloud.account", "")].Edges[toEntityKey("aws", "cloud.provider", "")].Relationship)
	// assert.Equal(t, HasResourcesInRegion, entities[toEntityKey("123456789012", "cloud.account", "")].Edges[toEntityKey("us-west-1", "cloud.region", "")].Relationship)
	// assert.Equal(t, BelongsToProvider, entities[toEntityKey("us-west-1", "cloud.region", "")].Edges[toEntityKey("aws", "cloud.provider", "")].Relationship)
	// assert.Equal(t, ContainsAvailabilityZone, entities[toEntityKey("us-west-1", "cloud.region", "")].Edges[toEntityKey("us-west-1a", "cloud.availability_zone", "")].Relationship)
	// assert.Equal(t, BelongsToAccount, entities[toEntityKey("us-west-1", "cloud.region", "")].Edges[toEntityKey("123456789012", "cloud.account", "")].Relationship)
	// assert.Equal(t, BelongsToRegion, entities[toEntityKey("us-west-1a", "cloud.availability_zone", "")].Edges[toEntityKey("us-west-1", "cloud.region", "")].Relationship)
}
