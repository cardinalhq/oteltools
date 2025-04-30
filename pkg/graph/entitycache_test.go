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
	"maps"
	"testing"

	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"

	"go.opentelemetry.io/collector/pdata/pcommon"

	"github.com/cardinalhq/oteltools/pkg/graph/graphpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertEntityExists(t *testing.T, entities map[string]*ResourceEntity, name, entityType string, otherAttributes ...string) *ResourceEntity {
	// convert otherAttributes to key/value map and then put to entityId.IdAttributes
	if len(otherAttributes)%2 != 0 {
		panic("otherAttributes must be key-value pairs")
	}
	attributes := make(map[string]string)
	for i := 0; i < len(otherAttributes); i += 2 {
		attributes[otherAttributes[i]] = otherAttributes[i+1]
	}
	entityId := ToEntityId(name, entityType, attributes)
	entity, exists := entities[entityId.Hash]
	require.True(t, exists, "Expected entity %s not found", entityId.Hash)
	return entity
}

func assertEdgeExists(t *testing.T, entity *ResourceEntity, target *EntityId, relationship string) {
	key := entity.EntityId.Hash + target.Hash
	actualRelationship, exists := entity.Edges[key]
	require.True(t, exists, "Expected edge %s not found", key)
	assert.Equal(t, relationship, actualRelationship.Relationship)
}

func assertEdgeFromExists(t *testing.T, entity *ResourceEntity, source *EntityId, relationship string) {
	key := source.Hash + entity.EntityId.Hash
	actualRelationship, exists := entity.Edges[key]
	require.True(t, exists, "Expected edge %s not found", key)
	assert.Equal(t, relationship, actualRelationship.Relationship)
}

func listify(m map[string]string) []string {
	var res []string
	for k, v := range m {
		res = append(res, k, v)
	}
	return res
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

	entities := ec._allEntities()

	clusterEntity := assertEntityExists(t, entities, "cluster-1", KubernetesCluster)
	assert.Equal(t, "us-east-1", clusterEntity.Attributes[string(semconv.K8SClusterUIDKey)])

	nodeEntity := assertEntityExists(t, entities, "node-1", Node, string(semconv.K8SClusterNameKey), "cluster-1")
	assert.Equal(t, "16", nodeEntity.Attributes[string(semconv.K8SNodeUIDKey)])

	namespaceEntity := assertEntityExists(t, entities, "default", KubernetesNamespace, string(semconv.K8SClusterNameKey), "cluster-1", string(semconv.K8SNamespaceNameKey), "default")

	podEntity := assertEntityExists(t, entities, "pod-1", KubernetesPod, string(semconv.K8SClusterNameKey), "cluster-1", string(semconv.K8SNamespaceNameKey), "default", string(semconv.K8SClusterNameKey), "cluster-1")
	assert.Equal(t, "pod-uid-1", podEntity.Attributes[string(semconv.K8SPodUIDKey)])
	assert.Equal(t, "127.0.0.1", podEntity.Attributes["k8s.pod.ip"])
	assert.Equal(t, "cardinal", podEntity.Attributes["k8s.pod.label.company-name"])

	serviceEntity := assertEntityExists(t, entities, "service-1", Service, string(semconv.K8SClusterNameKey), "cluster-1", string(semconv.K8SNamespaceNameKey), "default")
	deploymentEntity := assertEntityExists(t, entities, "deployment-1", KubernetesDeployment, string(semconv.K8SClusterNameKey), "cluster-1", string(semconv.K8SNamespaceNameKey), "default")
	replicaSetEntity := assertEntityExists(t, entities, "replicaset-1", KubernetesReplicaSet, string(semconv.K8SClusterNameKey), "cluster-1", string(semconv.K8SNamespaceNameKey), "default")

	assertEdgeExists(t, clusterEntity, nodeEntity.EntityId, HasNode)
	assertEdgeExists(t, nodeEntity, clusterEntity.EntityId, BelongsToCluster)
	assertEdgeExists(t, deploymentEntity, replicaSetEntity.EntityId, ManagesReplicaset)
	assertEdgeExists(t, replicaSetEntity, podEntity.EntityId, ContainsPod)
	assertEdgeExists(t, serviceEntity, namespaceEntity.EntityId, BelongsToNamespace)
	assertEdgeExists(t, serviceEntity, deploymentEntity.EntityId, IsManagedByDeployment)
	assertEdgeExists(t, namespaceEntity, serviceEntity.EntityId, ContainsService)
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

	entities := ec._allEntities()
	nodeEntity := assertEntityExists(t, entities, "node-1", Node, string(semconv.K8SClusterNameKey), "cluster-1")
	operatingSystemEntity := assertEntityExists(t, entities, "linux", OperatingSystem)
	assertEdgeExists(t, nodeEntity, operatingSystemEntity.EntityId, RunsOnOperatingSystem)
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

	entities := ec._allEntities()

	// Assert container entity
	container := assertEntityExists(t, entities, "my-container", Container)
	assert.Equal(t, "container-123", container.Attributes[string(semconv.ContainerIDKey)])
	assert.Equal(t, "/bin/bash", container.Attributes[string(semconv.ContainerCommandKey)])
	assert.Equal(t, "-c echo hello", container.Attributes[string(semconv.ContainerCommandArgsKey)])
	assert.Equal(t, "docker", container.Attributes[string(semconv.ContainerRuntimeKey)])
	assert.Equal(t, "/bin/bash -c echo hello", container.Attributes[string(semconv.ContainerCommandLineKey)])
	assert.Equal(t, "team-a", container.Attributes["container.label.owner"])

	// Assert process entity
	process := assertEntityExists(t, entities, "java", Process)
	assert.Equal(t, "java", process.Attributes[string(semconv.ProcessExecutableNameKey)])
	assert.Equal(t, "/usr/bin/java", process.Attributes[string(semconv.ProcessExecutablePathKey)])
	assert.Equal(t, "-jar myapp.jar", process.Attributes[string(semconv.ProcessCommandArgsKey)])
	assert.Equal(t, "java -jar myapp.jar", process.Attributes[string(semconv.ProcessCommandLineKey)])
	assert.Equal(t, "root", process.Attributes[string(semconv.ProcessOwnerKey)])
	assert.Equal(t, "1700000000", process.Attributes[string(semconv.ProcessCreationTimeKey)])
	assert.Equal(t, "voluntary", process.Attributes[string(semconv.ProcessContextSwitchTypeKey)])
	assert.Equal(t, "1001", process.Attributes[string(semconv.ProcessGroupLeaderPIDKey)])
	assert.Equal(t, "1000", process.Attributes[string(semconv.ProcessParentPIDKey)])
	assert.Equal(t, "2000", process.Attributes[string(semconv.ProcessPIDKey)])

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

	entities := ec._allEntities()
	dbEntity := assertEntityExists(t, entities, "glacier.tbl_", DatabaseCollection)
	serviceEntity := assertEntityExists(t, entities, "service-1", Service)
	assertEdgeExists(t, serviceEntity, dbEntity.EntityId, UsesDatabaseCollection)
}

func TestExtractEntityCallback(t *testing.T) {
	ec := NewResourceEntityCache()
	attributes := pcommon.NewMap()
	attributes.PutStr(string(semconv.ServiceNameKey), "service-1")
	globalEntityMap := ec.ProvisionResourceAttributes(attributes)

	recordAttributes := pcommon.NewMap()
	recordAttributes.PutStr(string(semconv.ServerAddressKey), "test-us-east-2-global-0.cryyasaiu1g7.us-east-2.rds.amazonaws.com")
	ec.ProvisionRecordAttributes(globalEntityMap, recordAttributes)

	entities := ec._allEntities()
	dbEntity := assertEntityExists(t, entities, "test-us-east-2-global", Service)
	serviceEntity := assertEntityExists(t, entities, "service-1", Service)
	assertEdgeExists(t, serviceEntity, dbEntity.EntityId, UsesDataBaseHostedOn)
}

func TestEndpointRelationships(t *testing.T) {
	ec := NewResourceEntityCache()
	attributes := pcommon.NewMap()
	attributes.PutStr(string(semconv.ServiceNameKey), "service-1")
	globalEntityMap := ec.ProvisionResourceAttributes(attributes)

	recordAttributes := pcommon.NewMap()
	recordAttributes.PutStr(string(semconv.URLTemplateKey), "/api/v1/hello")
	recordAttributes.PutStr(SpanKindString, "Server")
	recordAttributes.PutInt(string(semconv.HTTPResponseStatusCodeKey), 200)
	ec.ProvisionRecordAttributes(globalEntityMap, recordAttributes)

	entities := ec._allEntities()
	endpointEntity := assertEntityExists(t, entities, "/api/v1/hello", Endpoint)
	serviceEntity := assertEntityExists(t, entities, "service-1", Service)
	assertEdgeExists(t, serviceEntity, endpointEntity.EntityId, ServesEndpoint)
}

func TestMessagingConsumesFromRelationship(t *testing.T) {
	ec := NewResourceEntityCache()

	// Define the service entity
	attributes := pcommon.NewMap()
	attributes.PutStr(string(semconv.ServiceNameKey), "service-1")
	globalEntityMap := ec.ProvisionResourceAttributes(attributes)

	// Define the messaging record attributes
	recordAttributes := pcommon.NewMap()
	recordAttributes.PutStr(string(semconv.MessagingSystemKey), "kafka")
	recordAttributes.PutStr(string(semconv.MessagingDestinationNameKey), "topic-1")
	recordAttributes.PutStr(string(semconv.MessagingOperationNameKey), "process")
	recordAttributes.PutStr(string(semconv.MessagingConsumerGroupNameKey), "consumer-group-1")

	ec.ProvisionRecordAttributes(globalEntityMap, recordAttributes)

	// Fetch all entities
	entities := ec._allEntities()

	// Assert all expected entities exist
	service := assertEntityExists(t, entities, "service-1", Service)
	messagingDestination := assertEntityExists(t, entities, "topic-1", MessagingDestination)
	messagingConsumerGroup := assertEntityExists(t, entities, "consumer-group-1", MessagingConsumerGroup)

	// Assert edges from service to messaging entities
	assertEdgeExists(t, service, messagingDestination.EntityId, ConsumesFrom)
	assertEdgeExists(t, service, messagingConsumerGroup.EntityId, ConsumesFrom)
}

func TestMessagingProducesToRelationship(t *testing.T) {
	ec := NewResourceEntityCache()

	// Define the service entity
	attributes := pcommon.NewMap()
	attributes.PutStr(string(semconv.ServiceNameKey), "service-1")
	globalEntityMap := ec.ProvisionResourceAttributes(attributes)

	// Define the messaging record attributes (for a producer)
	recordAttributes := pcommon.NewMap()
	recordAttributes.PutStr(string(semconv.MessagingSystemKey), "kafka")
	recordAttributes.PutStr(string(semconv.MessagingDestinationNameKey), "topic-1")
	recordAttributes.PutStr(string(semconv.MessagingOperationTypeKey), "publish")

	ec.ProvisionRecordAttributes(globalEntityMap, recordAttributes)

	// Fetch all entities
	entities := ec._allEntities()

	// Assert service and messaging destination entities exist
	service := assertEntityExists(t, entities, "service-1", Service)
	messagingDestination := assertEntityExists(t, entities, "topic-1", MessagingDestination)

	// Assert that the service ProducesTo the messaging destination
	assertEdgeExists(t, service, messagingDestination.EntityId, ProducesTo)
}

func TestIfWeSkipBotTraffic(t *testing.T) {
	ec := NewResourceEntityCache()

	// Step 1: Set up the service entity
	attributes := pcommon.NewMap()
	attributes.PutStr(string(semconv.ServiceNameKey), "service-1")
	globalEntityMap := ec.ProvisionResourceAttributes(attributes)

	// Step 2: Simulate a request to /robots.txt with a 404 (should be skipped)
	recordAttributes := pcommon.NewMap()
	recordAttributes.PutStr(string(semconv.URLTemplateKey), "/robots.txt")
	recordAttributes.PutInt(string(semconv.HTTPResponseStatusCodeKey), 404)

	ec.ProvisionRecordAttributes(globalEntityMap, recordAttributes)

	entities := ec._allEntities()

	assert.Equal(t, 1, len(entities), "Expected only the service entity to be created")

	endpointId := &EntityId{
		Name: "robots.txt",
		Type: Endpoint,
	}
	_, endpointExists := entities[endpointId.Hash]
	assert.False(t, endpointExists, "Did not expect entity %s to exist", endpointId)

	recordAttributes.PutInt(string(semconv.HTTPResponseStatusCodeKey), 200)
	ec.ProvisionRecordAttributes(globalEntityMap, recordAttributes)
	entities = ec._allEntities()

	endpoint := assertEntityExists(t, entities, "/robots.txt", Endpoint)
	assert.NotNil(t, endpoint)
}

func TestCloudRelationships(t *testing.T) {
	ec := NewResourceEntityCache()

	// Step 1: Set up cloud-related resource attributes
	attributes := pcommon.NewMap()
	attributes.PutStr(string(semconv.CloudProviderKey), "aws")
	attributes.PutStr(string(semconv.CloudAccountIDKey), "123456789012")
	attributes.PutStr(string(semconv.CloudRegionKey), "us-west-1")
	attributes.PutStr(string(semconv.CloudAvailabilityZoneKey), "us-west-1a")

	ec.ProvisionResourceAttributes(attributes)
	entities := ec._allEntities()

	// Step 2: Assert all cloud entities exist
	provider := assertEntityExists(t, entities, "aws", CloudProvider)
	account := assertEntityExists(t, entities, "123456789012", CloudAccount)
	region := assertEntityExists(t, entities, "us-west-1", CloudRegion)
	zone := assertEntityExists(t, entities, "us-west-1a", CloudAvailabilityZone)

	// Step 3: Assert expected edges between cloud entities
	assertEdgeExists(t, provider, account.EntityId, ManagesAccount)
	assertEdgeExists(t, provider, region.EntityId, ContainsRegion)
	assertEdgeExists(t, provider, zone.EntityId, ContainsAvailabilityZone)

	assertEdgeExists(t, account, provider.EntityId, BelongsToProvider)
	assertEdgeExists(t, account, region.EntityId, HasResourcesInRegion)

	assertEdgeExists(t, region, provider.EntityId, BelongsToProvider)
	assertEdgeExists(t, region, zone.EntityId, ContainsAvailabilityZone)
	assertEdgeExists(t, region, account.EntityId, BelongsToAccount)

	assertEdgeExists(t, zone, region.EntityId, BelongsToRegion)
}

var (
	defaultIdentityAttributes = map[string]string{
		string(semconv.K8SClusterNameKey):   "cluster-1",
		string(semconv.K8SNamespaceNameKey): "default",
	}
)

func TestBuildEntities_HandlesSecretSummary(t *testing.T) {
	attrs := map[string]string{
		"key": "value",
	}
	maps.Copy(attrs, defaultIdentityAttributes)

	item := &graphpb.PackagedObject{
		ResourceAttributes: attrs,
		Object: &graphpb.PackagedObject_SecretSummary{
			SecretSummary: &graphpb.SecretSummary{
				BaseObject: &graphpb.BaseObject{
					Namespace:  defaultIdentityAttributes[string(semconv.K8SNamespaceNameKey)],
					Kind:       "Secret",
					ApiVersion: "v1",
					Name:       "test-secret",
				},
			},
		},
	}

	ec := NewResourceEntityCache()
	ec.ProvisionPackagedObject(item)
	entities := ec._allEntities()

	secret := assertEntityExists(t, entities, "test-secret", KubernetesSecret, listify(defaultIdentityAttributes)...)
	require.NotNil(t, secret)
	assert.Len(t, entities, 1)
	assert.Equal(t, attrs, secret.Attributes)
	assert.Len(t, secret.Edges, 0)
}

func TestBuildEntities_HandlesConfigMapSummary(t *testing.T) {
	attrs := map[string]string{
		"key": "value",
	}
	maps.Copy(attrs, defaultIdentityAttributes)

	item := &graphpb.PackagedObject{
		ResourceAttributes: attrs,
		Object: &graphpb.PackagedObject_ConfigMapSummary{
			ConfigMapSummary: &graphpb.ConfigMapSummary{
				BaseObject: &graphpb.BaseObject{
					Namespace:  attrs[string(semconv.K8SNamespaceNameKey)],
					Kind:       "ConfigMap",
					ApiVersion: "v1",
					Name:       "test-configmap",
				},
			},
		},
	}

	ec := NewResourceEntityCache()
	ec.ProvisionPackagedObject(item)
	entities := ec._allEntities()

	configmap := assertEntityExists(t, entities, "test-configmap", KubernetesConfigMap, listify(defaultIdentityAttributes)...)
	require.NotNil(t, configmap)
	assert.Len(t, entities, 1)
	assert.Equal(t, attrs, configmap.Attributes)
	assert.Len(t, configmap.Edges, 0)
}

func TestBuildEntities_HandlesPodSummary(t *testing.T) {
	attrs := map[string]string{
		"key": "value",
	}
	maps.Copy(attrs, defaultIdentityAttributes)

	item := &graphpb.PackagedObject{
		ResourceAttributes: attrs,
		Object: &graphpb.PackagedObject_PodSummary{
			PodSummary: &graphpb.PodSummary{
				BaseObject: &graphpb.BaseObject{
					Namespace:  defaultIdentityAttributes[string(semconv.K8SNamespaceNameKey)],
					Kind:       "Pod",
					ApiVersion: "v1",
					Name:       "test-pod",
				},
				Spec: &graphpb.PodSpec{
					Containers: []*graphpb.PodContainerSpec{},
				},
			},
		},
	}

	ec := NewResourceEntityCache()
	ec.ProvisionPackagedObject(item)
	entities := ec._allEntities()

	pod := assertEntityExists(t, entities, "test-pod", KubernetesPod, listify(defaultIdentityAttributes)...)
	require.NotNil(t, pod)
	assert.Len(t, entities, 1)
	assert.Equal(t, attrs, pod.Attributes)
	// With no containers, expect no edges.
	assert.Len(t, pod.Edges, 0)
}

func TestBuildEntities_HandlesPodSummaryWithContainers(t *testing.T) {
	attrs := map[string]string{
		"key": "value",
	}
	maps.Copy(attrs, defaultIdentityAttributes)

	item := &graphpb.PackagedObject{
		ResourceAttributes: attrs,
		Object: &graphpb.PackagedObject_PodSummary{
			PodSummary: &graphpb.PodSummary{
				BaseObject: &graphpb.BaseObject{
					Namespace:  "default",
					Kind:       "Pod",
					ApiVersion: "v1",
					Name:       "test-pod",
				},
				Spec: &graphpb.PodSpec{
					Containers: []*graphpb.PodContainerSpec{
						{
							Name:           "test-container",
							ConfigMapNames: []string{"test-configmap"},
							SecretNames:    []string{"test-secret"},
							Image:          "test-image",
						},
					},
				},
			},
		},
	}

	ec := NewResourceEntityCache()
	ec.ProvisionPackagedObject(item)
	entities := ec._allEntities()

	pod := assertEntityExists(t, entities, "test-pod", KubernetesPod, listify(defaultIdentityAttributes)...)
	require.NotNil(t, pod)
	assert.Len(t, pod.Attributes, 8)
	assert.Equal(t, "value", pod.Attributes["key"])
	assert.Equal(t, "test-image", pod.Attributes["container.image.name.test-container"])
	assert.Equal(t, "false", pod.Attributes[CrashLoopBackOff])
	assert.Equal(t, "false", pod.Attributes[OOMKilled])
	assert.Equal(t, "false", pod.Attributes[ImagePullBackOff])
	assert.Equal(t, "0", pod.Attributes["restart_count"])

	// Expect edges linking to the config map and secret.
	require.Len(t, pod.Edges, 4)
	assertEdgeExists(t, pod, ToEntityId("test-configmap", KubernetesConfigMap, defaultIdentityAttributes), UsesConfigMap)
	assertEdgeExists(t, pod, ToEntityId("test-secret", KubernetesSecret, defaultIdentityAttributes), UsesSecret)
	assertEdgeFromExists(t, pod, ToEntityId("test-configmap", KubernetesConfigMap, defaultIdentityAttributes), IsUsedByPod)
	assertEdgeFromExists(t, pod, ToEntityId("test-secret", KubernetesSecret, defaultIdentityAttributes), IsUsedByPod)
}

func TestBuildEntities_HandlesPodSummary_WithReplicaSetOwner(t *testing.T) {
	attrs := map[string]string{
		"key": "value",
	}
	maps.Copy(attrs, defaultIdentityAttributes)

	item := &graphpb.PackagedObject{
		ResourceAttributes: attrs,
		Object: &graphpb.PackagedObject_PodSummary{
			PodSummary: &graphpb.PodSummary{
				BaseObject: &graphpb.BaseObject{
					Namespace:  "default",
					Kind:       "Pod",
					ApiVersion: "v1",
					Name:       "test-pod",
					OwnerRef: []*graphpb.OwnerRef{
						{
							Kind: "ReplicaSet",
							Name: "test-replicaset",
						},
					},
				},
			},
		},
	}

	ec := NewResourceEntityCache()
	ec.ProvisionPackagedObject(item)
	entities := ec._allEntities()

	pod := assertEntityExists(t, entities, "test-pod", KubernetesPod, listify(defaultIdentityAttributes)...)
	require.NotNil(t, pod)
	assert.Equal(t, "value", pod.Attributes["key"])

	require.Len(t, pod.Edges, 2)
	assertEdgeExists(t, pod, ToEntityId("test-replicaset", KubernetesReplicaSet, defaultIdentityAttributes), IsManagedByReplicaSet)
	assertEdgeFromExists(t, pod, ToEntityId("test-replicaset", KubernetesReplicaSet, defaultIdentityAttributes), ManagesPod)
}

func TestBuildEntities_FormsEdges(t *testing.T) {
	attrs := map[string]string{"key": "value"}
	maps.Copy(attrs, defaultIdentityAttributes)

	secretItem := &graphpb.PackagedObject{
		ResourceAttributes: attrs,
		Object: &graphpb.PackagedObject_SecretSummary{
			SecretSummary: &graphpb.SecretSummary{
				BaseObject: &graphpb.BaseObject{
					Namespace:  "default",
					Kind:       "Secret",
					ApiVersion: "v1",
					Name:       "test-secret",
				},
			},
		},
	}

	configMapItem := &graphpb.PackagedObject{
		ResourceAttributes: attrs,
		Object: &graphpb.PackagedObject_ConfigMapSummary{
			ConfigMapSummary: &graphpb.ConfigMapSummary{
				BaseObject: &graphpb.BaseObject{
					Namespace:  "default",
					Kind:       "ConfigMap",
					ApiVersion: "v1",
					Name:       "test-configmap",
				},
			},
		},
	}

	podItem := &graphpb.PackagedObject{
		ResourceAttributes: attrs,
		Object: &graphpb.PackagedObject_PodSummary{
			PodSummary: &graphpb.PodSummary{
				BaseObject: &graphpb.BaseObject{
					Namespace:  "default",
					Kind:       "Pod",
					ApiVersion: "v1",
					Name:       "test-pod",
				},
				Spec: &graphpb.PodSpec{
					Containers: []*graphpb.PodContainerSpec{
						{
							Image:          "test-image",
							ConfigMapNames: []string{"test-configmap"},
							SecretNames:    []string{"test-secret"},
						},
					},
				},
			},
		},
	}

	ec := NewResourceEntityCache()
	ec.ProvisionPackagedObject(secretItem)
	ec.ProvisionPackagedObject(configMapItem)
	ec.ProvisionPackagedObject(podItem)
	entities := ec._allEntities()

	assert.Len(t, entities, 3)

	for _, entity := range entities {
		switch entity.EntityId.Name {
		case "test-secret":
			require.Len(t, entity.Edges, 0)
		case "test-configmap":
			require.Len(t, entity.Edges, 0)
		case "test-pod":
			require.Len(t, entity.Edges, 4)
			assertEdgeExists(t, entity, ToEntityId("test-configmap", KubernetesConfigMap, defaultIdentityAttributes), UsesConfigMap)
			assertEdgeExists(t, entity, ToEntityId("test-secret", KubernetesSecret, defaultIdentityAttributes), UsesSecret)
			assertEdgeFromExists(t, entity, ToEntityId("test-configmap", KubernetesConfigMap, defaultIdentityAttributes), IsUsedByPod)
			assertEdgeFromExists(t, entity, ToEntityId("test-secret", KubernetesSecret, defaultIdentityAttributes), IsUsedByPod)
		default:
			t.Errorf("Unexpected entity: %#v", entity.EntityId)
		}
	}
}

func TestBuildEntities_HandlesDaemonSetSummary(t *testing.T) {
	attrs := map[string]string{
		"key": "value",
	}
	maps.Copy(attrs, defaultIdentityAttributes)

	item := &graphpb.PackagedObject{
		ResourceAttributes: attrs,
		Object: &graphpb.PackagedObject_AppsDaemonSetSummary{
			AppsDaemonSetSummary: &graphpb.AppsDaemonSetSummary{
				BaseObject: &graphpb.BaseObject{
					Namespace:  defaultIdentityAttributes[string(semconv.K8SNamespaceNameKey)],
					Kind:       "DaemonSet",
					ApiVersion: "apps/v1",
					Name:       "test-daemonset",
				},
			},
		},
	}

	ec := NewResourceEntityCache()
	ec.ProvisionPackagedObject(item)
	entities := ec._allEntities()

	assert.Len(t, entities, 1)

	daemonSet := assertEntityExists(t, entities, "test-daemonset", KubernetesDaemonSet, listify(defaultIdentityAttributes)...)
	require.NotNil(t, daemonSet)
	assert.Equal(t, attrs, daemonSet.Attributes)
	assert.Len(t, daemonSet.Edges, 0)
}

func TestBuildEntities_HandlesDeploymentSummary(t *testing.T) {
	attrs := map[string]string{
		"key": "value",
	}
	maps.Copy(attrs, defaultIdentityAttributes)

	item := &graphpb.PackagedObject{
		ResourceAttributes: attrs,
		Object: &graphpb.PackagedObject_AppsDeploymentSummary{
			AppsDeploymentSummary: &graphpb.AppsDeploymentSummary{
				BaseObject: &graphpb.BaseObject{
					Namespace:  defaultIdentityAttributes[string(semconv.K8SNamespaceNameKey)],
					Kind:       "Deployment",
					ApiVersion: "apps/v1",
					Name:       "test-deployment",
				},
			},
		},
	}

	ec := NewResourceEntityCache()
	ec.ProvisionPackagedObject(item)
	entities := ec._allEntities()

	assert.Len(t, entities, 1)

	deployment := assertEntityExists(t, entities, "test-deployment", KubernetesDeployment, listify(defaultIdentityAttributes)...)
	require.NotNil(t, deployment)
	assert.Equal(t, attrs, deployment.Attributes)
	assert.Len(t, deployment.Edges, 0)
}

func TestBuildEntities_HandlesReplicaSetSummary(t *testing.T) {
	attrs := map[string]string{
		"key": "value",
	}
	maps.Copy(attrs, defaultIdentityAttributes)

	item := &graphpb.PackagedObject{
		ResourceAttributes: attrs,
		Object: &graphpb.PackagedObject_AppsReplicaSetSummary{
			AppsReplicaSetSummary: &graphpb.AppsReplicaSetSummary{
				BaseObject: &graphpb.BaseObject{
					Namespace:  defaultIdentityAttributes[string(semconv.K8SNamespaceNameKey)],
					Kind:       "ReplicaSet",
					ApiVersion: "apps/v1",
					Name:       "test-replicaset",
					OwnerRef: []*graphpb.OwnerRef{
						{
							Kind: "Deployment",
							Name: "test-deployment",
						},
					},
				},
			},
		},
	}

	ec := NewResourceEntityCache()
	ec.ProvisionPackagedObject(item)
	entities := ec._allEntities()

	assert.Len(t, entities, 1)

	replicaSet := assertEntityExists(t, entities, "test-replicaset", KubernetesReplicaSet, listify(defaultIdentityAttributes)...)
	require.NotNil(t, replicaSet)
	assert.Equal(t, attrs, replicaSet.Attributes)
	assert.Len(t, replicaSet.Edges, 2)
	assertEdgeExists(t, replicaSet, ToEntityId("test-deployment", KubernetesDeployment, defaultIdentityAttributes), IsManagedByDeployment)
	assertEdgeFromExists(t, replicaSet, ToEntityId("test-deployment", KubernetesDeployment, defaultIdentityAttributes), ManagesReplicaset)
}

func TestBuildEntities_HandlesStatefulSetSummary(t *testing.T) {
	attrs := map[string]string{
		"key": "value",
	}
	maps.Copy(attrs, defaultIdentityAttributes)

	item := &graphpb.PackagedObject{
		ResourceAttributes: attrs,
		Object: &graphpb.PackagedObject_AppsStatefulSetSummary{
			AppsStatefulSetSummary: &graphpb.AppsStatefulSetSummary{
				BaseObject: &graphpb.BaseObject{
					Namespace:  defaultIdentityAttributes[string(semconv.K8SNamespaceNameKey)],
					Kind:       "StatefulSet",
					ApiVersion: "apps/v1",
					Name:       "test-statefulset",
				},
			},
		},
	}

	ec := NewResourceEntityCache()
	ec.ProvisionPackagedObject(item)
	entities := ec._allEntities()

	assert.Len(t, entities, 1)

	statefulSet := assertEntityExists(t, entities, "test-statefulset", KubernetesStatefulSet, listify(defaultIdentityAttributes)...)
	require.NotNil(t, statefulSet)
	assert.Equal(t, attrs, statefulSet.Attributes)
	assert.Len(t, statefulSet.Edges, 0)
}
