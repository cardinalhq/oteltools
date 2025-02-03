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

import semconv "go.opentelemetry.io/otel/semconv/v1.27.0"

const (
	IsAPodForService         = "is a pod for service"
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
	ManagesStatefulSets      = "manages stateful sets"
	ManagesJobs              = "manages jobs"
	ManagesCronJobs          = "manages cron jobs"
	BelongsToCluster         = "belongs to cluster"
	ContainsService          = "contains service"
	SchedulesPod             = "schedules pod"
	RunsOnOperatingSystem    = "runs on operating system"
	ContainsPod              = "contains pod"
	ContainsDeployment       = "contains deployment"
	ContainsStatefulSet      = "contains statefulset"
	ContainsReplicaSet       = "contains replicaset"
	ContainsJob              = "contains job"
	ContainsCronJob          = "contains cronjob"
	IsPartOfDeployment       = "is part of deployment"
	IsPartOfStatefulSet      = "is part of statefulset"
	IsScheduledOnNode        = "is scheduled on node"
	RunsInPod                = "runs in pod"
	IsPartOfNamespace        = "is part of namespace"
	IsDeployedOnNode         = "is deployed on node"
	IsManagedByCluster       = "is managed by cluster"
	IsDeployedOnContainer    = "is deployed on container"
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
	IsDatabaseHostedOn       = "is a database hosted on"
	IsCollectionHostedOn     = "is a collection hosted on"
	ConsumesFrom             = "consumes from"
	ProducesTo               = "produces to"
	IsSpawnedByService       = "is spawned by service"
)

const (
	Service               = "service"
	KubernetesPod         = "k8s.pod"
	KubernetesContainer   = "k8s.container"
	KubernetesCluster     = "k8s.cluster"
	Node                  = "k8s.node"
	KubernetesNamespace   = "k8s.namespace"
	KubernetesReplicaSet  = "k8s.replicaset"
	KubernetesDeployment  = "k8s.deployment"
	KubernetesStatefulSet = "k8s.statefulset"
	KubernetesJob         = "k8s.job"
	KubernetesCronJob     = "k8s.cronjob"
	Container             = "container"
	ContainerImage        = "container.image"
	Os                    = "os"
	Process               = "process"
	AwsEcsContainer       = "aws.ecs.container"
	AwsEcsTask            = "aws.ecs.task"
	AwsEcsCluster         = "aws.ecs.cluster"
	AwsEksCluster         = "aws.eks.cluster"
	FaasInstance          = "faas.instance"
	FaasFunction          = "faas.function"
	CloudProvider         = "cloud.provider"
	CloudAccount          = "cloud.account"
	CloudRegion           = "cloud.region"
	CloudAvailabilityZone = "cloud.availability_zone"
	CloudResourceId       = "cloud.resource_id"
	Host                  = "host"
)

type EntityInfo struct {
	Type              string
	Relationships     map[string]string
	AttributeNames    []string
	AttributePrefixes []string
}

type RelationshipMap map[string]*EntityInfo

var EntityRelationships = RelationshipMap{
	// Service
	string(semconv.ServiceNameKey): {
		Type: Service,
		Relationships: map[string]string{
			string(semconv.K8SNamespaceNameKey):           BelongsToNamespace,
			string(semconv.K8SClusterNameKey):             IsPartOfCluster,
			string(semconv.K8SPodNameKey):                 IsDeployedOnPod,
			string(semconv.K8SNodeNameKey):                IsRunningOnNode,
			string(semconv.K8SDeploymentNameKey):          IsManagedByDeployment,
			string(semconv.K8SStatefulSetNameKey):         IsManagedByStatefulSet,
			string(semconv.K8SDaemonSetNameKey):           IsManagedByDeployment,
			string(semconv.K8SReplicaSetNameKey):          IsManagedByReplicaSet,
			string(semconv.DBSystemKey):                   UsesDatabase,
			string(semconv.MessagingConsumerGroupNameKey): ConsumesFrom,
			string(semconv.MessagingDestinationNameKey):   ProducesTo,
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
	},

	// Cluster
	string(semconv.K8SClusterNameKey): {
		Type: KubernetesCluster,
		Relationships: map[string]string{
			string(semconv.ServiceNameKey):        ContainsService,
			string(semconv.K8SNodeNameKey):        HasNode,
			string(semconv.K8SNamespaceNameKey):   HasNamespace,
			string(semconv.K8SDeploymentNameKey):  ManagesDeployments,
			string(semconv.K8SStatefulSetNameKey): ManagesStatefulSets,
			string(semconv.K8SJobNameKey):         ManagesJobs,
			string(semconv.K8SCronJobNameKey):     ManagesCronJobs,
		},
		AttributeNames: []string{
			string(semconv.K8SClusterUIDKey),
		},
	},

	// Node
	string(semconv.K8SNodeNameKey): {
		Type: Node,
		Relationships: map[string]string{
			string(semconv.ServiceNameKey):    ContainsService,
			string(semconv.K8SClusterNameKey): BelongsToCluster,
			string(semconv.K8SPodNameKey):     SchedulesPod,
			string(semconv.OSNameKey):         RunsOnOperatingSystem,
		},
		AttributeNames: []string{
			string(semconv.K8SNodeUIDKey),
		},
	},

	// Namespace
	string(semconv.K8SNamespaceNameKey): {
		Type: KubernetesNamespace,
		Relationships: map[string]string{
			string(semconv.ServiceNameKey):        ContainsService,
			string(semconv.K8SClusterNameKey):     BelongsToCluster,
			string(semconv.K8SPodNameKey):         ContainsPod,
			string(semconv.K8SDeploymentNameKey):  ContainsDeployment,
			string(semconv.K8SStatefulSetNameKey): ContainsStatefulSet,
			string(semconv.K8SReplicaSetNameKey):  ContainsReplicaSet,
			string(semconv.K8SJobNameKey):         ContainsJob,
			string(semconv.K8SCronJobNameKey):     ContainsCronJob,
		},
		AttributeNames:    []string{},
		AttributePrefixes: []string{},
	},

	// Pod
	string(semconv.K8SPodNameKey): {
		Type: KubernetesPod,
		Relationships: map[string]string{
			string(semconv.ServiceNameKey):        IsAPodForService,
			string(semconv.K8SNamespaceNameKey):   BelongsToNamespace,
			string(semconv.K8SNodeNameKey):        IsScheduledOnNode,
			string(semconv.K8SClusterNameKey):     IsPartOfCluster,
			string(semconv.K8SReplicaSetNameKey):  IsManagedByReplicaSet,
			string(semconv.K8SDeploymentNameKey):  IsPartOfDeployment,
			string(semconv.K8SStatefulSetNameKey): IsPartOfStatefulSet,
		},
		AttributeNames: []string{
			"k8s.pod.ip",
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
		},
		AttributePrefixes: []string{},
	},

	// ReplicaSet
	string(semconv.K8SReplicaSetNameKey): {
		Type: KubernetesReplicaSet,
		Relationships: map[string]string{
			string(semconv.K8SNamespaceNameKey):  BelongsToNamespace,
			string(semconv.K8SDeploymentNameKey): IsManagedByDeployment,
			string(semconv.K8SClusterNameKey):    IsPartOfCluster,
		},
		AttributeNames:    []string{string(semconv.K8SReplicaSetUIDKey)},
		AttributePrefixes: []string{},
	},

	// Deployment
	string(semconv.K8SDeploymentNameKey): {
		Type: KubernetesDeployment,
		Relationships: map[string]string{
			string(semconv.K8SNamespaceNameKey):  BelongsToNamespace,
			string(semconv.K8SClusterNameKey):    IsManagedByCluster,
			string(semconv.K8SReplicaSetNameKey): ManagesReplicaset,
		},
		AttributeNames:    []string{string(semconv.K8SDeploymentUIDKey)},
		AttributePrefixes: []string{},
	},

	// StatefulSet
	string(semconv.K8SStatefulSetNameKey): {
		Type: KubernetesStatefulSet,
		Relationships: map[string]string{
			string(semconv.K8SNamespaceNameKey): BelongsToNamespace,
			string(semconv.K8SClusterNameKey):   IsManagedByCluster,
		},
		AttributeNames:    []string{string(semconv.K8SStatefulSetUIDKey)},
		AttributePrefixes: []string{},
	},

	// Job
	string(semconv.K8SJobNameKey): {
		Type: KubernetesJob,
		Relationships: map[string]string{
			string(semconv.K8SNamespaceNameKey): BelongsToNamespace,
			string(semconv.K8SClusterNameKey):   IsManagedByCluster,
		},
		AttributeNames:    []string{string(semconv.K8SJobUIDKey)},
		AttributePrefixes: []string{},
	},

	// CronJob
	string(semconv.K8SCronJobNameKey): {
		Type: KubernetesCronJob,
		Relationships: map[string]string{
			string(semconv.K8SNamespaceNameKey): BelongsToNamespace,
			string(semconv.K8SClusterNameKey):   IsManagedByCluster,
		},
		AttributeNames:    []string{string(semconv.K8SCronJobUIDKey)},
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
