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
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/cardinalhq/oteltools/pkg/chqpb"
	"go.opentelemetry.io/collector/pdata/pcommon"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
	"golang.org/x/exp/slog"
	"google.golang.org/protobuf/proto"
)

type EdgeInfo struct {
	EntityId     *EntityId
	Relationship string
	LastSeen     int64
}

const (
	defaultExpiry = 600000 // 10 minutes
)

type EntityId struct {
	Name         string
	Type         string
	IdAttributes map[string]string
}

func (e *EntityId) toProto() *chqpb.ResourceEntityId {
	var tuples []*chqpb.AttributeTuple
	for k, v := range e.IdAttributes {
		tuples = append(tuples, &chqpb.AttributeTuple{
			Key:   k,
			Value: v,
		})
	}
	return &chqpb.ResourceEntityId{
		Name:            e.Name,
		Type:            e.Type,
		AttributeTuples: tuples,
	}
}

func (e *EntityId) toKey() string {
	var b strings.Builder

	// Start with Name and Type
	b.WriteString(e.Name)
	b.WriteString("|")
	b.WriteString(e.Type)

	keys := make([]string, 0, len(e.IdAttributes))
	for k := range e.IdAttributes {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := e.IdAttributes[k]
		b.WriteString("|")
		b.WriteString(k)
		b.WriteString("=")
		b.WriteString(v)
	}

	return b.String()
}

type ResourceEntity struct {
	EntityId      *EntityId
	AttributeName string
	Attributes    map[string]string
	Edges         map[string]*EdgeInfo
	lastSeen      int64
	mu            sync.Mutex
}

type ResourceEntityCache struct {
	entityMap sync.Map
}

func NewResourceEntityCache() *ResourceEntityCache {
	return &ResourceEntityCache{}
}

func (ec *ResourceEntityCache) PutEntity(attributeName string, entityId *EntityId, attributes map[string]string) (*ResourceEntity, bool) {
	if entity, exists := ec.entityMap.Load(entityId); exists {
		re := entity.(*ResourceEntity)
		re.mu.Lock()
		re.lastSeen = now()
		for key, value := range attributes {
			re.PutAttribute(key, value)
		}
		re.mu.Unlock()
		return re, false
	}

	newEntity := &ResourceEntity{
		AttributeName: attributeName,
		EntityId:      entityId,
		Attributes:    make(map[string]string),
		Edges:         make(map[string]*EdgeInfo),
		lastSeen:      now(),
	}

	entity, loaded := ec.entityMap.LoadOrStore(entityId.toKey(), newEntity)
	re := entity.(*ResourceEntity)
	re.mu.Lock()
	for key, value := range attributes {
		re.PutAttribute(key, value)
	}
	re.mu.Unlock()
	return re, !loaded
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

	currentTime := now()
	ec.entityMap.Range(func(key, value interface{}) bool {
		entity := value.(*ResourceEntity)
		entity.mu.Lock()
		edges := make(map[string]string)
		for k, v := range entity.Edges {
			if currentTime-v.LastSeen > defaultExpiry {
				delete(entity.Edges, k)
				continue
			}
			edges[k] = v.Relationship
		}
		if currentTime-entity.lastSeen > defaultExpiry {
			entity.mu.Unlock()
			ec.entityMap.Delete(key)
			return true
		}
		var edgesProtoList []*chqpb.EdgeProto
		for k, v := range edges {
			edgesProtoList = append(edgesProtoList, &chqpb.EdgeProto{
				Relationship: v,
				Id:           entity.Edges[k].EntityId.toProto(),
			})
		}
		protoEntity := &chqpb.ResourceEntityProto{
			Id:         entity.EntityId.toProto(),
			Attributes: entity.Attributes,
			Edges:      edgesProtoList,
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

func (re *ResourceEntity) AddEdge(entityId *EntityId, relationship string) {
	if re.Edges == nil {
		re.Edges = make(map[string]*EdgeInfo)
	}
	key := entityId.toKey()
	if edgeInfo, exists := re.Edges[key]; exists {
		edgeInfo.LastSeen = now()
	} else {
		re.Edges[key] = &EdgeInfo{
			Relationship: relationship,
			LastSeen:     now(),
			EntityId:     entityId,
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
			entityId := &EntityId{
				Name: entityName,
				Type: entityInfo.Type,
			}
			for _, otherIdAttribute := range entityInfo.OtherIDAttributes {
				otherAttribute, found := attributes.Get(otherIdAttribute)
				if found {
					if entityId.IdAttributes == nil {
						entityId.IdAttributes = make(map[string]string)
					}
					entityId.IdAttributes[otherIdAttribute] = otherAttribute.Str()
				}
			}
			entity, isNewEntity := ec.PutEntity(k, entityId, entityAttrs)
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
	return newEntities
}

func (ec *ResourceEntityCache) provisionRelationships(globalEntityMap map[string]*ResourceEntity, recordAttributes pcommon.Map) {
	for _, parentEntity := range globalEntityMap {
		if entityInfo, exists := EntityRelationships[parentEntity.AttributeName]; exists {

			parentEntity.mu.Lock()
			for childKey, relationship := range entityInfo.Relationships {
				if childEntity, childExists := globalEntityMap[childKey]; childExists {
					parentEntity.AddEdge(childEntity.EntityId, relationship)
				}
			}

			if entityInfo.DeriveRelationshipCallbacks != nil {
				for childKey := range entityInfo.DeriveRelationshipCallbacks {
					deriveRelationshipCallback := entityInfo.DeriveRelationshipCallbacks[childKey]
					if childEntity, childExists := globalEntityMap[childKey]; childExists {
						relationship := deriveRelationshipCallback(recordAttributes)
						if relationship != "" {
							parentEntity.AddEdge(childEntity.EntityId, relationship)
						}
					}
				}
			}
			parentEntity.mu.Unlock()
		}
	}
}
