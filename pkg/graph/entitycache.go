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
	AttributeName string               `json:"-"`
	Name          string               `json:"name"`
	Type          string               `json:"type"`
	Attributes    map[string]string    `json:"attributes"`
	Edges         map[string]*EdgeInfo `json:"edges"`
	lastSeen      int64
	mu            sync.Mutex
}

type ResourceEntityCache struct {
	entityMap sync.Map
}

func NewResourceEntityCache() *ResourceEntityCache {
	return &ResourceEntityCache{}
}

func toEntityId(name, entityType string) string {
	return name + "|" + entityType
}

func (ec *ResourceEntityCache) PutEntity(attributeName, entityName, entityType string, attributes map[string]string) (*ResourceEntity, bool) {
	entityId := toEntityId(entityName, entityType)

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
		Name:          entityName,
		Type:          entityType,
		Attributes:    make(map[string]string),
		Edges:         make(map[string]*EdgeInfo),
		lastSeen:      now(),
	}

	entity, loaded := ec.entityMap.LoadOrStore(entityId, newEntity)
	re := entity.(*ResourceEntity)
	re.mu.Lock()
	for key, value := range attributes {
		re.PutAttribute(key, value)
	}
	re.mu.Unlock()
	return re, !loaded
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
			parentEntity.mu.Unlock()
		}
	}
}

func (ec *ResourceEntityCache) ProvisionStatefulSet(statefulSet *K8SStatefulSetObject) {
	entityAttrs := make(map[string]string)
	for resourceName, resourceValue := range statefulSet.Attributes {
		entityAttrs[resourceName] = resourceValue
	}

	serviceEntity, _ := ec.PutEntity(string(semconv.ServiceNameKey), statefulSet.Name, Service, make(map[string]string))
	statefulSetEntity, _ := ec.PutEntity(string(semconv.K8SStatefulSetNameKey), statefulSet.Name, KubernetesStatefulSet, entityAttrs)
	if serviceEntity != nil && statefulSetEntity != nil {
		serviceEntity.mu.Lock()
		serviceEntity.AddEdge(statefulSet.Name, KubernetesStatefulSet, IsManagedByStatefulSet)
		serviceEntity.mu.Unlock()
	}
}

func (ec *ResourceEntityCache) ProvisionDeployment(deployment *K8SDeploymentObject) {
	if deployment.Name != "" {
		entityAttrs := make(map[string]string)
		entityAttrs["replicasAvailable"] = string(rune(deployment.ReplicasAvailable))
		entityAttrs["replicasDesired"] = string(rune(deployment.ReplicasDesired))
		entityAttrs["replicasUpdated"] = string(rune(deployment.ReplicasUpdated))
		entityAttrs["deploymentStatus"] = deployment.DeploymentStatus
		entityAttrs["progressMessage"] = deployment.ProgressMessage

		serviceEntity, _ := ec.PutEntity(string(semconv.ServiceNameKey), deployment.Name, Service, make(map[string]string))
		deploymentEntity, _ := ec.PutEntity(string(semconv.K8SDeploymentNameKey), deployment.Name, KubernetesDeployment, entityAttrs)
		if serviceEntity != nil && deploymentEntity != nil {
			serviceEntity.mu.Lock()
			serviceEntity.AddEdge(deployment.Name, KubernetesDeployment, IsManagedByDeployment)
			serviceEntity.mu.Unlock()
		}
	}
}

func (ec *ResourceEntityCache) ProvisionPod(podObject *K8SPodObject) {
	entityAttrs := make(map[string]string)
	for labelName, labelValue := range podObject.Labels {
		entityAttrs[labelName] = labelValue
	}
	for resourceName, resourceValue := range podObject.Resources {
		entityAttrs[resourceName] = resourceValue
	}
	entityAttrs[ContainerImage] = podObject.ImageID
	entityAttrs[K8SPodIp] = podObject.PodIP
	entityAttrs[HostIp] = podObject.HostIP
	entityAttrs[PodPhase] = podObject.Phase
	if podObject.PendingReason != "" {
		entityAttrs[PendingReason] = podObject.PendingReason
	}
	podEntity, _ := ec.PutEntity(string(semconv.K8SPodNameKey), podObject.Name, KubernetesPod, entityAttrs)

	var parentAttributeName, parentEntityType string
	switch podObject.OwnerRefKind {
	case "ReplicaSet":
		parentAttributeName = string(semconv.K8SReplicaSetNameKey)
		parentEntityType = KubernetesReplicaSet

	case "StatefulSet":
		parentAttributeName = string(semconv.K8SStatefulSetNameKey)
		parentEntityType = KubernetesStatefulSet

	case "DaemonSet":
		parentAttributeName = string(semconv.K8SDaemonSetNameKey)
		parentEntityType = KubernetesDaemonSet
	default:
	}

	if parentAttributeName != "" {
		parentEntity, _ := ec.PutEntity(parentAttributeName, podObject.OwnerRefName, parentEntityType, make(map[string]string))
		parentEntity.mu.Lock()
		podEntity.AddEdge(parentEntity.Name, parentEntity.Type, IsAPodFor)
		parentEntity.mu.Unlock()
	}
}
