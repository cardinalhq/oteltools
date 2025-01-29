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
	"strings"
	"sync"
)

type ResourceEntity struct {
	AttributeName string            `json:"-"`
	Name          string            `json:"name"`
	Type          string            `json:"type"`
	Attributes    map[string]string `json:"attributes"`
	Edges         map[string]string `json:"edges"`
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

func (ec *ResourceEntityCache) PutEntity(attributeName, entityName, entityType string, attributes map[string]string) *ResourceEntity {
	entityId := toEntityId(entityName, entityType)
	entityLock := ec.getOrCreateEntityLock(entityId)
	entityLock.Lock()
	defer entityLock.Unlock()

	ec.mapLock.Lock()
	entity, exists := ec.entityMap[entityName]
	if !exists {
		entity = &ResourceEntity{
			AttributeName: attributeName,
			Name:          entityName,
			Type:          entityType,
			Attributes:    make(map[string]string),
			Edges:         make(map[string]string),
		}
		ec.entityMap[entityId] = entity
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

func (ec *ResourceEntityCache) Provision(attributes pcommon.Map) {
	// Shared global entity map across all relationship maps
	globalEntityMap := make(map[string]*ResourceEntity)

	ec.provisionEntities(attributes, globalEntityMap)
	ec.provisionRelationships(globalEntityMap)
}

func (ec *ResourceEntityCache) provisionEntities(attributes pcommon.Map, entityMap map[string]*ResourceEntity) {
	attributes.Range(func(k string, v pcommon.Value) bool {
		entityValue := v.AsString()
		if entityInfo, exists := Relationships[k]; exists {
			entityAttrs := make(map[string]string)

			attributes.Range(func(attrKey string, attrValue pcommon.Value) bool {
				for _, prefix := range entityInfo.AttributePrefixes {
					if attrKey != k && strings.HasPrefix(attrKey, prefix) {
						entityAttrs[attrKey] = attrValue.AsString()
						break
					}
				}
				return true
			})

			entity := ec.PutEntity(k, entityValue, entityInfo.Type, entityAttrs)
			entityMap[k] = entity
		}
		return true
	})
}

func (ec *ResourceEntityCache) provisionRelationships(globalEntityMap map[string]*ResourceEntity) {
	for _, parentEntity := range globalEntityMap {
		if entityInfo, exists := Relationships[parentEntity.AttributeName]; exists {

			parentLock := ec.getOrCreateEntityLock(toEntityId(parentEntity.Name, parentEntity.Type))
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
}
