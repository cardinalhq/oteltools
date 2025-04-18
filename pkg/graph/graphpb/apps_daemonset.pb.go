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

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v5.29.3
// source: apps_daemonset.proto

package graphpb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// AppsDaemonSetSummary summarizes a Kubernetes apps/v1 DaemonSet.
type AppsDaemonSetSummary struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	BaseObject    *BaseObject            `protobuf:"bytes,1,opt,name=base_object,json=baseObject,proto3" json:"base_object,omitempty"`
	Spec          *AppsDaemonSetSpec     `protobuf:"bytes,2,opt,name=spec,proto3" json:"spec,omitempty"`
	Status        *AppDaemonSetStatus    `protobuf:"bytes,3,opt,name=status,proto3" json:"status,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *AppsDaemonSetSummary) Reset() {
	*x = AppsDaemonSetSummary{}
	mi := &file_apps_daemonset_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AppsDaemonSetSummary) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AppsDaemonSetSummary) ProtoMessage() {}

func (x *AppsDaemonSetSummary) ProtoReflect() protoreflect.Message {
	mi := &file_apps_daemonset_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AppsDaemonSetSummary.ProtoReflect.Descriptor instead.
func (*AppsDaemonSetSummary) Descriptor() ([]byte, []int) {
	return file_apps_daemonset_proto_rawDescGZIP(), []int{0}
}

func (x *AppsDaemonSetSummary) GetBaseObject() *BaseObject {
	if x != nil {
		return x.BaseObject
	}
	return nil
}

func (x *AppsDaemonSetSummary) GetSpec() *AppsDaemonSetSpec {
	if x != nil {
		return x.Spec
	}
	return nil
}

func (x *AppsDaemonSetSummary) GetStatus() *AppDaemonSetStatus {
	if x != nil {
		return x.Status
	}
	return nil
}

// AppsDaemonSetSummarySpec is the spec of a AppsDaemonSetSummary.
type AppsDaemonSetSpec struct {
	state    protoimpl.MessageState `protogen:"open.v1"`
	Replicas int32                  `protobuf:"varint,1,opt,name=replicas,proto3" json:"replicas,omitempty"`
	// "template" is the pod template of the DaemonSet.
	Template *AppsDaemonSetTemplate `protobuf:"bytes,2,opt,name=template,proto3" json:"template,omitempty"`
	// "service_name" is the name of the service that governs the DaemonSet.
	ServiceName   string `protobuf:"bytes,3,opt,name=service_name,json=serviceName,proto3" json:"service_name,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *AppsDaemonSetSpec) Reset() {
	*x = AppsDaemonSetSpec{}
	mi := &file_apps_daemonset_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AppsDaemonSetSpec) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AppsDaemonSetSpec) ProtoMessage() {}

func (x *AppsDaemonSetSpec) ProtoReflect() protoreflect.Message {
	mi := &file_apps_daemonset_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AppsDaemonSetSpec.ProtoReflect.Descriptor instead.
func (*AppsDaemonSetSpec) Descriptor() ([]byte, []int) {
	return file_apps_daemonset_proto_rawDescGZIP(), []int{1}
}

func (x *AppsDaemonSetSpec) GetReplicas() int32 {
	if x != nil {
		return x.Replicas
	}
	return 0
}

func (x *AppsDaemonSetSpec) GetTemplate() *AppsDaemonSetTemplate {
	if x != nil {
		return x.Template
	}
	return nil
}

func (x *AppsDaemonSetSpec) GetServiceName() string {
	if x != nil {
		return x.ServiceName
	}
	return ""
}

// AppsDaemonSetSummaryStatus is the status of a AppsDaemonSetSummary.
type AppsDaemonSetTemplate struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Metadata      *BaseObject            `protobuf:"bytes,1,opt,name=metadata,proto3" json:"metadata,omitempty"`
	PodSpec       *PodSpec               `protobuf:"bytes,2,opt,name=pod_spec,json=podSpec,proto3" json:"pod_spec,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *AppsDaemonSetTemplate) Reset() {
	*x = AppsDaemonSetTemplate{}
	mi := &file_apps_daemonset_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AppsDaemonSetTemplate) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AppsDaemonSetTemplate) ProtoMessage() {}

func (x *AppsDaemonSetTemplate) ProtoReflect() protoreflect.Message {
	mi := &file_apps_daemonset_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AppsDaemonSetTemplate.ProtoReflect.Descriptor instead.
func (*AppsDaemonSetTemplate) Descriptor() ([]byte, []int) {
	return file_apps_daemonset_proto_rawDescGZIP(), []int{2}
}

func (x *AppsDaemonSetTemplate) GetMetadata() *BaseObject {
	if x != nil {
		return x.Metadata
	}
	return nil
}

func (x *AppsDaemonSetTemplate) GetPodSpec() *PodSpec {
	if x != nil {
		return x.PodSpec
	}
	return nil
}

// AppDaemonSetStatus is the status of a Kubernetes apps/v1 DaemonSet.
type AppDaemonSetStatus struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// The number of nodes that are running at least 1
	// daemon pod and are supposed to run the daemon pod.
	// More info: https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/
	CurrentNumberScheduled int32 `protobuf:"varint,1,opt,name=current_number_scheduled,json=currentNumberScheduled,proto3" json:"current_number_scheduled,omitempty"`
	// The number of nodes that are running the daemon pod, but are
	// not supposed to run the daemon pod.
	// More info: https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/
	NumberMisscheduled int32 `protobuf:"varint,2,opt,name=number_misscheduled,json=numberMisscheduled,proto3" json:"number_misscheduled,omitempty"`
	// The total number of nodes that should be running the daemon
	// pod (including nodes correctly running the daemon pod).
	// More info: https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/
	DesiredNumberScheduled int32 `protobuf:"varint,3,opt,name=desired_number_scheduled,json=desiredNumberScheduled,proto3" json:"desired_number_scheduled,omitempty"`
	// numberReady is the number of nodes that should be running the daemon pod and have one
	// or more of the daemon pod running with a Ready Condition.
	NumberReady int32 `protobuf:"varint,4,opt,name=number_ready,json=numberReady,proto3" json:"number_ready,omitempty"`
	// The total number of nodes that are running updated daemon pod
	UpdatedNumberScheduled int32 `protobuf:"varint,6,opt,name=updated_number_scheduled,json=updatedNumberScheduled,proto3" json:"updated_number_scheduled,omitempty"`
	// The number of nodes that should be running the
	// daemon pod and have one or more of the daemon pod running and
	// available (ready for at least spec.minReadySeconds)
	NumberAvailable int32 `protobuf:"varint,7,opt,name=number_available,json=numberAvailable,proto3" json:"number_available,omitempty"`
	// The number of nodes that should be running the
	// daemon pod and have none of the daemon pod running and available
	// (ready for at least spec.minReadySeconds)
	NumberUnavailable int32 `protobuf:"varint,8,opt,name=number_unavailable,json=numberUnavailable,proto3" json:"number_unavailable,omitempty"`
	unknownFields     protoimpl.UnknownFields
	sizeCache         protoimpl.SizeCache
}

func (x *AppDaemonSetStatus) Reset() {
	*x = AppDaemonSetStatus{}
	mi := &file_apps_daemonset_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AppDaemonSetStatus) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AppDaemonSetStatus) ProtoMessage() {}

func (x *AppDaemonSetStatus) ProtoReflect() protoreflect.Message {
	mi := &file_apps_daemonset_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AppDaemonSetStatus.ProtoReflect.Descriptor instead.
func (*AppDaemonSetStatus) Descriptor() ([]byte, []int) {
	return file_apps_daemonset_proto_rawDescGZIP(), []int{3}
}

func (x *AppDaemonSetStatus) GetCurrentNumberScheduled() int32 {
	if x != nil {
		return x.CurrentNumberScheduled
	}
	return 0
}

func (x *AppDaemonSetStatus) GetNumberMisscheduled() int32 {
	if x != nil {
		return x.NumberMisscheduled
	}
	return 0
}

func (x *AppDaemonSetStatus) GetDesiredNumberScheduled() int32 {
	if x != nil {
		return x.DesiredNumberScheduled
	}
	return 0
}

func (x *AppDaemonSetStatus) GetNumberReady() int32 {
	if x != nil {
		return x.NumberReady
	}
	return 0
}

func (x *AppDaemonSetStatus) GetUpdatedNumberScheduled() int32 {
	if x != nil {
		return x.UpdatedNumberScheduled
	}
	return 0
}

func (x *AppDaemonSetStatus) GetNumberAvailable() int32 {
	if x != nil {
		return x.NumberAvailable
	}
	return 0
}

func (x *AppDaemonSetStatus) GetNumberUnavailable() int32 {
	if x != nil {
		return x.NumberUnavailable
	}
	return 0
}

var File_apps_daemonset_proto protoreflect.FileDescriptor

const file_apps_daemonset_proto_rawDesc = "" +
	"\n" +
	"\x14apps_daemonset.proto\x12\agraphpb\x1a\n" +
	"base.proto\x1a\tpod.proto\"\xb1\x01\n" +
	"\x14AppsDaemonSetSummary\x124\n" +
	"\vbase_object\x18\x01 \x01(\v2\x13.graphpb.BaseObjectR\n" +
	"baseObject\x12.\n" +
	"\x04spec\x18\x02 \x01(\v2\x1a.graphpb.AppsDaemonSetSpecR\x04spec\x123\n" +
	"\x06status\x18\x03 \x01(\v2\x1b.graphpb.AppDaemonSetStatusR\x06status\"\x8e\x01\n" +
	"\x11AppsDaemonSetSpec\x12\x1a\n" +
	"\breplicas\x18\x01 \x01(\x05R\breplicas\x12:\n" +
	"\btemplate\x18\x02 \x01(\v2\x1e.graphpb.AppsDaemonSetTemplateR\btemplate\x12!\n" +
	"\fservice_name\x18\x03 \x01(\tR\vserviceName\"u\n" +
	"\x15AppsDaemonSetTemplate\x12/\n" +
	"\bmetadata\x18\x01 \x01(\v2\x13.graphpb.BaseObjectR\bmetadata\x12+\n" +
	"\bpod_spec\x18\x02 \x01(\v2\x10.graphpb.PodSpecR\apodSpec\"\xf0\x02\n" +
	"\x12AppDaemonSetStatus\x128\n" +
	"\x18current_number_scheduled\x18\x01 \x01(\x05R\x16currentNumberScheduled\x12/\n" +
	"\x13number_misscheduled\x18\x02 \x01(\x05R\x12numberMisscheduled\x128\n" +
	"\x18desired_number_scheduled\x18\x03 \x01(\x05R\x16desiredNumberScheduled\x12!\n" +
	"\fnumber_ready\x18\x04 \x01(\x05R\vnumberReady\x128\n" +
	"\x18updated_number_scheduled\x18\x06 \x01(\x05R\x16updatedNumberScheduled\x12)\n" +
	"\x10number_available\x18\a \x01(\x05R\x0fnumberAvailable\x12-\n" +
	"\x12number_unavailable\x18\b \x01(\x05R\x11numberUnavailableB3Z1github.com/cardinalhq/oteltools/pkg/graph/graphpbb\x06proto3"

var (
	file_apps_daemonset_proto_rawDescOnce sync.Once
	file_apps_daemonset_proto_rawDescData []byte
)

func file_apps_daemonset_proto_rawDescGZIP() []byte {
	file_apps_daemonset_proto_rawDescOnce.Do(func() {
		file_apps_daemonset_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_apps_daemonset_proto_rawDesc), len(file_apps_daemonset_proto_rawDesc)))
	})
	return file_apps_daemonset_proto_rawDescData
}

var file_apps_daemonset_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_apps_daemonset_proto_goTypes = []any{
	(*AppsDaemonSetSummary)(nil),  // 0: graphpb.AppsDaemonSetSummary
	(*AppsDaemonSetSpec)(nil),     // 1: graphpb.AppsDaemonSetSpec
	(*AppsDaemonSetTemplate)(nil), // 2: graphpb.AppsDaemonSetTemplate
	(*AppDaemonSetStatus)(nil),    // 3: graphpb.AppDaemonSetStatus
	(*BaseObject)(nil),            // 4: graphpb.BaseObject
	(*PodSpec)(nil),               // 5: graphpb.PodSpec
}
var file_apps_daemonset_proto_depIdxs = []int32{
	4, // 0: graphpb.AppsDaemonSetSummary.base_object:type_name -> graphpb.BaseObject
	1, // 1: graphpb.AppsDaemonSetSummary.spec:type_name -> graphpb.AppsDaemonSetSpec
	3, // 2: graphpb.AppsDaemonSetSummary.status:type_name -> graphpb.AppDaemonSetStatus
	2, // 3: graphpb.AppsDaemonSetSpec.template:type_name -> graphpb.AppsDaemonSetTemplate
	4, // 4: graphpb.AppsDaemonSetTemplate.metadata:type_name -> graphpb.BaseObject
	5, // 5: graphpb.AppsDaemonSetTemplate.pod_spec:type_name -> graphpb.PodSpec
	6, // [6:6] is the sub-list for method output_type
	6, // [6:6] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_apps_daemonset_proto_init() }
func file_apps_daemonset_proto_init() {
	if File_apps_daemonset_proto != nil {
		return
	}
	file_base_proto_init()
	file_pod_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_apps_daemonset_proto_rawDesc), len(file_apps_daemonset_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_apps_daemonset_proto_goTypes,
		DependencyIndexes: file_apps_daemonset_proto_depIdxs,
		MessageInfos:      file_apps_daemonset_proto_msgTypes,
	}.Build()
	File_apps_daemonset_proto = out.File
	file_apps_daemonset_proto_goTypes = nil
	file_apps_daemonset_proto_depIdxs = nil
}
