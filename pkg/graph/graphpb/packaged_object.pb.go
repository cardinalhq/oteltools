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
// 	protoc-gen-go v1.36.5
// 	protoc        v5.29.3
// source: packaged_object.proto

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

// PackagedObject represents an object with arbitrary attributes.
type PackagedObject struct {
	state              protoimpl.MessageState `protogen:"open.v1"`
	ResourceAttributes map[string]string      `protobuf:"bytes,1,rep,name=resource_attributes,json=resourceAttributes,proto3" json:"resource_attributes,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	RecordAttributes   map[string]string      `protobuf:"bytes,2,rep,name=record_attributes,json=recordAttributes,proto3" json:"record_attributes,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	// Types that are valid to be assigned to Object:
	//
	//	*PackagedObject_PodSummary
	//	*PackagedObject_SecretSummary
	//	*PackagedObject_ConfigMapSummary
	//	*PackagedObject_AppsDeploymentSummary
	//	*PackagedObject_AppsStatefulSetSummary
	//	*PackagedObject_AppsDaemonSetSummary
	//	*PackagedObject_AppsReplicaSetSummary
	Object        isPackagedObject_Object `protobuf_oneof:"object"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PackagedObject) Reset() {
	*x = PackagedObject{}
	mi := &file_packaged_object_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PackagedObject) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PackagedObject) ProtoMessage() {}

func (x *PackagedObject) ProtoReflect() protoreflect.Message {
	mi := &file_packaged_object_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PackagedObject.ProtoReflect.Descriptor instead.
func (*PackagedObject) Descriptor() ([]byte, []int) {
	return file_packaged_object_proto_rawDescGZIP(), []int{0}
}

func (x *PackagedObject) GetResourceAttributes() map[string]string {
	if x != nil {
		return x.ResourceAttributes
	}
	return nil
}

func (x *PackagedObject) GetRecordAttributes() map[string]string {
	if x != nil {
		return x.RecordAttributes
	}
	return nil
}

func (x *PackagedObject) GetObject() isPackagedObject_Object {
	if x != nil {
		return x.Object
	}
	return nil
}

func (x *PackagedObject) GetPodSummary() *PodSummary {
	if x != nil {
		if x, ok := x.Object.(*PackagedObject_PodSummary); ok {
			return x.PodSummary
		}
	}
	return nil
}

func (x *PackagedObject) GetSecretSummary() *SecretSummary {
	if x != nil {
		if x, ok := x.Object.(*PackagedObject_SecretSummary); ok {
			return x.SecretSummary
		}
	}
	return nil
}

func (x *PackagedObject) GetConfigMapSummary() *ConfigMapSummary {
	if x != nil {
		if x, ok := x.Object.(*PackagedObject_ConfigMapSummary); ok {
			return x.ConfigMapSummary
		}
	}
	return nil
}

func (x *PackagedObject) GetAppsDeploymentSummary() *AppsDeploymentSummary {
	if x != nil {
		if x, ok := x.Object.(*PackagedObject_AppsDeploymentSummary); ok {
			return x.AppsDeploymentSummary
		}
	}
	return nil
}

func (x *PackagedObject) GetAppsStatefulSetSummary() *AppsStatefulSetSummary {
	if x != nil {
		if x, ok := x.Object.(*PackagedObject_AppsStatefulSetSummary); ok {
			return x.AppsStatefulSetSummary
		}
	}
	return nil
}

func (x *PackagedObject) GetAppsDaemonSetSummary() *AppsDaemonSetSummary {
	if x != nil {
		if x, ok := x.Object.(*PackagedObject_AppsDaemonSetSummary); ok {
			return x.AppsDaemonSetSummary
		}
	}
	return nil
}

func (x *PackagedObject) GetAppsReplicaSetSummary() *AppsReplicaSetSummary {
	if x != nil {
		if x, ok := x.Object.(*PackagedObject_AppsReplicaSetSummary); ok {
			return x.AppsReplicaSetSummary
		}
	}
	return nil
}

type isPackagedObject_Object interface {
	isPackagedObject_Object()
}

type PackagedObject_PodSummary struct {
	PodSummary *PodSummary `protobuf:"bytes,3,opt,name=pod_summary,json=podSummary,proto3,oneof"`
}

type PackagedObject_SecretSummary struct {
	SecretSummary *SecretSummary `protobuf:"bytes,4,opt,name=secret_summary,json=secretSummary,proto3,oneof"`
}

type PackagedObject_ConfigMapSummary struct {
	ConfigMapSummary *ConfigMapSummary `protobuf:"bytes,5,opt,name=config_map_summary,json=configMapSummary,proto3,oneof"`
}

type PackagedObject_AppsDeploymentSummary struct {
	AppsDeploymentSummary *AppsDeploymentSummary `protobuf:"bytes,6,opt,name=apps_deployment_summary,json=appsDeploymentSummary,proto3,oneof"`
}

type PackagedObject_AppsStatefulSetSummary struct {
	AppsStatefulSetSummary *AppsStatefulSetSummary `protobuf:"bytes,7,opt,name=apps_stateful_set_summary,json=appsStatefulSetSummary,proto3,oneof"`
}

type PackagedObject_AppsDaemonSetSummary struct {
	AppsDaemonSetSummary *AppsDaemonSetSummary `protobuf:"bytes,8,opt,name=apps_daemon_set_summary,json=appsDaemonSetSummary,proto3,oneof"`
}

type PackagedObject_AppsReplicaSetSummary struct {
	AppsReplicaSetSummary *AppsReplicaSetSummary `protobuf:"bytes,9,opt,name=apps_replica_set_summary,json=appsReplicaSetSummary,proto3,oneof"`
}

func (*PackagedObject_PodSummary) isPackagedObject_Object() {}

func (*PackagedObject_SecretSummary) isPackagedObject_Object() {}

func (*PackagedObject_ConfigMapSummary) isPackagedObject_Object() {}

func (*PackagedObject_AppsDeploymentSummary) isPackagedObject_Object() {}

func (*PackagedObject_AppsStatefulSetSummary) isPackagedObject_Object() {}

func (*PackagedObject_AppsDaemonSetSummary) isPackagedObject_Object() {}

func (*PackagedObject_AppsReplicaSetSummary) isPackagedObject_Object() {}

type PackagedObjectList struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Items         []*PackagedObject      `protobuf:"bytes,1,rep,name=items,proto3" json:"items,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PackagedObjectList) Reset() {
	*x = PackagedObjectList{}
	mi := &file_packaged_object_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PackagedObjectList) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PackagedObjectList) ProtoMessage() {}

func (x *PackagedObjectList) ProtoReflect() protoreflect.Message {
	mi := &file_packaged_object_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PackagedObjectList.ProtoReflect.Descriptor instead.
func (*PackagedObjectList) Descriptor() ([]byte, []int) {
	return file_packaged_object_proto_rawDescGZIP(), []int{1}
}

func (x *PackagedObjectList) GetItems() []*PackagedObject {
	if x != nil {
		return x.Items
	}
	return nil
}

var File_packaged_object_proto protoreflect.FileDescriptor

var file_packaged_object_proto_rawDesc = string([]byte{
	0x0a, 0x15, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x64, 0x5f, 0x6f, 0x62, 0x6a, 0x65, 0x63,
	0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x67, 0x72, 0x61, 0x70, 0x68, 0x70, 0x62,
	0x1a, 0x14, 0x61, 0x70, 0x70, 0x73, 0x5f, 0x64, 0x61, 0x65, 0x6d, 0x6f, 0x6e, 0x73, 0x65, 0x74,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x15, 0x61, 0x70, 0x70, 0x73, 0x5f, 0x64, 0x65, 0x70,
	0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x15, 0x61,
	0x70, 0x70, 0x73, 0x5f, 0x72, 0x65, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x73, 0x65, 0x74, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x16, 0x61, 0x70, 0x70, 0x73, 0x5f, 0x73, 0x74, 0x61, 0x74, 0x65,
	0x66, 0x75, 0x6c, 0x73, 0x65, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0f, 0x63, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x6d, 0x61, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x09, 0x70,
	0x6f, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0c, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x93, 0x07, 0x0a, 0x0e, 0x50, 0x61, 0x63, 0x6b, 0x61,
	0x67, 0x65, 0x64, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x12, 0x60, 0x0a, 0x13, 0x72, 0x65, 0x73,
	0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f, 0x61, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x73,
	0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2f, 0x2e, 0x67, 0x72, 0x61, 0x70, 0x68, 0x70, 0x62,
	0x2e, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x64, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x2e,
	0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x41, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74,
	0x65, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x12, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63,
	0x65, 0x41, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x73, 0x12, 0x5a, 0x0a, 0x11, 0x72,
	0x65, 0x63, 0x6f, 0x72, 0x64, 0x5f, 0x61, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x73,
	0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2d, 0x2e, 0x67, 0x72, 0x61, 0x70, 0x68, 0x70, 0x62,
	0x2e, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x64, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x2e,
	0x52, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x41, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x73,
	0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x10, 0x72, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x41, 0x74, 0x74,
	0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x73, 0x12, 0x36, 0x0a, 0x0b, 0x70, 0x6f, 0x64, 0x5f, 0x73,
	0x75, 0x6d, 0x6d, 0x61, 0x72, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x67,
	0x72, 0x61, 0x70, 0x68, 0x70, 0x62, 0x2e, 0x50, 0x6f, 0x64, 0x53, 0x75, 0x6d, 0x6d, 0x61, 0x72,
	0x79, 0x48, 0x00, 0x52, 0x0a, 0x70, 0x6f, 0x64, 0x53, 0x75, 0x6d, 0x6d, 0x61, 0x72, 0x79, 0x12,
	0x3f, 0x0a, 0x0e, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74, 0x5f, 0x73, 0x75, 0x6d, 0x6d, 0x61, 0x72,
	0x79, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x67, 0x72, 0x61, 0x70, 0x68, 0x70,
	0x62, 0x2e, 0x53, 0x65, 0x63, 0x72, 0x65, 0x74, 0x53, 0x75, 0x6d, 0x6d, 0x61, 0x72, 0x79, 0x48,
	0x00, 0x52, 0x0d, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74, 0x53, 0x75, 0x6d, 0x6d, 0x61, 0x72, 0x79,
	0x12, 0x49, 0x0a, 0x12, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x5f, 0x6d, 0x61, 0x70, 0x5f, 0x73,
	0x75, 0x6d, 0x6d, 0x61, 0x72, 0x79, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x67,
	0x72, 0x61, 0x70, 0x68, 0x70, 0x62, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x4d, 0x61, 0x70,
	0x53, 0x75, 0x6d, 0x6d, 0x61, 0x72, 0x79, 0x48, 0x00, 0x52, 0x10, 0x63, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x4d, 0x61, 0x70, 0x53, 0x75, 0x6d, 0x6d, 0x61, 0x72, 0x79, 0x12, 0x58, 0x0a, 0x17, 0x61,
	0x70, 0x70, 0x73, 0x5f, 0x64, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x73,
	0x75, 0x6d, 0x6d, 0x61, 0x72, 0x79, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x67,
	0x72, 0x61, 0x70, 0x68, 0x70, 0x62, 0x2e, 0x41, 0x70, 0x70, 0x73, 0x44, 0x65, 0x70, 0x6c, 0x6f,
	0x79, 0x6d, 0x65, 0x6e, 0x74, 0x53, 0x75, 0x6d, 0x6d, 0x61, 0x72, 0x79, 0x48, 0x00, 0x52, 0x15,
	0x61, 0x70, 0x70, 0x73, 0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x53, 0x75,
	0x6d, 0x6d, 0x61, 0x72, 0x79, 0x12, 0x5c, 0x0a, 0x19, 0x61, 0x70, 0x70, 0x73, 0x5f, 0x73, 0x74,
	0x61, 0x74, 0x65, 0x66, 0x75, 0x6c, 0x5f, 0x73, 0x65, 0x74, 0x5f, 0x73, 0x75, 0x6d, 0x6d, 0x61,
	0x72, 0x79, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x67, 0x72, 0x61, 0x70, 0x68,
	0x70, 0x62, 0x2e, 0x41, 0x70, 0x70, 0x73, 0x53, 0x74, 0x61, 0x74, 0x65, 0x66, 0x75, 0x6c, 0x53,
	0x65, 0x74, 0x53, 0x75, 0x6d, 0x6d, 0x61, 0x72, 0x79, 0x48, 0x00, 0x52, 0x16, 0x61, 0x70, 0x70,
	0x73, 0x53, 0x74, 0x61, 0x74, 0x65, 0x66, 0x75, 0x6c, 0x53, 0x65, 0x74, 0x53, 0x75, 0x6d, 0x6d,
	0x61, 0x72, 0x79, 0x12, 0x56, 0x0a, 0x17, 0x61, 0x70, 0x70, 0x73, 0x5f, 0x64, 0x61, 0x65, 0x6d,
	0x6f, 0x6e, 0x5f, 0x73, 0x65, 0x74, 0x5f, 0x73, 0x75, 0x6d, 0x6d, 0x61, 0x72, 0x79, 0x18, 0x08,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x67, 0x72, 0x61, 0x70, 0x68, 0x70, 0x62, 0x2e, 0x41,
	0x70, 0x70, 0x73, 0x44, 0x61, 0x65, 0x6d, 0x6f, 0x6e, 0x53, 0x65, 0x74, 0x53, 0x75, 0x6d, 0x6d,
	0x61, 0x72, 0x79, 0x48, 0x00, 0x52, 0x14, 0x61, 0x70, 0x70, 0x73, 0x44, 0x61, 0x65, 0x6d, 0x6f,
	0x6e, 0x53, 0x65, 0x74, 0x53, 0x75, 0x6d, 0x6d, 0x61, 0x72, 0x79, 0x12, 0x59, 0x0a, 0x18, 0x61,
	0x70, 0x70, 0x73, 0x5f, 0x72, 0x65, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x5f, 0x73, 0x65, 0x74, 0x5f,
	0x73, 0x75, 0x6d, 0x6d, 0x61, 0x72, 0x79, 0x18, 0x09, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1e, 0x2e,
	0x67, 0x72, 0x61, 0x70, 0x68, 0x70, 0x62, 0x2e, 0x41, 0x70, 0x70, 0x73, 0x52, 0x65, 0x70, 0x6c,
	0x69, 0x63, 0x61, 0x53, 0x65, 0x74, 0x53, 0x75, 0x6d, 0x6d, 0x61, 0x72, 0x79, 0x48, 0x00, 0x52,
	0x15, 0x61, 0x70, 0x70, 0x73, 0x52, 0x65, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x53, 0x65, 0x74, 0x53,
	0x75, 0x6d, 0x6d, 0x61, 0x72, 0x79, 0x1a, 0x45, 0x0a, 0x17, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72,
	0x63, 0x65, 0x41, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x73, 0x45, 0x6e, 0x74, 0x72,
	0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03,
	0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x1a, 0x43, 0x0a,
	0x15, 0x52, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x41, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65,
	0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02,
	0x38, 0x01, 0x42, 0x08, 0x0a, 0x06, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x22, 0x43, 0x0a, 0x12,
	0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x64, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x4c, 0x69,
	0x73, 0x74, 0x12, 0x2d, 0x0a, 0x05, 0x69, 0x74, 0x65, 0x6d, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x17, 0x2e, 0x67, 0x72, 0x61, 0x70, 0x68, 0x70, 0x62, 0x2e, 0x50, 0x61, 0x63, 0x6b,
	0x61, 0x67, 0x65, 0x64, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x52, 0x05, 0x69, 0x74, 0x65, 0x6d,
	0x73, 0x42, 0x33, 0x5a, 0x31, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x63, 0x61, 0x72, 0x64, 0x69, 0x6e, 0x61, 0x6c, 0x68, 0x71, 0x2f, 0x6f, 0x74, 0x65, 0x6c, 0x74,
	0x6f, 0x6f, 0x6c, 0x73, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x67, 0x72, 0x61, 0x70, 0x68, 0x2f, 0x67,
	0x72, 0x61, 0x70, 0x68, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
})

var (
	file_packaged_object_proto_rawDescOnce sync.Once
	file_packaged_object_proto_rawDescData []byte
)

func file_packaged_object_proto_rawDescGZIP() []byte {
	file_packaged_object_proto_rawDescOnce.Do(func() {
		file_packaged_object_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_packaged_object_proto_rawDesc), len(file_packaged_object_proto_rawDesc)))
	})
	return file_packaged_object_proto_rawDescData
}

var file_packaged_object_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_packaged_object_proto_goTypes = []any{
	(*PackagedObject)(nil),         // 0: graphpb.PackagedObject
	(*PackagedObjectList)(nil),     // 1: graphpb.PackagedObjectList
	nil,                            // 2: graphpb.PackagedObject.ResourceAttributesEntry
	nil,                            // 3: graphpb.PackagedObject.RecordAttributesEntry
	(*PodSummary)(nil),             // 4: graphpb.PodSummary
	(*SecretSummary)(nil),          // 5: graphpb.SecretSummary
	(*ConfigMapSummary)(nil),       // 6: graphpb.ConfigMapSummary
	(*AppsDeploymentSummary)(nil),  // 7: graphpb.AppsDeploymentSummary
	(*AppsStatefulSetSummary)(nil), // 8: graphpb.AppsStatefulSetSummary
	(*AppsDaemonSetSummary)(nil),   // 9: graphpb.AppsDaemonSetSummary
	(*AppsReplicaSetSummary)(nil),  // 10: graphpb.AppsReplicaSetSummary
}
var file_packaged_object_proto_depIdxs = []int32{
	2,  // 0: graphpb.PackagedObject.resource_attributes:type_name -> graphpb.PackagedObject.ResourceAttributesEntry
	3,  // 1: graphpb.PackagedObject.record_attributes:type_name -> graphpb.PackagedObject.RecordAttributesEntry
	4,  // 2: graphpb.PackagedObject.pod_summary:type_name -> graphpb.PodSummary
	5,  // 3: graphpb.PackagedObject.secret_summary:type_name -> graphpb.SecretSummary
	6,  // 4: graphpb.PackagedObject.config_map_summary:type_name -> graphpb.ConfigMapSummary
	7,  // 5: graphpb.PackagedObject.apps_deployment_summary:type_name -> graphpb.AppsDeploymentSummary
	8,  // 6: graphpb.PackagedObject.apps_stateful_set_summary:type_name -> graphpb.AppsStatefulSetSummary
	9,  // 7: graphpb.PackagedObject.apps_daemon_set_summary:type_name -> graphpb.AppsDaemonSetSummary
	10, // 8: graphpb.PackagedObject.apps_replica_set_summary:type_name -> graphpb.AppsReplicaSetSummary
	0,  // 9: graphpb.PackagedObjectList.items:type_name -> graphpb.PackagedObject
	10, // [10:10] is the sub-list for method output_type
	10, // [10:10] is the sub-list for method input_type
	10, // [10:10] is the sub-list for extension type_name
	10, // [10:10] is the sub-list for extension extendee
	0,  // [0:10] is the sub-list for field type_name
}

func init() { file_packaged_object_proto_init() }
func file_packaged_object_proto_init() {
	if File_packaged_object_proto != nil {
		return
	}
	file_apps_daemonset_proto_init()
	file_apps_deployment_proto_init()
	file_apps_replicaset_proto_init()
	file_apps_statefulset_proto_init()
	file_configmap_proto_init()
	file_pod_proto_init()
	file_secret_proto_init()
	file_packaged_object_proto_msgTypes[0].OneofWrappers = []any{
		(*PackagedObject_PodSummary)(nil),
		(*PackagedObject_SecretSummary)(nil),
		(*PackagedObject_ConfigMapSummary)(nil),
		(*PackagedObject_AppsDeploymentSummary)(nil),
		(*PackagedObject_AppsStatefulSetSummary)(nil),
		(*PackagedObject_AppsDaemonSetSummary)(nil),
		(*PackagedObject_AppsReplicaSetSummary)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_packaged_object_proto_rawDesc), len(file_packaged_object_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_packaged_object_proto_goTypes,
		DependencyIndexes: file_packaged_object_proto_depIdxs,
		MessageInfos:      file_packaged_object_proto_msgTypes,
	}.Build()
	File_packaged_object_proto = out.File
	file_packaged_object_proto_goTypes = nil
	file_packaged_object_proto_depIdxs = nil
}
