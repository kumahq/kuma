// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.23.0
// 	protoc        v3.14.0
// source: pkg/plugins/ca/builtin/config/builtin_ca_config.proto

package config

import (
	proto "github.com/golang/protobuf/proto"
	wrappers "github.com/golang/protobuf/ptypes/wrappers"
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

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

// BuiltinCertificateAuthorityConfig defines configuration for Builtin CA
// plugin
type BuiltinCertificateAuthorityConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Configuration of CA Certificate
	CaCert *BuiltinCertificateAuthorityConfig_CaCert `protobuf:"bytes,1,opt,name=caCert,proto3" json:"caCert,omitempty"`
}

func (x *BuiltinCertificateAuthorityConfig) Reset() {
	*x = BuiltinCertificateAuthorityConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_plugins_ca_builtin_config_builtin_ca_config_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BuiltinCertificateAuthorityConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BuiltinCertificateAuthorityConfig) ProtoMessage() {}

func (x *BuiltinCertificateAuthorityConfig) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_plugins_ca_builtin_config_builtin_ca_config_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BuiltinCertificateAuthorityConfig.ProtoReflect.Descriptor instead.
func (*BuiltinCertificateAuthorityConfig) Descriptor() ([]byte, []int) {
	return file_pkg_plugins_ca_builtin_config_builtin_ca_config_proto_rawDescGZIP(), []int{0}
}

func (x *BuiltinCertificateAuthorityConfig) GetCaCert() *BuiltinCertificateAuthorityConfig_CaCert {
	if x != nil {
		return x.CaCert
	}
	return nil
}

// CaCert defines configuration for Certificate of CA.
type BuiltinCertificateAuthorityConfig_CaCert struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// RSAbits of the certificate
	RSAbits *wrappers.UInt32Value `protobuf:"bytes,1,opt,name=RSAbits,proto3" json:"RSAbits,omitempty"`
	// Expiration time of the certificate
	Expiration string `protobuf:"bytes,2,opt,name=expiration,proto3" json:"expiration,omitempty"`
}

func (x *BuiltinCertificateAuthorityConfig_CaCert) Reset() {
	*x = BuiltinCertificateAuthorityConfig_CaCert{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_plugins_ca_builtin_config_builtin_ca_config_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BuiltinCertificateAuthorityConfig_CaCert) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BuiltinCertificateAuthorityConfig_CaCert) ProtoMessage() {}

func (x *BuiltinCertificateAuthorityConfig_CaCert) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_plugins_ca_builtin_config_builtin_ca_config_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BuiltinCertificateAuthorityConfig_CaCert.ProtoReflect.Descriptor instead.
func (*BuiltinCertificateAuthorityConfig_CaCert) Descriptor() ([]byte, []int) {
	return file_pkg_plugins_ca_builtin_config_builtin_ca_config_proto_rawDescGZIP(), []int{0, 0}
}

func (x *BuiltinCertificateAuthorityConfig_CaCert) GetRSAbits() *wrappers.UInt32Value {
	if x != nil {
		return x.RSAbits
	}
	return nil
}

func (x *BuiltinCertificateAuthorityConfig_CaCert) GetExpiration() string {
	if x != nil {
		return x.Expiration
	}
	return ""
}

var File_pkg_plugins_ca_builtin_config_builtin_ca_config_proto protoreflect.FileDescriptor

var file_pkg_plugins_ca_builtin_config_builtin_ca_config_proto_rawDesc = []byte{
	0x0a, 0x35, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x2f, 0x63, 0x61,
	0x2f, 0x62, 0x75, 0x69, 0x6c, 0x74, 0x69, 0x6e, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2f,
	0x62, 0x75, 0x69, 0x6c, 0x74, 0x69, 0x6e, 0x5f, 0x63, 0x61, 0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0f, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x70, 0x6c,
	0x75, 0x67, 0x69, 0x6e, 0x73, 0x2e, 0x63, 0x61, 0x1a, 0x1e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x77, 0x72, 0x61, 0x70, 0x70, 0x65,
	0x72, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xd8, 0x01, 0x0a, 0x21, 0x42, 0x75, 0x69,
	0x6c, 0x74, 0x69, 0x6e, 0x43, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x41,
	0x75, 0x74, 0x68, 0x6f, 0x72, 0x69, 0x74, 0x79, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x51,
	0x0a, 0x06, 0x63, 0x61, 0x43, 0x65, 0x72, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x39,
	0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x2e, 0x63, 0x61,
	0x2e, 0x42, 0x75, 0x69, 0x6c, 0x74, 0x69, 0x6e, 0x43, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63,
	0x61, 0x74, 0x65, 0x41, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x69, 0x74, 0x79, 0x43, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x2e, 0x43, 0x61, 0x43, 0x65, 0x72, 0x74, 0x52, 0x06, 0x63, 0x61, 0x43, 0x65, 0x72,
	0x74, 0x1a, 0x60, 0x0a, 0x06, 0x43, 0x61, 0x43, 0x65, 0x72, 0x74, 0x12, 0x36, 0x0a, 0x07, 0x52,
	0x53, 0x41, 0x62, 0x69, 0x74, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x55,
	0x49, 0x6e, 0x74, 0x33, 0x32, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x07, 0x52, 0x53, 0x41, 0x62,
	0x69, 0x74, 0x73, 0x12, 0x1e, 0x0a, 0x0a, 0x65, 0x78, 0x70, 0x69, 0x72, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x65, 0x78, 0x70, 0x69, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x42, 0x2a, 0x5a, 0x28, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x6b, 0x75, 0x6d, 0x61, 0x68, 0x71, 0x2f, 0x6b, 0x75, 0x6d, 0x61, 0x2f, 0x70, 0x6c,
	0x75, 0x67, 0x69, 0x6e, 0x73, 0x2f, 0x63, 0x61, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pkg_plugins_ca_builtin_config_builtin_ca_config_proto_rawDescOnce sync.Once
	file_pkg_plugins_ca_builtin_config_builtin_ca_config_proto_rawDescData = file_pkg_plugins_ca_builtin_config_builtin_ca_config_proto_rawDesc
)

func file_pkg_plugins_ca_builtin_config_builtin_ca_config_proto_rawDescGZIP() []byte {
	file_pkg_plugins_ca_builtin_config_builtin_ca_config_proto_rawDescOnce.Do(func() {
		file_pkg_plugins_ca_builtin_config_builtin_ca_config_proto_rawDescData = protoimpl.X.CompressGZIP(file_pkg_plugins_ca_builtin_config_builtin_ca_config_proto_rawDescData)
	})
	return file_pkg_plugins_ca_builtin_config_builtin_ca_config_proto_rawDescData
}

var file_pkg_plugins_ca_builtin_config_builtin_ca_config_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_pkg_plugins_ca_builtin_config_builtin_ca_config_proto_goTypes = []interface{}{
	(*BuiltinCertificateAuthorityConfig)(nil),        // 0: kuma.plugins.ca.BuiltinCertificateAuthorityConfig
	(*BuiltinCertificateAuthorityConfig_CaCert)(nil), // 1: kuma.plugins.ca.BuiltinCertificateAuthorityConfig.CaCert
	(*wrappers.UInt32Value)(nil),                     // 2: google.protobuf.UInt32Value
}
var file_pkg_plugins_ca_builtin_config_builtin_ca_config_proto_depIdxs = []int32{
	1, // 0: kuma.plugins.ca.BuiltinCertificateAuthorityConfig.caCert:type_name -> kuma.plugins.ca.BuiltinCertificateAuthorityConfig.CaCert
	2, // 1: kuma.plugins.ca.BuiltinCertificateAuthorityConfig.CaCert.RSAbits:type_name -> google.protobuf.UInt32Value
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_pkg_plugins_ca_builtin_config_builtin_ca_config_proto_init() }
func file_pkg_plugins_ca_builtin_config_builtin_ca_config_proto_init() {
	if File_pkg_plugins_ca_builtin_config_builtin_ca_config_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pkg_plugins_ca_builtin_config_builtin_ca_config_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BuiltinCertificateAuthorityConfig); i {
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
		file_pkg_plugins_ca_builtin_config_builtin_ca_config_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BuiltinCertificateAuthorityConfig_CaCert); i {
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
			RawDescriptor: file_pkg_plugins_ca_builtin_config_builtin_ca_config_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_pkg_plugins_ca_builtin_config_builtin_ca_config_proto_goTypes,
		DependencyIndexes: file_pkg_plugins_ca_builtin_config_builtin_ca_config_proto_depIdxs,
		MessageInfos:      file_pkg_plugins_ca_builtin_config_builtin_ca_config_proto_msgTypes,
	}.Build()
	File_pkg_plugins_ca_builtin_config_builtin_ca_config_proto = out.File
	file_pkg_plugins_ca_builtin_config_builtin_ca_config_proto_rawDesc = nil
	file_pkg_plugins_ca_builtin_config_builtin_ca_config_proto_goTypes = nil
	file_pkg_plugins_ca_builtin_config_builtin_ca_config_proto_depIdxs = nil
}
