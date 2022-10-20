// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.20.0
// source: pkg/plugins/policies/meshaccesslog/api/v1alpha1/meshaccesslog.proto

package v1alpha1

import (
	v1alpha1 "github.com/kumahq/kuma/api/common/v1alpha1"
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

// MeshAccessLog defines access log policies between different data plane
// proxies entities.
type MeshAccessLog struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// TargetRef is a reference to the resource the policy takes an effect on.
	// The resource could be either a real store object or virtual resource
	// defined inplace.
	TargetRef *v1alpha1.TargetRef `protobuf:"bytes,1,opt,name=targetRef,proto3" json:"targetRef,omitempty"`
	// From is a list of pairs – a group of clients and action applied for it
	// +optional
	// +nullable
	From []*MeshAccessLog_From `protobuf:"bytes,3,rep,name=from,proto3" json:"from"`
	// +optional
	// +nullable
	To   []*MeshAccessLog_To   `protobuf:"bytes,4,rep,name=to,proto3" json:"to"`
}

func (x *MeshAccessLog) Reset() {
	*x = MeshAccessLog{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MeshAccessLog) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MeshAccessLog) ProtoMessage() {}

func (x *MeshAccessLog) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MeshAccessLog.ProtoReflect.Descriptor instead.
func (*MeshAccessLog) Descriptor() ([]byte, []int) {
	return file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_rawDescGZIP(), []int{0}
}

func (x *MeshAccessLog) GetTargetRef() *v1alpha1.TargetRef {
	if x != nil {
		return x.TargetRef
	}
	return nil
}

func (x *MeshAccessLog) GetFrom() []*MeshAccessLog_From {
	if x != nil {
		return x.From
	}
	return nil
}

func (x *MeshAccessLog) GetTo() []*MeshAccessLog_To {
	if x != nil {
		return x.To
	}
	return nil
}

type MeshAccessLog_Format struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Plain string                            `protobuf:"bytes,1,opt,name=plain,proto3" json:"plain,omitempty"`
	// +optional
	// +nullable
	Json  []*MeshAccessLog_Format_JsonValue `protobuf:"bytes,2,rep,name=json,proto3" json:"json"`
}

func (x *MeshAccessLog_Format) Reset() {
	*x = MeshAccessLog_Format{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MeshAccessLog_Format) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MeshAccessLog_Format) ProtoMessage() {}

func (x *MeshAccessLog_Format) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MeshAccessLog_Format.ProtoReflect.Descriptor instead.
func (*MeshAccessLog_Format) Descriptor() ([]byte, []int) {
	return file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_rawDescGZIP(), []int{0, 0}
}

func (x *MeshAccessLog_Format) GetPlain() string {
	if x != nil {
		return x.Plain
	}
	return ""
}

func (x *MeshAccessLog_Format) GetJson() []*MeshAccessLog_Format_JsonValue {
	if x != nil {
		return x.Json
	}
	return nil
}

// Backend defines logging backend.
type MeshAccessLog_TCPBackend struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Format of access logs. Placeholders available on
	// https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log
	Format *MeshAccessLog_Format `protobuf:"bytes,1,opt,name=format,proto3" json:"format,omitempty"`
	// Type of the backend (Kuma ships with 'tcp' and 'file')
	Address string `protobuf:"bytes,2,opt,name=address,proto3" json:"address,omitempty"`
}

func (x *MeshAccessLog_TCPBackend) Reset() {
	*x = MeshAccessLog_TCPBackend{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MeshAccessLog_TCPBackend) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MeshAccessLog_TCPBackend) ProtoMessage() {}

func (x *MeshAccessLog_TCPBackend) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MeshAccessLog_TCPBackend.ProtoReflect.Descriptor instead.
func (*MeshAccessLog_TCPBackend) Descriptor() ([]byte, []int) {
	return file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_rawDescGZIP(), []int{0, 1}
}

func (x *MeshAccessLog_TCPBackend) GetFormat() *MeshAccessLog_Format {
	if x != nil {
		return x.Format
	}
	return nil
}

func (x *MeshAccessLog_TCPBackend) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

// FileBackend defines configuration for file based access logs
type MeshAccessLog_FileBackend struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Format of access logs. Placeholders available on
	// https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log
	Format *MeshAccessLog_Format `protobuf:"bytes,1,opt,name=format,proto3" json:"format,omitempty"`
	// Path to a file that logs will be written to
	Path string `protobuf:"bytes,2,opt,name=path,proto3" json:"path,omitempty"`
}

func (x *MeshAccessLog_FileBackend) Reset() {
	*x = MeshAccessLog_FileBackend{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MeshAccessLog_FileBackend) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MeshAccessLog_FileBackend) ProtoMessage() {}

func (x *MeshAccessLog_FileBackend) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MeshAccessLog_FileBackend.ProtoReflect.Descriptor instead.
func (*MeshAccessLog_FileBackend) Descriptor() ([]byte, []int) {
	return file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_rawDescGZIP(), []int{0, 2}
}

func (x *MeshAccessLog_FileBackend) GetFormat() *MeshAccessLog_Format {
	if x != nil {
		return x.Format
	}
	return nil
}

func (x *MeshAccessLog_FileBackend) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

type MeshAccessLog_Backend struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Tcp  *MeshAccessLog_TCPBackend  `protobuf:"bytes,1,opt,name=tcp,proto3" json:"tcp,omitempty"`
	File *MeshAccessLog_FileBackend `protobuf:"bytes,2,opt,name=file,proto3" json:"file,omitempty"`
}

func (x *MeshAccessLog_Backend) Reset() {
	*x = MeshAccessLog_Backend{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MeshAccessLog_Backend) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MeshAccessLog_Backend) ProtoMessage() {}

func (x *MeshAccessLog_Backend) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MeshAccessLog_Backend.ProtoReflect.Descriptor instead.
func (*MeshAccessLog_Backend) Descriptor() ([]byte, []int) {
	return file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_rawDescGZIP(), []int{0, 3}
}

func (x *MeshAccessLog_Backend) GetTcp() *MeshAccessLog_TCPBackend {
	if x != nil {
		return x.Tcp
	}
	return nil
}

func (x *MeshAccessLog_Backend) GetFile() *MeshAccessLog_FileBackend {
	if x != nil {
		return x.File
	}
	return nil
}

type MeshAccessLog_Conf struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// +optional
	// +nullable
	Backends []*MeshAccessLog_Backend `protobuf:"bytes,1,rep,name=backends,proto3" json:"backends"`
}

func (x *MeshAccessLog_Conf) Reset() {
	*x = MeshAccessLog_Conf{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MeshAccessLog_Conf) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MeshAccessLog_Conf) ProtoMessage() {}

func (x *MeshAccessLog_Conf) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MeshAccessLog_Conf.ProtoReflect.Descriptor instead.
func (*MeshAccessLog_Conf) Descriptor() ([]byte, []int) {
	return file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_rawDescGZIP(), []int{0, 4}
}

func (x *MeshAccessLog_Conf) GetBackends() []*MeshAccessLog_Backend {
	if x != nil {
		return x.Backends
	}
	return nil
}

type MeshAccessLog_From struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// TargetRef is a reference to the resource that represents a group of
	// clients.
	TargetRef *v1alpha1.TargetRef `protobuf:"bytes,1,opt,name=targetRef,proto3" json:"targetRef,omitempty"`
	// Default is a configuration specific to the group of clients referenced in
	// 'targetRef'
	Default *MeshAccessLog_Conf `protobuf:"bytes,2,opt,name=default,proto3" json:"default,omitempty"`
}

func (x *MeshAccessLog_From) Reset() {
	*x = MeshAccessLog_From{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MeshAccessLog_From) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MeshAccessLog_From) ProtoMessage() {}

func (x *MeshAccessLog_From) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MeshAccessLog_From.ProtoReflect.Descriptor instead.
func (*MeshAccessLog_From) Descriptor() ([]byte, []int) {
	return file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_rawDescGZIP(), []int{0, 5}
}

func (x *MeshAccessLog_From) GetTargetRef() *v1alpha1.TargetRef {
	if x != nil {
		return x.TargetRef
	}
	return nil
}

func (x *MeshAccessLog_From) GetDefault() *MeshAccessLog_Conf {
	if x != nil {
		return x.Default
	}
	return nil
}

type MeshAccessLog_To struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// TargetRef is a reference to the resource that represents a group of
	// clients.
	TargetRef *v1alpha1.TargetRef `protobuf:"bytes,1,opt,name=targetRef,proto3" json:"targetRef,omitempty"`
	// Default is a configuration specific to the group of clients referenced in
	// 'targetRef'
	Default *MeshAccessLog_Conf `protobuf:"bytes,2,opt,name=default,proto3" json:"default,omitempty"`
}

func (x *MeshAccessLog_To) Reset() {
	*x = MeshAccessLog_To{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MeshAccessLog_To) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MeshAccessLog_To) ProtoMessage() {}

func (x *MeshAccessLog_To) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MeshAccessLog_To.ProtoReflect.Descriptor instead.
func (*MeshAccessLog_To) Descriptor() ([]byte, []int) {
	return file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_rawDescGZIP(), []int{0, 6}
}

func (x *MeshAccessLog_To) GetTargetRef() *v1alpha1.TargetRef {
	if x != nil {
		return x.TargetRef
	}
	return nil
}

func (x *MeshAccessLog_To) GetDefault() *MeshAccessLog_Conf {
	if x != nil {
		return x.Default
	}
	return nil
}

type MeshAccessLog_Format_JsonValue struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key   string `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Value string `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *MeshAccessLog_Format_JsonValue) Reset() {
	*x = MeshAccessLog_Format_JsonValue{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MeshAccessLog_Format_JsonValue) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MeshAccessLog_Format_JsonValue) ProtoMessage() {}

func (x *MeshAccessLog_Format_JsonValue) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MeshAccessLog_Format_JsonValue.ProtoReflect.Descriptor instead.
func (*MeshAccessLog_Format_JsonValue) Descriptor() ([]byte, []int) {
	return file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_rawDescGZIP(), []int{0, 0, 0}
}

func (x *MeshAccessLog_Format_JsonValue) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *MeshAccessLog_Format_JsonValue) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

var File_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto protoreflect.FileDescriptor

var file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_rawDesc = []byte{
	0x0a, 0x43, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x2f, 0x70, 0x6f,
	0x6c, 0x69, 0x63, 0x69, 0x65, 0x73, 0x2f, 0x6d, 0x65, 0x73, 0x68, 0x61, 0x63, 0x63, 0x65, 0x73,
	0x73, 0x6c, 0x6f, 0x67, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61,
	0x31, 0x2f, 0x6d, 0x65, 0x73, 0x68, 0x61, 0x63, 0x63, 0x65, 0x73, 0x73, 0x6c, 0x6f, 0x67, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x2c, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x70, 0x6c, 0x75, 0x67,
	0x69, 0x6e, 0x73, 0x2e, 0x70, 0x6f, 0x6c, 0x69, 0x63, 0x69, 0x65, 0x73, 0x2e, 0x6d, 0x65, 0x73,
	0x68, 0x61, 0x63, 0x63, 0x65, 0x73, 0x73, 0x6c, 0x6f, 0x67, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70,
	0x68, 0x61, 0x31, 0x1a, 0x12, 0x6d, 0x65, 0x73, 0x68, 0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f,
	0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x72,
	0x65, 0x66, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x15, 0x6b, 0x75, 0x6d, 0x61, 0x2d, 0x64,
	0x6f, 0x63, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0xed, 0x0a, 0x0a, 0x0d, 0x4d, 0x65, 0x73, 0x68, 0x41, 0x63, 0x63, 0x65, 0x73, 0x73, 0x4c, 0x6f,
	0x67, 0x12, 0x3d, 0x0a, 0x09, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x52, 0x65, 0x66, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x63, 0x6f, 0x6d, 0x6d,
	0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x54, 0x61, 0x72, 0x67,
	0x65, 0x74, 0x52, 0x65, 0x66, 0x52, 0x09, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x52, 0x65, 0x66,
	0x12, 0x54, 0x0a, 0x04, 0x66, 0x72, 0x6f, 0x6d, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x40,
	0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x2e, 0x70, 0x6f,
	0x6c, 0x69, 0x63, 0x69, 0x65, 0x73, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x61, 0x63, 0x63, 0x65, 0x73,
	0x73, 0x6c, 0x6f, 0x67, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x4d, 0x65,
	0x73, 0x68, 0x41, 0x63, 0x63, 0x65, 0x73, 0x73, 0x4c, 0x6f, 0x67, 0x2e, 0x46, 0x72, 0x6f, 0x6d,
	0x52, 0x04, 0x66, 0x72, 0x6f, 0x6d, 0x12, 0x4e, 0x0a, 0x02, 0x74, 0x6f, 0x18, 0x04, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x3e, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e,
	0x73, 0x2e, 0x70, 0x6f, 0x6c, 0x69, 0x63, 0x69, 0x65, 0x73, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x61,
	0x63, 0x63, 0x65, 0x73, 0x73, 0x6c, 0x6f, 0x67, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61,
	0x31, 0x2e, 0x4d, 0x65, 0x73, 0x68, 0x41, 0x63, 0x63, 0x65, 0x73, 0x73, 0x4c, 0x6f, 0x67, 0x2e,
	0x54, 0x6f, 0x52, 0x02, 0x74, 0x6f, 0x1a, 0xc1, 0x01, 0x0a, 0x06, 0x46, 0x6f, 0x72, 0x6d, 0x61,
	0x74, 0x12, 0x14, 0x0a, 0x05, 0x70, 0x6c, 0x61, 0x69, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x05, 0x70, 0x6c, 0x61, 0x69, 0x6e, 0x12, 0x60, 0x0a, 0x04, 0x6a, 0x73, 0x6f, 0x6e, 0x18,
	0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x4c, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x70, 0x6c, 0x75,
	0x67, 0x69, 0x6e, 0x73, 0x2e, 0x70, 0x6f, 0x6c, 0x69, 0x63, 0x69, 0x65, 0x73, 0x2e, 0x6d, 0x65,
	0x73, 0x68, 0x61, 0x63, 0x63, 0x65, 0x73, 0x73, 0x6c, 0x6f, 0x67, 0x2e, 0x76, 0x31, 0x61, 0x6c,
	0x70, 0x68, 0x61, 0x31, 0x2e, 0x4d, 0x65, 0x73, 0x68, 0x41, 0x63, 0x63, 0x65, 0x73, 0x73, 0x4c,
	0x6f, 0x67, 0x2e, 0x46, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x2e, 0x4a, 0x73, 0x6f, 0x6e, 0x56, 0x61,
	0x6c, 0x75, 0x65, 0x52, 0x04, 0x6a, 0x73, 0x6f, 0x6e, 0x1a, 0x3f, 0x0a, 0x09, 0x4a, 0x73, 0x6f,
	0x6e, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x16, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x42, 0x04, 0x88, 0xb5, 0x18, 0x01, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x1a,
	0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x04, 0x88,
	0xb5, 0x18, 0x01, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x1a, 0x8e, 0x01, 0x0a, 0x0a, 0x54,
	0x43, 0x50, 0x42, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x12, 0x60, 0x0a, 0x06, 0x66, 0x6f, 0x72,
	0x6d, 0x61, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x42, 0x2e, 0x6b, 0x75, 0x6d, 0x61,
	0x2e, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x2e, 0x70, 0x6f, 0x6c, 0x69, 0x63, 0x69, 0x65,
	0x73, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x61, 0x63, 0x63, 0x65, 0x73, 0x73, 0x6c, 0x6f, 0x67, 0x2e,
	0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x4d, 0x65, 0x73, 0x68, 0x41, 0x63, 0x63,
	0x65, 0x73, 0x73, 0x4c, 0x6f, 0x67, 0x2e, 0x46, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x42, 0x04, 0x88,
	0xb5, 0x18, 0x01, 0x52, 0x06, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x12, 0x1e, 0x0a, 0x07, 0x61,
	0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x04, 0x88, 0xb5,
	0x18, 0x01, 0x52, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x1a, 0x89, 0x01, 0x0a, 0x0b,
	0x46, 0x69, 0x6c, 0x65, 0x42, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x12, 0x60, 0x0a, 0x06, 0x66,
	0x6f, 0x72, 0x6d, 0x61, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x42, 0x2e, 0x6b, 0x75,
	0x6d, 0x61, 0x2e, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x2e, 0x70, 0x6f, 0x6c, 0x69, 0x63,
	0x69, 0x65, 0x73, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x61, 0x63, 0x63, 0x65, 0x73, 0x73, 0x6c, 0x6f,
	0x67, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x4d, 0x65, 0x73, 0x68, 0x41,
	0x63, 0x63, 0x65, 0x73, 0x73, 0x4c, 0x6f, 0x67, 0x2e, 0x46, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x42,
	0x04, 0x88, 0xb5, 0x18, 0x01, 0x52, 0x06, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x12, 0x18, 0x0a,
	0x04, 0x70, 0x61, 0x74, 0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x04, 0x88, 0xb5, 0x18,
	0x01, 0x52, 0x04, 0x70, 0x61, 0x74, 0x68, 0x1a, 0xc0, 0x01, 0x0a, 0x07, 0x42, 0x61, 0x63, 0x6b,
	0x65, 0x6e, 0x64, 0x12, 0x58, 0x0a, 0x03, 0x74, 0x63, 0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x46, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x2e,
	0x70, 0x6f, 0x6c, 0x69, 0x63, 0x69, 0x65, 0x73, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x61, 0x63, 0x63,
	0x65, 0x73, 0x73, 0x6c, 0x6f, 0x67, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e,
	0x4d, 0x65, 0x73, 0x68, 0x41, 0x63, 0x63, 0x65, 0x73, 0x73, 0x4c, 0x6f, 0x67, 0x2e, 0x54, 0x43,
	0x50, 0x42, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x52, 0x03, 0x74, 0x63, 0x70, 0x12, 0x5b, 0x0a,
	0x04, 0x66, 0x69, 0x6c, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x47, 0x2e, 0x6b, 0x75,
	0x6d, 0x61, 0x2e, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x2e, 0x70, 0x6f, 0x6c, 0x69, 0x63,
	0x69, 0x65, 0x73, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x61, 0x63, 0x63, 0x65, 0x73, 0x73, 0x6c, 0x6f,
	0x67, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x4d, 0x65, 0x73, 0x68, 0x41,
	0x63, 0x63, 0x65, 0x73, 0x73, 0x4c, 0x6f, 0x67, 0x2e, 0x46, 0x69, 0x6c, 0x65, 0x42, 0x61, 0x63,
	0x6b, 0x65, 0x6e, 0x64, 0x52, 0x04, 0x66, 0x69, 0x6c, 0x65, 0x1a, 0x6d, 0x0a, 0x04, 0x43, 0x6f,
	0x6e, 0x66, 0x12, 0x65, 0x0a, 0x08, 0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x73, 0x18, 0x01,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x43, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x70, 0x6c, 0x75, 0x67,
	0x69, 0x6e, 0x73, 0x2e, 0x70, 0x6f, 0x6c, 0x69, 0x63, 0x69, 0x65, 0x73, 0x2e, 0x6d, 0x65, 0x73,
	0x68, 0x61, 0x63, 0x63, 0x65, 0x73, 0x73, 0x6c, 0x6f, 0x67, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70,
	0x68, 0x61, 0x31, 0x2e, 0x4d, 0x65, 0x73, 0x68, 0x41, 0x63, 0x63, 0x65, 0x73, 0x73, 0x4c, 0x6f,
	0x67, 0x2e, 0x42, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x42, 0x04, 0x88, 0xb5, 0x18, 0x01, 0x52,
	0x08, 0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x73, 0x1a, 0xad, 0x01, 0x0a, 0x04, 0x46, 0x72,
	0x6f, 0x6d, 0x12, 0x43, 0x0a, 0x09, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x52, 0x65, 0x66, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x63, 0x6f, 0x6d,
	0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x54, 0x61, 0x72,
	0x67, 0x65, 0x74, 0x52, 0x65, 0x66, 0x42, 0x04, 0x88, 0xb5, 0x18, 0x01, 0x52, 0x09, 0x74, 0x61,
	0x72, 0x67, 0x65, 0x74, 0x52, 0x65, 0x66, 0x12, 0x60, 0x0a, 0x07, 0x64, 0x65, 0x66, 0x61, 0x75,
	0x6c, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x40, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e,
	0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x2e, 0x70, 0x6f, 0x6c, 0x69, 0x63, 0x69, 0x65, 0x73,
	0x2e, 0x6d, 0x65, 0x73, 0x68, 0x61, 0x63, 0x63, 0x65, 0x73, 0x73, 0x6c, 0x6f, 0x67, 0x2e, 0x76,
	0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x4d, 0x65, 0x73, 0x68, 0x41, 0x63, 0x63, 0x65,
	0x73, 0x73, 0x4c, 0x6f, 0x67, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x42, 0x04, 0x88, 0xb5, 0x18, 0x01,
	0x52, 0x07, 0x64, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x1a, 0xab, 0x01, 0x0a, 0x02, 0x54, 0x6f,
	0x12, 0x43, 0x0a, 0x09, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x52, 0x65, 0x66, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f,
	0x6e, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x54, 0x61, 0x72, 0x67, 0x65,
	0x74, 0x52, 0x65, 0x66, 0x42, 0x04, 0x88, 0xb5, 0x18, 0x01, 0x52, 0x09, 0x74, 0x61, 0x72, 0x67,
	0x65, 0x74, 0x52, 0x65, 0x66, 0x12, 0x60, 0x0a, 0x07, 0x64, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x40, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x70, 0x6c,
	0x75, 0x67, 0x69, 0x6e, 0x73, 0x2e, 0x70, 0x6f, 0x6c, 0x69, 0x63, 0x69, 0x65, 0x73, 0x2e, 0x6d,
	0x65, 0x73, 0x68, 0x61, 0x63, 0x63, 0x65, 0x73, 0x73, 0x6c, 0x6f, 0x67, 0x2e, 0x76, 0x31, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x4d, 0x65, 0x73, 0x68, 0x41, 0x63, 0x63, 0x65, 0x73, 0x73,
	0x4c, 0x6f, 0x67, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x42, 0x04, 0x88, 0xb5, 0x18, 0x01, 0x52, 0x07,
	0x64, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x3a, 0x06, 0xb2, 0x8c, 0x89, 0xa6, 0x01, 0x00, 0x42,
	0x6e, 0x5a, 0x46, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6b, 0x75,
	0x6d, 0x61, 0x68, 0x71, 0x2f, 0x6b, 0x75, 0x6d, 0x61, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x6c,
	0x75, 0x67, 0x69, 0x6e, 0x73, 0x2f, 0x70, 0x6f, 0x6c, 0x69, 0x63, 0x69, 0x65, 0x73, 0x2f, 0x6d,
	0x65, 0x73, 0x68, 0x61, 0x63, 0x63, 0x65, 0x73, 0x73, 0x6c, 0x6f, 0x67, 0x2f, 0x61, 0x70, 0x69,
	0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x8a, 0xb5, 0x18, 0x22, 0x50, 0x01, 0xa2,
	0x01, 0x0d, 0x4d, 0x65, 0x73, 0x68, 0x41, 0x63, 0x63, 0x65, 0x73, 0x73, 0x4c, 0x6f, 0x67, 0xf2,
	0x01, 0x0d, 0x6d, 0x65, 0x73, 0x68, 0x61, 0x63, 0x63, 0x65, 0x73, 0x73, 0x6c, 0x6f, 0x67, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_rawDescOnce sync.Once
	file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_rawDescData = file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_rawDesc
)

func file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_rawDescGZIP() []byte {
	file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_rawDescOnce.Do(func() {
		file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_rawDescData = protoimpl.X.CompressGZIP(file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_rawDescData)
	})
	return file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_rawDescData
}

var file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_msgTypes = make([]protoimpl.MessageInfo, 9)
var file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_goTypes = []interface{}{
	(*MeshAccessLog)(nil),                  // 0: kuma.plugins.policies.meshaccesslog.v1alpha1.MeshAccessLog
	(*MeshAccessLog_Format)(nil),           // 1: kuma.plugins.policies.meshaccesslog.v1alpha1.MeshAccessLog.Format
	(*MeshAccessLog_TCPBackend)(nil),       // 2: kuma.plugins.policies.meshaccesslog.v1alpha1.MeshAccessLog.TCPBackend
	(*MeshAccessLog_FileBackend)(nil),      // 3: kuma.plugins.policies.meshaccesslog.v1alpha1.MeshAccessLog.FileBackend
	(*MeshAccessLog_Backend)(nil),          // 4: kuma.plugins.policies.meshaccesslog.v1alpha1.MeshAccessLog.Backend
	(*MeshAccessLog_Conf)(nil),             // 5: kuma.plugins.policies.meshaccesslog.v1alpha1.MeshAccessLog.Conf
	(*MeshAccessLog_From)(nil),             // 6: kuma.plugins.policies.meshaccesslog.v1alpha1.MeshAccessLog.From
	(*MeshAccessLog_To)(nil),               // 7: kuma.plugins.policies.meshaccesslog.v1alpha1.MeshAccessLog.To
	(*MeshAccessLog_Format_JsonValue)(nil), // 8: kuma.plugins.policies.meshaccesslog.v1alpha1.MeshAccessLog.Format.JsonValue
	(*v1alpha1.TargetRef)(nil),             // 9: kuma.common.v1alpha1.TargetRef
}
var file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_depIdxs = []int32{
	9,  // 0: kuma.plugins.policies.meshaccesslog.v1alpha1.MeshAccessLog.targetRef:type_name -> kuma.common.v1alpha1.TargetRef
	6,  // 1: kuma.plugins.policies.meshaccesslog.v1alpha1.MeshAccessLog.from:type_name -> kuma.plugins.policies.meshaccesslog.v1alpha1.MeshAccessLog.From
	7,  // 2: kuma.plugins.policies.meshaccesslog.v1alpha1.MeshAccessLog.to:type_name -> kuma.plugins.policies.meshaccesslog.v1alpha1.MeshAccessLog.To
	8,  // 3: kuma.plugins.policies.meshaccesslog.v1alpha1.MeshAccessLog.Format.json:type_name -> kuma.plugins.policies.meshaccesslog.v1alpha1.MeshAccessLog.Format.JsonValue
	1,  // 4: kuma.plugins.policies.meshaccesslog.v1alpha1.MeshAccessLog.TCPBackend.format:type_name -> kuma.plugins.policies.meshaccesslog.v1alpha1.MeshAccessLog.Format
	1,  // 5: kuma.plugins.policies.meshaccesslog.v1alpha1.MeshAccessLog.FileBackend.format:type_name -> kuma.plugins.policies.meshaccesslog.v1alpha1.MeshAccessLog.Format
	2,  // 6: kuma.plugins.policies.meshaccesslog.v1alpha1.MeshAccessLog.Backend.tcp:type_name -> kuma.plugins.policies.meshaccesslog.v1alpha1.MeshAccessLog.TCPBackend
	3,  // 7: kuma.plugins.policies.meshaccesslog.v1alpha1.MeshAccessLog.Backend.file:type_name -> kuma.plugins.policies.meshaccesslog.v1alpha1.MeshAccessLog.FileBackend
	4,  // 8: kuma.plugins.policies.meshaccesslog.v1alpha1.MeshAccessLog.Conf.backends:type_name -> kuma.plugins.policies.meshaccesslog.v1alpha1.MeshAccessLog.Backend
	9,  // 9: kuma.plugins.policies.meshaccesslog.v1alpha1.MeshAccessLog.From.targetRef:type_name -> kuma.common.v1alpha1.TargetRef
	5,  // 10: kuma.plugins.policies.meshaccesslog.v1alpha1.MeshAccessLog.From.default:type_name -> kuma.plugins.policies.meshaccesslog.v1alpha1.MeshAccessLog.Conf
	9,  // 11: kuma.plugins.policies.meshaccesslog.v1alpha1.MeshAccessLog.To.targetRef:type_name -> kuma.common.v1alpha1.TargetRef
	5,  // 12: kuma.plugins.policies.meshaccesslog.v1alpha1.MeshAccessLog.To.default:type_name -> kuma.plugins.policies.meshaccesslog.v1alpha1.MeshAccessLog.Conf
	13, // [13:13] is the sub-list for method output_type
	13, // [13:13] is the sub-list for method input_type
	13, // [13:13] is the sub-list for extension type_name
	13, // [13:13] is the sub-list for extension extendee
	0,  // [0:13] is the sub-list for field type_name
}

func init() { file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_init() }
func file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_init() {
	if File_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MeshAccessLog); i {
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
		file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MeshAccessLog_Format); i {
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
		file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MeshAccessLog_TCPBackend); i {
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
		file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MeshAccessLog_FileBackend); i {
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
		file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MeshAccessLog_Backend); i {
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
		file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MeshAccessLog_Conf); i {
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
		file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MeshAccessLog_From); i {
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
		file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MeshAccessLog_To); i {
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
		file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MeshAccessLog_Format_JsonValue); i {
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
			RawDescriptor: file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   9,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_goTypes,
		DependencyIndexes: file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_depIdxs,
		MessageInfos:      file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_msgTypes,
	}.Build()
	File_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto = out.File
	file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_rawDesc = nil
	file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_goTypes = nil
	file_pkg_plugins_policies_meshaccesslog_api_v1alpha1_meshaccesslog_proto_depIdxs = nil
}
