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
	"testing"

	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"

	"go.opentelemetry.io/collector/pdata/pcommon"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertHasEdge(t *testing.T, entities map[string]*ResourceEntity, clusterName, namespaceName, fromName, fromType, toName, toType string, relationship string) {
	entityID := toEntityId(clusterName, namespaceName, fromName, fromType)
	entity := entities[entityID]
	require.NotNil(t, entity, "Expected entity %s not found", entityID)
	edgeID := toEntityId(clusterName, namespaceName, toName, toType)
	edge := entity.Edges[edgeID]
	require.NotNil(t, edge, "Expected edge from %s to %s not found", entityID, edgeID)
	assert.Equal(t, relationship, edge.Relationship, "Incorrect relationship for edge from %s to %s", fromName, toName)
}

func TestKubernetesEntityRelationships(t *testing.T) {
	clusterName := "cluster-1"
	namespaceName := "default"

	attributes := pcommon.NewMap()

	// Resource attributes
	attributes.PutStr("k8s.cluster.name", clusterName)
	attributes.PutStr("k8s.namespace.name", namespaceName)

	attributes.PutStr(string(semconv.K8SClusterUIDKey), "us-east-1")
	attributes.PutStr(string(semconv.K8SNodeNameKey), "node-1")
	attributes.PutStr(string(semconv.K8SNodeUIDKey), "16")
	attributes.PutStr(string(semconv.K8SPodNameKey), "pod-1")
	attributes.PutStr(string(semconv.K8SPodUIDKey), "pod-uid-1")
	attributes.PutStr("k8s.pod.ip", "127.0.0.1")
	attributes.PutStr("k8s.pod.label.company-name", "cardinal")
	attributes.PutStr(string(semconv.ServiceNameKey), "service-1")
	attributes.PutStr(string(semconv.K8SDeploymentNameKey), "deployment-1")
	attributes.PutStr(string(semconv.K8SReplicaSetNameKey), "replicaset-1")

	ec := NewResourceEntityCache()
	ec.ProvisionResourceAttributes(attributes)
	entities := ec._allEntities()

	expectedEntities := map[string]string{
		clusterName:    "k8s.cluster",
		"node-1":       "k8s.node",
		namespaceName:  "k8s.namespace",
		"pod-1":        "k8s.pod",
		"service-1":    "service",
		"deployment-1": "k8s.deployment",
		"replicaset-1": "k8s.replicaset",
	}

	for name, entityType := range expectedEntities {
		id := toEntityId(clusterName, namespaceName, name, entityType)
		entity, exists := entities[id]
		require.True(t, exists, "Expected entity %s not found", name)
		assert.Equal(t, entityType, entity.Type, "Incorrect type for entity %s", name)
	}

	assert.Equal(t, "us-east-1", entities[toEntityId(clusterName, namespaceName, clusterName, "k8s.cluster")].Attributes[string(semconv.K8SClusterUIDKey)])
	assert.Equal(t, "16", entities[toEntityId(clusterName, namespaceName, "node-1", "k8s.node")].Attributes[string(semconv.K8SNodeUIDKey)])
	assert.Equal(t, "pod-uid-1", entities[toEntityId(clusterName, namespaceName, "pod-1", "k8s.pod")].Attributes[string(semconv.K8SPodUIDKey)])
	assert.Equal(t, "127.0.0.1", entities[toEntityId(clusterName, namespaceName, "pod-1", "k8s.pod")].Attributes["k8s.pod.ip"])
	assert.Equal(t, "cardinal", entities[toEntityId(clusterName, namespaceName, "pod-1", "k8s.pod")].Attributes["k8s.pod.label.company-name"])

	assertHasEdge(t, entities, clusterName, namespaceName, "cluster-1", "k8s.cluster", "node-1", "k8s.node", HasNode)
	assertHasEdge(t, entities, clusterName, namespaceName, "node-1", "k8s.node", "cluster-1", "k8s.cluster", BelongsToCluster)
	assertHasEdge(t, entities, clusterName, namespaceName, "default", "k8s.namespace", "service-1", "service", ContainsService)
	assertHasEdge(t, entities, clusterName, namespaceName, "service-1", "service", "default", "k8s.namespace", BelongsToNamespace)
	assertHasEdge(t, entities, clusterName, namespaceName, "service-1", "service", "deployment-1", "k8s.deployment", IsManagedByDeployment)
	assertHasEdge(t, entities, clusterName, namespaceName, "deployment-1", "k8s.deployment", "replicaset-1", "k8s.replicaset", ManagesReplicaset)
	assertHasEdge(t, entities, clusterName, namespaceName, "replicaset-1", "k8s.replicaset", "pod-1", "k8s.pod", ContainsPod)
}

func TestInterdependencyBetweenRelationshipMaps(t *testing.T) {
	clusterName := "cluster-1"
	namespaceName := "default"

	attributes := pcommon.NewMap()

	// Resource attributes
	attributes.PutStr("k8s.cluster.name", clusterName)
	attributes.PutStr("k8s.namespace.name", namespaceName)

	attributes.PutStr(string(semconv.K8SClusterNameKey), "cluster-1")
	attributes.PutStr(string(semconv.K8SNodeNameKey), "node-1")
	attributes.PutStr(string(semconv.OSNameKey), "linux")
	attributes.PutStr(string(semconv.K8SNamespaceNameKey), "default")
	attributes.PutStr(string(semconv.K8SPodNameKey), "pod-1")
	attributes.PutStr(string(semconv.ServiceNameKey), "service1")

	ec := NewResourceEntityCache()
	ec.ProvisionResourceAttributes(attributes)
	entities := ec._allEntities()

	expectedEntities := map[string]string{
		clusterName:   "k8s.cluster",
		"node-1":      "k8s.node",
		namespaceName: "k8s.namespace",
		"pod-1":       "k8s.pod",
		"linux":       "os",
		"service1":    "service",
	}

	for name, entityType := range expectedEntities {
		entityId := toEntityId(clusterName, namespaceName, name, entityType)
		entity, exists := entities[entityId]
		require.True(t, exists, "Expected entity %s not found", entityId)
		assert.Equal(t, entityType, entity.Type, "Incorrect type for entity %s", entityId)
	}

	assertHasEdge(t, entities, clusterName, namespaceName, "cluster-1", "k8s.cluster", "node-1", "k8s.node", HasNode)
	assertHasEdge(t, entities, clusterName, namespaceName, "node-1", "k8s.node", "cluster-1", "k8s.cluster", BelongsToCluster)
	assertHasEdge(t, entities, clusterName, namespaceName, "default", "k8s.namespace", "service1", "service", ContainsService)
	assertHasEdge(t, entities, clusterName, namespaceName, "service1", "service", "default", "k8s.namespace", BelongsToNamespace)
	assertHasEdge(t, entities, clusterName, namespaceName, "node-1", "k8s.node", "linux", "os", RunsOnOperatingSystem)
}

func TestContainerRelationships(t *testing.T) {
	clusterName := "cluster-1"
	namespaceName := "default"

	attributes := pcommon.NewMap()

	// Resource attributes
	attributes.PutStr("k8s.cluster.name", clusterName)
	attributes.PutStr("k8s.namespace.name", namespaceName)

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

	ec := NewResourceEntityCache()
	ec.ProvisionResourceAttributes(attributes)
	entities := ec._allEntities()

	expectedEntities := map[string]string{
		"my-container": "container",
		"java":         "process",
	}

	for name, entityType := range expectedEntities {
		entityID := toEntityId(clusterName, namespaceName, name, entityType)
		entity, exists := entities[entityID]
		require.True(t, exists, "Expected entity %s not found", entityID)
		assert.Equal(t, entityType, entity.Type, "Incorrect type for entity %s", entityID)
	}

	assert.Equal(t, "container-123", entities[toEntityId(clusterName, namespaceName, "my-container", "container")].Attributes[string(semconv.ContainerIDKey)])
	assert.Equal(t, "/bin/bash", entities[toEntityId(clusterName, namespaceName, "my-container", "container")].Attributes[string(semconv.ContainerCommandKey)])
	assert.Equal(t, "-c echo hello", entities[toEntityId(clusterName, namespaceName, "my-container", "container")].Attributes[string(semconv.ContainerCommandArgsKey)])
	assert.Equal(t, "docker", entities[toEntityId(clusterName, namespaceName, "my-container", "container")].Attributes[string(semconv.ContainerRuntimeKey)])
	assert.Equal(t, "/bin/bash -c echo hello", entities[toEntityId(clusterName, namespaceName, "my-container", "container")].Attributes[string(semconv.ContainerCommandLineKey)])
	assert.Equal(t, "team-a", entities[toEntityId(clusterName, namespaceName, "my-container", "container")].Attributes["container.label.owner"])

	assert.Equal(t, "java", entities[toEntityId(clusterName, namespaceName, "java", "process")].Attributes[string(semconv.ProcessExecutableNameKey)])
	assert.Equal(t, "/usr/bin/java", entities[toEntityId(clusterName, namespaceName, "java", "process")].Attributes[string(semconv.ProcessExecutablePathKey)])
	assert.Equal(t, "-jar myapp.jar", entities[toEntityId(clusterName, namespaceName, "java", "process")].Attributes[string(semconv.ProcessCommandArgsKey)])
	assert.Equal(t, "java -jar myapp.jar", entities[toEntityId(clusterName, namespaceName, "java", "process")].Attributes[string(semconv.ProcessCommandLineKey)])
	assert.Equal(t, "root", entities[toEntityId(clusterName, namespaceName, "java", "process")].Attributes[string(semconv.ProcessOwnerKey)])
	assert.Equal(t, "1700000000", entities[toEntityId(clusterName, namespaceName, "java", "process")].Attributes[string(semconv.ProcessCreationTimeKey)])
	assert.Equal(t, "voluntary", entities[toEntityId(clusterName, namespaceName, "java", "process")].Attributes[string(semconv.ProcessContextSwitchTypeKey)])
	assert.Equal(t, "1001", entities[toEntityId(clusterName, namespaceName, "java", "process")].Attributes[string(semconv.ProcessGroupLeaderPIDKey)])
	assert.Equal(t, "1000", entities[toEntityId(clusterName, namespaceName, "java", "process")].Attributes[string(semconv.ProcessParentPIDKey)])
	assert.Equal(t, "2000", entities[toEntityId(clusterName, namespaceName, "java", "process")].Attributes[string(semconv.ProcessPIDKey)])

}

func TestDBRelationships(t *testing.T) {
	clusterName := "cluster-1"
	namespaceName := "default"

	attributes := pcommon.NewMap()
	attributes.PutStr(string(semconv.ServiceNameKey), "service-1")

	ec := NewResourceEntityCache()
	globalEntityMap := ec.ProvisionResourceAttributes(attributes)

	recordAttributes := pcommon.NewMap()
	recordAttributes.PutStr(string(semconv.DBSystemKey), "mysql")
	recordAttributes.PutStr(string(semconv.DBCollectionNameKey), "glacier.tbl_17665234232")

	ec.ProvisionRecordAttributes(globalEntityMap, recordAttributes)

	dbCollectionEntityId := toEntityId(clusterName, namespaceName, "glacier.tbl_", "database.collection")

	entities := ec._allEntities()
	_, exists := entities[dbCollectionEntityId]
	require.True(t, exists, "Expected entity %s not found", dbCollectionEntityId)
	assert.Equal(t, UsesDatabaseCollection, entities[toEntityId(clusterName, namespaceName, "service-1", "service")].Edges[toEntityId(clusterName, namespaceName, "glacier.tbl_", "database.collection")].Relationship)
}

func TestMessagingConsumesFromRelationship(t *testing.T) {
	clusterName := "cluster-1"
	namespaceName := "default"

	attributes := pcommon.NewMap()
	attributes.PutStr(string(semconv.ServiceNameKey), "service-1")

	ec := NewResourceEntityCache()
	globalEntityMap := ec.ProvisionResourceAttributes(attributes)

	recordAttributes := pcommon.NewMap()
	recordAttributes.PutStr(string(semconv.MessagingSystemKey), "kafka")
	recordAttributes.PutStr(string(semconv.MessagingDestinationNameKey), "topic-1")
	recordAttributes.PutStr(string(semconv.MessagingOperationNameKey), "process")
	recordAttributes.PutStr(string(semconv.MessagingConsumerGroupNameKey), "consumer-group-1")

	ec.ProvisionRecordAttributes(globalEntityMap, recordAttributes)

	entities := ec._allEntities()

	messagingDestinationEntityId := toEntityId(clusterName, namespaceName, "topic-1", MessagingDestination)
	messagingConsumerGroupEntityId := toEntityId(clusterName, namespaceName, "consumer-group-1", MessagingConsumerGroup)

	_, messagingDestinationExists := entities[messagingDestinationEntityId]
	require.True(t, messagingDestinationExists, "Expected entity %s not found", messagingDestinationEntityId)

	_, messagingConsumerGroupEntityExists := entities[messagingConsumerGroupEntityId]
	require.True(t, messagingConsumerGroupEntityExists, "Expected entity %s not found", messagingConsumerGroupEntityId)

	// assert edge between service and messaging destination, and assert the edge relationship
	assertHasEdge(t, entities, clusterName, namespaceName, "service-1", "service", "topic-1", MessagingDestination, ConsumesFrom)
	assertHasEdge(t, entities, clusterName, namespaceName, "service-1", "service", "consumer-group-1", MessagingConsumerGroup, ConsumesFrom)
}

func TestMessagingProducesToRelationship(t *testing.T) {
	clusterName := "cluster-1"
	namespaceName := "default"

	ec := NewResourceEntityCache()
	attributes := pcommon.NewMap()
	attributes.PutStr(string(semconv.ServiceNameKey), "service-1")
	globalEntityMap := ec.ProvisionResourceAttributes(attributes)

	recordAttributes := pcommon.NewMap()
	recordAttributes.PutStr(string(semconv.MessagingSystemKey), "kafka")
	recordAttributes.PutStr(string(semconv.MessagingDestinationNameKey), "topic-1")
	recordAttributes.PutStr(string(semconv.MessagingOperationTypeKey), "publish")

	ec.ProvisionRecordAttributes(globalEntityMap, recordAttributes)
	entities := ec._allEntities()
	messagingDestinationEntityId := toEntityId(clusterName, namespaceName, "topic-1", MessagingDestination)

	_, messagingDestinationExists := entities[messagingDestinationEntityId]
	require.True(t, messagingDestinationExists, "Expected entity %s not found", messagingDestinationEntityId)

	// assert edge between service and messaging destination, and assert the edge relationship
	assertHasEdge(t, entities, clusterName, namespaceName, "service-1", "service", "topic-1", MessagingDestination, ProducesTo)
}

func TestIfWeSkipBotTraffic(t *testing.T) {
	clusterName := "cluster-1"
	namespaceName := "default"

	ec := NewResourceEntityCache()
	attributes := pcommon.NewMap()
	attributes.PutStr(string(semconv.ServiceNameKey), "service-1")
	globalEntityMap := ec.ProvisionResourceAttributes(attributes)

	recordAttributes := pcommon.NewMap()
	recordAttributes.PutStr(string(semconv.URLTemplateKey), "/robots.txt")
	recordAttributes.PutInt(string(semconv.HTTPResponseStatusCodeKey), 404)

	ec.ProvisionRecordAttributes(globalEntityMap, recordAttributes)
	entities := ec._allEntities()

	// assert that we skip the entity creation for the record
	assert.Equal(t, 1, len(entities))

	// assert that there is no entity of Endpoint type
	_, endpointExists := entities[toEntityId(clusterName, namespaceName, "/robots.txt", Endpoint)]
	assert.False(t, endpointExists, "Unwanted entity %s found", toEntityId(clusterName, namespaceName, "/robots.txt", Endpoint))

	// Now let's test if the status code was 200, we would indeed create the entity.
	recordAttributes.PutInt(string(semconv.HTTPResponseStatusCodeKey), 200)
	ec.ProvisionRecordAttributes(globalEntityMap, recordAttributes)
	entities = ec._allEntities()
	_, endpointExists = entities[toEntityId(clusterName, namespaceName, "/robots.txt", Endpoint)]
	assert.True(t, endpointExists, "Expected entity %s not found", toEntityId(clusterName, namespaceName, "/robots.txt", Endpoint))
}

func TestCloudRelationships(t *testing.T) {
	clusterName := "cluster-1"
	namespaceName := "default"

	attributes := pcommon.NewMap()
	attributes.PutStr(string(semconv.CloudProviderKey), "aws")
	attributes.PutStr(string(semconv.CloudAccountIDKey), "123456789012")
	attributes.PutStr(string(semconv.CloudRegionKey), "us-west-1")
	attributes.PutStr(string(semconv.CloudAvailabilityZoneKey), "us-west-1a")

	ec := NewResourceEntityCache()
	ec.ProvisionResourceAttributes(attributes)
	entities := ec._allEntities()

	expectedEntities := map[string]string{
		"aws":          "cloud.provider",
		"123456789012": "cloud.account",
		"us-west-1":    "cloud.region",
		"us-west-1a":   "cloud.availability_zone",
	}

	// Check that expected entities exist
	for name, entityType := range expectedEntities {
		entityId := toEntityId(clusterName, namespaceName, name, entityType)
		entity, exists := entities[entityId]
		require.True(t, exists, "Expected entity %s not found", entityId)
		assert.Equal(t, entityType, entity.Type, "Incorrect type for entity %s", entityId)
	}

	// Validate relationships
	assertHasEdge(t, entities, clusterName, namespaceName, "aws", "cloud.provider", "123456789012", "cloud.account", ManagesAccount)
	assertHasEdge(t, entities, clusterName, namespaceName, "aws", "cloud.provider", "us-west-1", "cloud.region", ContainsRegion)
	assertHasEdge(t, entities, clusterName, namespaceName, "us-west-1", "cloud.region", "us-west-1a", "cloud.availability_zone", ContainsAvailabilityZone)
	assertHasEdge(t, entities, clusterName, namespaceName, "123456789012", "cloud.account", "aws", "cloud.provider", BelongsToProvider)
	assertHasEdge(t, entities, clusterName, namespaceName, "123456789012", "cloud.account", "us-west-1", "cloud.region", HasResourcesInRegion)
	assertHasEdge(t, entities, clusterName, namespaceName, "us-west-1", "cloud.region", "aws", "cloud.provider", BelongsToProvider)
	assertHasEdge(t, entities, clusterName, namespaceName, "us-west-1", "cloud.region", "us-west-1a", "cloud.availability_zone", ContainsAvailabilityZone)
	assertHasEdge(t, entities, clusterName, namespaceName, "us-west-1", "cloud.region", "123456789012", "cloud.account", BelongsToAccount)
	assertHasEdge(t, entities, clusterName, namespaceName, "us-west-1a", "cloud.availability_zone", "us-west-1", "cloud.region", BelongsToRegion)
}
