// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.20.0
// source: api/system/v1alpha1/config.proto

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

// Config is a entity that represents dynamic configuration that is stored in
// underlying storage. For now it's used only for internal mechanisms.
type Config struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// configuration that is stored (ex. in JSON)
	Config string `protobuf:"bytes,1,opt,name=config,proto3" json:"config,omitempty"`
}

func (x *Config) Reset() {
	*x = Config{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_system_v1alpha1_config_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_api_system_v1alpha1_config_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Config.ProtoReflect.Descriptor instead.
func (*Config) Descriptor() ([]byte, []int) {
	return file_api_system_v1alpha1_config_proto_rawDescGZIP(), []int{0}
}

func (x *Config) GetConfig() string {
	if x != nil {
		return x.Config
	}
	return ""
}

var File_api_system_v1alpha1_config_proto protoreflect.FileDescriptor

var file_api_system_v1alpha1_config_proto_rawDesc = []byte{
	0x0a, 0x20, 0x61, 0x70, 0x69, 0x2f, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x2f, 0x76, 0x31, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x14, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x2e,
	0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x1a, 0x16, 0x61, 0x70, 0x69, 0x2f, 0x6d, 0x65,
	0x73, 0x68, 0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0x6e, 0x0a, 0x06, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x16, 0x0a, 0x06, 0x63, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x63, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x3a, 0x4c, 0xaa, 0x8c, 0x89, 0xa6, 0x01, 0x10, 0x0a, 0x0e, 0x43, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0xaa, 0x8c, 0x89, 0xa6, 0x01, 0x08,
	0x12, 0x06, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0xaa, 0x8c, 0x89, 0xa6, 0x01, 0x08, 0x22, 0x06,
	0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0xaa, 0x8c, 0x89, 0xa6, 0x01, 0x02, 0x18, 0x01, 0xaa, 0x8c,
	0x89, 0xa6, 0x01, 0x02, 0x60, 0x01, 0xaa, 0x8c, 0x89, 0xa6, 0x01, 0x04, 0x52, 0x02, 0x10, 0x01,
	0x42, 0x2c, 0x5a, 0x2a, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6b,
	0x75, 0x6d, 0x61, 0x68, 0x71, 0x2f, 0x6b, 0x75, 0x6d, 0x61, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x73,
	0x79, 0x73, 0x74, 0x65, 0x6d, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_system_v1alpha1_config_proto_rawDescOnce sync.Once
	file_api_system_v1alpha1_config_proto_rawDescData = file_api_system_v1alpha1_config_proto_rawDesc
)

func file_api_system_v1alpha1_config_proto_rawDescGZIP() []byte {
	file_api_system_v1alpha1_config_proto_rawDescOnce.Do(func() {
		file_api_system_v1alpha1_config_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_system_v1alpha1_config_proto_rawDescData)
	})
	return file_api_system_v1alpha1_config_proto_rawDescData
}

var file_api_system_v1alpha1_config_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_api_system_v1alpha1_config_proto_goTypes = []interface{}{
	(*Config)(nil), // 0: kuma.system.v1alpha1.Config
}
var file_api_system_v1alpha1_config_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_api_system_v1alpha1_config_proto_init() }
func file_api_system_v1alpha1_config_proto_init() {
	if File_api_system_v1alpha1_config_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_api_system_v1alpha1_config_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Config); i {
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
			RawDescriptor: file_api_system_v1alpha1_config_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_system_v1alpha1_config_proto_goTypes,
		DependencyIndexes: file_api_system_v1alpha1_config_proto_depIdxs,
		MessageInfos:      file_api_system_v1alpha1_config_proto_msgTypes,
	}.Build()
	File_api_system_v1alpha1_config_proto = out.File
	file_api_system_v1alpha1_config_proto_rawDesc = nil
	file_api_system_v1alpha1_config_proto_goTypes = nil
	file_api_system_v1alpha1_config_proto_depIdxs = nil
}
