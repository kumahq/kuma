// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.14.0
// source: mesh/v1alpha1/traffic_permission.proto

package v1alpha1

import (
	_ "github.com/kumahq/kuma/api/mesh"
	_ "github.com/kumahq/protoc-gen-kumadoc/proto"
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

// TrafficPermission defines permission for traffic between dataplanes.
type TrafficPermission struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// List of selectors to match dataplanes that are sources of traffic.
	Sources []*Selector `protobuf:"bytes,1,rep,name=sources,proto3" json:"sources,omitempty"`
	// List of selectors to match services that are destinations of traffic.
	Destinations []*Selector `protobuf:"bytes,2,rep,name=destinations,proto3" json:"destinations,omitempty"`
}

func (x *TrafficPermission) Reset() {
	*x = TrafficPermission{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mesh_v1alpha1_traffic_permission_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TrafficPermission) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TrafficPermission) ProtoMessage() {}

func (x *TrafficPermission) ProtoReflect() protoreflect.Message {
	mi := &file_mesh_v1alpha1_traffic_permission_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TrafficPermission.ProtoReflect.Descriptor instead.
func (*TrafficPermission) Descriptor() ([]byte, []int) {
	return file_mesh_v1alpha1_traffic_permission_proto_rawDescGZIP(), []int{0}
}

func (x *TrafficPermission) GetSources() []*Selector {
	if x != nil {
		return x.Sources
	}
	return nil
}

func (x *TrafficPermission) GetDestinations() []*Selector {
	if x != nil {
		return x.Destinations
	}
	return nil
}

var File_mesh_v1alpha1_traffic_permission_proto protoreflect.FileDescriptor

var file_mesh_v1alpha1_traffic_permission_proto_rawDesc = []byte{
	0x0a, 0x26, 0x6d, 0x65, 0x73, 0x68, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f,
	0x74, 0x72, 0x61, 0x66, 0x66, 0x69, 0x63, 0x5f, 0x70, 0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69,
	0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x12, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x6d,
	0x65, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x1a, 0x12, 0x6d, 0x65,
	0x73, 0x68, 0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x1c, 0x6d, 0x65, 0x73, 0x68, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f,
	0x73, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0c,
	0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xf1, 0x01, 0x0a,
	0x11, 0x54, 0x72, 0x61, 0x66, 0x66, 0x69, 0x63, 0x50, 0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69,
	0x6f, 0x6e, 0x12, 0x36, 0x0a, 0x07, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x18, 0x01, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e,
	0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x53, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f,
	0x72, 0x52, 0x07, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x12, 0x40, 0x0a, 0x0c, 0x64, 0x65,
	0x73, 0x74, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x1c, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x53, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x52, 0x0c,
	0x64, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x3a, 0x62, 0xaa, 0x8c,
	0x89, 0xa6, 0x01, 0x1b, 0x0a, 0x19, 0x54, 0x72, 0x61, 0x66, 0x66, 0x69, 0x63, 0x50, 0x65, 0x72,
	0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0xaa,
	0x8c, 0x89, 0xa6, 0x01, 0x13, 0x12, 0x11, 0x54, 0x72, 0x61, 0x66, 0x66, 0x69, 0x63, 0x50, 0x65,
	0x72, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0xaa, 0x8c, 0x89, 0xa6, 0x01, 0x06, 0x22, 0x04,
	0x6d, 0x65, 0x73, 0x68, 0xaa, 0x8c, 0x89, 0xa6, 0x01, 0x16, 0x3a, 0x14, 0x0a, 0x12, 0x74, 0x72,
	0x61, 0x66, 0x66, 0x69, 0x63, 0x2d, 0x70, 0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e,
	0x42, 0x5b, 0x5a, 0x28, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6b,
	0x75, 0x6d, 0x61, 0x68, 0x71, 0x2f, 0x6b, 0x75, 0x6d, 0x61, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x6d,
	0x65, 0x73, 0x68, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x8a, 0xb5, 0x18, 0x2d,
	0x50, 0x01, 0xa2, 0x01, 0x12, 0x54, 0x72, 0x61, 0x66, 0x66, 0x69, 0x63, 0x50, 0x65, 0x72, 0x6d,
	0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0xf2, 0x01, 0x13, 0x74, 0x72, 0x61, 0x66, 0x66, 0x69,
	0x63, 0x2d, 0x70, 0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_mesh_v1alpha1_traffic_permission_proto_rawDescOnce sync.Once
	file_mesh_v1alpha1_traffic_permission_proto_rawDescData = file_mesh_v1alpha1_traffic_permission_proto_rawDesc
)

func file_mesh_v1alpha1_traffic_permission_proto_rawDescGZIP() []byte {
	file_mesh_v1alpha1_traffic_permission_proto_rawDescOnce.Do(func() {
		file_mesh_v1alpha1_traffic_permission_proto_rawDescData = protoimpl.X.CompressGZIP(file_mesh_v1alpha1_traffic_permission_proto_rawDescData)
	})
	return file_mesh_v1alpha1_traffic_permission_proto_rawDescData
}

var file_mesh_v1alpha1_traffic_permission_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_mesh_v1alpha1_traffic_permission_proto_goTypes = []interface{}{
	(*TrafficPermission)(nil), // 0: kuma.mesh.v1alpha1.TrafficPermission
	(*Selector)(nil),          // 1: kuma.mesh.v1alpha1.Selector
}
var file_mesh_v1alpha1_traffic_permission_proto_depIdxs = []int32{
	1, // 0: kuma.mesh.v1alpha1.TrafficPermission.sources:type_name -> kuma.mesh.v1alpha1.Selector
	1, // 1: kuma.mesh.v1alpha1.TrafficPermission.destinations:type_name -> kuma.mesh.v1alpha1.Selector
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_mesh_v1alpha1_traffic_permission_proto_init() }
func file_mesh_v1alpha1_traffic_permission_proto_init() {
	if File_mesh_v1alpha1_traffic_permission_proto != nil {
		return
	}
	file_mesh_v1alpha1_selector_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_mesh_v1alpha1_traffic_permission_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TrafficPermission); i {
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
			RawDescriptor: file_mesh_v1alpha1_traffic_permission_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_mesh_v1alpha1_traffic_permission_proto_goTypes,
		DependencyIndexes: file_mesh_v1alpha1_traffic_permission_proto_depIdxs,
		MessageInfos:      file_mesh_v1alpha1_traffic_permission_proto_msgTypes,
	}.Build()
	File_mesh_v1alpha1_traffic_permission_proto = out.File
	file_mesh_v1alpha1_traffic_permission_proto_rawDesc = nil
	file_mesh_v1alpha1_traffic_permission_proto_goTypes = nil
	file_mesh_v1alpha1_traffic_permission_proto_depIdxs = nil
}
