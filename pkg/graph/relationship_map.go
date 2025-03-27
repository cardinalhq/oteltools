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
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"

	"github.com/cardinalhq/oteltools/pkg/ottl/functions"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

const (
	BelongsToAccount         = "belongs to account"
	BelongsToCluster         = "belongs to cluster"
	BelongsToNamespace       = "belongs to namespace"
	BelongsToProvider        = "belongs to provider"
	BelongsToRegion          = "belongs to region"
	BelongsToZone            = "belongs to zone"
	CallsEndpoint            = "calls endpoint"
	ConsumesFrom             = "consumes from"
	ExecutesQuery            = "executes query"
	ContainsAvailabilityZone = "contains availability zone"
	ContainsContainer        = "contains container"
	ContainsNamespace        = "contains namespace"
	ContainsPod              = "contains pod"
	ContainsRegion           = "contains region"
	ContainsService          = "contains service"
	ContainsTask             = "contains task"
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
	IsCalledByService        = "is called by service"
	IsCollectionHostedOn     = "is a collection hosted on"
	IsConsumedFromByService  = "is consumed from by service"
	IsDaemonSetFor           = "is daemonset for"
	IsDatabaseHostedOn       = "is a database hosted on"
	IsDeployedOnCluster      = "is deployed on cluster"
	IsDeployedOnContainer    = "is deployed on container"
	IsDeployedOnNode         = "is deployed on node"
	IsDeployedOnPod          = "is deployed on pod"
	IsHostedOnNode           = "is hosted on node"
	IsInstanceOfFunction     = "is instance of function"
	IsManagedByCronJob       = "is managed by cronjob"
	IsManagedByDaemonSet     = "is managed by daemonset"
	IsManagedByDeployment    = "is managed by deployment"
	IsManagedByJob           = "is managed by job"
	IsManagedByReplicaSet    = "is managed by replicaset"
	IsManagedByStatefulSet   = "is managed by statefulset"
	IsPartOfCluster          = "is part of cluster"
	IsPartOfDatabase         = "is part of database"
	IsPartOfNamespace        = "is part of namespace"
	IsRunningOnNode          = "is running on node"
	IsServedByService        = "is served by service"
	IsSpawnedByService       = "is spawned by service"
	IsStatefulSetFor         = "is statefulset for"
	IsUsedByContainer        = "is used by container"
	IsUsedByCronJob          = "is used by cronjob"
	IsUsedByDaemonSet        = "is used by daemonset"
	IsUsedByDeployment       = "is used by deployment"
	IsUsedByJob              = "is used by job"
	IsUsedByPod              = "is used by pod"
	IsUsedByReplicaSet       = "is used by replicaset"
	IsUsedByService          = "is used by service"
	IsUsedByStatefulSet      = "is used by statefulset"
	ManagesAccount           = "manages account"
	ManagesReplicaset        = "manages replicaset"
	ManagesPod               = "manages pod"
	ManagesJob               = "manages job"
	ProducesTo               = "produces to"
	RunsInPod                = "runs in pod"
	RunsOnOperatingSystem    = "runs on operating system"
	RunsProcess              = "runs process"
	ServesEndpoint           = "serves endpoint"
	SettlesMessagesTo        = "settles messages to"
	UsesConfigMap            = "uses configmap"
	UsesDatabase             = "uses database"
	UsesDatabaseCollection   = "uses database collection"
	UsesContainerImage       = "uses container image"
	UsesMessagingDestination = "uses messaging destination"
	UsesSecret               = "uses secret"
	DataHashPrefix           = "data.hash."
	Replicas                 = "replicas"
	ReadyReplicas            = "ready.replicas"
	CurrentReplicas          = "current.replicas"
	UpdatedReplicas          = "updated.replicas"
	AvailableReplicas        = "available.replicas"
	UnavailableReplicas      = "unavailable.replicas"
	Deployment               = "Deployment"
	StatefulSet              = "StatefulSet"
	DaemonSet                = "DaemonSet"
	ReplicaSet               = "ReplicaSet"
	Job                      = "Job"
	CronJob                  = "CronJob"
)

const (
	DBQuerySummary           = "db.query.summary"
	DBStatement              = "db.statement"
	SpanKindString           = "span.kind.string"
	NetPeerName              = "net.peer.name"
	AwsEcsCluster            = "aws.ecs.cluster"
	AwsEcsContainer          = "aws.ecs.container"
	AwsEcsTask               = "aws.ecs.task"
	AwsEksCluster            = "aws.eks.cluster"
	CloudAccount             = "cloud.account"
	CloudAvailabilityZone    = "cloud.availability_zone"
	CloudProvider            = "cloud.provider"
	CloudRegion              = "cloud.region"
	CloudResourceId          = "cloud.resource_id"
	Container                = "container"
	ContainerImage           = "container.image"
	ContainerImageName       = string(semconv.ContainerImageNameKey)
	ContainerImageID         = string(semconv.ContainerImageIDKey)
	Database                 = "database"
	DatabaseCollection       = "database.collection"
	Endpoint                 = "endpoint"
	FaasFunction             = "faas.function"
	FaasInstance             = "faas.instance"
	Host                     = "host"
	HostIp                   = "host.ip"
	K8SPodIp                 = "k8s.pod.ip"
	KubernetesCluster        = "k8s.cluster"
	KubernetesConfigMap      = "k8s.configmap"
	KubernetesContainer      = "k8s.container"
	KubernetesCronJob        = "k8s.cronjob"
	KubernetesDaemonSet      = "k8s.daemonset"
	KubernetesDeployment     = "k8s.deployment"
	KubernetesJob            = "k8s.job"
	KubernetesNamespace      = "k8s.namespace"
	KubernetesPod            = "k8s.pod"
	KubernetesReplicaSet     = "k8s.replicaset"
	KubernetesSecret         = "k8s.secret"
	KubernetesStatefulSet    = "k8s.statefulset"
	MessagingConsumerGroup   = "messaging.consumer.group"
	MessagingDestination     = "messaging.destination"
	Node                     = "k8s.node"
	OperatingSystem          = "os"
	PendingReason            = "pending.reason"
	PodPhase                 = "pod.phase"
	Process                  = "process"
	Service                  = "service"
	ContainerImageNamePrefix = "container.image.name."
	ContainerImageIDPrefix   = "container.image.id."
	CrashLoopBackOff         = "CrashLoopBackOff"
	ImagePullBackOff         = "ImagePullBackOff"
	OOMKilled                = "OOMKilled"
)

type EntityInfo struct {
	Type                        string
	Relationships               map[string]string
	AttributeNames              []string
	AttributePrefixes           []string
	NameTransformer             func(string) string
	DeriveRelationshipCallbacks map[string]func(pcommon.Map) string
	ShouldCreateCallBack        func(p pcommon.Map) bool
	OtherIDAttributes           []string
}

type RelationshipMap map[string]*EntityInfo

var messagingPublishOperations = map[string]struct{}{
	"publish": {},
	"create":  {},
	"send":    {},
}

var messagingConsumeOperations = map[string]struct{}{
	"consume": {},
	"receive": {},
	"process": {},
}

var EntityRelationships = RelationshipMap{
	// Cluster
	string(semconv.K8SClusterNameKey): {
		Type: KubernetesCluster,
		Relationships: map[string]string{
			string(semconv.K8SNodeNameKey):      HasNode,
			string(semconv.K8SNamespaceNameKey): ContainsNamespace,
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
		OtherIDAttributes: []string{
			string(semconv.K8SClusterNameKey),
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
		OtherIDAttributes: []string{
			string(semconv.K8SNamespaceNameKey),
			string(semconv.K8SClusterNameKey),
		},
	},

	// HTTP endpoints
	string(semconv.URLTemplateKey): {
		Type: Endpoint,
		DeriveRelationshipCallbacks: map[string]func(pcommon.Map) string{
			string(semconv.ServiceNameKey): func(m pcommon.Map) string {
				spanKindCode, spanKindCodeFound := m.Get(SpanKindString)
				if spanKindCodeFound {
					spanKind := spanKindCode.AsString()
					if spanKind == ptrace.SpanKindServer.String() {
						return IsServedByService
					} else if spanKind == ptrace.SpanKindClient.String() {
						return IsCalledByService
					}
				}
				return ""
			},
		},
		ShouldCreateCallBack: func(m pcommon.Map) bool {
			httpStatusCode, httpStatusCodeFound := m.Get(string(semconv.HTTPResponseStatusCodeKey))
			return httpStatusCodeFound && httpStatusCode.Int() != 404
		},
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
			string(semconv.MessagingConsumerGroupNameKey): ConsumesFrom,
			string(semconv.K8SNodeNameKey):                IsHostedOnNode,
			string(semconv.K8SClusterNameKey):             IsDeployedOnCluster,
			string(semconv.ProcessCommandKey):             RunsProcess,
		},
		AttributeNames: []string{
			string(semconv.ServiceInstanceIDKey),
			string(semconv.ServiceVersionKey),
			string(semconv.ServiceNamespaceKey),
			string(semconv.TelemetrySDKNameKey),
			string(semconv.TelemetrySDKLanguageKey),
			string(semconv.TelemetrySDKVersionKey),
			string(semconv.ContainerImageNameKey),
		},
		AttributePrefixes: []string{},
		DeriveRelationshipCallbacks: map[string]func(m pcommon.Map) string{
			string(semconv.DBCollectionNameKey): func(m pcommon.Map) string {
				dbQuery, dbQueryFound := m.Get(string(semconv.DBQuerySummaryKey))
				if dbQueryFound {
					return fmt.Sprintf("%s %s", ExecutesQuery, dbQuery.AsString())
				}
				return UsesDatabaseCollection
			},
			string(semconv.MessagingDestinationNameKey): func(m pcommon.Map) string {
				_, mcgVFound := m.Get(string(semconv.MessagingConsumerGroupNameKey))
				if mcgVFound {
					return ConsumesFrom
				}

				monV, monVFound := m.Get(string(semconv.MessagingOperationTypeKey))
				if monVFound {
					operation := monV.AsString()
					if _, ok := messagingPublishOperations[operation]; ok {
						return ProducesTo
					}
					if _, ok := messagingConsumeOperations[operation]; ok {
						return ConsumesFrom
					}
					if operation == "settle" {
						return SettlesMessagesTo
					}
				}
				return UsesMessagingDestination
			},
			string(semconv.URLTemplateKey): func(m pcommon.Map) string {
				spanKindCode, spanKindCodeFound := m.Get(SpanKindString)
				if spanKindCodeFound {
					spanKind := spanKindCode.AsString()
					if spanKind == ptrace.SpanKindServer.String() {
						return ServesEndpoint
					} else if spanKind == ptrace.SpanKindClient.String() {
						return CallsEndpoint
					}
				}
				return ""
			},
		},
		OtherIDAttributes: []string{
			string(semconv.K8SClusterNameKey),
			string(semconv.K8SNamespaceNameKey),
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
		OtherIDAttributes: []string{
			string(semconv.K8SClusterNameKey),
			string(semconv.K8SNamespaceNameKey),
		},
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
		OtherIDAttributes: []string{
			string(semconv.K8SClusterNameKey),
			string(semconv.K8SNamespaceNameKey),
		},
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
		OtherIDAttributes: []string{
			string(semconv.K8SClusterNameKey),
			string(semconv.K8SNamespaceNameKey),
		},
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
		OtherIDAttributes: []string{
			string(semconv.K8SClusterNameKey),
			string(semconv.K8SNamespaceNameKey),
		},
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
		OtherIDAttributes: []string{
			string(semconv.K8SClusterNameKey),
			string(semconv.K8SNamespaceNameKey),
		},
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
		OtherIDAttributes: []string{
			string(semconv.K8SClusterNameKey),
			string(semconv.K8SNamespaceNameKey),
		},
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
		OtherIDAttributes: []string{
			string(semconv.K8SClusterNameKey),
			string(semconv.K8SNamespaceNameKey),
		},
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
		OtherIDAttributes: []string{
			string(semconv.K8SClusterNameKey),
			string(semconv.K8SNamespaceNameKey),
		},
	},

	// Docker Container
	string(semconv.ContainerNameKey): {
		Type: Container,
		Relationships: map[string]string{
			string(semconv.ServiceNameKey): IsDeployedOnContainer,
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

	string(semconv.OSNameKey): {
		Type:          OperatingSystem,
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
			string(semconv.DBSystemNameKey),
		},
		AttributePrefixes: []string{},
		OtherIDAttributes: []string{
			string(semconv.ServerAddressKey),
		},
	},

	// Database Collection (e.g. Mongo Collection, Postgres Table)
	string(semconv.DBCollectionNameKey): {
		Type: DatabaseCollection,
		Relationships: map[string]string{
			string(semconv.ServiceNameKey): IsUsedByService,
			string(semconv.DBNamespaceKey): IsPartOfDatabase,
		},
		AttributeNames: []string{
			string(semconv.DBSystemNameKey),
		},
		AttributePrefixes: []string{},
		NameTransformer: func(s string) string {
			return functions.ScrubWord(s)
		},
		OtherIDAttributes: []string{
			string(semconv.DBNamespaceKey),
			string(semconv.ServerAddressKey),
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
