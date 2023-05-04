// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.20.0
// source: api/system/v1alpha1/inter_cp_ping.proto

package v1alpha1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type PingRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	InstanceId  string `protobuf:"bytes,1,opt,name=instance_id,json=instanceId,proto3" json:"instance_id,omitempty"`
	Address     string `protobuf:"bytes,2,opt,name=address,proto3" json:"address,omitempty"`
	InterCpPort uint32 `protobuf:"varint,3,opt,name=inter_cp_port,json=interCpPort,proto3" json:"inter_cp_port,omitempty"`
	Ready       bool   `protobuf:"varint,4,opt,name=ready,proto3" json:"ready,omitempty"`
}

func (x *PingRequest) Reset() {
	*x = PingRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_system_v1alpha1_inter_cp_ping_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PingRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PingRequest) ProtoMessage() {}

func (x *PingRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_system_v1alpha1_inter_cp_ping_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PingRequest.ProtoReflect.Descriptor instead.
func (*PingRequest) Descriptor() ([]byte, []int) {
	return file_api_system_v1alpha1_inter_cp_ping_proto_rawDescGZIP(), []int{0}
}

func (x *PingRequest) GetInstanceId() string {
	if x != nil {
		return x.InstanceId
	}
	return ""
}

func (x *PingRequest) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *PingRequest) GetInterCpPort() uint32 {
	if x != nil {
		return x.InterCpPort
	}
	return 0
}

func (x *PingRequest) GetReady() bool {
	if x != nil {
		return x.Ready
	}
	return false
}

type PingResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Leader bool `protobuf:"varint,1,opt,name=leader,proto3" json:"leader,omitempty"`
}

func (x *PingResponse) Reset() {
	*x = PingResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_system_v1alpha1_inter_cp_ping_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PingResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PingResponse) ProtoMessage() {}

func (x *PingResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_system_v1alpha1_inter_cp_ping_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PingResponse.ProtoReflect.Descriptor instead.
func (*PingResponse) Descriptor() ([]byte, []int) {
	return file_api_system_v1alpha1_inter_cp_ping_proto_rawDescGZIP(), []int{1}
}

func (x *PingResponse) GetLeader() bool {
	if x != nil {
		return x.Leader
	}
	return false
}

var File_api_system_v1alpha1_inter_cp_ping_proto protoreflect.FileDescriptor

var file_api_system_v1alpha1_inter_cp_ping_proto_rawDesc = []byte{
	0x0a, 0x27, 0x61, 0x70, 0x69, 0x2f, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x2f, 0x76, 0x31, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x5f, 0x63, 0x70, 0x5f, 0x70,
	0x69, 0x6e, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x14, 0x6b, 0x75, 0x6d, 0x61, 0x2e,
	0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x22,
	0x82, 0x01, 0x0a, 0x0b, 0x50, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x1f, 0x0a, 0x0b, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x49, 0x64,
	0x12, 0x18, 0x0a, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x22, 0x0a, 0x0d, 0x69, 0x6e,
	0x74, 0x65, 0x72, 0x5f, 0x63, 0x70, 0x5f, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x0d, 0x52, 0x0b, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x43, 0x70, 0x50, 0x6f, 0x72, 0x74, 0x12, 0x14,
	0x0a, 0x05, 0x72, 0x65, 0x61, 0x64, 0x79, 0x18, 0x04, 0x20, 0x01, 0x28, 0x08, 0x52, 0x05, 0x72,
	0x65, 0x61, 0x64, 0x79, 0x22, 0x26, 0x0a, 0x0c, 0x50, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x6c, 0x65, 0x61, 0x64, 0x65, 0x72, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x6c, 0x65, 0x61, 0x64, 0x65, 0x72, 0x32, 0x63, 0x0a, 0x12,
	0x49, 0x6e, 0x74, 0x65, 0x72, 0x43, 0x70, 0x50, 0x69, 0x6e, 0x67, 0x53, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x12, 0x4d, 0x0a, 0x04, 0x50, 0x69, 0x6e, 0x67, 0x12, 0x21, 0x2e, 0x6b, 0x75, 0x6d,
	0x61, 0x2e, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61,
	0x31, 0x2e, 0x50, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x22, 0x2e,
	0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x2e, 0x76, 0x31, 0x61, 0x6c,
	0x70, 0x68, 0x61, 0x31, 0x2e, 0x50, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x42, 0x2c, 0x5a, 0x2a, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x6b, 0x75, 0x6d, 0x61, 0x68, 0x71, 0x2f, 0x6b, 0x75, 0x6d, 0x61, 0x2f, 0x61, 0x70, 0x69, 0x2f,
	0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_system_v1alpha1_inter_cp_ping_proto_rawDescOnce sync.Once
	file_api_system_v1alpha1_inter_cp_ping_proto_rawDescData = file_api_system_v1alpha1_inter_cp_ping_proto_rawDesc
)

func file_api_system_v1alpha1_inter_cp_ping_proto_rawDescGZIP() []byte {
	file_api_system_v1alpha1_inter_cp_ping_proto_rawDescOnce.Do(func() {
		file_api_system_v1alpha1_inter_cp_ping_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_system_v1alpha1_inter_cp_ping_proto_rawDescData)
	})
	return file_api_system_v1alpha1_inter_cp_ping_proto_rawDescData
}

var file_api_system_v1alpha1_inter_cp_ping_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_api_system_v1alpha1_inter_cp_ping_proto_goTypes = []interface{}{
	(*PingRequest)(nil),  // 0: kuma.system.v1alpha1.PingRequest
	(*PingResponse)(nil), // 1: kuma.system.v1alpha1.PingResponse
}
var file_api_system_v1alpha1_inter_cp_ping_proto_depIdxs = []int32{
	0, // 0: kuma.system.v1alpha1.InterCpPingService.Ping:input_type -> kuma.system.v1alpha1.PingRequest
	1, // 1: kuma.system.v1alpha1.InterCpPingService.Ping:output_type -> kuma.system.v1alpha1.PingResponse
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_api_system_v1alpha1_inter_cp_ping_proto_init() }
func file_api_system_v1alpha1_inter_cp_ping_proto_init() {
	if File_api_system_v1alpha1_inter_cp_ping_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_api_system_v1alpha1_inter_cp_ping_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PingRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_system_v1alpha1_inter_cp_ping_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PingResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_api_system_v1alpha1_inter_cp_ping_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_api_system_v1alpha1_inter_cp_ping_proto_goTypes,
		DependencyIndexes: file_api_system_v1alpha1_inter_cp_ping_proto_depIdxs,
		MessageInfos:      file_api_system_v1alpha1_inter_cp_ping_proto_msgTypes,
	}.Build()
	File_api_system_v1alpha1_inter_cp_ping_proto = out.File
	file_api_system_v1alpha1_inter_cp_ping_proto_rawDesc = nil
	file_api_system_v1alpha1_inter_cp_ping_proto_goTypes = nil
	file_api_system_v1alpha1_inter_cp_ping_proto_depIdxs = nil
}
