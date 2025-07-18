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
	"hash/fnv"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/cardinalhq/oteltools/pkg/graph/graphpb"

	"github.com/cardinalhq/oteltools/pkg/chqpb"
	"go.opentelemetry.io/collector/pdata/pcommon"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
	"golang.org/x/exp/slog"
	"google.golang.org/protobuf/proto"
)

type EdgeInfo struct {
	Source       *EntityId
	Target       *EntityId
	Relationship string
	LastSeen     int64
	LastEmitted  int64
}

const (
	defaultExpiry  = 600000 // 10 minutes
	defaultCoolOff = 60 * 1000 * 20
)

type EntityId struct {
	Name         string
	Type         string
	IdAttributes map[string]string
	Hash         string
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
		Hash:            e.Hash,
		Name:            e.Name,
		Type:            e.Type,
		AttributeTuples: tuples,
	}
}

func (e *EntityId) computeHash() string {
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

	h := fnv.New64a()
	_, _ = h.Write([]byte(b.String()))
	return fmt.Sprintf("%x", h.Sum64())
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

func ToEntityId(name, entityType string, idAttributes map[string]string) *EntityId {
	entityId := &EntityId{
		Name:         name,
		Type:         entityType,
		IdAttributes: idAttributes,
	}
	entityId.Hash = entityId.computeHash()
	return entityId
}

func ToKubernetesEntityId(name, entityType, namespace, clusterName string) *EntityId {
	return ToEntityId(name, entityType, map[string]string{
		string(semconv.K8SNamespaceNameKey): namespace,
		string(semconv.K8SClusterNameKey):   clusterName,
	})
}

func clearAttributes(attributes map[string]string, prefix string) {
	for key := range attributes {
		if strings.HasPrefix(key, prefix) {
			delete(attributes, key)
		}
	}
}

func (ec *ResourceEntityCache) PutEntity(attributeName string, entityId *EntityId, attributes map[string]string) (*ResourceEntity, bool) {
	if entity, exists := ec.entityMap.Load(entityId.Hash); exists {
		re := entity.(*ResourceEntity)
		re.mu.Lock()
		re.lastSeen = now()
		switch re.EntityId.Type {
		case KubernetesPod:
			clearAttributes(re.Attributes, ContainerImageNamePrefix)
			clearAttributes(re.Attributes, ContainerImageIDPrefix)
		case KubernetesDeployment:
			clearAttributes(re.Attributes, ContainerImageNamePrefix)
		case KubernetesStatefulSet:
			clearAttributes(re.Attributes, ContainerImageNamePrefix)
		case KubernetesDaemonSet:
			clearAttributes(re.Attributes, ContainerImageNamePrefix)
		case KubernetesReplicaSet:
			clearAttributes(re.Attributes, ContainerImageNamePrefix)
		case KubernetesSecret, KubernetesConfigMap:
			clearAttributes(re.Attributes, DataHashPrefix)
		}
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

	entity, loaded := ec.entityMap.LoadOrStore(entityId.Hash, newEntity)
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
			if currentTime-v.LastEmitted < defaultCoolOff {
				continue
			}
			edges[k] = v.Relationship
		}
		if currentTime-entity.lastSeen > defaultExpiry {
			entity.mu.Unlock()
			ec.entityMap.Delete(key)
			return true
		}
		if len(edges) == 0 {
			entity.mu.Unlock()
			return true
		}
		var edgesProtoList []*chqpb.EdgeProto
		for k, v := range edges {
			edge := entity.Edges[k]
			edge.LastEmitted = currentTime
			edgesProtoList = append(edgesProtoList, &chqpb.EdgeProto{
				Relationship: v,
				Source:       edge.Source.toProto(),
				Target:       edge.Target.toProto(),
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

// AddEdge adds an edge from this entity to the target entity with the given relationship.
func (re *ResourceEntity) AddEdge(target *EntityId, relationship string) {
	if re.Edges == nil {
		re.Edges = make(map[string]*EdgeInfo)
	}
	key := re.EntityId.Hash + target.Hash
	if edgeInfo, exists := re.Edges[key]; exists {
		edgeInfo.LastSeen = now()
	} else {
		re.Edges[key] = &EdgeInfo{
			Relationship: relationship,
			LastSeen:     now(),
			Source:       re.EntityId,
			Target:       target,
		}
	}
}

// AddEdgeBacklink adds an edge from the source entity back to this entity with the given relationship.
func (re *ResourceEntity) AddEdgeBacklink(source *EntityId, relationship string) {
	if re.Edges == nil {
		re.Edges = make(map[string]*EdgeInfo)
	}
	key := source.Hash + re.EntityId.Hash
	if edgeInfo, exists := re.Edges[key]; exists {
		edgeInfo.LastSeen = now()
	} else {
		re.Edges[key] = &EdgeInfo{
			Relationship: relationship,
			LastSeen:     now(),
			Source:       source,
			Target:       re.EntityId,
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
	provisionedEntities := ec.provisionEntities(recordAttributes, resourceEntityMap)
	if serviceEntity, exists := resourceEntityMap[string(semconv.ServiceNameKey)]; exists {
		provisionedEntities[string(semconv.ServiceNameKey)] = serviceEntity
	}
	if len(provisionedEntities) > 0 {
		ec.provisionRelationships(provisionedEntities, recordAttributes)
	}
}

func (ec *ResourceEntityCache) provisionEntities(attributes pcommon.Map, entityMap map[string]*ResourceEntity) map[string]*ResourceEntity {
	matches := make(map[*EntityInfo]*ResourceEntity)
	provisionedEntities := make(map[string]*ResourceEntity)
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
			idAttributes := make(map[string]string)
			for _, otherIdAttribute := range entityInfo.OtherIDAttributes {
				otherAttribute, found := attributes.Get(otherIdAttribute)
				if found {
					oav := otherAttribute.Str()
					entityAttrs[otherIdAttribute] = oav // add id attributes to the entity attributes as well.
					idAttributes[otherIdAttribute] = oav
				}
			}
			entityId := ToEntityId(entityName, entityInfo.Type, idAttributes)
			entity, _ := ec.PutEntity(k, entityId, entityAttrs)
			provisionedEntities[k] = entity
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
	return provisionedEntities
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

			if entityInfo.CreateEntityRelationshipsCallback != nil {
				entityRelationships := entityInfo.CreateEntityRelationshipsCallback(recordAttributes)
				for _, entityRelationship := range entityRelationships {
					targetEntityId := ToEntityId(entityRelationship.EntityName, entityRelationship.EntityType, nil)
					ec.PutEntity(entityRelationship.EntityType, targetEntityId, entityRelationship.EntityAttributes)
					parentEntity.AddEdge(targetEntityId, entityRelationship.Relationship)
					if entityRelationship.ReverseRelationship != "" {
						parentEntity.AddEdgeBacklink(targetEntityId, entityRelationship.ReverseRelationship)
					}
				}
			}
			parentEntity.mu.Unlock()
		}
	}
}

func ownedByRelationshipForType(kind string) string {
	switch kind {
	case Deployment:
		return IsManagedByDeployment
	case StatefulSet:
		return IsManagedByStatefulSet
	case DaemonSet:
		return IsManagedByDaemonSet
	case ReplicaSet:
		return IsManagedByReplicaSet
	case Job:
		return IsManagedByCronJob
	case CronJob:
		return IsManagedByJob
	}
	return ""
}

func ownerByKind(kind string) string {
	switch kind {
	case Deployment:
		return KubernetesDeployment
	case StatefulSet:
		return KubernetesStatefulSet
	case DaemonSet:
		return KubernetesDaemonSet
	case ReplicaSet:
		return KubernetesReplicaSet
	case Job:
		return KubernetesJob
	case CronJob:
		return KubernetesCronJob
	}
	return ""
}

func addEdgesFromPodSpec(entity *ResourceEntity, podSpec *graphpb.PodSpec, namespace, clusterName string) {
	if podSpec == nil {
		return
	}
	entity.mu.Lock()
	defer entity.mu.Unlock()
	for _, container := range podSpec.Containers {
		for _, configMapName := range container.ConfigMapNames {
			configMapEntityId := ToKubernetesEntityId(configMapName, KubernetesConfigMap, namespace, clusterName)
			entity.AddEdge(configMapEntityId, UsesConfigMap)
			entity.AddEdgeBacklink(configMapEntityId, IsUsedByPod)
		}
		for _, secretName := range container.SecretNames {
			secretEntityId := ToKubernetesEntityId(secretName, KubernetesSecret, namespace, clusterName)
			entity.AddEdge(secretEntityId, UsesSecret)
			entity.AddEdgeBacklink(secretEntityId, IsUsedByPod)
		}
		if container.Image != "" {
			entity.PutAttribute(ContainerImageNamePrefix+container.Name, container.Image)
		}
	}
}

func ipv4FromList(ipList []string) string {
	for _, ip := range ipList {
		if strings.Contains(ip, ":") {
			continue
		}
		return ip
	}
	return ""
}

func inverseOfManaged(relationship string) string {
	switch relationship {
	case IsManagedByDeployment:
		return ManagesReplicaset
	case IsManagedByStatefulSet:
		return ManagesPod
	case IsManagedByDaemonSet:
		return ManagesPod
	case IsManagedByReplicaSet:
		return ManagesPod
	case IsManagedByCronJob:
		return ManagesJob
	case IsManagedByJob:
		return ManagesPod
	}
	return ""
}

func addEdgesForOwnerRefs(entity *ResourceEntity, bo *graphpb.BaseObject, clusterName string) {
	for _, owner := range bo.OwnerRef {
		relationship := ownedByRelationshipForType(owner.Kind)
		if relationship == "" {
			continue
		}
		ownerType := ownerByKind(owner.Kind)
		if ownerType == "" {
			continue
		}
		ownerEntityId := ToKubernetesEntityId(owner.Name, ownerType, bo.Namespace, clusterName)
		entity.mu.Lock()
		entity.AddEdge(ownerEntityId, relationship)
		reverseRelationship := inverseOfManaged(relationship)
		if reverseRelationship != "" {
			entity.AddEdgeBacklink(ownerEntityId, reverseRelationship)
		}
		entity.mu.Unlock()
	}
}

// trueFalse returns "true" if the value is true, otherwise "false".
func trueFalse(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

// ProvisionPackagedObject does the obvious.
// It also MODIFIES the passed in PackagedObject, specificially the ResourceAttributes,
// so this makes the PackagedObject not safe for concurrent use.
func (ec *ResourceEntityCache) ProvisionPackagedObject(po *graphpb.PackagedObject) {
	rattr := po.GetResourceAttributes()
	clusterName, found := rattr[string(semconv.K8SClusterNameKey)]
	if !found {
		return
	}
	namespace := po.GetBaseObject().Namespace

	switch obj := po.Object.(type) {
	case *graphpb.PackagedObject_PodSummary:
		podSummary := obj.PodSummary
		crashLoopBackoff := false
		imagePullBackoff := false
		oomKilled := false
		restartCount := 0
		if podSummary.Status != nil {
			for _, containerStatus := range podSummary.Status.ContainerStatus {
				crashLoopBackoff = crashLoopBackoff || containerStatus.IsCrashLoopBackOff
				imagePullBackoff = imagePullBackoff || containerStatus.IsImagePullBackOff
				oomKilled = oomKilled || containerStatus.WasOomKilled
				if containerStatus.Image != nil {
					if containerStatus.Image.Image != "" {
						rattr[ContainerImageNamePrefix+containerStatus.Name] = containerStatus.Image.Image
					}
					if containerStatus.Image.ImageId != "" {
						rattr[ContainerImageIDPrefix+containerStatus.Name] = containerStatus.Image.ImageId
					}
				}
				if containerStatus.RestartCount != 0 {
					restartCount += int(containerStatus.RestartCount)
				}
			}
		}
		rattr[CrashLoopBackOff] = trueFalse(crashLoopBackoff)
		rattr[ImagePullBackOff] = trueFalse(imagePullBackoff)
		rattr[OOMKilled] = trueFalse(oomKilled)
		rattr[RestartCount] = fmt.Sprintf("%d", restartCount)

		// port handling
		addContainerPorts(rattr, podSummary.Spec)

		podEntityId := ToKubernetesEntityId(podSummary.BaseObject.Name, KubernetesPod, namespace, clusterName)
		if podSummary.Status != nil {
			rattr[PodPhase] = podSummary.Status.Phase
			rattr[K8SPodIp] = ipv4FromList(podSummary.Status.PodIps)
			rattr[HostIp] = ipv4FromList(podSummary.Status.HostIps)
			if podSummary.Status.PhaseMessage != "" {
				rattr[PendingReason] = podSummary.Status.PhaseMessage
			}
		}
		pe, _ := ec.PutEntity(KubernetesPod, podEntityId, rattr)
		addEdgesForOwnerRefs(pe, podSummary.BaseObject, clusterName)
		if podSummary.Spec != nil {
			addEdgesFromPodSpec(pe, podSummary.Spec, namespace, clusterName)
		}

	case *graphpb.PackagedObject_SecretSummary:
		for name, hash := range obj.SecretSummary.Hashes {
			rattr[DataHashPrefix+name] = hash
		}
		secretId := ToKubernetesEntityId(obj.SecretSummary.BaseObject.Name, KubernetesSecret, namespace, clusterName)
		ec.PutEntity(KubernetesSecret, secretId, rattr)

	case *graphpb.PackagedObject_ConfigMapSummary:
		for name, hash := range obj.ConfigMapSummary.Hashes {
			rattr[DataHashPrefix+name] = hash
		}
		secretId := ToKubernetesEntityId(obj.ConfigMapSummary.BaseObject.Name, KubernetesConfigMap, namespace, clusterName)
		ec.PutEntity(KubernetesConfigMap, secretId, rattr)

	case *graphpb.PackagedObject_AppsDeploymentSummary:
		deploymentSummary := obj.AppsDeploymentSummary
		if deploymentSummary.Spec != nil {
			rattr[Replicas] = fmt.Sprintf("%d", deploymentSummary.Spec.Replicas)
		}
		if deploymentSummary.Status != nil {
			rattr[ReadyReplicas] = fmt.Sprintf("%d", deploymentSummary.Status.ReadyReplicas)
			rattr[AvailableReplicas] = fmt.Sprintf("%d", deploymentSummary.Status.AvailableReplicas)
			rattr[UnavailableReplicas] = fmt.Sprintf("%d", deploymentSummary.Status.UnavailableReplicas)
		}

		if deploymentSummary.Spec != nil && deploymentSummary.Spec.Template != nil && deploymentSummary.Spec.Template.PodSpec != nil {
			addContainerPorts(rattr, deploymentSummary.Spec.Template.PodSpec)
		}

		deploymentId := ToKubernetesEntityId(deploymentSummary.BaseObject.Name, KubernetesDeployment, namespace, clusterName)
		deploymentEntity, _ := ec.PutEntity(KubernetesDeployment, deploymentId, rattr)
		if deploymentSummary.Spec != nil && deploymentSummary.Spec.Template != nil && deploymentSummary.Spec.Template.PodSpec != nil {
			addEdgesFromPodSpec(deploymentEntity, deploymentSummary.Spec.Template.PodSpec, namespace, clusterName)
		}

	case *graphpb.PackagedObject_AppsStatefulSetSummary:
		statefulSetSummary := obj.AppsStatefulSetSummary
		if statefulSetSummary.Spec != nil {
			rattr[Replicas] = fmt.Sprintf("%d", statefulSetSummary.Spec.Replicas)
		}
		if statefulSetSummary.Status != nil {
			rattr[ReadyReplicas] = fmt.Sprintf("%d", statefulSetSummary.Status.ReadyReplicas)
			rattr[CurrentReplicas] = fmt.Sprintf("%d", statefulSetSummary.Status.CurrentReplicas)
			rattr[UpdatedReplicas] = fmt.Sprintf("%d", statefulSetSummary.Status.UpdatedReplicas)
		}

		if statefulSetSummary.Spec != nil && statefulSetSummary.Spec.Template != nil && statefulSetSummary.Spec.Template.PodSpec != nil {
			addContainerPorts(rattr, statefulSetSummary.Spec.Template.PodSpec)
		}

		statefulSetId := ToKubernetesEntityId(statefulSetSummary.BaseObject.Name, KubernetesStatefulSet, namespace, clusterName)
		statefulSetEntity, _ := ec.PutEntity(KubernetesStatefulSet, statefulSetId, rattr)
		if statefulSetSummary.Spec != nil && statefulSetSummary.Spec.Template != nil {
			addEdgesFromPodSpec(statefulSetEntity, statefulSetSummary.Spec.Template.PodSpec, namespace, clusterName)
		}

	case *graphpb.PackagedObject_AppsDaemonSetSummary:
		daemonSetSummary := obj.AppsDaemonSetSummary

		if daemonSetSummary.Spec != nil && daemonSetSummary.Spec.Template != nil && daemonSetSummary.Spec.Template.PodSpec != nil {
			addContainerPorts(rattr, daemonSetSummary.Spec.Template.PodSpec)
		}
		if daemonSetSummary.Spec != nil {
			rattr[Replicas] = fmt.Sprintf("%d", daemonSetSummary.Spec.Replicas)
		}

		daemonSetId := ToKubernetesEntityId(daemonSetSummary.BaseObject.Name, KubernetesDaemonSet, namespace, clusterName)
		daemonsetEntity, _ := ec.PutEntity(KubernetesDaemonSet, daemonSetId, rattr)
		if daemonSetSummary.Spec != nil && daemonSetSummary.Spec.Template != nil {
			addEdgesFromPodSpec(daemonsetEntity, daemonSetSummary.Spec.Template.PodSpec, namespace, clusterName)
		}

	case *graphpb.PackagedObject_AppsReplicaSetSummary:
		replicaSetSummary := obj.AppsReplicaSetSummary
		if replicaSetSummary.Spec != nil {
			rattr[Replicas] = fmt.Sprintf("%d", replicaSetSummary.Spec.Replicas)
		}
		if replicaSetSummary.Spec != nil && replicaSetSummary.Spec.Template != nil && replicaSetSummary.Spec.Template.PodSpec != nil {
			addContainerPorts(rattr, replicaSetSummary.Spec.Template.PodSpec)
		}
		replicaSetId := ToKubernetesEntityId(replicaSetSummary.BaseObject.Name, KubernetesReplicaSet, namespace, clusterName)
		replicaSetEntity, _ := ec.PutEntity(KubernetesReplicaSet, replicaSetId, rattr)
		addEdgesForOwnerRefs(replicaSetEntity, replicaSetSummary.BaseObject, clusterName)
		if replicaSetSummary.Spec != nil && replicaSetSummary.Spec.Template != nil {
			addEdgesFromPodSpec(replicaSetEntity, replicaSetSummary.Spec.Template.PodSpec, namespace, clusterName)
		}

	case *graphpb.PackagedObject_AutoscalingHpaSummary:
		hpaSummary := obj.AutoscalingHpaSummary
		if hpaSummary.Status != nil {
			rattr[CurrentReplicas] = fmt.Sprintf("%d", hpaSummary.Status.CurrentReplicas)
			rattr[DesiredReplicas] = fmt.Sprintf("%d", hpaSummary.Status.DesiredReplicas)

			for _, condition := range hpaSummary.Status.Conditions {
				if condition.Type == "ScalingLimited" {
					rattr[ScalingLimitedStatusAttribute] = trueFalse(condition.Status == "True")
					rattr[ScalingLimitedReasonAttribute] = condition.Reason
					rattr[ScalingLimitedMessageAttribute] = condition.Message
				}
			}
		}
		if hpaSummary.Spec != nil {
			rattr[MinReplicas] = fmt.Sprintf("%d", hpaSummary.Spec.MinReplicas)
			rattr[MaxReplicas] = fmt.Sprintf("%d", hpaSummary.Spec.MaxReplicas)
			if hpaSummary.Spec.Target != nil {
				targetKind := hpaSummary.Spec.Target.Kind
				targetApiVersion := hpaSummary.Spec.Target.ApiVersion
				switch targetApiVersion + "/" + targetKind {
				case "apps/v1/Deployment":
					rattr[ScalingTargetName] = hpaSummary.Spec.Target.Name
					rattr[ScalingTargetKind] = KubernetesDeployment
					rattr[ScalingTargetAPIVersion] = targetApiVersion
				case "apps/v1/StatefulSet":
					rattr[ScalingTargetName] = hpaSummary.Spec.Target.Name
					rattr[ScalingTargetKind] = KubernetesStatefulSet
					rattr[ScalingTargetAPIVersion] = targetApiVersion
				case "apps/v1/ReplicaSet":
					rattr[ScalingTargetName] = hpaSummary.Spec.Target.Name
					rattr[ScalingTargetKind] = KubernetesReplicaSet
					rattr[ScalingTargetAPIVersion] = targetApiVersion
				}
			}
		}

		hpaId := ToKubernetesEntityId(hpaSummary.BaseObject.Name, KubernetesHPA, namespace, clusterName)
		hpaEntity, _ := ec.PutEntity(KubernetesHPA, hpaId, rattr)
		addEdgesForOwnerRefs(hpaEntity, hpaSummary.BaseObject, clusterName)
		if hpaSummary.Spec != nil && hpaSummary.Spec.Target != nil {
			targetName := hpaSummary.Spec.Target.Name
			targetApiVersion := hpaSummary.Spec.Target.ApiVersion
			targetKind := hpaSummary.Spec.Target.Kind
			var graphTargetKind string
			switch targetApiVersion + "/" + targetKind {
			case "apps/v1/Deployment":
				graphTargetKind = KubernetesDeployment
			case "apps/v1/StatefulSet":
				graphTargetKind = KubernetesStatefulSet
			default:
				graphTargetKind = ""
			}
			if graphTargetKind != "" {
				targetId := ToKubernetesEntityId(targetName, graphTargetKind, namespace, clusterName)
				hpaEntity.AddEdge(targetId, HorizontallyScales)
				hpaEntity.AddEdgeBacklink(targetId, IsHorizontallyScaledBy)
			}
		}

	default:
		slog.Info("Received unknown object", slog.Any("object", po))
	}
}

func addContainerPorts(rattr map[string]string, podSpec *graphpb.PodSpec) {
	if podSpec == nil || podSpec.Containers == nil {
		return
	}
	for _, container := range podSpec.Containers {
		for _, port := range container.Ports {
			if port.ContainerPort != 0 {
				portName := port.Name
				if portName == "" {
					portName = fmt.Sprintf("%d", port.ContainerPort)
				}
				rattr[ContainerPortPrefix+portName] = fmt.Sprintf("%d", port.ContainerPort)
			}
		}
	}
}
