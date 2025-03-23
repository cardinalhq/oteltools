package main

import (
	"fmt"
	"os"
)

type NodeType string

const (
	NodeTypeK8SPod         NodeType = "k8s.pod"
	NodeTypeK8SReplicaSet  NodeType = "k8s.replicaset"
	NodeTypeK8SDeployment  NodeType = "k8s.deployment"
	NodeTypeK8SStatefulSet NodeType = "k8s.statefulset"
	NodeTypeK8SDaemonSet   NodeType = "k8s.daemonset"
	NodeTypeK8SJob         NodeType = "k8s.job"
	NodeTypeK8SCronJob     NodeType = "k8s.cronjob"
	NodeTypeK8SConfigMap   NodeType = "k8s.configmap"
	NodeTypeK8SSecret      NodeType = "k8s.secret"
	NodeTypeK8SNamespace   NodeType = "k8s.namespace"
	NodeTypeK8SNode        NodeType = "k8s.node"
	NodeTypeCHQService     NodeType = "chq.service"
)

type NodeStyle string

const (
	NodeStyleReal    NodeStyle = "real"
	NodeStyleVirtual NodeStyle = "virtual"
)

type Relationship string

const (
	RelationshipManages   Relationship = "manages"
	RelationshipManagedBy Relationship = "managed by"

	RelationshipHostedOn Relationship = "hosted on"
	RelationshipHosts    Relationship = "hosts"

	RelationshipUses   Relationship = "uses"
	RelationshipUsedBy Relationship = "used by"

	RelationshipPartOf   Relationship = "part of"
	RelationshipIsPartOf Relationship = "is part of"

	RelationshipContains Relationship = "contains"
)

type DetectionType string

const (
	// There is a drect relationship between the two nodes
	// e.g., an OwnerReference in a K8s object
	DetectionTypeDirect DetectionType = "direct"

	// The relationship is calculated based on some other information,
	// e.g., a ReplicaSet manages a Pod if the Pod has this ReplicaSet
	// in its OwnerReferences
	DetectionTypeCalculated DetectionType = "calculated"
)

type NodeStyleMap struct {
	NodeType NodeType
	Style    NodeStyle
}

var NodeStyles = map[NodeType]NodeStyle{
	NodeTypeK8SPod:         NodeStyleReal,
	NodeTypeK8SReplicaSet:  NodeStyleReal,
	NodeTypeK8SDeployment:  NodeStyleReal,
	NodeTypeK8SStatefulSet: NodeStyleReal,
	NodeTypeK8SDaemonSet:   NodeStyleReal,
	NodeTypeK8SJob:         NodeStyleReal,
	NodeTypeK8SCronJob:     NodeStyleReal,
	NodeTypeK8SConfigMap:   NodeStyleReal,
	NodeTypeK8SSecret:      NodeStyleReal,
	NodeTypeK8SNamespace:   NodeStyleVirtual,
	NodeTypeK8SNode:        NodeStyleReal,

	NodeTypeCHQService: NodeStyleVirtual,
}

type Edge struct {
	SourceType      NodeType
	DestinationType NodeType
	Relationship    Relationship
	DetectionType   DetectionType
}

var Edges = []Edge{
	{NodeTypeK8SPod, NodeTypeK8SReplicaSet, RelationshipManagedBy, DetectionTypeDirect},
	{NodeTypeK8SPod, NodeTypeK8SStatefulSet, RelationshipManagedBy, DetectionTypeDirect},
	{NodeTypeK8SPod, NodeTypeK8SDaemonSet, RelationshipManagedBy, DetectionTypeDirect},
	{NodeTypeK8SPod, NodeTypeK8SJob, RelationshipManagedBy, DetectionTypeDirect},
	{NodeTypeK8SPod, NodeTypeK8SConfigMap, RelationshipUses, DetectionTypeDirect},
	{NodeTypeK8SPod, NodeTypeK8SSecret, RelationshipUses, DetectionTypeDirect},
	{NodeTypeK8SPod, NodeTypeCHQService, RelationshipPartOf, DetectionTypeDirect},
	{NodeTypeK8SPod, NodeTypeK8SNamespace, RelationshipPartOf, DetectionTypeDirect},
	{NodeTypeK8SPod, NodeTypeK8SNode, RelationshipHostedOn, DetectionTypeDirect},

	{NodeTypeK8SReplicaSet, NodeTypeK8SDeployment, RelationshipManagedBy, DetectionTypeDirect},
	{NodeTypeK8SReplicaSet, NodeTypeK8SPod, RelationshipManages, DetectionTypeCalculated},
	{NodeTypeK8SReplicaSet, NodeTypeK8SConfigMap, RelationshipUses, DetectionTypeDirect},
	{NodeTypeK8SReplicaSet, NodeTypeK8SSecret, RelationshipUses, DetectionTypeDirect},
	{NodeTypeK8SReplicaSet, NodeTypeCHQService, RelationshipPartOf, DetectionTypeDirect},
	{NodeTypeK8SReplicaSet, NodeTypeK8SNamespace, RelationshipPartOf, DetectionTypeDirect},

	{NodeTypeK8SDeployment, NodeTypeK8SReplicaSet, RelationshipManages, DetectionTypeCalculated},
	{NodeTypeK8SDeployment, NodeTypeK8SConfigMap, RelationshipUses, DetectionTypeDirect},
	{NodeTypeK8SDeployment, NodeTypeK8SSecret, RelationshipUses, DetectionTypeDirect},
	{NodeTypeK8SDeployment, NodeTypeCHQService, RelationshipPartOf, DetectionTypeDirect},
	{NodeTypeK8SDeployment, NodeTypeK8SNamespace, RelationshipPartOf, DetectionTypeDirect},

	{NodeTypeK8SStatefulSet, NodeTypeK8SPod, RelationshipManages, DetectionTypeCalculated},
	{NodeTypeK8SStatefulSet, NodeTypeK8SConfigMap, RelationshipUses, DetectionTypeDirect},
	{NodeTypeK8SStatefulSet, NodeTypeK8SSecret, RelationshipUses, DetectionTypeDirect},
	{NodeTypeK8SStatefulSet, NodeTypeCHQService, RelationshipPartOf, DetectionTypeDirect},
	{NodeTypeK8SStatefulSet, NodeTypeK8SNamespace, RelationshipPartOf, DetectionTypeDirect},

	{NodeTypeK8SDaemonSet, NodeTypeK8SPod, RelationshipManages, DetectionTypeCalculated},
	{NodeTypeK8SDaemonSet, NodeTypeK8SConfigMap, RelationshipUses, DetectionTypeDirect},
	{NodeTypeK8SDaemonSet, NodeTypeK8SSecret, RelationshipUses, DetectionTypeDirect},
	{NodeTypeK8SDaemonSet, NodeTypeCHQService, RelationshipPartOf, DetectionTypeDirect},
	{NodeTypeK8SDaemonSet, NodeTypeK8SNamespace, RelationshipPartOf, DetectionTypeDirect},

	{NodeTypeK8SJob, NodeTypeK8SPod, RelationshipManages, DetectionTypeCalculated},
	{NodeTypeK8SJob, NodeTypeK8SCronJob, RelationshipManagedBy, DetectionTypeDirect},
	{NodeTypeK8SJob, NodeTypeK8SConfigMap, RelationshipUses, DetectionTypeDirect},
	{NodeTypeK8SJob, NodeTypeK8SSecret, RelationshipUses, DetectionTypeDirect},
	{NodeTypeK8SJob, NodeTypeCHQService, RelationshipPartOf, DetectionTypeDirect},
	{NodeTypeK8SJob, NodeTypeK8SNamespace, RelationshipPartOf, DetectionTypeDirect},

	{NodeTypeK8SCronJob, NodeTypeK8SJob, RelationshipManages, DetectionTypeCalculated},
	{NodeTypeK8SCronJob, NodeTypeK8SConfigMap, RelationshipUses, DetectionTypeDirect},
	{NodeTypeK8SCronJob, NodeTypeK8SSecret, RelationshipUses, DetectionTypeDirect},
	{NodeTypeK8SCronJob, NodeTypeCHQService, RelationshipPartOf, DetectionTypeDirect},
	{NodeTypeK8SCronJob, NodeTypeK8SNamespace, RelationshipPartOf, DetectionTypeDirect},

	{NodeTypeK8SConfigMap, NodeTypeK8SPod, RelationshipUsedBy, DetectionTypeCalculated},
	{NodeTypeK8SConfigMap, NodeTypeK8SReplicaSet, RelationshipUsedBy, DetectionTypeCalculated},
	{NodeTypeK8SConfigMap, NodeTypeK8SDeployment, RelationshipUsedBy, DetectionTypeCalculated},
	{NodeTypeK8SConfigMap, NodeTypeK8SStatefulSet, RelationshipUsedBy, DetectionTypeCalculated},
	{NodeTypeK8SConfigMap, NodeTypeK8SDaemonSet, RelationshipUsedBy, DetectionTypeCalculated},
	{NodeTypeK8SConfigMap, NodeTypeK8SJob, RelationshipUsedBy, DetectionTypeCalculated},
	{NodeTypeK8SConfigMap, NodeTypeK8SCronJob, RelationshipUsedBy, DetectionTypeCalculated},
	{NodeTypeK8SConfigMap, NodeTypeK8SNamespace, RelationshipPartOf, DetectionTypeDirect},

	{NodeTypeK8SSecret, NodeTypeK8SPod, RelationshipUsedBy, DetectionTypeCalculated},
	{NodeTypeK8SSecret, NodeTypeK8SReplicaSet, RelationshipUsedBy, DetectionTypeCalculated},
	{NodeTypeK8SSecret, NodeTypeK8SDeployment, RelationshipUsedBy, DetectionTypeCalculated},
	{NodeTypeK8SSecret, NodeTypeK8SStatefulSet, RelationshipUsedBy, DetectionTypeCalculated},
	{NodeTypeK8SSecret, NodeTypeK8SDaemonSet, RelationshipUsedBy, DetectionTypeCalculated},
	{NodeTypeK8SSecret, NodeTypeK8SJob, RelationshipUsedBy, DetectionTypeCalculated},
	{NodeTypeK8SSecret, NodeTypeK8SCronJob, RelationshipUsedBy, DetectionTypeCalculated},
	{NodeTypeK8SSecret, NodeTypeK8SNamespace, RelationshipPartOf, DetectionTypeDirect},

	{NodeTypeCHQService, NodeTypeK8SPod, RelationshipIsPartOf, DetectionTypeCalculated},
	{NodeTypeCHQService, NodeTypeK8SReplicaSet, RelationshipIsPartOf, DetectionTypeCalculated},
	{NodeTypeCHQService, NodeTypeK8SDeployment, RelationshipIsPartOf, DetectionTypeCalculated},
	{NodeTypeCHQService, NodeTypeK8SStatefulSet, RelationshipIsPartOf, DetectionTypeCalculated},
	{NodeTypeCHQService, NodeTypeK8SDaemonSet, RelationshipIsPartOf, DetectionTypeCalculated},
	{NodeTypeCHQService, NodeTypeK8SJob, RelationshipIsPartOf, DetectionTypeCalculated},
	{NodeTypeCHQService, NodeTypeK8SCronJob, RelationshipIsPartOf, DetectionTypeCalculated},

	{NodeTypeK8SNamespace, NodeTypeK8SPod, RelationshipContains, DetectionTypeCalculated},
	{NodeTypeK8SNamespace, NodeTypeK8SReplicaSet, RelationshipContains, DetectionTypeCalculated},
	{NodeTypeK8SNamespace, NodeTypeK8SDeployment, RelationshipContains, DetectionTypeCalculated},
	{NodeTypeK8SNamespace, NodeTypeK8SStatefulSet, RelationshipContains, DetectionTypeCalculated},
	{NodeTypeK8SNamespace, NodeTypeK8SDaemonSet, RelationshipContains, DetectionTypeCalculated},
	{NodeTypeK8SNamespace, NodeTypeK8SJob, RelationshipContains, DetectionTypeCalculated},
	{NodeTypeK8SNamespace, NodeTypeK8SCronJob, RelationshipContains, DetectionTypeCalculated},
	{NodeTypeK8SNamespace, NodeTypeK8SConfigMap, RelationshipContains, DetectionTypeCalculated},
	{NodeTypeK8SNamespace, NodeTypeK8SSecret, RelationshipContains, DetectionTypeCalculated},

	{NodeTypeK8SNode, NodeTypeK8SPod, RelationshipHosts, DetectionTypeCalculated},
}

type edgeKey struct {
	A NodeType
	B NodeType
}

type edgeData struct {
	ForwardDirect     bool
	ReverseCalculated bool
	OnlyCalculated    bool
	ForwardLabel      string
	ReverseLabel      string
}

const renderRelationships = false

func main() {
	fmt.Println("digraph G {")
	fmt.Println(`  overlap=false;`)
	fmt.Println(`  fontname="sans-serif";`)
	fmt.Println(`  nodesep=0.6;`)
	fmt.Println(`  ranksep=0.8;`)
	fmt.Println(`  node [fontname="sans-serif", width=1.5, height=0.8, fixedsize=true];`)
	fmt.Println(`  edge [fontname="sans-serif", penwidth=2, arrowsize=1.5];`)
	fmt.Println()

	// Define node shapes
	for nodeType, style := range NodeStyles {
		shape := "rectangle"
		if style == NodeStyleVirtual {
			shape = "ellipse"
		}
		fmt.Printf("  \"%s\" [shape=%s];\n", nodeType, shape)
	}
	fmt.Println()

	// Build edge map
	edgeMap := map[edgeKey]*edgeData{}
	for _, edge := range Edges {
		k := edgeKey{A: edge.SourceType, B: edge.DestinationType}
		if edge.DetectionType == DetectionTypeDirect {
			if edgeMap[k] == nil {
				edgeMap[k] = &edgeData{}
			}
			edgeMap[k].ForwardDirect = true
			edgeMap[k].ForwardLabel = string(edge.Relationship)
		} else {
			reverseKey := edgeKey{A: edge.DestinationType, B: edge.SourceType}
			if edgeMap[reverseKey] != nil && edgeMap[reverseKey].ForwardDirect {
				if edgeMap[reverseKey] == nil {
					edgeMap[reverseKey] = &edgeData{}
				}
				edgeMap[reverseKey].ReverseCalculated = true
				edgeMap[reverseKey].ReverseLabel = string(edge.Relationship)
			} else {
				fmt.Fprintf(os.Stderr, "Warning: Only calculated edge between %s -> %s\n", edge.SourceType, edge.DestinationType)
				if edgeMap[k] == nil {
					edgeMap[k] = &edgeData{}
				}
				edgeMap[k].OnlyCalculated = true
				edgeMap[k].ForwardLabel = string(edge.Relationship)
			}
		}
	}

	// Render edges
	for k, data := range edgeMap {
		if data.OnlyCalculated {
			var label string
			if renderRelationships {
				label = data.ForwardLabel
			}
			fmt.Printf("  \"%s\" -> \"%s\" [label=\"%s\", color=darkorange, fontsize=10];\n",
				k.A, k.B, label)
		} else if data.ForwardDirect && data.ReverseCalculated {
			var label string
			if renderRelationships {
				label = fmt.Sprintf("%s / %s", data.ForwardLabel, data.ReverseLabel)
			}
			fmt.Printf("  \"%s\" -> \"%s\" [label=\"%s\", dir=both, arrowhead=normal, arrowtail=vee, color=darkblue, fontsize=10];\n",
				k.A, k.B, label)
		} else if data.ForwardDirect {
			var label string
			if renderRelationships {
				label = data.ForwardLabel
			}
			fmt.Printf("  \"%s\" -> \"%s\" [label=\"%s\", color=darkgreen, fontsize=10];\n",
				k.A, k.B, label)
		}
	}

	fmt.Println("}")
}
