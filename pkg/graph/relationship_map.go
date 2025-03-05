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
	"github.com/cardinalhq/oteltools/pkg/ottl/functions"
	"go.opentelemetry.io/collector/pdata/pcommon"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
)

const (
	BelongsToAccount         = "belongs to account"
	BelongsToCluster         = "belongs to cluster"
	BelongsToNamespace       = "belongs to namespace"
	BelongsToProvider        = "belongs to provider"
	BelongsToRegion          = "belongs to region"
	BelongsToZone            = "belongs to zone"
	ConsumesFrom             = "consumes from"
	ContainsAvailabilityZone = "contains availability zone"
	ContainsNamespace        = "contains namespace"
	ContainsPod              = "contains pod"
	ContainsRegion           = "contains region"
	ContainsService          = "contains service"
	ContainsTask             = "contains task"
	DBQuerySummary           = "db.query.summary"
	DBStatement              = "db.statement"
	HasCollection            = "has a collection"
	HasInstance              = "has instance"
	HasNamespace             = "has a namespace"
	HasNode                  = "has a node"
	HasResourcesInRegion     = "has resources in region"
	HostedOnNode             = "hosted on node"
	HostsCluster             = "hosts cluster"
	HostsPod                 = "hosts pod"
	HostsService             = "hosts service"
	IsACronJobFor            = "is a cronjob for"
	IsAJobFor                = "is a job for"
	IsAPodFor                = "is a pod for"
	IsAPodForService         = "is a pod for service"
	IsAssociatedWith         = "is associated with"
	IsAssociatedWithCluster  = "is associated with cluster"
	IsAssociatedWithNode     = "is associated with node"
	IsAssociatedWithTask     = "is associated with task"
	IsCollectionHostedOn     = "is a collection hosted on"
	IsDaemonSetFor           = "is daemonset for"
	IsDatabaseHostedOn       = "is a database hosted on"
	IsDeployedOnContainer    = "is deployed on container"
	IsDeployedOnNode         = "is deployed on node"
	IsDeployedOnPod          = "is deployed on pod"
	IsInstanceOfFunction     = "is instance of function"
	IsManagedByCronJob       = "is managed by cronjob"
	IsManagedByDaemonSet     = "is managed by daemonset"
	IsManagedByDeployment    = "is managed by deployment"
	IsManagedByJob           = "is managed by job"
	IsManagedByReplicaSet    = "is managed by replicaset"
	IsManagedByStatefulSet   = "is managed by statefulset"
	IsPartOfCluster          = "is part of cluster"
	IsPartOfNamespace        = "is part of namespace"
	IsPartOfDatabase         = "is part of database"
	IsRunningOnNode          = "is running on node"
	IsSpawnedByService       = "is spawned by service"
	IsStatefulSetFor         = "is statefulset for"
	IsUsedByContainer        = "is used by container"
	IsUsedByService          = "is used by service"
	IsConsumedFromByService  = "is consumed from by service"
	ManagesAccount           = "manages account"
	ManagesReplicaset        = "manages replicaset"
	NetPeerName              = "net.peer.name"
	ProducesTo               = "produces to"
	RunsInPod                = "runs in pod"
	RunsOnOperatingSystem    = "runs on operating system"
	UsesDatabase             = "uses database"
	UsesDatabaseCollection   = "uses database collection"
	UsesImage                = "uses image"
	UsesMessagingDestination = "uses messaging destination"
)

const (
	AwsEcsCluster          = "aws.ecs.cluster"
	AwsEcsContainer        = "aws.ecs.container"
	AwsEcsTask             = "aws.ecs.task"
	AwsEksCluster          = "aws.eks.cluster"
	CloudAccount           = "cloud.account"
	CloudAvailabilityZone  = "cloud.availability_zone"
	CloudProvider          = "cloud.provider"
	CloudRegion            = "cloud.region"
	CloudResourceId        = "cloud.resource_id"
	Container              = "container"
	ContainerImage         = "container.image"
	FaasFunction           = "faas.function"
	FaasInstance           = "faas.instance"
	Host                   = "host"
	KubernetesCluster      = "k8s.cluster"
	KubernetesContainer    = "k8s.container"
	KubernetesCronJob      = "k8s.cronjob"
	KubernetesDaemonSet    = "k8s.daemonset"
	KubernetesDeployment   = "k8s.deployment"
	KubernetesJob          = "k8s.job"
	KubernetesNamespace    = "k8s.namespace"
	KubernetesPod          = "k8s.pod"
	KubernetesReplicaSet   = "k8s.replicaset"
	KubernetesStatefulSet  = "k8s.statefulset"
	Node                   = "k8s.node"
	Os                     = "os"
	Process                = "process"
	Service                = "service"
	Database               = "database"
	DatabaseCollection     = "database.collection"
	MessagingDestination   = "messaging.destination"
	MessagingConsumerGroup = "messaging.consumer.group"
	K8SPodIp               = "k8s.pod.ip"
	HostIp                 = "host.ip"
	PodPhase               = "pod.phase"
	PendingReason          = "pending.reason"
)

type EntityInfo struct {
	Type                        string
	Relationships               map[string]string
	AttributeNames              []string
	AttributePrefixes           []string
	NameTransformer             func(string) string
	DeriveRelationshipCallbacks map[string]func(pcommon.Map) string
}

type RelationshipMap map[string]*EntityInfo

var EntityRelationships = RelationshipMap{
	// Cluster
	string(semconv.K8SClusterNameKey): {
		Type: KubernetesCluster,
		Relationships: map[string]string{
			string(semconv.K8SNodeNameKey): HasNode,
		},
		AttributeNames: []string{
			string(semconv.K8SClusterUIDKey),
		},
	},

	// Node
	string(semconv.K8SNodeNameKey): {
		Type: Node,
		Relationships: map[string]string{
			string(semconv.K8SNamespaceNameKey): ContainsNamespace,
			string(semconv.K8SClusterNameKey):   BelongsToCluster,
			string(semconv.OSNameKey):           RunsOnOperatingSystem,
		},
		AttributeNames: []string{
			string(semconv.K8SNodeUIDKey),
		},
	},

	// Namespace
	string(semconv.K8SNamespaceNameKey): {
		Type: KubernetesNamespace,
		Relationships: map[string]string{
			string(semconv.ServiceNameKey): ContainsService,
			string(semconv.K8SNodeNameKey): HostedOnNode,
		},
		AttributeNames: []string{
			string(semconv.K8SNodeNameKey),
			string(semconv.K8SClusterNameKey),
		},
		AttributePrefixes: []string{},
	},

	// Service
	string(semconv.ServiceNameKey): {
		Type: Service,
		Relationships: map[string]string{
			string(semconv.K8SDaemonSetNameKey):           IsManagedByDaemonSet,
			string(semconv.K8SDeploymentNameKey):          IsManagedByDeployment,
			string(semconv.K8SStatefulSetNameKey):         IsManagedByStatefulSet,
			string(semconv.K8SCronJobNameKey):             IsManagedByCronJob,
			string(semconv.K8SJobNameKey):                 IsManagedByJob,
			string(semconv.K8SNamespaceNameKey):           BelongsToNamespace,
			string(semconv.DBNamespaceKey):                UsesDatabase,
			string(semconv.DBCollectionNameKey):           UsesDatabaseCollection,
			string(semconv.MessagingConsumerGroupNameKey): ConsumesFrom,
		},
		AttributeNames: []string{
			string(semconv.ServiceInstanceIDKey),
			string(semconv.ServiceVersionKey),
			string(semconv.ServiceNamespaceKey),
			string(semconv.TelemetrySDKNameKey),
			string(semconv.TelemetrySDKLanguageKey),
			string(semconv.TelemetrySDKVersionKey),
		},
		AttributePrefixes: []string{},
		DeriveRelationshipCallbacks: map[string]func(m pcommon.Map) string{
			string(semconv.MessagingDestinationNameKey): func(m pcommon.Map) string {
				_, mcgVFound := m.Get(string(semconv.MessagingConsumerGroupNameKey))
				if mcgVFound {
					return ConsumesFrom
				}

				monV, monVFound := m.Get(string(semconv.MessagingOperationNameKey))
				if monVFound {
					operation := monV.AsString()
					if operation == "publish" {
						return ProducesTo
					}
					if operation == "process" {
						return ConsumesFrom
					}
				}
				return UsesMessagingDestination
			},
		},
	},

	// DaemonSet
	string(semconv.K8SDaemonSetNameKey): {
		Type: KubernetesDaemonSet,
		Relationships: map[string]string{
			string(semconv.K8SPodNameKey):  ContainsPod,
			string(semconv.ServiceNameKey): IsDaemonSetFor,
		},
		AttributeNames:    []string{string(semconv.K8SStatefulSetUIDKey)},
		AttributePrefixes: []string{},
	},

	// Deployment
	string(semconv.K8SDeploymentNameKey): {
		Type: KubernetesDeployment,
		Relationships: map[string]string{
			string(semconv.K8SReplicaSetNameKey): ManagesReplicaset,
			string(semconv.ServiceNameKey):       IsManagedByDeployment,
		},
		AttributeNames: []string{
			string(semconv.K8SDeploymentUIDKey),
			string(semconv.K8SNamespaceNameKey),
			string(semconv.K8SNodeNameKey),
			string(semconv.ServiceNameKey),
			string(semconv.K8SClusterNameKey),
		},
		AttributePrefixes: []string{},
	},

	// ReplicaSet
	string(semconv.K8SReplicaSetNameKey): {
		Type: KubernetesReplicaSet,
		Relationships: map[string]string{
			string(semconv.K8SPodNameKey):        ContainsPod,
			string(semconv.K8SDeploymentNameKey): IsManagedByDeployment,
		},
		AttributeNames: []string{
			string(semconv.K8SReplicaSetUIDKey),
			string(semconv.K8SDeploymentNameKey),
			string(semconv.K8SNamespaceNameKey),
			string(semconv.ServiceNameKey),
			string(semconv.K8SNodeNameKey),
			string(semconv.K8SClusterNameKey),
		},
		AttributePrefixes: []string{},
	},

	// StatefulSet
	string(semconv.K8SStatefulSetNameKey): {
		Type: KubernetesStatefulSet,
		Relationships: map[string]string{
			string(semconv.K8SPodNameKey):  ContainsPod,
			string(semconv.ServiceNameKey): IsStatefulSetFor,
		},
		AttributeNames: []string{
			string(semconv.K8SStatefulSetUIDKey),
			string(semconv.K8SNamespaceNameKey),
			string(semconv.K8SNodeNameKey),
			string(semconv.ServiceNameKey),
			string(semconv.K8SClusterNameKey),
		},
		AttributePrefixes: []string{},
	},

	// Job
	string(semconv.K8SJobNameKey): {
		Type: KubernetesJob,
		Relationships: map[string]string{
			string(semconv.K8SPodNameKey):  ContainsPod,
			string(semconv.ServiceNameKey): IsAJobFor,
		},
		AttributeNames: []string{
			string(semconv.K8SJobUIDKey),
			string(semconv.K8SNamespaceNameKey),
			string(semconv.K8SNodeNameKey),
			string(semconv.ServiceNameKey),
			string(semconv.K8SClusterNameKey),
		},
		AttributePrefixes: []string{},
	},

	// CronJob
	string(semconv.K8SCronJobNameKey): {
		Type: KubernetesCronJob,
		Relationships: map[string]string{
			string(semconv.K8SPodNameKey):  ContainsPod,
			string(semconv.ServiceNameKey): IsACronJobFor,
		},
		AttributeNames: []string{
			string(semconv.K8SCronJobUIDKey),
			string(semconv.K8SNamespaceNameKey),
			string(semconv.K8SNodeNameKey),
			string(semconv.ServiceNameKey),
			string(semconv.K8SClusterNameKey),
		},
		AttributePrefixes: []string{},
	},

	// Pod
	string(semconv.K8SPodNameKey): {
		Type: KubernetesPod,
		Relationships: map[string]string{
			string(semconv.K8SContainerNameKey):   RunsInPod,
			string(semconv.K8SCronJobNameKey):     IsAPodFor,
			string(semconv.K8SDaemonSetNameKey):   IsAPodFor,
			string(semconv.K8SJobNameKey):         IsAPodFor,
			string(semconv.K8SStatefulSetNameKey): IsAPodFor,
			string(semconv.K8SReplicaSetNameKey):  IsAPodFor,
		},
		AttributeNames: []string{
			K8SPodIp,
			string(semconv.K8SNamespaceNameKey),
			string(semconv.K8SNodeNameKey),
			string(semconv.ServiceNameKey),
			string(semconv.K8SClusterNameKey),
			string(semconv.K8SPodUIDKey),
		},
		AttributePrefixes: []string{"k8s.pod.label.", "k8s.pod.annotation."},
	},

	// Container
	string(semconv.K8SContainerNameKey): {
		Type: KubernetesContainer,
		Relationships: map[string]string{
			string(semconv.K8SPodNameKey):       RunsInPod,
			string(semconv.K8SNamespaceNameKey): IsPartOfNamespace,
			string(semconv.K8SNodeNameKey):      IsDeployedOnNode,
		},
		AttributeNames: []string{
			string(semconv.K8SContainerRestartCountKey),
			string(semconv.K8SContainerStatusLastTerminatedReasonKey),
			string(semconv.K8SNamespaceNameKey),
			string(semconv.K8SNodeNameKey),
			string(semconv.ServiceNameKey),
			string(semconv.K8SClusterNameKey),
		},
		AttributePrefixes: []string{},
	},

	// Docker Container
	string(semconv.ContainerNameKey): {
		Type: Container,
		Relationships: map[string]string{
			string(semconv.ServiceNameKey):        IsDeployedOnContainer,
			string(semconv.ContainerImageNameKey): UsesImage,
		},
		AttributeNames: []string{
			string(semconv.ContainerIDKey),
			string(semconv.ContainerCommandKey),
			string(semconv.ContainerCommandArgsKey),
			string(semconv.ContainerRuntimeKey),
			string(semconv.ContainerCommandLineKey),
		},
		AttributePrefixes: []string{"container.label"},
	},

	// Container Image Entity
	string(semconv.ContainerImageNameKey): {
		Type: ContainerImage,
		Relationships: map[string]string{
			string(semconv.ContainerNameKey): IsUsedByContainer,
		},
		AttributeNames: []string{
			string(semconv.ContainerImageIDKey),
			string(semconv.ContainerImageTagsKey),
			string(semconv.ContainerImageRepoDigestsKey),
		},
		AttributePrefixes: []string{},
	},

	string(semconv.OSNameKey): {
		Type:          Os,
		Relationships: map[string]string{},
		AttributeNames: []string{
			string(semconv.OSVersionKey),
			string(semconv.OSDescriptionKey),
			string(semconv.OSBuildIDKey),
		},
		AttributePrefixes: []string{},
	},

	string(semconv.ProcessCommandKey): {
		Type: Process,
		Relationships: map[string]string{
			string(semconv.ServiceNameKey): IsSpawnedByService,
		},
		AttributeNames: []string{
			string(semconv.ProcessExecutableNameKey),
			string(semconv.ProcessExecutablePathKey),
			string(semconv.ProcessCommandKey),
			string(semconv.ProcessCommandArgsKey),
			string(semconv.ProcessCommandLineKey),
			string(semconv.ProcessOwnerKey),
			string(semconv.ProcessCreationTimeKey),
			string(semconv.ProcessContextSwitchTypeKey),
			string(semconv.ProcessGroupLeaderPIDKey),
			string(semconv.ProcessParentPIDKey),
			string(semconv.ProcessPIDKey),
		},
		AttributePrefixes: []string{},
	},

	// Database (e.g. Mongo, Postgres, Redis)
	string(semconv.DBNamespaceKey): {
		Type: Database,
		Relationships: map[string]string{
			string(semconv.ServiceNameKey):      IsUsedByService,
			string(semconv.DBCollectionNameKey): HasCollection,
		},
		AttributeNames: []string{
			string(semconv.DBSystemKey),
		},
		AttributePrefixes: []string{},
	},

	// Database Collection (e.g. Mongo Collection, Postgres Table)
	string(semconv.DBCollectionNameKey): {
		Type: DatabaseCollection,
		Relationships: map[string]string{
			string(semconv.ServiceNameKey): IsUsedByService,
			string(semconv.DBNamespaceKey): IsPartOfDatabase,
		},
		AttributeNames: []string{
			string(semconv.DBSystemKey),
		},
		AttributePrefixes: []string{},
		NameTransformer: func(s string) string {
			return functions.ScrubWord(s)
		},
	},

	// Messaging System Destination (e.g. Kafka Topic, RabbitMQ Queue)
	string(semconv.MessagingDestinationNameKey): {
		Type: MessagingDestination,
		Relationships: map[string]string{
			string(semconv.ServiceNameKey): IsUsedByService,
		},
		AttributeNames: []string{
			string(semconv.MessagingSystemKey),
		},
		AttributePrefixes: []string{},
	},

	// Messaging Consumer Group (for Kafka)
	string(semconv.MessagingConsumerGroupNameKey): {
		Type: MessagingConsumerGroup,
		Relationships: map[string]string{
			string(semconv.ServiceNameKey): IsConsumedFromByService,
		},
		AttributeNames: []string{
			string(semconv.MessagingSystemKey),
			string(semconv.MessagingDestinationNameKey),
		},
		AttributePrefixes: []string{},
	},

	// ECS Container
	string(semconv.AWSECSContainerARNKey): {
		Type: AwsEcsContainer,
		Relationships: map[string]string{
			string(semconv.AWSECSClusterARNKey): IsPartOfCluster,
			string(semconv.AWSECSTaskARNKey):    IsAssociatedWithTask,
		},
		AttributeNames:    []string{},
		AttributePrefixes: []string{},
	},

	// ECS Task
	string(semconv.AWSECSTaskARNKey): {
		Type: AwsEcsTask,
		Relationships: map[string]string{
			string(semconv.AWSECSClusterARNKey): IsPartOfCluster,
		},
		AttributeNames: []string{
			string(semconv.AWSECSTaskFamilyKey),
			string(semconv.AWSECSTaskRevisionKey),
			string(semconv.AWSECSTaskIDKey),
		},
		AttributePrefixes: []string{},
	},

	// ECS Cluster
	string(semconv.AWSECSClusterARNKey): {
		Type: AwsEcsCluster,
		Relationships: map[string]string{
			string(semconv.AWSECSTaskIDKey): ContainsTask,
		},
		AttributeNames: []string{
			string(semconv.AWSECSLaunchtypeKey),
		},
		AttributePrefixes: []string{},
	},

	// EKS Cluster
	string(semconv.AWSEKSClusterARNKey): {
		Type: AwsEksCluster,
		Relationships: map[string]string{
			string(semconv.K8SClusterNameKey): IsAssociatedWithCluster,
			string(semconv.K8SNodeNameKey):    IsAssociatedWithNode,
		},
		AttributeNames:    []string{},
		AttributePrefixes: []string{},
	},

	// FaaS Instance
	string(semconv.FaaSInstanceKey): {
		Type: FaasInstance,
		Relationships: map[string]string{
			string(semconv.FaaSNameKey): IsInstanceOfFunction,
		},
		AttributeNames:    []string{},
		AttributePrefixes: []string{},
	},

	// FaaS Function
	string(semconv.FaaSNameKey): {
		Type: FaasFunction,
		Relationships: map[string]string{
			string(semconv.FaaSInstanceKey): HasInstance,
		},
		AttributeNames: []string{
			string(semconv.FaaSColdstartKey),
			string(semconv.FaaSCronKey),
			string(semconv.FaaSDocumentCollectionKey),
			string(semconv.FaaSDocumentOperationKey),
			string(semconv.FaaSDocumentNameKey),
			string(semconv.FaaSDocumentTimeKey),
			string(semconv.FaaSInvokedRegionKey),
			string(semconv.FaaSMaxMemoryKey),
			string(semconv.FaaSInvokedNameKey),
		},
		AttributePrefixes: []string{},
	},

	// Cloud Provider (e.g., AWS, GCP, Azure)
	string(semconv.CloudProviderKey): {
		Type: CloudProvider,
		Relationships: map[string]string{
			string(semconv.CloudAccountIDKey):        ManagesAccount,
			string(semconv.CloudRegionKey):           ContainsRegion,
			string(semconv.CloudAvailabilityZoneKey): ContainsAvailabilityZone,
		},
		AttributeNames:    []string{},
		AttributePrefixes: []string{},
	},

	// Cloud Account
	string(semconv.CloudAccountIDKey): {
		Type: CloudAccount,
		Relationships: map[string]string{
			string(semconv.CloudProviderKey):         BelongsToProvider,
			string(semconv.CloudRegionKey):           HasResourcesInRegion,
			string(semconv.CloudAvailabilityZoneKey): ContainsAvailabilityZone,
		},
		AttributeNames:    []string{},
		AttributePrefixes: []string{},
	},

	// Cloud Region
	string(semconv.CloudRegionKey): {
		Type: CloudRegion,
		Relationships: map[string]string{
			string(semconv.CloudProviderKey):         BelongsToProvider,
			string(semconv.CloudAvailabilityZoneKey): ContainsAvailabilityZone,
			string(semconv.CloudAccountIDKey):        BelongsToAccount,
		},
		AttributeNames:    []string{},
		AttributePrefixes: []string{},
	},

	// Cloud Availability Zone
	string(semconv.CloudAvailabilityZoneKey): {
		Type: CloudAvailabilityZone,
		Relationships: map[string]string{
			string(semconv.CloudRegionKey):    BelongsToRegion,
			string(semconv.CloudAccountIDKey): BelongsToAccount,
			string(semconv.CloudProviderKey):  BelongsToProvider,
		},
		AttributeNames:    []string{},
		AttributePrefixes: []string{},
	},

	// Cloud Availability Zone
	string(semconv.CloudResourceIDKey): {
		Type: CloudResourceId,
		Relationships: map[string]string{
			string(semconv.CloudRegionKey):           BelongsToRegion,
			string(semconv.CloudAvailabilityZoneKey): BelongsToZone,
			string(semconv.CloudAccountIDKey):        BelongsToAccount,
			string(semconv.CloudProviderKey):         BelongsToProvider,
		},
		AttributeNames:    []string{},
		AttributePrefixes: []string{},
	},

	string(semconv.HostNameKey): {
		Type: Host,
		Relationships: map[string]string{
			string(semconv.ServiceNameKey):        HostsService,
			string(semconv.K8SClusterNameKey):     HostsCluster,
			string(semconv.K8SPodNameKey):         HostsPod,
			string(semconv.K8SStatefulSetNameKey): HostsPod,
			string(semconv.K8SDaemonSetNameKey):   HostsPod,
		},
		AttributeNames: []string{
			string(semconv.HostIDKey),
			string(semconv.HostTypeKey),
			string(semconv.HostImageIDKey),
			string(semconv.HostImageNameKey),
			string(semconv.HostImageVersionKey),
			string(semconv.HostCPUFamilyKey),
			string(semconv.HostCPUSteppingKey),
			string(semconv.HostCPUCacheL2SizeKey),
			string(semconv.HostCPUVendorIDKey),
			string(semconv.HostArchKey),
			string(semconv.HostIPKey),
		},
		AttributePrefixes: []string{},
	},
}

var RestrictedServices = map[string]bool{
	"kubernetes-nodes":          true,
	"kubernetes-nodes-cadvisor": true,
}

var RestrictedNamespaces = map[string]bool{
	"datadog-agent": true,
}
