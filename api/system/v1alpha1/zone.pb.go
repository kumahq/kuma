// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.20.0
// source: api/system/v1alpha1/zone.proto

package v1alpha1

import (
	_ "github.com/kumahq/kuma/api/mesh"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Zone defines the Zone configuration used at the Global Control Plane
// within a distributed deployment
type Zone struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// enable allows to turn the zone on/off and exclude the whole zone from
	// balancing traffic on it
	Enabled *wrapperspb.BoolValue `protobuf:"bytes,1,opt,name=enabled,proto3" json:"enabled,omitempty"`
}

func (x *Zone) Reset() {
	*x = Zone{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_system_v1alpha1_zone_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Zone) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Zone) ProtoMessage() {}

func (x *Zone) ProtoReflect() protoreflect.Message {
	mi := &file_api_system_v1alpha1_zone_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Zone.ProtoReflect.Descriptor instead.
func (*Zone) Descriptor() ([]byte, []int) {
	return file_api_system_v1alpha1_zone_proto_rawDescGZIP(), []int{0}
}

func (x *Zone) GetEnabled() *wrapperspb.BoolValue {
	if x != nil {
		return x.Enabled
	}
	return nil
}

var File_api_system_v1alpha1_zone_proto protoreflect.FileDescriptor

var file_api_system_v1alpha1_zone_proto_rawDesc = []byte{
	0x0a, 0x1e, 0x61, 0x70, 0x69, 0x2f, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x2f, 0x76, 0x31, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x7a, 0x6f, 0x6e, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x14, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x2e, 0x76, 0x31,
	0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x1a, 0x16, 0x61, 0x70, 0x69, 0x2f, 0x6d, 0x65, 0x73, 0x68,
	0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f,
	0x77, 0x72, 0x61, 0x70, 0x70, 0x65, 0x72, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x82,
	0x01, 0x0a, 0x04, 0x5a, 0x6f, 0x6e, 0x65, 0x12, 0x34, 0x0a, 0x07, 0x65, 0x6e, 0x61, 0x62, 0x6c,
	0x65, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x42, 0x6f, 0x6f, 0x6c, 0x56,
	0x61, 0x6c, 0x75, 0x65, 0x52, 0x07, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x3a, 0x44, 0xaa,
	0x8c, 0x89, 0xa6, 0x01, 0x0e, 0x0a, 0x0c, 0x5a, 0x6f, 0x6e, 0x65, 0x52, 0x65, 0x73, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0xaa, 0x8c, 0x89, 0xa6, 0x01, 0x06, 0x12, 0x04, 0x5a, 0x6f, 0x6e, 0x65, 0xaa,
	0x8c, 0x89, 0xa6, 0x01, 0x08, 0x22, 0x06, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0xaa, 0x8c, 0x89,
	0xa6, 0x01, 0x02, 0x18, 0x01, 0xaa, 0x8c, 0x89, 0xa6, 0x01, 0x08, 0x3a, 0x06, 0x0a, 0x04, 0x7a,
	0x6f, 0x6e, 0x65, 0x42, 0x2c, 0x5a, 0x2a, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x6b, 0x75, 0x6d, 0x61, 0x68, 0x71, 0x2f, 0x6b, 0x75, 0x6d, 0x61, 0x2f, 0x61, 0x70,
	0x69, 0x2f, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61,
	0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_system_v1alpha1_zone_proto_rawDescOnce sync.Once
	file_api_system_v1alpha1_zone_proto_rawDescData = file_api_system_v1alpha1_zone_proto_rawDesc
)

func file_api_system_v1alpha1_zone_proto_rawDescGZIP() []byte {
	file_api_system_v1alpha1_zone_proto_rawDescOnce.Do(func() {
		file_api_system_v1alpha1_zone_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_system_v1alpha1_zone_proto_rawDescData)
	})
	return file_api_system_v1alpha1_zone_proto_rawDescData
}

var file_api_system_v1alpha1_zone_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_api_system_v1alpha1_zone_proto_goTypes = []interface{}{
	(*Zone)(nil),                 // 0: kuma.system.v1alpha1.Zone
	(*wrapperspb.BoolValue)(nil), // 1: google.protobuf.BoolValue
}
var file_api_system_v1alpha1_zone_proto_depIdxs = []int32{
	1, // 0: kuma.system.v1alpha1.Zone.enabled:type_name -> google.protobuf.BoolValue
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_api_system_v1alpha1_zone_proto_init() }
func file_api_system_v1alpha1_zone_proto_init() {
	if File_api_system_v1alpha1_zone_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_api_system_v1alpha1_zone_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Zone); i {
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
			RawDescriptor: file_api_system_v1alpha1_zone_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_system_v1alpha1_zone_proto_goTypes,
		DependencyIndexes: file_api_system_v1alpha1_zone_proto_depIdxs,
		MessageInfos:      file_api_system_v1alpha1_zone_proto_msgTypes,
	}.Build()
	File_api_system_v1alpha1_zone_proto = out.File
	file_api_system_v1alpha1_zone_proto_rawDesc = nil
	file_api_system_v1alpha1_zone_proto_goTypes = nil
	file_api_system_v1alpha1_zone_proto_depIdxs = nil
}
