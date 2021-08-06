// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.14.0
// source: mesh/v1alpha1/zone_ingress_insight.proto

package v1alpha1

import (
	_ "github.com/kumahq/kuma/api/mesh"
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

// ZoneIngressInsight defines the observed state of a Zone Ingress.
type ZoneIngressInsight struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// List of ADS subscriptions created by a given Zone Kuma CP.
	Subscriptions []*DiscoverySubscription `protobuf:"bytes,1,rep,name=subscriptions,proto3" json:"subscriptions,omitempty"`
}

func (x *ZoneIngressInsight) Reset() {
	*x = ZoneIngressInsight{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mesh_v1alpha1_zone_ingress_insight_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ZoneIngressInsight) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ZoneIngressInsight) ProtoMessage() {}

func (x *ZoneIngressInsight) ProtoReflect() protoreflect.Message {
	mi := &file_mesh_v1alpha1_zone_ingress_insight_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ZoneIngressInsight.ProtoReflect.Descriptor instead.
func (*ZoneIngressInsight) Descriptor() ([]byte, []int) {
	return file_mesh_v1alpha1_zone_ingress_insight_proto_rawDescGZIP(), []int{0}
}

func (x *ZoneIngressInsight) GetSubscriptions() []*DiscoverySubscription {
	if x != nil {
		return x.Subscriptions
	}
	return nil
}

var File_mesh_v1alpha1_zone_ingress_insight_proto protoreflect.FileDescriptor

var file_mesh_v1alpha1_zone_ingress_insight_proto_rawDesc = []byte{
	0x0a, 0x28, 0x6d, 0x65, 0x73, 0x68, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f,
	0x7a, 0x6f, 0x6e, 0x65, 0x5f, 0x69, 0x6e, 0x67, 0x72, 0x65, 0x73, 0x73, 0x5f, 0x69, 0x6e, 0x73,
	0x69, 0x67, 0x68, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x12, 0x6b, 0x75, 0x6d, 0x61,
	0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x1a, 0x12,
	0x6d, 0x65, 0x73, 0x68, 0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x25, 0x6d, 0x65, 0x73, 0x68, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61,
	0x31, 0x2f, 0x64, 0x61, 0x74, 0x61, 0x70, 0x6c, 0x61, 0x6e, 0x65, 0x5f, 0x69, 0x6e, 0x73, 0x69,
	0x67, 0x68, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xf2, 0x01, 0x0a, 0x12, 0x5a, 0x6f,
	0x6e, 0x65, 0x49, 0x6e, 0x67, 0x72, 0x65, 0x73, 0x73, 0x49, 0x6e, 0x73, 0x69, 0x67, 0x68, 0x74,
	0x12, 0x4f, 0x0a, 0x0d, 0x73, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x29, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x6d,
	0x65, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x44, 0x69, 0x73,
	0x63, 0x6f, 0x76, 0x65, 0x72, 0x79, 0x53, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69,
	0x6f, 0x6e, 0x52, 0x0d, 0x73, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x3a, 0x8a, 0x01, 0xaa, 0x8c, 0x89, 0xa6, 0x01, 0x1c, 0x0a, 0x1a, 0x5a, 0x6f, 0x6e, 0x65,
	0x49, 0x6e, 0x67, 0x72, 0x65, 0x73, 0x73, 0x49, 0x6e, 0x73, 0x69, 0x67, 0x68, 0x74, 0x52, 0x65,
	0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0xaa, 0x8c, 0x89, 0xa6, 0x01, 0x14, 0x12, 0x12, 0x5a, 0x6f,
	0x6e, 0x65, 0x49, 0x6e, 0x67, 0x72, 0x65, 0x73, 0x73, 0x49, 0x6e, 0x73, 0x69, 0x67, 0x68, 0x74,
	0xaa, 0x8c, 0x89, 0xa6, 0x01, 0x06, 0x22, 0x04, 0x6d, 0x65, 0x73, 0x68, 0xaa, 0x8c, 0x89, 0xa6,
	0x01, 0x02, 0x18, 0x01, 0xaa, 0x8c, 0x89, 0xa6, 0x01, 0x02, 0x28, 0x01, 0xaa, 0x8c, 0x89, 0xa6,
	0x01, 0x04, 0x52, 0x02, 0x08, 0x01, 0xaa, 0x8c, 0x89, 0xa6, 0x01, 0x18, 0x3a, 0x16, 0x0a, 0x14,
	0x7a, 0x6f, 0x6e, 0x65, 0x2d, 0x69, 0x6e, 0x67, 0x72, 0x65, 0x73, 0x73, 0x2d, 0x69, 0x6e, 0x73,
	0x69, 0x67, 0x68, 0x74, 0xaa, 0x8c, 0x89, 0xa6, 0x01, 0x04, 0x3a, 0x02, 0x18, 0x01, 0x42, 0x2a,
	0x5a, 0x28, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6b, 0x75, 0x6d,
	0x61, 0x68, 0x71, 0x2f, 0x6b, 0x75, 0x6d, 0x61, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x6d, 0x65, 0x73,
	0x68, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_mesh_v1alpha1_zone_ingress_insight_proto_rawDescOnce sync.Once
	file_mesh_v1alpha1_zone_ingress_insight_proto_rawDescData = file_mesh_v1alpha1_zone_ingress_insight_proto_rawDesc
)

func file_mesh_v1alpha1_zone_ingress_insight_proto_rawDescGZIP() []byte {
	file_mesh_v1alpha1_zone_ingress_insight_proto_rawDescOnce.Do(func() {
		file_mesh_v1alpha1_zone_ingress_insight_proto_rawDescData = protoimpl.X.CompressGZIP(file_mesh_v1alpha1_zone_ingress_insight_proto_rawDescData)
	})
	return file_mesh_v1alpha1_zone_ingress_insight_proto_rawDescData
}

var file_mesh_v1alpha1_zone_ingress_insight_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_mesh_v1alpha1_zone_ingress_insight_proto_goTypes = []interface{}{
	(*ZoneIngressInsight)(nil),    // 0: kuma.mesh.v1alpha1.ZoneIngressInsight
	(*DiscoverySubscription)(nil), // 1: kuma.mesh.v1alpha1.DiscoverySubscription
}
var file_mesh_v1alpha1_zone_ingress_insight_proto_depIdxs = []int32{
	1, // 0: kuma.mesh.v1alpha1.ZoneIngressInsight.subscriptions:type_name -> kuma.mesh.v1alpha1.DiscoverySubscription
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_mesh_v1alpha1_zone_ingress_insight_proto_init() }
func file_mesh_v1alpha1_zone_ingress_insight_proto_init() {
	if File_mesh_v1alpha1_zone_ingress_insight_proto != nil {
		return
	}
	file_mesh_v1alpha1_dataplane_insight_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_mesh_v1alpha1_zone_ingress_insight_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ZoneIngressInsight); i {
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
			RawDescriptor: file_mesh_v1alpha1_zone_ingress_insight_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_mesh_v1alpha1_zone_ingress_insight_proto_goTypes,
		DependencyIndexes: file_mesh_v1alpha1_zone_ingress_insight_proto_depIdxs,
		MessageInfos:      file_mesh_v1alpha1_zone_ingress_insight_proto_msgTypes,
	}.Build()
	File_mesh_v1alpha1_zone_ingress_insight_proto = out.File
	file_mesh_v1alpha1_zone_ingress_insight_proto_rawDesc = nil
	file_mesh_v1alpha1_zone_ingress_insight_proto_goTypes = nil
	file_mesh_v1alpha1_zone_ingress_insight_proto_depIdxs = nil
}
