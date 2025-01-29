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
	"strings"
	"sync"
)

type ResourceEntity struct {
	Name       string
	Type       string
	Attributes map[string]string
	Edges      map[string]string // Key: (entityName:entityType), Value: relationship info
}

type ResourceEntityCache struct {
	entityMap   map[string]*ResourceEntity
	entityLocks map[string]*sync.RWMutex
	mapLock     sync.RWMutex // Global lock for entityMap
}

func NewResourceEntityCache() *ResourceEntityCache {
	return &ResourceEntityCache{
		entityMap:   make(map[string]*ResourceEntity),
		entityLocks: make(map[string]*sync.RWMutex),
	}
}

func (ec *ResourceEntityCache) getOrCreateEntityLock(name string) *sync.RWMutex {
	ec.mapLock.Lock()
	defer ec.mapLock.Unlock()

	if _, exists := ec.entityLocks[name]; !exists {
		ec.entityLocks[name] = &sync.RWMutex{}
	}
	return ec.entityLocks[name]
}

func toEntityId(name, entityType string) string {
	return name + ":" + entityType
}

func (ec *ResourceEntityCache) PutEntity(name, entityType string, attributes map[string]string) *ResourceEntity {
	entityLock := ec.getOrCreateEntityLock(name)
	entityLock.Lock()
	defer entityLock.Unlock()

	ec.mapLock.Lock()
	entity, exists := ec.entityMap[name]
	if !exists {
		entity = &ResourceEntity{
			Name:       name,
			Type:       entityType,
			Attributes: make(map[string]string),
			Edges:      make(map[string]string),
		}
		ec.entityMap[toEntityId(name, entityType)] = entity
	}
	ec.mapLock.Unlock()

	// Update attributes
	for key, value := range attributes {
		entity.Attributes[key] = value
	}
	return entity
}

func (ec *ResourceEntityCache) GetAllEntities() map[string]*ResourceEntity {
	ec.mapLock.Lock()
	defer ec.mapLock.Unlock()

	entities := ec.entityMap
	ec.entityMap = make(map[string]*ResourceEntity)
	ec.entityLocks = make(map[string]*sync.RWMutex)
	return entities
}

func (re *ResourceEntity) AddEdge(targetName, targetType, relationship string) {
	if re.Edges == nil {
		re.Edges = make(map[string]string)
	}
	re.Edges[toEntityId(targetName, targetType)] = relationship
}

type EntityInfo struct {
	Type              string
	Relationships     map[string]string
	AttributePrefixes []string
}

type RelationshipMap map[string]*EntityInfo

const (
	BelongsToNamespace           = "belongs to namespace"
	IsPartOfCluster              = "is part of cluster"
	IsDeployedOnPod              = "is deployed on pod"
	IsRunningOnNode              = "is running on node"
	IsManagedByDeployment        = "is managed by deployment"
	IsManagedByStatefulSet       = "is managed by statefulset"
	IsManagedByReplicaSet        = "is managed by replicaset"
	HasNode                      = "has a node"
	HasNamespace                 = "has a namespace"
	ManagesDeployments           = "manages deployments"
	ManagesDaemonSets            = "manages daemon sets"
	ManagesStatefulSets          = "manages stateful sets"
	ManagesJobs                  = "manages jobs"
	ManagesCronJobs              = "manages cron jobs"
	BelongsToCluster             = "belongs to cluster"
	SchedulesPod                 = "schedules pod"
	RunsOnOperatingSystem        = "runs on operating system"
	ContainsPod                  = "contains pod"
	ContainsDeployment           = "contains deployment"
	ContainsStatefulSet          = "contains statefulset"
	ContainsDaemonSet            = "contains daemonset"
	ContainsReplicaSet           = "contains replicaset"
	ContainsJob                  = "contains job"
	ContainsCronJob              = "contains cronjob"
	IsPartOfDeployment           = "is part of deployment"
	IsPartOfStatefulSet          = "is part of statefulset"
	IsPartOfDaemonSet            = "is part of daemonset"
	IsScheduledOnNode            = "is scheduled on node"
	RunsInPod                    = "runs in pod"
	IsPartOfNamespace            = "is part of namespace"
	IsDeployedOnNode             = "is deployed on node"
	IsManagedByCluster           = "is managed by cluster"
	ManagesReplicaset            = "manages replicaset"
	UsesImage                    = "uses image"
	IsUsedByContainer            = "is used by container"
	IsAssociatedWithTask         = "is associated with task"
	IsAssociatedWithCluster      = "is associated with cluster"
	IsAssociatedWithNode         = "is associated with node"
	IsInstanceOfFunction         = "is instance of function"
	HasInstance                  = "has instance"
	ContainsTask                 = "contains task"
	ManagesAccount               = "manages account"
	ContainsRegion               = "contains region"
	ContainsAvailabilityZone     = "contains availability zone"
	BelongsToProvider            = "belongs to provider"
	HasResourcesInRegion         = "has resources in region"
	ContainsResourcesFromAccount = "contains resources from account"
	BelongsToRegion              = "belongs to region"
)

var KubernetesRelationships = RelationshipMap{
	// Service
	string(semconv.ServiceNameKey): {
		Type: "service",
		Relationships: map[string]string{
			string(semconv.K8SNamespaceNameKey):   BelongsToNamespace,
			string(semconv.K8SClusterNameKey):     IsPartOfCluster,
			string(semconv.K8SPodNameKey):         IsDeployedOnPod,
			string(semconv.K8SNodeNameKey):        IsRunningOnNode,
			string(semconv.K8SDeploymentNameKey):  IsManagedByDeployment,
			string(semconv.K8SStatefulSetNameKey): IsManagedByStatefulSet,
			string(semconv.K8SDaemonSetNameKey):   IsManagedByDeployment,
			string(semconv.K8SReplicaSetNameKey):  IsManagedByReplicaSet,
		},
		AttributePrefixes: []string{"service."},
	},

	// Cluster
	string(semconv.K8SClusterNameKey): {
		Type: "k8s.cluster",
		Relationships: map[string]string{
			string(semconv.K8SNodeNameKey):        HasNode,
			string(semconv.K8SNamespaceNameKey):   HasNamespace,
			string(semconv.K8SDeploymentNameKey):  ManagesDeployments,
			string(semconv.K8SDaemonSetNameKey):   ManagesDaemonSets,
			string(semconv.K8SStatefulSetNameKey): ManagesStatefulSets,
			string(semconv.K8SJobNameKey):         ManagesJobs,
			string(semconv.K8SCronJobNameKey):     ManagesCronJobs,
		},
		AttributePrefixes: []string{"k8s.cluster."},
	},

	// Node
	string(semconv.K8SNodeNameKey): {
		Type: "k8s.node",
		Relationships: map[string]string{
			string(semconv.K8SClusterNameKey): BelongsToCluster,
			string(semconv.K8SPodNameKey):     SchedulesPod,
			string(semconv.OSNameKey):         RunsOnOperatingSystem,
		},
		AttributePrefixes: []string{"k8s.node."},
	},

	// Namespace
	string(semconv.K8SNamespaceNameKey): {
		Type: "k8s.namespace",
		Relationships: map[string]string{
			string(semconv.K8SClusterNameKey):     BelongsToCluster,
			string(semconv.K8SPodNameKey):         ContainsPod,
			string(semconv.K8SDeploymentNameKey):  ContainsDeployment,
			string(semconv.K8SStatefulSetNameKey): ContainsStatefulSet,
			string(semconv.K8SDaemonSetNameKey):   ContainsDaemonSet,
			string(semconv.K8SReplicaSetNameKey):  ContainsReplicaSet,
			string(semconv.K8SJobNameKey):         ContainsJob,
			string(semconv.K8SCronJobNameKey):     ContainsCronJob,
		},
		AttributePrefixes: []string{"k8s.namespace."},
	},

	// Pod
	string(semconv.K8SPodNameKey): {
		Type: "k8s.pod",
		Relationships: map[string]string{
			string(semconv.K8SNamespaceNameKey):   BelongsToNamespace,
			string(semconv.K8SNodeNameKey):        IsScheduledOnNode,
			string(semconv.K8SClusterNameKey):     IsPartOfCluster,
			string(semconv.K8SReplicaSetNameKey):  IsManagedByReplicaSet,
			string(semconv.K8SDeploymentNameKey):  IsPartOfDeployment,
			string(semconv.K8SStatefulSetNameKey): IsPartOfStatefulSet,
			string(semconv.K8SDaemonSetNameKey):   IsPartOfDaemonSet,
		},
		AttributePrefixes: []string{"k8s.pod."},
	},

	// Container
	string(semconv.K8SContainerNameKey): {
		Type: "k8s.container",
		Relationships: map[string]string{
			string(semconv.K8SPodNameKey):       RunsInPod,
			string(semconv.K8SNamespaceNameKey): IsPartOfNamespace,
			string(semconv.K8SNodeNameKey):      IsDeployedOnNode,
		},
		AttributePrefixes: []string{"k8s.container."},
	},

	// ReplicaSet
	string(semconv.K8SReplicaSetNameKey): {
		Type: "k8s.replicaset",
		Relationships: map[string]string{
			string(semconv.K8SNamespaceNameKey):  BelongsToNamespace,
			string(semconv.K8SDeploymentNameKey): IsManagedByDeployment,
			string(semconv.K8SClusterNameKey):    IsPartOfCluster,
		},
		AttributePrefixes: []string{"k8s.replicaset."},
	},

	// Deployment
	string(semconv.K8SDeploymentNameKey): {
		Type: "k8s.deployment",
		Relationships: map[string]string{
			string(semconv.K8SNamespaceNameKey):  BelongsToNamespace,
			string(semconv.K8SClusterNameKey):    IsManagedByCluster,
			string(semconv.K8SReplicaSetNameKey): ManagesReplicaset,
		},
		AttributePrefixes: []string{"k8s.deployment."},
	},

	// DaemonSet
	string(semconv.K8SDaemonSetNameKey): {
		Type: "k8s.daemonset",
		Relationships: map[string]string{
			string(semconv.K8SNamespaceNameKey): BelongsToNamespace,
			string(semconv.K8SClusterNameKey):   IsManagedByCluster,
		},
		AttributePrefixes: []string{"k8s.daemonset."},
	},

	// StatefulSet
	string(semconv.K8SStatefulSetNameKey): {
		Type: "k8s.statefulset",
		Relationships: map[string]string{
			string(semconv.K8SNamespaceNameKey): BelongsToNamespace,
			string(semconv.K8SClusterNameKey):   IsManagedByCluster,
		},
		AttributePrefixes: []string{"k8s.statefulset."},
	},

	// Job
	string(semconv.K8SJobNameKey): {
		Type: "k8s.job",
		Relationships: map[string]string{
			string(semconv.K8SNamespaceNameKey): BelongsToNamespace,
			string(semconv.K8SClusterNameKey):   IsManagedByCluster,
		},
		AttributePrefixes: []string{"k8s.job."},
	},

	// CronJob
	string(semconv.K8SCronJobNameKey): {
		Type: "k8s.cronjob",
		Relationships: map[string]string{
			string(semconv.K8SNamespaceNameKey): BelongsToNamespace,
			string(semconv.K8SClusterNameKey):   IsManagedByCluster,
		},
		AttributePrefixes: []string{"k8s.cronjob."},
	},
}

var ContainerRelationships = RelationshipMap{
	// Container Entity
	string(semconv.ContainerNameKey): {
		Type: "container",
		Relationships: map[string]string{
			string(semconv.ContainerImageNameKey): UsesImage,
		},
		AttributePrefixes: []string{
			"container.command",
			"container.",
			"oci.",
		},
	},

	// Container Image Entity
	string(semconv.ContainerImageNameKey): {
		Type: "container.image",
		Relationships: map[string]string{
			string(semconv.ContainerNameKey): IsUsedByContainer,
		},
		AttributePrefixes: []string{
			"container.image",
		},
	},
}

var OSRelationships = RelationshipMap{
	string(semconv.OSNameKey): {
		Type:              "os",
		Relationships:     map[string]string{},
		AttributePrefixes: []string{"os."},
	},
}

var ProcessRelationships = RelationshipMap{
	string(semconv.ProcessCommandKey): {
		Type:              "process",
		Relationships:     map[string]string{},
		AttributePrefixes: []string{"process."},
	},
}

var AWSECSRelationships = RelationshipMap{
	// ECS Container
	string(semconv.AWSECSContainerARNKey): {
		Type: "aws.ecs.container",
		Relationships: map[string]string{
			string(semconv.AWSECSClusterARNKey): IsPartOfCluster,
			string(semconv.AWSECSTaskIDKey):     IsAssociatedWithTask,
		},
		AttributePrefixes: []string{
			"aws.ecs.task.",
		},
	},

	// ECS Task
	string(semconv.AWSECSTaskIDKey): {
		Type: "aws.ecs.task",
		Relationships: map[string]string{
			string(semconv.AWSECSClusterARNKey): IsPartOfCluster,
		},
		AttributePrefixes: []string{
			"aws.ecs.task.",
		},
	},

	// ECS Cluster
	string(semconv.AWSECSClusterARNKey): {
		Type: "aws.ecs.cluster",
		Relationships: map[string]string{
			string(semconv.AWSECSTaskIDKey): ContainsTask,
		},
		AttributePrefixes: []string{
			"aws.ecs.launchtype.",
		},
	},
}

var AWSEKSRelationships = RelationshipMap{
	// EKS Cluster
	string(semconv.AWSEKSClusterARNKey): {
		Type: "aws.eks.cluster",
		Relationships: map[string]string{
			string(semconv.K8SClusterNameKey): IsAssociatedWithCluster,
			string(semconv.K8SNodeNameKey):    IsAssociatedWithNode,
		},
		AttributePrefixes: []string{
			"aws.eks.",
		},
	},
}

var FAASRelationships = RelationshipMap{
	// FaaS Instance
	string(semconv.FaaSInstanceKey): {
		Type: "faas.instance",
		Relationships: map[string]string{
			string(semconv.FaaSNameKey): IsInstanceOfFunction,
		},
		AttributePrefixes: []string{"faas.instance."},
	},

	// FaaS Function
	string(semconv.FaaSNameKey): {
		Type: "faas.function",
		Relationships: map[string]string{
			string(semconv.FaaSInstanceKey): HasInstance,
		},
		AttributePrefixes: []string{"faas."},
	},
}

var CloudRelationships = RelationshipMap{
	// Cloud Provider (e.g., AWS, GCP, Azure)
	string(semconv.CloudProviderKey): {
		Type: "cloud.provider",
		Relationships: map[string]string{
			string(semconv.CloudAccountIDKey):        ManagesAccount,
			string(semconv.CloudRegionKey):           ContainsRegion,
			string(semconv.CloudAvailabilityZoneKey): ContainsAvailabilityZone,
		},
		AttributePrefixes: []string{"cloud."},
	},

	// Cloud Account
	string(semconv.CloudAccountIDKey): {
		Type: "cloud.account",
		Relationships: map[string]string{
			string(semconv.CloudProviderKey): BelongsToProvider,
			string(semconv.CloudRegionKey):   HasResourcesInRegion,
		},
		AttributePrefixes: []string{"cloud.account."},
	},

	// Cloud Region
	string(semconv.CloudRegionKey): {
		Type: "cloud.region",
		Relationships: map[string]string{
			string(semconv.CloudProviderKey):         BelongsToProvider,
			string(semconv.CloudAvailabilityZoneKey): ContainsAvailabilityZone,
			string(semconv.CloudAccountIDKey):        ContainsResourcesFromAccount,
		},
		AttributePrefixes: []string{"cloud.region."},
	},

	// Cloud Availability Zone
	string(semconv.CloudAvailabilityZoneKey): {
		Type: "cloud.availability_zone",
		Relationships: map[string]string{
			string(semconv.CloudRegionKey): BelongsToRegion,
		},
		AttributePrefixes: []string{"cloud.availability_zone."},
	},
}

var AllRelationships = []RelationshipMap{
	KubernetesRelationships,
	ContainerRelationships,
	OSRelationships,
	ProcessRelationships,
	AWSECSRelationships,
	AWSEKSRelationships,
	FAASRelationships,
	CloudRelationships,
}

func (ec *ResourceEntityCache) Provision(attributes pcommon.Map) {
	// Shared global entity map across all relationship maps
	globalEntityMap := make(map[string]*ResourceEntity)

	for _, relationship := range AllRelationships {
		ec.provisionEntities(attributes, relationship, globalEntityMap)
	}

	for _, relationship := range AllRelationships {
		ec.provisionRelationships(relationship, globalEntityMap)
	}
}

func (ec *ResourceEntityCache) provisionRelationships(relationships RelationshipMap, globalEntityMap map[string]*ResourceEntity) {
	for parentKey, entityInfo := range relationships {
		parentEntity, parentExists := globalEntityMap[parentKey]
		if !parentExists {
			continue
		}

		parentLock := ec.getOrCreateEntityLock(parentEntity.Name)
		parentLock.Lock()

		for childKey, relationship := range entityInfo.Relationships {
			childEntity, childExists := globalEntityMap[childKey]
			if childExists {
				parentEntity.AddEdge(childEntity.Name, childEntity.Type, relationship)
			}
		}

		parentLock.Unlock()
	}
}

func (ec *ResourceEntityCache) provisionEntities(attributes pcommon.Map, relationships RelationshipMap, entityMap map[string]*ResourceEntity) {
	attributes.Range(func(k string, v pcommon.Value) bool {
		entityValue := v.AsString()
		if entityInfo, exists := relationships[k]; exists {
			entityAttrs := make(map[string]string)

			attributes.Range(func(attrKey string, attrValue pcommon.Value) bool {
				for _, prefix := range entityInfo.AttributePrefixes {
					if strings.HasPrefix(attrKey, prefix) {
						entityAttrs[attrKey] = attrValue.AsString()
						break
					}
				}
				return true
			})

			entity := ec.PutEntity(entityValue, entityInfo.Type, entityAttrs)
			entityMap[k] = entity
		}
		return true
	})
}
