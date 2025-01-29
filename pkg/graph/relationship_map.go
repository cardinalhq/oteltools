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

type EntityInfo struct {
	Type              string
	Relationships     map[string]string
	AttributePrefixes []string
}

type RelationshipMap map[string]*EntityInfo

var Relationships = RelationshipMap{
	// Service
	string(semconv.ServiceNameKey): {
		Type: "service",
		Relationships: map[string]string{
			string(semconv.K8SNamespaceNameKey):   BelongsToNamespace,
			string(semconv.K8SClusterNameKey):     IsPartOfCluster,
			string(semconv.K8SPodNameKey):         IsDeployedOnPod,
			string(semconv.K8SNodeNameKey):        IsRunningOnNode,
			string(semconv.K8SDeploymentNameKey):  IsManagedByDeployment,
			string(semconv.K8SStatefulSetNameKey): IsManagedByStatefulSet,
			string(semconv.K8SDaemonSetNameKey):   IsManagedByDeployment,
			string(semconv.K8SReplicaSetNameKey):  IsManagedByReplicaSet,
		},
		AttributePrefixes: []string{"service."},
	},

	// Cluster
	string(semconv.K8SClusterNameKey): {
		Type: "k8s.cluster",
		Relationships: map[string]string{
			string(semconv.K8SNodeNameKey):        HasNode,
			string(semconv.K8SNamespaceNameKey):   HasNamespace,
			string(semconv.K8SDeploymentNameKey):  ManagesDeployments,
			string(semconv.K8SDaemonSetNameKey):   ManagesDaemonSets,
			string(semconv.K8SStatefulSetNameKey): ManagesStatefulSets,
			string(semconv.K8SJobNameKey):         ManagesJobs,
			string(semconv.K8SCronJobNameKey):     ManagesCronJobs,
		},
		AttributePrefixes: []string{"k8s.cluster."},
	},

	// Node
	string(semconv.K8SNodeNameKey): {
		Type: "k8s.node",
		Relationships: map[string]string{
			string(semconv.K8SClusterNameKey): BelongsToCluster,
			string(semconv.K8SPodNameKey):     SchedulesPod,
			string(semconv.OSNameKey):         RunsOnOperatingSystem,
		},
		AttributePrefixes: []string{"k8s.node."},
	},

	// Namespace
	string(semconv.K8SNamespaceNameKey): {
		Type: "k8s.namespace",
		Relationships: map[string]string{
			string(semconv.K8SClusterNameKey):     BelongsToCluster,
			string(semconv.K8SPodNameKey):         ContainsPod,
			string(semconv.K8SDeploymentNameKey):  ContainsDeployment,
			string(semconv.K8SStatefulSetNameKey): ContainsStatefulSet,
			string(semconv.K8SDaemonSetNameKey):   ContainsDaemonSet,
			string(semconv.K8SReplicaSetNameKey):  ContainsReplicaSet,
			string(semconv.K8SJobNameKey):         ContainsJob,
			string(semconv.K8SCronJobNameKey):     ContainsCronJob,
		},
		AttributePrefixes: []string{"k8s.namespace."},
	},

	// Pod
	string(semconv.K8SPodNameKey): {
		Type: "k8s.pod",
		Relationships: map[string]string{
			string(semconv.K8SNamespaceNameKey):   BelongsToNamespace,
			string(semconv.K8SNodeNameKey):        IsScheduledOnNode,
			string(semconv.K8SClusterNameKey):     IsPartOfCluster,
			string(semconv.K8SReplicaSetNameKey):  IsManagedByReplicaSet,
			string(semconv.K8SDeploymentNameKey):  IsPartOfDeployment,
			string(semconv.K8SStatefulSetNameKey): IsPartOfStatefulSet,
			string(semconv.K8SDaemonSetNameKey):   IsPartOfDaemonSet,
		},
		AttributePrefixes: []string{"k8s.pod."},
	},

	// Container
	string(semconv.K8SContainerNameKey): {
		Type: "k8s.container",
		Relationships: map[string]string{
			string(semconv.K8SPodNameKey):       RunsInPod,
			string(semconv.K8SNamespaceNameKey): IsPartOfNamespace,
			string(semconv.K8SNodeNameKey):      IsDeployedOnNode,
		},
		AttributePrefixes: []string{"k8s.container."},
	},

	// ReplicaSet
	string(semconv.K8SReplicaSetNameKey): {
		Type: "k8s.replicaset",
		Relationships: map[string]string{
			string(semconv.K8SNamespaceNameKey):  BelongsToNamespace,
			string(semconv.K8SDeploymentNameKey): IsManagedByDeployment,
			string(semconv.K8SClusterNameKey):    IsPartOfCluster,
		},
		AttributePrefixes: []string{"k8s.replicaset."},
	},

	// Deployment
	string(semconv.K8SDeploymentNameKey): {
		Type: "k8s.deployment",
		Relationships: map[string]string{
			string(semconv.K8SNamespaceNameKey):  BelongsToNamespace,
			string(semconv.K8SClusterNameKey):    IsManagedByCluster,
			string(semconv.K8SReplicaSetNameKey): ManagesReplicaset,
		},
		AttributePrefixes: []string{"k8s.deployment."},
	},

	// DaemonSet
	string(semconv.K8SDaemonSetNameKey): {
		Type: "k8s.daemonset",
		Relationships: map[string]string{
			string(semconv.K8SNamespaceNameKey): BelongsToNamespace,
			string(semconv.K8SClusterNameKey):   IsManagedByCluster,
		},
		AttributePrefixes: []string{"k8s.daemonset."},
	},

	// StatefulSet
	string(semconv.K8SStatefulSetNameKey): {
		Type: "k8s.statefulset",
		Relationships: map[string]string{
			string(semconv.K8SNamespaceNameKey): BelongsToNamespace,
			string(semconv.K8SClusterNameKey):   IsManagedByCluster,
		},
		AttributePrefixes: []string{"k8s.statefulset."},
	},

	// Job
	string(semconv.K8SJobNameKey): {
		Type: "k8s.job",
		Relationships: map[string]string{
			string(semconv.K8SNamespaceNameKey): BelongsToNamespace,
			string(semconv.K8SClusterNameKey):   IsManagedByCluster,
		},
		AttributePrefixes: []string{"k8s.job."},
	},

	// CronJob
	string(semconv.K8SCronJobNameKey): {
		Type: "k8s.cronjob",
		Relationships: map[string]string{
			string(semconv.K8SNamespaceNameKey): BelongsToNamespace,
			string(semconv.K8SClusterNameKey):   IsManagedByCluster,
		},
		AttributePrefixes: []string{"k8s.cronjob."},
	},

	// Container Entity
	string(semconv.ContainerNameKey): {
		Type: "container",
		Relationships: map[string]string{
			string(semconv.ContainerImageNameKey): UsesImage,
		},
		AttributePrefixes: []string{
			"container.command",
			"container.",
			"oci.",
		},
	},

	// Container Image Entity
	string(semconv.ContainerImageNameKey): {
		Type: "container.image",
		Relationships: map[string]string{
			string(semconv.ContainerNameKey): IsUsedByContainer,
		},
		AttributePrefixes: []string{
			"container.image",
		},
	},
	string(semconv.OSNameKey): {
		Type:              "os",
		Relationships:     map[string]string{},
		AttributePrefixes: []string{"os."},
	},
	string(semconv.ProcessCommandKey): {
		Type:              "process",
		Relationships:     map[string]string{},
		AttributePrefixes: []string{"process."},
	},
	// ECS Container
	string(semconv.AWSECSContainerARNKey): {
		Type: "aws.ecs.container",
		Relationships: map[string]string{
			string(semconv.AWSECSClusterARNKey): IsPartOfCluster,
			string(semconv.AWSECSTaskIDKey):     IsAssociatedWithTask,
		},
		AttributePrefixes: []string{
			"aws.ecs.task.",
		},
	},

	// ECS Task
	string(semconv.AWSECSTaskIDKey): {
		Type: "aws.ecs.task",
		Relationships: map[string]string{
			string(semconv.AWSECSClusterARNKey): IsPartOfCluster,
		},
		AttributePrefixes: []string{
			"aws.ecs.task.",
		},
	},

	// ECS Cluster
	string(semconv.AWSECSClusterARNKey): {
		Type: "aws.ecs.cluster",
		Relationships: map[string]string{
			string(semconv.AWSECSTaskIDKey): ContainsTask,
		},
		AttributePrefixes: []string{
			"aws.ecs.launchtype.",
		},
	},
	// EKS Cluster
	string(semconv.AWSEKSClusterARNKey): {
		Type: "aws.eks.cluster",
		Relationships: map[string]string{
			string(semconv.K8SClusterNameKey): IsAssociatedWithCluster,
			string(semconv.K8SNodeNameKey):    IsAssociatedWithNode,
		},
		AttributePrefixes: []string{
			"aws.eks.",
		},
	},
	// FaaS Instance
	string(semconv.FaaSInstanceKey): {
		Type: "faas.instance",
		Relationships: map[string]string{
			string(semconv.FaaSNameKey): IsInstanceOfFunction,
		},
		AttributePrefixes: []string{"faas.instance."},
	},

	// FaaS Function
	string(semconv.FaaSNameKey): {
		Type: "faas.function",
		Relationships: map[string]string{
			string(semconv.FaaSInstanceKey): HasInstance,
		},
		AttributePrefixes: []string{"faas."},
	},
	// Cloud Provider (e.g., AWS, GCP, Azure)
	string(semconv.CloudProviderKey): {
		Type: "cloud.provider",
		Relationships: map[string]string{
			string(semconv.CloudAccountIDKey):        ManagesAccount,
			string(semconv.CloudRegionKey):           ContainsRegion,
			string(semconv.CloudAvailabilityZoneKey): ContainsAvailabilityZone,
		},
		AttributePrefixes: []string{"cloud."},
	},

	// Cloud Account
	string(semconv.CloudAccountIDKey): {
		Type: "cloud.account",
		Relationships: map[string]string{
			string(semconv.CloudProviderKey): BelongsToProvider,
			string(semconv.CloudRegionKey):   HasResourcesInRegion,
		},
		AttributePrefixes: []string{"cloud.account."},
	},

	// Cloud Region
	string(semconv.CloudRegionKey): {
		Type: "cloud.region",
		Relationships: map[string]string{
			string(semconv.CloudProviderKey):         BelongsToProvider,
			string(semconv.CloudAvailabilityZoneKey): ContainsAvailabilityZone,
			string(semconv.CloudAccountIDKey):        ContainsResourcesFromAccount,
		},
		AttributePrefixes: []string{"cloud.region."},
	},

	// Cloud Availability Zone
	string(semconv.CloudAvailabilityZoneKey): {
		Type: "cloud.availability_zone",
		Relationships: map[string]string{
			string(semconv.CloudRegionKey): BelongsToRegion,
		},
		AttributePrefixes: []string{"cloud.availability_zone."},
	},
}
