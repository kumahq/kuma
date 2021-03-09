// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.14.0
// source: system/v1alpha1/secret.proto

package v1alpha1

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

// Secret defines encrypted value in Kuma
type Secret struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Value of the secret
	Data *wrappers.BytesValue `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *Secret) Reset() {
	*x = Secret{}
	if protoimpl.UnsafeEnabled {
		mi := &file_system_v1alpha1_secret_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Secret) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Secret) ProtoMessage() {}

func (x *Secret) ProtoReflect() protoreflect.Message {
	mi := &file_system_v1alpha1_secret_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Secret.ProtoReflect.Descriptor instead.
func (*Secret) Descriptor() ([]byte, []int) {
	return file_system_v1alpha1_secret_proto_rawDescGZIP(), []int{0}
}

func (x *Secret) GetData() *wrappers.BytesValue {
	if x != nil {
		return x.Data
	}
	return nil
}

var File_system_v1alpha1_secret_proto protoreflect.FileDescriptor

var file_system_v1alpha1_secret_proto_rawDesc = []byte{
	0x0a, 0x1c, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61,
	0x31, 0x2f, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x14,
	0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x2e, 0x76, 0x31, 0x61, 0x6c,
	0x70, 0x68, 0x61, 0x31, 0x1a, 0x1e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x77, 0x72, 0x61, 0x70, 0x70, 0x65, 0x72, 0x73, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0x39, 0x0a, 0x06, 0x53, 0x65, 0x63, 0x72, 0x65, 0x74, 0x12, 0x2f,
	0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x42,
	0x79, 0x74, 0x65, 0x73, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x42,
	0x2c, 0x5a, 0x2a, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6b, 0x75,
	0x6d, 0x61, 0x68, 0x71, 0x2f, 0x6b, 0x75, 0x6d, 0x61, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x73, 0x79,
	0x73, 0x74, 0x65, 0x6d, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_system_v1alpha1_secret_proto_rawDescOnce sync.Once
	file_system_v1alpha1_secret_proto_rawDescData = file_system_v1alpha1_secret_proto_rawDesc
)

func file_system_v1alpha1_secret_proto_rawDescGZIP() []byte {
	file_system_v1alpha1_secret_proto_rawDescOnce.Do(func() {
		file_system_v1alpha1_secret_proto_rawDescData = protoimpl.X.CompressGZIP(file_system_v1alpha1_secret_proto_rawDescData)
	})
	return file_system_v1alpha1_secret_proto_rawDescData
}

var file_system_v1alpha1_secret_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_system_v1alpha1_secret_proto_goTypes = []interface{}{
	(*Secret)(nil),              // 0: kuma.system.v1alpha1.Secret
	(*wrappers.BytesValue)(nil), // 1: google.protobuf.BytesValue
}
var file_system_v1alpha1_secret_proto_depIdxs = []int32{
	1, // 0: kuma.system.v1alpha1.Secret.data:type_name -> google.protobuf.BytesValue
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_system_v1alpha1_secret_proto_init() }
func file_system_v1alpha1_secret_proto_init() {
	if File_system_v1alpha1_secret_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_system_v1alpha1_secret_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Secret); i {
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
			RawDescriptor: file_system_v1alpha1_secret_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_system_v1alpha1_secret_proto_goTypes,
		DependencyIndexes: file_system_v1alpha1_secret_proto_depIdxs,
		MessageInfos:      file_system_v1alpha1_secret_proto_msgTypes,
	}.Build()
	File_system_v1alpha1_secret_proto = out.File
	file_system_v1alpha1_secret_proto_rawDesc = nil
	file_system_v1alpha1_secret_proto_goTypes = nil
	file_system_v1alpha1_secret_proto_depIdxs = nil
}
