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
	"go.opentelemetry.io/collector/pdata/plog"
)

type K8SPodObject struct {
	Name               string            `json:"name"`
	Labels             map[string]string `json:"labels"`
	OwnerRefKind       string            `json:"ownerRefKind"`
	OwnerRefName       string            `json:"ownerRefName"`
	Resources          map[string]string `json:"resources"`
	Phase              string            `json:"phase"`
	StartedAt          string            `json:"startedAt"`
	PendingReason      string            `json:"pendingReason"`
	IsImagePullBackOff bool              `json:"isImagePullBackOff"`
	IsCrashLoopBackOff bool              `json:"isCrashLoopBackOff"`
	IsOOMKilled        bool              `json:"isOOMKilled"`
	PodIP              string            `json:"podIP"`
	HostIP             string            `json:"hostIP"`
	ImageID            string            `json:"imageID"`
}

func ExtractPodObject(lr plog.LogRecord) *K8SPodObject {
	rawValue := lr.Body().AsRaw()

	objectMap, ok := rawValue.(map[string]any)
	if !ok {
		return nil
	}

	if kind, ok := objectMap["kind"].(string); !ok || kind != "Pod" {
		return nil
	}

	metadata, metadataPresent := objectMap["metadata"].(map[string]any)
	if !metadataPresent {
		return nil
	}

	name, _ := metadata["name"].(string)

	labels := make(map[string]string)
	if labelMap, exists := metadata["labels"].(map[string]any); exists {
		for k, v := range labelMap {
			if str, ok := v.(string); ok {
				labels[k] = str
			}
		}
	}

	// Extract owner reference
	ownerKind, ownerName := "", ""
	if ownerRefs, exists := metadata["ownerReferences"].([]any); exists && len(ownerRefs) > 0 {
		if firstOwnerRef, ok := ownerRefs[0].(map[string]any); ok {
			ownerKind, _ = firstOwnerRef["kind"].(string)
			ownerName, _ = firstOwnerRef["name"].(string)
		}
	}

	// Extract resources from spec
	resourceMap := make(map[string]string)
	if spec, exists := objectMap["spec"].(map[string]any); exists {
		if containers, exists := spec["containers"].([]any); exists {
			for _, container := range containers {
				if containerMap, ok := container.(map[string]any); ok {
					if resources, exists := containerMap["resources"].(map[string]any); exists {
						if requests, exists := resources["requests"].(map[string]any); exists {
							for k, v := range requests {
								if str, ok := v.(string); ok {
									resourceMap["requests."+k] = str
								}
							}
						}
						if limits, exists := resources["limits"].(map[string]any); exists {
							for k, v := range limits {
								if str, ok := v.(string); ok {
									resourceMap["limits."+k] = str
								}
							}
						}
					}
				}
			}
		}
	}

	phase, podIP, hostIP, imageID, startedAt := "", "", "", "", ""
	pendingReason, isImagePullBackOff, isCrashLoopBackOff, isOOMKilled := "", false, false, false

	if status, exists := objectMap["status"].(map[string]any); exists {
		phase, _ = status["phase"].(string)
		podIP, _ = status["podIP"].(string)
		hostIP, _ = status["hostIP"].(string)
		startedAt, _ = status["startTime"].(string) // Extract `startTime`

		if phase == "Pending" {
			if conditions, exists := status["conditions"].([]any); exists {
				for _, cond := range conditions {
					if condMap, ok := cond.(map[string]any); ok {
						if condType, exists := condMap["type"].(string); exists && condType == "PodScheduled" {
							pendingReason, _ = condMap["reason"].(string)
							break
						}
					}
				}
			}
		} else {
			pendingReason = "N/A"
		}

		if containerStatuses, exists := status["containerStatuses"].([]any); exists {
			for _, container := range containerStatuses {
				if containerMap, ok := container.(map[string]any); ok {
					imageID, _ = containerMap["imageID"].(string)

					if state, exists := containerMap["state"].(map[string]any); exists {
						if waiting, exists := state["waiting"].(map[string]any); exists {
							if reason, exists := waiting["reason"].(string); exists {
								if reason == "ImagePullBackOff" {
									isImagePullBackOff = true
								} else if reason == "CrashLoopBackOff" {
									isCrashLoopBackOff = true
								}
							}
						}
					}

					// Check last terminated state for OOMKilled
					if lastState, exists := containerMap["lastState"].(map[string]any); exists {
						if terminated, exists := lastState["terminated"].(map[string]any); exists {
							if reason, exists := terminated["reason"].(string); exists && reason == "OOMKilled" {
								isOOMKilled = true
							}
						}
					}
				}
			}
		}
	}

	return &K8SPodObject{
		Name:               name,
		Labels:             labels,
		OwnerRefKind:       ownerKind,
		OwnerRefName:       ownerName,
		Resources:          resourceMap,
		Phase:              phase,
		StartedAt:          startedAt,
		PendingReason:      pendingReason,
		IsImagePullBackOff: isImagePullBackOff,
		IsCrashLoopBackOff: isCrashLoopBackOff,
		IsOOMKilled:        isOOMKilled,
		PodIP:              podIP,
		HostIP:             hostIP,
		ImageID:            imageID,
	}
}

type K8SDeploymentObject struct {
	Name              string
	Namespace         string
	ReplicasAvailable int
	ReplicasDesired   int
	ReplicasUpdated   int
	DeploymentStatus  string
	ProgressMessage   string
}

func ExtractDeploymentObject(lr plog.LogRecord) *K8SDeploymentObject {
	rawValue := lr.Body().AsRaw()
	if bodyMap, ok := rawValue.(map[string]any); ok {
		if kind, exists := bodyMap["kind"].(string); !exists || kind != "Deployment" {
			return nil
		}

		metadata, metadataPresent := bodyMap["metadata"].(map[string]any)
		if !metadataPresent {
			return nil
		}

		name := metadata["name"].(string)
		namespace := metadata["namespace"].(string)

		// Extract replica information
		replicasAvailable := 0
		replicasDesired := 0
		replicasUpdated := 0
		progressMessage := ""

		if status, exists := bodyMap["status"].(map[string]any); exists {
			if available, exists := status["availableReplicas"].(float64); exists {
				replicasAvailable = int(available)
			}
			if desired, exists := status["replicas"].(float64); exists {
				replicasDesired = int(desired)
			}
			if updated, exists := status["updatedReplicas"].(float64); exists {
				replicasUpdated = int(updated)
			}
		}

		// Extract deployment progress status
		deploymentStatus := "Unknown"
		if status, exists := bodyMap["status"].(map[string]any); exists {
			if conditions, exists := status["conditions"].([]any); exists {
				for _, condition := range conditions {
					if conditionMap, ok := condition.(map[string]any); ok {
						condType, _ := conditionMap["type"].(string)
						condStatus, _ := conditionMap["status"].(string)
						if condType == "Progressing" && condStatus == "True" {
							deploymentStatus = "Progressing"
							if msg, exists := conditionMap["message"].(string); exists {
								progressMessage = msg
							}
						}
						if condType == "Available" && condStatus == "True" {
							deploymentStatus = "Available"
						}
					}
				}
			}
		}

		return &K8SDeploymentObject{
			Name:              name,
			Namespace:         namespace,
			ReplicasAvailable: replicasAvailable,
			ReplicasDesired:   replicasDesired,
			ReplicasUpdated:   replicasUpdated,
			DeploymentStatus:  deploymentStatus,
			ProgressMessage:   progressMessage,
		}
	}
	return nil
}

type K8SStatefulSetObject struct {
	Name       string
	Namespace  string
	Attributes map[string]string
}

// Safe extraction helper
func getStringFromMap(m map[string]any, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func ExtractStatefulSetObject(lr plog.LogRecord) *K8SStatefulSetObject {
	rawValue := lr.Body().AsRaw()
	bodyMap, ok := rawValue.(map[string]any)
	if !ok {
		return nil
	}

	metadata, metadataPresent := bodyMap["metadata"].(map[string]any)
	if !metadataPresent {
		return nil
	}

	name := getStringFromMap(metadata, "name")
	namespace := getStringFromMap(metadata, "namespace")

	status, statusPresent := bodyMap["status"].(map[string]any)
	if !statusPresent {
		return nil
	}

	attributes := map[string]string{
		"ReplicasDesired":   fmt.Sprintf("%v", status["replicas"]),
		"ReplicasAvailable": fmt.Sprintf("%v", status["availableReplicas"]),
		"ReplicasUpdated":   fmt.Sprintf("%v", status["updatedReplicas"]),
		"CurrentRevision":   getStringFromMap(status, "currentRevision"),
	}

	if volumeTemplates, exists := bodyMap["spec"].(map[string]any)["volumeClaimTemplates"]; exists {
		if volumeList, ok := volumeTemplates.([]any); ok && len(volumeList) > 0 {
			if firstVolume, ok := volumeList[0].(map[string]any); ok {
				attributes["PersistentVolume"] = getStringFromMap(firstVolume, "metadata.name")
				if spec, exists := firstVolume["status"].(map[string]any); exists {
					attributes["VolumeStatus"] = getStringFromMap(spec, "phase")
				}
			}
		}
	}

	return &K8SStatefulSetObject{
		Name:       name,
		Namespace:  namespace,
		Attributes: attributes,
	}
}
