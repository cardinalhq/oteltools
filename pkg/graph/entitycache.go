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
	entityLocks sync.Map
	mapLock     sync.RWMutex // Global lock for entityMap
}

func NewResourceEntityCache() *ResourceEntityCache {
	return &ResourceEntityCache{
		entityMap: make(map[string]*ResourceEntity),
	}
}

func (ec *ResourceEntityCache) getOrCreateEntityLock(entityId string) *sync.RWMutex {
	lock, _ := ec.entityLocks.LoadOrStore(entityId, &sync.RWMutex{})
	return lock.(*sync.RWMutex)
}

func toEntityId(name, entityType string) string {
	return name + ":" + entityType
}

func (ec *ResourceEntityCache) PutEntity(attributeName, entityName, entityType string, attributes map[string]string) *ResourceEntity {
	entityId := toEntityId(entityName, entityType)

	ec.mapLock.RLock()
	entity, exists := ec.entityMap[entityId]
	ec.mapLock.RUnlock()

	if !exists {
		ec.mapLock.Lock()
		if entity, exists = ec.entityMap[entityId]; !exists {
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
	}

	for key, value := range attributes {
		entity.Attributes[key] = value
	}

	return entity
}

func (ec *ResourceEntityCache) PutEntityObject(entity *ResourceEntity) {
	entityId := toEntityId(entity.Name, entity.Type)

	ec.mapLock.RLock()
	_, exists := ec.entityMap[entityId]
	ec.mapLock.RUnlock()

	if !exists {
		ec.mapLock.Lock()
		if _, exists := ec.entityMap[entityId]; !exists {
			ec.entityMap[entityId] = entity
		}
		ec.mapLock.Unlock()
	}
}

func (ec *ResourceEntityCache) GetAllEntities() map[string]*ResourceEntity {
	ec.mapLock.Lock()
	defer ec.mapLock.Unlock()

	entities := ec.entityMap
	ec.entityMap = make(map[string]*ResourceEntity)
	return entities
}

func (re *ResourceEntity) AddEdge(targetName, targetType, relationship string) {
	if re.Edges == nil {
		re.Edges = make(map[string]string)
	}
	re.Edges[toEntityId(targetName, targetType)] = relationship
}

const (
	BelongsToNamespace       = "belongs to namespace"
	IsPartOfCluster          = "is part of cluster"
	IsDeployedOnPod          = "is deployed on pod"
	IsRunningOnNode          = "is running on node"
	IsManagedByDeployment    = "is managed by deployment"
	IsManagedByStatefulSet   = "is managed by statefulset"
	IsManagedByReplicaSet    = "is managed by replicaset"
	HasNode                  = "has a node"
	HasNamespace             = "has a namespace"
	HasCollection            = "has a collection"
	ManagesDeployments       = "manages deployments"
	ManagesDaemonSets        = "manages daemon sets"
	ManagesStatefulSets      = "manages stateful sets"
	ManagesJobs              = "manages jobs"
	ManagesCronJobs          = "manages cron jobs"
	BelongsToCluster         = "belongs to cluster"
	SchedulesPod             = "schedules pod"
	RunsOnOperatingSystem    = "runs on operating system"
	ContainsPod              = "contains pod"
	ContainsDeployment       = "contains deployment"
	ContainsStatefulSet      = "contains statefulset"
	ContainsDaemonSet        = "contains daemonset"
	ContainsReplicaSet       = "contains replicaset"
	ContainsJob              = "contains job"
	ContainsCronJob          = "contains cronjob"
	IsPartOfDeployment       = "is part of deployment"
	IsPartOfStatefulSet      = "is part of statefulset"
	IsPartOfDaemonSet        = "is part of daemonset"
	IsScheduledOnNode        = "is scheduled on node"
	RunsInPod                = "runs in pod"
	IsPartOfNamespace        = "is part of namespace"
	IsDeployedOnNode         = "is deployed on node"
	IsManagedByCluster       = "is managed by cluster"
	ManagesReplicaset        = "manages replicaset"
	UsesImage                = "uses image"
	IsUsedByContainer        = "is used by container"
	IsAssociatedWithTask     = "is associated with task"
	IsAssociatedWithCluster  = "is associated with cluster"
	IsAssociatedWithNode     = "is associated with node"
	IsInstanceOfFunction     = "is instance of function"
	HasInstance              = "has instance"
	ContainsTask             = "contains task"
	ManagesAccount           = "manages account"
	ContainsRegion           = "contains region"
	ContainsAvailabilityZone = "contains availability zone"
	BelongsToProvider        = "belongs to provider"
	HasResourcesInRegion     = "has resources in region"
	BelongsToRegion          = "belongs to region"
	BelongsToZone            = "belongs to zone"
	BelongsToAccount         = "belongs to account"
	HostsService             = "hosts service"
	HostsPod                 = "hosts pod"
	HostsCluster             = "hosts cluster"
	IsAssociatedWith         = "is associated with"
	NetPeerName              = "net.peer.name"
	DBQuerySummary           = "db.query.summary"
	DBStatement              = "db.statement"
	UsesDatabase             = "uses database"
	IsUsedByDatabase         = "is used by database"
	IsDatabaseHostedOn       = "is a database hosted on"
	IsCollectionHostedOn     = "is a collection hosted on"
	MessagingProducer        = "messaging.producer"
	MessagingConsumer        = "messaging.consumer"
	ConsumesFrom             = "consumes from"
	ProducesTo               = "produces to"
)

func (ec *ResourceEntityCache) Provision(resourceAttributes pcommon.Map, attributes pcommon.Map) {
	// Shared global entity map across all relationship maps
	globalEntityMap := make(map[string]*ResourceEntity)

	ec.provisionEntities(resourceAttributes, globalEntityMap)
	dbEntities := toDBEntities(attributes)
	if len(dbEntities) > 0 {
		for _, v := range dbEntities {
			globalEntityMap[v.AttributeName] = v
			ec.PutEntityObject(v)
		}
	}
	messagingEntities := toMessagingEntities(attributes)
	if len(messagingEntities) > 0 {
		for _, v := range messagingEntities {
			globalEntityMap[v.AttributeName] = v
			ec.PutEntityObject(v)
		}
	}
	ec.provisionRelationships(globalEntityMap)
}

func (ec *ResourceEntityCache) provisionEntities(attributes pcommon.Map, entityMap map[string]*ResourceEntity) {
	matches := make(map[*EntityInfo]*ResourceEntity)
	attributes.Range(func(k string, v pcommon.Value) bool {
		entityValue := v.AsString()
		if entityInfo, exists := EntityRelationships[k]; exists {
			entityAttrs := make(map[string]string)
			entity := ec.PutEntity(k, entityValue, entityInfo.Type, entityAttrs)
			matches[entityInfo] = entity
			entityMap[k] = entity
		}
		return true
	})

	for entityInfo, entity := range matches {
		for _, attributeName := range entityInfo.AttributeNames {
			if v, exists := attributes.Get(attributeName); exists {
				entity.Attributes[attributeName] = v.AsString()
			}
		}
		if len(entityInfo.AttributePrefixes) > 0 {
			attributes.Range(func(k string, v pcommon.Value) bool {
				for _, prefix := range entityInfo.AttributePrefixes {
					if k == entity.AttributeName {
						return true
					}
					if !strings.HasPrefix(k, prefix) {
						return true
					}
					entity.Attributes[k] = v.AsString()
				}
				return true
			})
		}
	}
}

func (ec *ResourceEntityCache) provisionRelationships(globalEntityMap map[string]*ResourceEntity) {
	var unlinkedEntities []*ResourceEntity

	// Lock per entity to avoid contention
	for _, parentEntity := range globalEntityMap {
		if entityInfo, exists := EntityRelationships[parentEntity.AttributeName]; exists {
			parentLock := ec.getOrCreateEntityLock(toEntityId(parentEntity.Name, parentEntity.Type))
			parentLock.Lock()

			foundLinkage := false
			for childKey, relationship := range entityInfo.Relationships {
				if childEntity, childExists := globalEntityMap[childKey]; childExists {
					parentEntity.AddEdge(childEntity.Name, childEntity.Type, relationship)
					foundLinkage = true
				}
			}

			if !foundLinkage {
				unlinkedEntities = append(unlinkedEntities, parentEntity)
			}

			parentLock.Unlock()
		}
	}

	for _, entity := range unlinkedEntities {
		entityLock := ec.getOrCreateEntityLock(toEntityId(entity.Name, entity.Type))
		entityLock.Lock()

		for _, otherEntity := range globalEntityMap {
			if entity == otherEntity {
				continue
			}
			entity.AddEdge(otherEntity.Name, otherEntity.Type, IsAssociatedWith)
			break
		}
		entityLock.Unlock()
	}
}
