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
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/plog"
)

func TestExtractPodObject(t *testing.T) {
	filePath := "testdata/pod-normal.json"
	data, err := os.ReadFile(filePath)
	assert.NoError(t, err, "Failed to read test data file")

	var body map[string]interface{}
	err = json.Unmarshal(data, &body)
	assert.NoError(t, err, "Failed to unmarshal JSON")

	logs := plog.NewLogs()
	resourceLogs := logs.ResourceLogs().AppendEmpty()
	scopeLogs := resourceLogs.ScopeLogs().AppendEmpty()
	logRecord := scopeLogs.LogRecords().AppendEmpty()

	err = logRecord.Body().SetEmptyMap().FromRaw(body)
	assert.NoError(t, err, "Failed to set log record body")

	podObject := extractPodObject(logRecord)
	assert.NotNil(t, podObject, "K8SPodObject should not be nil")

	expectedPod := &K8SPodObject{
		Name: "aws-test-us-east-2-cl-polling-57d97bbf5f-gj9qx",
		Labels: map[string]string{
			"app.kubernetes.io/component":  "collector",
			"app.kubernetes.io/instance":   "aws-test-us-east-2-cl-polling",
			"app.kubernetes.io/managed-by": "cardinalhq-operator",
			"app.kubernetes.io/name":       "cardinalhq-otel-collector",
			"pod-template-hash":            "57d97bbf5f",
		},
		OwnerRefKind: "ReplicaSet",
		OwnerRefName: "aws-test-us-east-2-cl-polling-57d97bbf5f",
		Resources: map[string]string{
			"requests.cpu":    "2",
			"requests.memory": "2Gi",
			"limits.cpu":      "2",
			"limits.memory":   "2Gi",
		},
		Phase:   "Running",
		PodIP:   "10.181.7.227",
		HostIP:  "10.181.4.236",
		ImageID: "public.ecr.aws/cardinalhq.io/cardinalhq-otel-collector@sha256:c969517e8e5dfffd51fdd20168e71d8dcdcd47b187ffd5b26436cdf4df65a8c9",
	}

	assert.Equal(t, expectedPod.Name, podObject.Name, "Pod name mismatch")
	assert.Equal(t, expectedPod.Labels, podObject.Labels, "Pod labels mismatch")
	assert.Equal(t, expectedPod.OwnerRefKind, podObject.OwnerRefKind, "OwnerRefKind mismatch")
	assert.Equal(t, expectedPod.OwnerRefName, podObject.OwnerRefName, "OwnerRefName mismatch")
	assert.Equal(t, expectedPod.Resources, podObject.Resources, "Resources mismatch")
	assert.Equal(t, expectedPod.Phase, podObject.Phase, "Phase mismatch")
	assert.Equal(t, expectedPod.PodIP, podObject.PodIP, "PodIP mismatch")
	assert.Equal(t, expectedPod.HostIP, podObject.HostIP, "HostIP mismatch")
	assert.Equal(t, expectedPod.ImageID, podObject.ImageID, "ImageID mismatch")
}

func TestExtractPodObject_CrashLoopBackOff(t *testing.T) {
	filePath := "testdata/pod-crashloop.json"
	data, err := os.ReadFile(filePath)
	assert.NoError(t, err, "Failed to read test data file")

	var body map[string]interface{}
	err = json.Unmarshal(data, &body)
	assert.NoError(t, err, "Failed to unmarshal JSON")

	logs := plog.NewLogs()
	resourceLogs := logs.ResourceLogs().AppendEmpty()
	scopeLogs := resourceLogs.ScopeLogs().AppendEmpty()
	logRecord := scopeLogs.LogRecords().AppendEmpty()

	err = logRecord.Body().SetEmptyMap().FromRaw(body)
	assert.NoError(t, err, "Failed to set log record body")

	podObject := extractPodObject(logRecord)
	assert.NotNil(t, podObject, "K8SPodObject should not be nil")

	expectedPod := &K8SPodObject{
		Name: "aws-test-us-east-2-cl-polling-57d97bbf5f-gj9qx",
		Labels: map[string]string{
			"app.kubernetes.io/component":  "collector",
			"app.kubernetes.io/instance":   "aws-test-us-east-2-cl-polling",
			"app.kubernetes.io/managed-by": "cardinalhq-operator",
			"app.kubernetes.io/name":       "cardinalhq-otel-collector",
			"pod-template-hash":            "57d97bbf5f",
		},
		OwnerRefKind: "ReplicaSet",
		OwnerRefName: "aws-test-us-east-2-cl-polling-57d97bbf5f",
		Resources: map[string]string{
			"requests.cpu":    "2",
			"requests.memory": "2Gi",
			"limits.cpu":      "2",
			"limits.memory":   "2Gi",
		},
		Phase:              "Running",
		PodIP:              "10.181.7.227",
		HostIP:             "10.181.4.236",
		ImageID:            "public.ecr.aws/cardinalhq.io/cardinalhq-otel-collector@sha256:c969517e8e5dfffd51fdd20168e71d8dcdcd47b187ffd5b26436cdf4df65a8c9",
		IsCrashLoopBackOff: true,
	}

	assert.Equal(t, expectedPod.Name, podObject.Name, "Pod name mismatch")
	assert.Equal(t, expectedPod.Labels, podObject.Labels, "Pod labels mismatch")
	assert.Equal(t, expectedPod.OwnerRefKind, podObject.OwnerRefKind, "OwnerRefKind mismatch")
	assert.Equal(t, expectedPod.OwnerRefName, podObject.OwnerRefName, "OwnerRefName mismatch")
	assert.Equal(t, expectedPod.Resources, podObject.Resources, "Resources mismatch")
	assert.Equal(t, expectedPod.Phase, podObject.Phase, "Phase mismatch")
	assert.Equal(t, expectedPod.PodIP, podObject.PodIP, "PodIP mismatch")
	assert.Equal(t, expectedPod.HostIP, podObject.HostIP, "HostIP mismatch")
	assert.Equal(t, expectedPod.ImageID, podObject.ImageID, "ImageID mismatch")
	assert.True(t, podObject.IsCrashLoopBackOff, "Expected IsCrashLoopBackOff to be true")
}

func TestExtractPodObject_ImagePullBackOff(t *testing.T) {
	filePath := "testdata/pod-image-pull-back-off.json"
	data, err := os.ReadFile(filePath)
	assert.NoError(t, err, "Failed to read test data file")

	var body map[string]interface{}
	err = json.Unmarshal(data, &body)
	assert.NoError(t, err, "Failed to unmarshal JSON")

	logs := plog.NewLogs()
	resourceLogs := logs.ResourceLogs().AppendEmpty()
	scopeLogs := resourceLogs.ScopeLogs().AppendEmpty()
	logRecord := scopeLogs.LogRecords().AppendEmpty()

	err = logRecord.Body().SetEmptyMap().FromRaw(body)
	assert.NoError(t, err, "Failed to set log record body")

	podObject := extractPodObject(logRecord)
	assert.NotNil(t, podObject, "K8SPodObject should not be nil")

	expectedPod := &K8SPodObject{
		Name: "aws-test-us-east-2-cl-polling-57d97bbf5f-gj9qx",
		Labels: map[string]string{
			"app.kubernetes.io/component":  "collector",
			"app.kubernetes.io/instance":   "aws-test-us-east-2-cl-polling",
			"app.kubernetes.io/managed-by": "cardinalhq-operator",
			"app.kubernetes.io/name":       "cardinalhq-otel-collector",
			"pod-template-hash":            "57d97bbf5f",
		},
		OwnerRefKind: "ReplicaSet",
		OwnerRefName: "aws-test-us-east-2-cl-polling-57d97bbf5f",
		Resources: map[string]string{
			"requests.cpu":    "2",
			"requests.memory": "2Gi",
			"limits.cpu":      "2",
			"limits.memory":   "2Gi",
		},
		Phase:              "Pending",
		PodIP:              "",
		HostIP:             "10.181.4.236",
		ImageID:            "",
		IsImagePullBackOff: true,
		IsCrashLoopBackOff: false,
	}

	assert.Equal(t, expectedPod.Name, podObject.Name, "Pod name mismatch")
	assert.Equal(t, expectedPod.Labels, podObject.Labels, "Pod labels mismatch")
	assert.Equal(t, expectedPod.OwnerRefKind, podObject.OwnerRefKind, "OwnerRefKind mismatch")
	assert.Equal(t, expectedPod.OwnerRefName, podObject.OwnerRefName, "OwnerRefName mismatch")
	assert.Equal(t, expectedPod.Resources, podObject.Resources, "Resources mismatch")
	assert.Equal(t, expectedPod.Phase, podObject.Phase, "Phase mismatch")
	assert.Equal(t, expectedPod.PodIP, podObject.PodIP, "PodIP mismatch")
	assert.Equal(t, expectedPod.HostIP, podObject.HostIP, "HostIP mismatch")
	assert.Equal(t, expectedPod.ImageID, podObject.ImageID, "ImageID mismatch")

	assert.True(t, podObject.IsImagePullBackOff, "Pod should be in ImagePullBackOff state")
	assert.False(t, podObject.IsCrashLoopBackOff, "Pod should not be in CrashLoopBackOff state")
}

func TestExtractPodObject_PendingPod(t *testing.T) {
	filePath := "testdata/pod-pending.json"
	data, err := os.ReadFile(filePath)
	assert.NoError(t, err, "Failed to read test data file")

	var body map[string]interface{}
	err = json.Unmarshal(data, &body)
	assert.NoError(t, err, "Failed to unmarshal JSON")

	logs := plog.NewLogs()
	resourceLogs := logs.ResourceLogs().AppendEmpty()
	scopeLogs := resourceLogs.ScopeLogs().AppendEmpty()
	logRecord := scopeLogs.LogRecords().AppendEmpty()

	err = logRecord.Body().SetEmptyMap().FromRaw(body)
	assert.NoError(t, err, "Failed to set log record body")

	podObject := extractPodObject(logRecord)
	assert.NotNil(t, podObject, "K8SPodObject should not be nil")

	// Expected pod object values
	expectedPod := &K8SPodObject{
		Name: "pending-pod-12345",
		Labels: map[string]string{
			"app":         "test-app",
			"environment": "staging",
		},
		OwnerRefKind: "ReplicaSet",
		OwnerRefName: "pending-pod-replicaset",
		Resources: map[string]string{
			"requests.cpu":    "500m",
			"requests.memory": "512Mi",
			"limits.cpu":      "1000m",
			"limits.memory":   "1024Mi",
		},
		Phase:         "Pending",
		PodIP:         "",
		HostIP:        "",
		PendingReason: "Unschedulable",
	}

	assert.Equal(t, expectedPod.Name, podObject.Name, "Pod name mismatch")
	assert.Equal(t, expectedPod.Labels, podObject.Labels, "Pod labels mismatch")
	assert.Equal(t, expectedPod.OwnerRefKind, podObject.OwnerRefKind, "OwnerRefKind mismatch")
	assert.Equal(t, expectedPod.OwnerRefName, podObject.OwnerRefName, "OwnerRefName mismatch")
	assert.Equal(t, expectedPod.Resources, podObject.Resources, "Resources mismatch")
	assert.Equal(t, expectedPod.Phase, podObject.Phase, "Phase mismatch")
	assert.Equal(t, expectedPod.PodIP, podObject.PodIP, "PodIP mismatch")
	assert.Equal(t, expectedPod.HostIP, podObject.HostIP, "HostIP mismatch")
	assert.Equal(t, expectedPod.PendingReason, podObject.PendingReason, "PendingReason mismatch")
}

func TestExtractDeploymentObject(t *testing.T) {
	filePath := "testdata/deployment.json"
	data, err := os.ReadFile(filePath)
	assert.NoError(t, err, "Failed to read test data file")

	var body map[string]interface{}
	err = json.Unmarshal(data, &body)
	assert.NoError(t, err, "Failed to unmarshal JSON")

	logs := plog.NewLogs()
	resourceLogs := logs.ResourceLogs().AppendEmpty()
	scopeLogs := resourceLogs.ScopeLogs().AppendEmpty()
	logRecord := scopeLogs.LogRecords().AppendEmpty()

	err = logRecord.Body().SetEmptyMap().FromRaw(body)
	assert.NoError(t, err, "Failed to set log record body")

	deploymentObject := extractDeploymentObject(logRecord)
	assert.NotNil(t, deploymentObject, "K8SDeploymentObject should not be nil")

	expectedDeployment := &K8SDeploymentObject{
		Name:              "cert-manager-webhook",
		Namespace:         "cert-manager",
		ReplicasAvailable: 1,
		ReplicasDesired:   1,
		ReplicasUpdated:   1,
		DeploymentStatus:  "Available",
		ProgressMessage:   "ReplicaSet \"cert-manager-webhook-856b854bfb\" has successfully progressed.",
	}

	assert.Equal(t, expectedDeployment.Name, deploymentObject.Name, "Deployment name mismatch")
	assert.Equal(t, expectedDeployment.Namespace, deploymentObject.Namespace, "Namespace mismatch")
	assert.Equal(t, expectedDeployment.ReplicasAvailable, deploymentObject.ReplicasAvailable, "ReplicasAvailable mismatch")
	assert.Equal(t, expectedDeployment.ReplicasDesired, deploymentObject.ReplicasDesired, "ReplicasDesired mismatch")
	assert.Equal(t, expectedDeployment.ReplicasUpdated, deploymentObject.ReplicasUpdated, "ReplicasUpdated mismatch")
	assert.Equal(t, expectedDeployment.DeploymentStatus, deploymentObject.DeploymentStatus, "DeploymentStatus mismatch")
	assert.Equal(t, expectedDeployment.ProgressMessage, deploymentObject.ProgressMessage, "ProgressMessage mismatch")
}

func TestExtractStatefulSetObject(t *testing.T) {
	filePath := "testdata/statefulset.json"
	data, err := os.ReadFile(filePath)
	assert.NoError(t, err, "Failed to read test data file")

	var body map[string]interface{}
	err = json.Unmarshal(data, &body)
	assert.NoError(t, err, "Failed to unmarshal JSON")

	logs := plog.NewLogs()
	resourceLogs := logs.ResourceLogs().AppendEmpty()
	scopeLogs := resourceLogs.ScopeLogs().AppendEmpty()
	logRecord := scopeLogs.LogRecords().AppendEmpty()

	err = logRecord.Body().SetEmptyMap().FromRaw(body)
	assert.NoError(t, err, "Failed to set log record body")

	statefulSetObject := extractStatefulSetObject(logRecord)

	assert.NotNil(t, statefulSetObject, "K8SStatefulSetObject should not be nil")

	expectedObject := &K8SStatefulSetObject{
		Name:      "prometheus-alertmanager",
		Namespace: "chq-demo-apps",
		Attributes: map[string]string{
			"ReplicasDesired":   "1",
			"ReplicasAvailable": "1",
			"ReplicasUpdated":   "1",
			"CurrentRevision":   "prometheus-alertmanager-7c6d54f5bc",
			"PersistentVolume":  "",
			"VolumeStatus":      "Pending",
		},
	}

	assert.Equal(t, expectedObject.Name, statefulSetObject.Name, "StatefulSet name mismatch")
	assert.Equal(t, expectedObject.Namespace, statefulSetObject.Namespace, "Namespace mismatch")
	assert.Equal(t, expectedObject.Attributes, statefulSetObject.Attributes, "Attributes mismatch")
}
