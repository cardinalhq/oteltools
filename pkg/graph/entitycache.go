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

	"github.com/cardinalhq/oteltools/pkg/chqpb"
	"go.opentelemetry.io/collector/pdata/pcommon"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
	"golang.org/x/exp/slog"
	"google.golang.org/protobuf/proto"
)

type ResourceEntity struct {
	AttributeName string            `json:"-"`
	Name          string            `json:"name"`
	Type          string            `json:"type"`
	Attributes    map[string]string `json:"attributes"`
	Edges         map[string]string `json:"edges"`
	mu            sync.Mutex
}

type ResourceEntityCache struct {
	entityMap sync.Map
}

func NewResourceEntityCache() *ResourceEntityCache {
	return &ResourceEntityCache{}
}

func toEntityId(name, entityType string) string {
	return name + ":" + entityType
}

func (ec *ResourceEntityCache) PutEntity(attributeName, entityName, entityType string, attributes map[string]string) *ResourceEntity {
	entityId := toEntityId(entityName, entityType)

	if entity, exists := ec.entityMap.Load(entityId); exists {
		re := entity.(*ResourceEntity)
		re.mu.Lock()
		for key, value := range attributes {
			re.PutAttribute(key, value)
		}
		re.mu.Unlock()
		return re
	}

	newEntity := &ResourceEntity{
		AttributeName: attributeName,
		Name:          entityName,
		Type:          entityType,
		Attributes:    make(map[string]string),
		Edges:         make(map[string]string),
	}

	entity, _ := ec.entityMap.LoadOrStore(entityId, newEntity)
	re := entity.(*ResourceEntity)
	re.mu.Lock()
	for key, value := range attributes {
		re.PutAttribute(key, value)
	}
	re.mu.Unlock()
	return re
}

func (ec *ResourceEntityCache) PutEntityObject(entity *ResourceEntity) {
	entityId := toEntityId(entity.Name, entity.Type)
	ec.entityMap.Store(entityId, entity)
}

func (ec *ResourceEntityCache) _allEntities() map[string]*ResourceEntity {
	entities := make(map[string]*ResourceEntity)

	ec.entityMap.Range(func(key, value interface{}) bool {
		entities[key.(string)] = value.(*ResourceEntity)
		return true
	})

	return entities
}

func (ec *ResourceEntityCache) GetAllEntities() []byte {
	var batch [][]byte

	ec.entityMap.Range(func(key, value interface{}) bool {
		entity := value.(*ResourceEntity)
		entity.mu.Lock()
		protoEntity := &chqpb.ResourceEntityProto{
			Name:       entity.Name,
			Type:       entity.Type,
			Attributes: entity.Attributes,
			Edges:      entity.Edges,
		}
		serialized, err := proto.Marshal(protoEntity)
		if err == nil {
			batch = append(batch, serialized)
		}
		entity.mu.Unlock()
		return true
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
		re.Edges = make(map[string]string)
	}
	re.Edges[toEntityId(targetName, targetType)] = relationship
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
	ec.provisionRelationships(entityMap)
	return entityMap
}

func (ec *ResourceEntityCache) ProvisionRecordAttributes(resourceEntityMap map[string]*ResourceEntity, recordAttributes pcommon.Map) {
	if serviceEntity, exists := resourceEntityMap[string(semconv.ServiceNameKey)]; exists {
		entityMap := map[string]*ResourceEntity{string(semconv.ServiceNameKey): serviceEntity}

		dbEntities := toDBEntities(recordAttributes)
		for _, v := range dbEntities {
			entityMap[v.AttributeName] = v
			ec.PutEntityObject(v)
		}

		messagingEntities := toMessagingEntities(recordAttributes)
		for _, v := range messagingEntities {
			entityMap[v.AttributeName] = v
			ec.PutEntityObject(v)
		}

		ec.provisionRelationships(entityMap)
	}
}

func (ec *ResourceEntityCache) provisionEntities(attributes pcommon.Map, entityMap map[string]*ResourceEntity) {
	matches := make(map[*EntityInfo]*ResourceEntity)
	attributes.Range(func(k string, v pcommon.Value) bool {
		entityValue := v.AsString()
		if entityInfo, exists := EntityRelationships[k]; exists && entityValue != "" {
			entityAttrs := make(map[string]string)
			entity := ec.PutEntity(k, entityValue, entityInfo.Type, entityAttrs)
			matches[entityInfo] = entity
			entityMap[k] = entity
		}
		return true
	})

	for entityInfo, entity := range matches {
		for _, attributeName := range entityInfo.AttributeNames {
			entity.mu.Lock()
			if v, exists := attributes.Get(attributeName); exists {
				entity.PutAttribute(attributeName, v.AsString())
			}
			entity.mu.Unlock()
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

	for _, parentEntity := range globalEntityMap {
		if entityInfo, exists := EntityRelationships[parentEntity.AttributeName]; exists {

			foundLinkage := false
			parentEntity.mu.Lock()
			for childKey, relationship := range entityInfo.Relationships {
				if childEntity, childExists := globalEntityMap[childKey]; childExists {
					parentEntity.AddEdge(childEntity.Name, childEntity.Type, relationship)
					foundLinkage = true
				}
			}
			parentEntity.mu.Unlock()

			if !foundLinkage {
				unlinkedEntities = append(unlinkedEntities, parentEntity)
			}
		}
	}

	for _, entity := range unlinkedEntities {
		entity.mu.Lock()
		for _, otherEntity := range globalEntityMap {
			if entity == otherEntity {
				continue
			}
			entity.AddEdge(otherEntity.Name, otherEntity.Type, IsAssociatedWith)
			break
		}
		entity.mu.Unlock()
	}
}
