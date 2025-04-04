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
// source: secret.proto

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

// SecretSummary summarizes a Kubernetes Secret.
type SecretSummary struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	BaseObject    *BaseObject            `protobuf:"bytes,1,opt,name=base_object,json=baseObject,proto3" json:"base_object,omitempty"`
	Type          string                 `protobuf:"bytes,2,opt,name=type,proto3" json:"type,omitempty"`
	Hashes        map[string]string      `protobuf:"bytes,3,rep,name=hashes,proto3" json:"hashes,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SecretSummary) Reset() {
	*x = SecretSummary{}
	mi := &file_secret_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SecretSummary) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SecretSummary) ProtoMessage() {}

func (x *SecretSummary) ProtoReflect() protoreflect.Message {
	mi := &file_secret_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SecretSummary.ProtoReflect.Descriptor instead.
func (*SecretSummary) Descriptor() ([]byte, []int) {
	return file_secret_proto_rawDescGZIP(), []int{0}
}

func (x *SecretSummary) GetBaseObject() *BaseObject {
	if x != nil {
		return x.BaseObject
	}
	return nil
}

func (x *SecretSummary) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *SecretSummary) GetHashes() map[string]string {
	if x != nil {
		return x.Hashes
	}
	return nil
}

var File_secret_proto protoreflect.FileDescriptor

const file_secret_proto_rawDesc = "" +
	"\n" +
	"\fsecret.proto\x12\agraphpb\x1a\n" +
	"base.proto\"\xd0\x01\n" +
	"\rSecretSummary\x124\n" +
	"\vbase_object\x18\x01 \x01(\v2\x13.graphpb.BaseObjectR\n" +
	"baseObject\x12\x12\n" +
	"\x04type\x18\x02 \x01(\tR\x04type\x12:\n" +
	"\x06hashes\x18\x03 \x03(\v2\".graphpb.SecretSummary.HashesEntryR\x06hashes\x1a9\n" +
	"\vHashesEntry\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n" +
	"\x05value\x18\x02 \x01(\tR\x05value:\x028\x01B3Z1github.com/cardinalhq/oteltools/pkg/graph/graphpbb\x06proto3"

var (
	file_secret_proto_rawDescOnce sync.Once
	file_secret_proto_rawDescData []byte
)

func file_secret_proto_rawDescGZIP() []byte {
	file_secret_proto_rawDescOnce.Do(func() {
		file_secret_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_secret_proto_rawDesc), len(file_secret_proto_rawDesc)))
	})
	return file_secret_proto_rawDescData
}

var file_secret_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_secret_proto_goTypes = []any{
	(*SecretSummary)(nil), // 0: graphpb.SecretSummary
	nil,                   // 1: graphpb.SecretSummary.HashesEntry
	(*BaseObject)(nil),    // 2: graphpb.BaseObject
}
var file_secret_proto_depIdxs = []int32{
	2, // 0: graphpb.SecretSummary.base_object:type_name -> graphpb.BaseObject
	1, // 1: graphpb.SecretSummary.hashes:type_name -> graphpb.SecretSummary.HashesEntry
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_secret_proto_init() }
func file_secret_proto_init() {
	if File_secret_proto != nil {
		return
	}
	file_base_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_secret_proto_rawDesc), len(file_secret_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_secret_proto_goTypes,
		DependencyIndexes: file_secret_proto_depIdxs,
		MessageInfos:      file_secret_proto_msgTypes,
	}.Build()
	File_secret_proto = out.File
	file_secret_proto_goTypes = nil
	file_secret_proto_depIdxs = nil
}
