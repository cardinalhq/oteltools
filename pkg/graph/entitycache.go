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
	"strings"
	"sync"
	"time"

	"github.com/cardinalhq/oteltools/pkg/chqpb"
	"github.com/cardinalhq/oteltools/pkg/syncmap"
	"go.opentelemetry.io/collector/pdata/pcommon"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
	"golang.org/x/exp/slog"
	"google.golang.org/protobuf/proto"
)

type EdgeInfo struct {
	Relationship string
	LastSeen     int64
}

const (
	defaultExpiry = 600000 // 10 minutes
)

type ResourceEntity struct {
	sync.Mutex
	AttributeName string               `json:"-"`
	Name          string               `json:"name"`
	Type          string               `json:"type"`
	Attributes    map[string]string    `json:"attributes"`
	Edges         map[string]*EdgeInfo `json:"edges"`
	lastSeen      int64
}

type ResourceEntityCache struct {
	entityMap syncmap.SyncMap[string, *ResourceEntity]
}

func NewResourceEntityCache() *ResourceEntityCache {
	return &ResourceEntityCache{}
}

func toEntityId(name, entityType string) string {
	return name + "|" + entityType
}

func (ec *ResourceEntityCache) PutEntity(attributeName, entityName, entityType string, attributes map[string]string) (*ResourceEntity, bool) {
	entityId := toEntityId(entityName, entityType)

	if current, replaced, _ := ec.entityMap.ReplaceFunc(entityId, func(current *ResourceEntity) (*ResourceEntity, error) {
		current.Lock()
		current.lastSeen = now()
		for key, value := range attributes {
			current.PutAttribute(key, value)
		}
		current.Unlock()
		return current, nil
	}); replaced {
		return current, false
	}

	entity, created, _ := ec.entityMap.LoadOrStoreFunc(entityId, func() (*ResourceEntity, error) {
		newEntity := &ResourceEntity{
			AttributeName: attributeName,
			Name:          entityName,
			Type:          entityType,
			Attributes:    make(map[string]string),
			Edges:         make(map[string]*EdgeInfo),
			lastSeen:      now(),
		}
		for key, value := range attributes {
			newEntity.PutAttribute(key, value)
		}
		return newEntity, nil
	})
	return entity, created
}

func (ec *ResourceEntityCache) PutEntityObject(entity *ResourceEntity) {
	entityId := toEntityId(entity.Name, entity.Type)
	ec.entityMap.Store(entityId, entity)
}

func (ec *ResourceEntityCache) _allEntities() map[string]*ResourceEntity {
	entities := make(map[string]*ResourceEntity)

	ec.entityMap.Range(func(key string, value *ResourceEntity) bool {
		entities[key] = value
		return true
	})

	return entities
}

func (ec *ResourceEntityCache) GetAllEntities() []byte {
	var batch [][]byte

	currentTime := now()
	ec.entityMap.RemoveIf(func(key string, entity *ResourceEntity) bool {
		entity.Lock()
		defer entity.Unlock()
		edges := make(map[string]string)
		for k, v := range entity.Edges {
			if currentTime-v.LastSeen > defaultExpiry {
				delete(entity.Edges, k)
				continue
			}
			edges[k] = v.Relationship
		}
		if currentTime-entity.lastSeen > defaultExpiry {
			return true
		}
		protoEntity := &chqpb.ResourceEntityProto{
			Name:       entity.Name,
			Type:       entity.Type,
			Attributes: entity.Attributes,
			Edges:      edges,
		}
		serialized, err := proto.Marshal(protoEntity)
		if err == nil {
			batch = append(batch, serialized)
		}
		return false
	})

	serialized, err := proto.Marshal(&chqpb.ResourceEntityProtoList{Entities: batch})
	if err == nil {
		return serialized
	}
	slog.Error("Error marshaling entities", slog.String("error", err.Error()))
	return []byte{}
}

func (re *ResourceEntity) AddEdge(targetName, targetType, relationship string) {
	if re.Edges == nil {
		re.Edges = make(map[string]*EdgeInfo)
	}
	if edgeInfo, exists := re.Edges[toEntityId(targetName, targetType)]; exists {
		edgeInfo.LastSeen = now()
	} else {
		re.Edges[toEntityId(targetName, targetType)] = &EdgeInfo{
			Relationship: relationship,
			LastSeen:     now(),
		}
	}
}

func now() int64 {
	return time.Now().UnixMilli()
}

func (re *ResourceEntity) PutAttribute(k, v string) {
	re.Attributes[k] = v
}

func (ec *ResourceEntityCache) ProvisionResourceAttributes(attributes pcommon.Map) map[string]*ResourceEntity {
	entityMap := make(map[string]*ResourceEntity)
	serviceName, serviceNameFound := attributes.Get(string(semconv.ServiceNameKey))
	if serviceNameFound && RestrictedServices[serviceName.AsString()] {
		return entityMap
	}
	namespace, namespaceFound := attributes.Get(string(semconv.K8SNamespaceNameKey))
	if namespaceFound && RestrictedNamespaces[namespace.AsString()] {
		return entityMap
	}

	ec.provisionEntities(attributes, entityMap)
	ec.provisionRelationships(entityMap, attributes)
	return entityMap
}

func (ec *ResourceEntityCache) ProvisionRecordAttributes(resourceEntityMap map[string]*ResourceEntity, recordAttributes pcommon.Map) {
	newEntityMap := ec.provisionEntities(recordAttributes, resourceEntityMap)
	if serviceEntity, exists := resourceEntityMap[string(semconv.ServiceNameKey)]; exists {
		newEntityMap[string(semconv.ServiceNameKey)] = serviceEntity
	}
	if len(newEntityMap) > 0 {
		ec.provisionRelationships(newEntityMap, recordAttributes)
	}
}

func (ec *ResourceEntityCache) provisionEntities(attributes pcommon.Map, entityMap map[string]*ResourceEntity) map[string]*ResourceEntity {
	matches := make(map[*EntityInfo]*ResourceEntity)
	newEntities := make(map[string]*ResourceEntity)
	attributes.Range(func(k string, v pcommon.Value) bool {
		entityName := v.AsString()
		if entityInfo, exists := EntityRelationships[k]; exists && entityName != "" {
			if entityInfo.ShouldCreateCallBack != nil {
				if !entityInfo.ShouldCreateCallBack(attributes) {
					return true
				}
			}
			entityAttrs := make(map[string]string)
			if entityInfo.NameTransformer != nil {
				entityName = entityInfo.NameTransformer(entityName)
			}
			entity, isNewEntity := ec.PutEntity(k, entityName, entityInfo.Type, entityAttrs)
			if isNewEntity {
				newEntities[k] = entity
			}
			matches[entityInfo] = entity
			entityMap[k] = entity
		}
		return true
	})

	for entityInfo, entity := range matches {
		for _, attributeName := range entityInfo.AttributeNames {
			entity.Lock()
			if v, exists := attributes.Get(attributeName); exists {
				entity.PutAttribute(attributeName, v.AsString())
			}
			entity.Unlock()
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
	return newEntities
}

func (ec *ResourceEntityCache) provisionRelationships(globalEntityMap map[string]*ResourceEntity, recordAttributes pcommon.Map) {
	for _, parentEntity := range globalEntityMap {
		if entityInfo, exists := EntityRelationships[parentEntity.AttributeName]; exists {

			parentEntity.Lock()
			for childKey, relationship := range entityInfo.Relationships {
				if childEntity, childExists := globalEntityMap[childKey]; childExists {
					parentEntity.AddEdge(childEntity.Name, childEntity.Type, relationship)
				}
			}

			if entityInfo.DeriveRelationshipCallbacks != nil {
				for childKey := range entityInfo.DeriveRelationshipCallbacks {
					deriveRelationshipCallback := entityInfo.DeriveRelationshipCallbacks[childKey]
					if childEntity, childExists := globalEntityMap[childKey]; childExists {
						relationship := deriveRelationshipCallback(recordAttributes)
						if relationship != "" {
							parentEntity.AddEdge(childEntity.Name, childEntity.Type, relationship)
						}
					}
				}
			}
			parentEntity.Unlock()
		}
	}
}
