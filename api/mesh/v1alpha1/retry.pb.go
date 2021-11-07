// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.14.0
// source: mesh/v1alpha1/retry.proto

package v1alpha1

import (
	_ "github.com/envoyproxy/protoc-gen-validate/validate"
	_ "github.com/kumahq/kuma/api/mesh"
	_ "github.com/kumahq/protoc-gen-kumadoc/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	durationpb "google.golang.org/protobuf/types/known/durationpb"
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

type Retry_Conf_Http_Method int32

const (
	Retry_Conf_Http_CONNECT Retry_Conf_Http_Method = 0
	Retry_Conf_Http_DELETE  Retry_Conf_Http_Method = 1
	Retry_Conf_Http_GET     Retry_Conf_Http_Method = 2
	Retry_Conf_Http_HEAD    Retry_Conf_Http_Method = 3
	Retry_Conf_Http_OPTIONS Retry_Conf_Http_Method = 4
	Retry_Conf_Http_PATCH   Retry_Conf_Http_Method = 5
	Retry_Conf_Http_POST    Retry_Conf_Http_Method = 6
	Retry_Conf_Http_PUT     Retry_Conf_Http_Method = 7
	Retry_Conf_Http_TRACE   Retry_Conf_Http_Method = 8
)

// Enum value maps for Retry_Conf_Http_Method.
var (
	Retry_Conf_Http_Method_name = map[int32]string{
		0: "CONNECT",
		1: "DELETE",
		2: "GET",
		3: "HEAD",
		4: "OPTIONS",
		5: "PATCH",
		6: "POST",
		7: "PUT",
		8: "TRACE",
	}
	Retry_Conf_Http_Method_value = map[string]int32{
		"CONNECT": 0,
		"DELETE":  1,
		"GET":     2,
		"HEAD":    3,
		"OPTIONS": 4,
		"PATCH":   5,
		"POST":    6,
		"PUT":     7,
		"TRACE":   8,
	}
)

func (x Retry_Conf_Http_Method) Enum() *Retry_Conf_Http_Method {
	p := new(Retry_Conf_Http_Method)
	*p = x
	return p
}

func (x Retry_Conf_Http_Method) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Retry_Conf_Http_Method) Descriptor() protoreflect.EnumDescriptor {
	return file_mesh_v1alpha1_retry_proto_enumTypes[0].Descriptor()
}

func (Retry_Conf_Http_Method) Type() protoreflect.EnumType {
	return &file_mesh_v1alpha1_retry_proto_enumTypes[0]
}

func (x Retry_Conf_Http_Method) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Retry_Conf_Http_Method.Descriptor instead.
func (Retry_Conf_Http_Method) EnumDescriptor() ([]byte, []int) {
	return file_mesh_v1alpha1_retry_proto_rawDescGZIP(), []int{0, 0, 1, 0}
}

type Retry_Conf_Grpc_RetryOn int32

const (
	Retry_Conf_Grpc_cancelled          Retry_Conf_Grpc_RetryOn = 0
	Retry_Conf_Grpc_deadline_exceeded  Retry_Conf_Grpc_RetryOn = 1
	Retry_Conf_Grpc_internal           Retry_Conf_Grpc_RetryOn = 2
	Retry_Conf_Grpc_resource_exhausted Retry_Conf_Grpc_RetryOn = 3
	Retry_Conf_Grpc_unavailable        Retry_Conf_Grpc_RetryOn = 4
)

// Enum value maps for Retry_Conf_Grpc_RetryOn.
var (
	Retry_Conf_Grpc_RetryOn_name = map[int32]string{
		0: "cancelled",
		1: "deadline_exceeded",
		2: "internal",
		3: "resource_exhausted",
		4: "unavailable",
	}
	Retry_Conf_Grpc_RetryOn_value = map[string]int32{
		"cancelled":          0,
		"deadline_exceeded":  1,
		"internal":           2,
		"resource_exhausted": 3,
		"unavailable":        4,
	}
)

func (x Retry_Conf_Grpc_RetryOn) Enum() *Retry_Conf_Grpc_RetryOn {
	p := new(Retry_Conf_Grpc_RetryOn)
	*p = x
	return p
}

func (x Retry_Conf_Grpc_RetryOn) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Retry_Conf_Grpc_RetryOn) Descriptor() protoreflect.EnumDescriptor {
	return file_mesh_v1alpha1_retry_proto_enumTypes[1].Descriptor()
}

func (Retry_Conf_Grpc_RetryOn) Type() protoreflect.EnumType {
	return &file_mesh_v1alpha1_retry_proto_enumTypes[1]
}

func (x Retry_Conf_Grpc_RetryOn) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Retry_Conf_Grpc_RetryOn.Descriptor instead.
func (Retry_Conf_Grpc_RetryOn) EnumDescriptor() ([]byte, []int) {
	return file_mesh_v1alpha1_retry_proto_rawDescGZIP(), []int{0, 0, 3, 0}
}

type Retry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// List of selectors to match dataplanes that retry policy should be
	// configured for
	Sources []*Selector `protobuf:"bytes,1,rep,name=sources,proto3" json:"sources,omitempty"`
	// List of selectors to match services that need to be health checked.
	Destinations []*Selector `protobuf:"bytes,2,rep,name=destinations,proto3" json:"destinations,omitempty"`
	//  +required
	Conf *Retry_Conf `protobuf:"bytes,3,opt,name=conf,proto3" json:"conf,omitempty"`
}

func (x *Retry) Reset() {
	*x = Retry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mesh_v1alpha1_retry_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Retry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Retry) ProtoMessage() {}

func (x *Retry) ProtoReflect() protoreflect.Message {
	mi := &file_mesh_v1alpha1_retry_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Retry.ProtoReflect.Descriptor instead.
func (*Retry) Descriptor() ([]byte, []int) {
	return file_mesh_v1alpha1_retry_proto_rawDescGZIP(), []int{0}
}

func (x *Retry) GetSources() []*Selector {
	if x != nil {
		return x.Sources
	}
	return nil
}

func (x *Retry) GetDestinations() []*Selector {
	if x != nil {
		return x.Destinations
	}
	return nil
}

func (x *Retry) GetConf() *Retry_Conf {
	if x != nil {
		return x.Conf
	}
	return nil
}

type Retry_Conf struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Http *Retry_Conf_Http `protobuf:"bytes,1,opt,name=http,proto3" json:"http,omitempty"`
	Tcp  *Retry_Conf_Tcp  `protobuf:"bytes,2,opt,name=tcp,proto3" json:"tcp,omitempty"`
	Grpc *Retry_Conf_Grpc `protobuf:"bytes,3,opt,name=grpc,proto3" json:"grpc,omitempty"`
}

func (x *Retry_Conf) Reset() {
	*x = Retry_Conf{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mesh_v1alpha1_retry_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Retry_Conf) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Retry_Conf) ProtoMessage() {}

func (x *Retry_Conf) ProtoReflect() protoreflect.Message {
	mi := &file_mesh_v1alpha1_retry_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Retry_Conf.ProtoReflect.Descriptor instead.
func (*Retry_Conf) Descriptor() ([]byte, []int) {
	return file_mesh_v1alpha1_retry_proto_rawDescGZIP(), []int{0, 0}
}

func (x *Retry_Conf) GetHttp() *Retry_Conf_Http {
	if x != nil {
		return x.Http
	}
	return nil
}

func (x *Retry_Conf) GetTcp() *Retry_Conf_Tcp {
	if x != nil {
		return x.Tcp
	}
	return nil
}

func (x *Retry_Conf) GetGrpc() *Retry_Conf_Grpc {
	if x != nil {
		return x.Grpc
	}
	return nil
}

type Retry_Conf_BackOff struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//  +required
	BaseInterval *durationpb.Duration `protobuf:"bytes,1,opt,name=base_interval,json=baseInterval,proto3" json:"base_interval,omitempty"`
	//  +optional
	MaxInterval *durationpb.Duration `protobuf:"bytes,2,opt,name=max_interval,json=maxInterval,proto3" json:"max_interval,omitempty"`
}

func (x *Retry_Conf_BackOff) Reset() {
	*x = Retry_Conf_BackOff{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mesh_v1alpha1_retry_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Retry_Conf_BackOff) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Retry_Conf_BackOff) ProtoMessage() {}

func (x *Retry_Conf_BackOff) ProtoReflect() protoreflect.Message {
	mi := &file_mesh_v1alpha1_retry_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Retry_Conf_BackOff.ProtoReflect.Descriptor instead.
func (*Retry_Conf_BackOff) Descriptor() ([]byte, []int) {
	return file_mesh_v1alpha1_retry_proto_rawDescGZIP(), []int{0, 0, 0}
}

func (x *Retry_Conf_BackOff) GetBaseInterval() *durationpb.Duration {
	if x != nil {
		return x.BaseInterval
	}
	return nil
}

func (x *Retry_Conf_BackOff) GetMaxInterval() *durationpb.Duration {
	if x != nil {
		return x.MaxInterval
	}
	return nil
}

type Retry_Conf_Http struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//  +optional
	NumRetries *wrapperspb.UInt32Value `protobuf:"bytes,2,opt,name=num_retries,json=numRetries,proto3" json:"num_retries,omitempty"`
	//  +optional
	PerTryTimeout *durationpb.Duration `protobuf:"bytes,3,opt,name=per_try_timeout,json=perTryTimeout,proto3" json:"per_try_timeout,omitempty"`
	//  +optional
	BackOff *Retry_Conf_BackOff `protobuf:"bytes,4,opt,name=back_off,json=backOff,proto3" json:"back_off,omitempty"`
	//  +optional
	RetriableStatusCodes []uint32 `protobuf:"varint,5,rep,packed,name=retriable_status_codes,json=retriableStatusCodes,proto3" json:"retriable_status_codes,omitempty"`
	//  +optional
	RetriableMethods []Retry_Conf_Http_Method `protobuf:"varint,6,rep,packed,name=retriable_methods,json=retriableMethods,proto3,enum=kuma.mesh.v1alpha1.Retry_Conf_Http_Method" json:"retriable_methods,omitempty"`
}

func (x *Retry_Conf_Http) Reset() {
	*x = Retry_Conf_Http{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mesh_v1alpha1_retry_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Retry_Conf_Http) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Retry_Conf_Http) ProtoMessage() {}

func (x *Retry_Conf_Http) ProtoReflect() protoreflect.Message {
	mi := &file_mesh_v1alpha1_retry_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Retry_Conf_Http.ProtoReflect.Descriptor instead.
func (*Retry_Conf_Http) Descriptor() ([]byte, []int) {
	return file_mesh_v1alpha1_retry_proto_rawDescGZIP(), []int{0, 0, 1}
}

func (x *Retry_Conf_Http) GetNumRetries() *wrapperspb.UInt32Value {
	if x != nil {
		return x.NumRetries
	}
	return nil
}

func (x *Retry_Conf_Http) GetPerTryTimeout() *durationpb.Duration {
	if x != nil {
		return x.PerTryTimeout
	}
	return nil
}

func (x *Retry_Conf_Http) GetBackOff() *Retry_Conf_BackOff {
	if x != nil {
		return x.BackOff
	}
	return nil
}

func (x *Retry_Conf_Http) GetRetriableStatusCodes() []uint32 {
	if x != nil {
		return x.RetriableStatusCodes
	}
	return nil
}

func (x *Retry_Conf_Http) GetRetriableMethods() []Retry_Conf_Http_Method {
	if x != nil {
		return x.RetriableMethods
	}
	return nil
}

type Retry_Conf_Tcp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//  +optional
	MaxConnectAttempts uint32 `protobuf:"varint,1,opt,name=max_connect_attempts,json=maxConnectAttempts,proto3" json:"max_connect_attempts,omitempty"`
}

func (x *Retry_Conf_Tcp) Reset() {
	*x = Retry_Conf_Tcp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mesh_v1alpha1_retry_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Retry_Conf_Tcp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Retry_Conf_Tcp) ProtoMessage() {}

func (x *Retry_Conf_Tcp) ProtoReflect() protoreflect.Message {
	mi := &file_mesh_v1alpha1_retry_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Retry_Conf_Tcp.ProtoReflect.Descriptor instead.
func (*Retry_Conf_Tcp) Descriptor() ([]byte, []int) {
	return file_mesh_v1alpha1_retry_proto_rawDescGZIP(), []int{0, 0, 2}
}

func (x *Retry_Conf_Tcp) GetMaxConnectAttempts() uint32 {
	if x != nil {
		return x.MaxConnectAttempts
	}
	return 0
}

type Retry_Conf_Grpc struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//  +optional
	RetryOn []Retry_Conf_Grpc_RetryOn `protobuf:"varint,1,rep,packed,name=retry_on,json=retryOn,proto3,enum=kuma.mesh.v1alpha1.Retry_Conf_Grpc_RetryOn" json:"retry_on,omitempty"`
	//  +optional
	NumRetries *wrapperspb.UInt32Value `protobuf:"bytes,2,opt,name=num_retries,json=numRetries,proto3" json:"num_retries,omitempty"`
	//  +optional
	PerTryTimeout *durationpb.Duration `protobuf:"bytes,3,opt,name=per_try_timeout,json=perTryTimeout,proto3" json:"per_try_timeout,omitempty"`
	//  +optional
	BackOff *Retry_Conf_BackOff `protobuf:"bytes,4,opt,name=back_off,json=backOff,proto3" json:"back_off,omitempty"`
}

func (x *Retry_Conf_Grpc) Reset() {
	*x = Retry_Conf_Grpc{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mesh_v1alpha1_retry_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Retry_Conf_Grpc) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Retry_Conf_Grpc) ProtoMessage() {}

func (x *Retry_Conf_Grpc) ProtoReflect() protoreflect.Message {
	mi := &file_mesh_v1alpha1_retry_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Retry_Conf_Grpc.ProtoReflect.Descriptor instead.
func (*Retry_Conf_Grpc) Descriptor() ([]byte, []int) {
	return file_mesh_v1alpha1_retry_proto_rawDescGZIP(), []int{0, 0, 3}
}

func (x *Retry_Conf_Grpc) GetRetryOn() []Retry_Conf_Grpc_RetryOn {
	if x != nil {
		return x.RetryOn
	}
	return nil
}

func (x *Retry_Conf_Grpc) GetNumRetries() *wrapperspb.UInt32Value {
	if x != nil {
		return x.NumRetries
	}
	return nil
}

func (x *Retry_Conf_Grpc) GetPerTryTimeout() *durationpb.Duration {
	if x != nil {
		return x.PerTryTimeout
	}
	return nil
}

func (x *Retry_Conf_Grpc) GetBackOff() *Retry_Conf_BackOff {
	if x != nil {
		return x.BackOff
	}
	return nil
}

var File_mesh_v1alpha1_retry_proto protoreflect.FileDescriptor

var file_mesh_v1alpha1_retry_proto_rawDesc = []byte{
	0x0a, 0x19, 0x6d, 0x65, 0x73, 0x68, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f,
	0x72, 0x65, 0x74, 0x72, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x12, 0x6b, 0x75, 0x6d,
	0x61, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x1a,
	0x12, 0x6d, 0x65, 0x73, 0x68, 0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x6d, 0x65, 0x73, 0x68, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68,
	0x61, 0x31, 0x2f, 0x73, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x1e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2f, 0x64, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x1e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2f, 0x77, 0x72, 0x61, 0x70, 0x70, 0x65, 0x72, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x17, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x76, 0x61, 0x6c, 0x69,
	0x64, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0c, 0x63, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xf2, 0x0b, 0x0a, 0x05, 0x52, 0x65, 0x74,
	0x72, 0x79, 0x12, 0x44, 0x0a, 0x07, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x18, 0x01, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e,
	0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x53, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f,
	0x72, 0x42, 0x0c, 0xfa, 0x42, 0x05, 0x92, 0x01, 0x02, 0x08, 0x01, 0x88, 0xb5, 0x18, 0x01, 0x52,
	0x07, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x12, 0x4e, 0x0a, 0x0c, 0x64, 0x65, 0x73, 0x74,
	0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1c,
	0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70,
	0x68, 0x61, 0x31, 0x2e, 0x53, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x42, 0x0c, 0xfa, 0x42,
	0x05, 0x92, 0x01, 0x02, 0x08, 0x01, 0x88, 0xb5, 0x18, 0x01, 0x52, 0x0c, 0x64, 0x65, 0x73, 0x74,
	0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x38, 0x0a, 0x04, 0x63, 0x6f, 0x6e, 0x66,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x6d, 0x65,
	0x73, 0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x52, 0x65, 0x74, 0x72,
	0x79, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x42, 0x04, 0x88, 0xb5, 0x18, 0x01, 0x52, 0x04, 0x63, 0x6f,
	0x6e, 0x66, 0x1a, 0xbe, 0x09, 0x0a, 0x04, 0x43, 0x6f, 0x6e, 0x66, 0x12, 0x37, 0x0a, 0x04, 0x68,
	0x74, 0x74, 0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x23, 0x2e, 0x6b, 0x75, 0x6d, 0x61,
	0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x52,
	0x65, 0x74, 0x72, 0x79, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x2e, 0x48, 0x74, 0x74, 0x70, 0x52, 0x04,
	0x68, 0x74, 0x74, 0x70, 0x12, 0x34, 0x0a, 0x03, 0x74, 0x63, 0x70, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x22, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76, 0x31,
	0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x52, 0x65, 0x74, 0x72, 0x79, 0x2e, 0x43, 0x6f, 0x6e,
	0x66, 0x2e, 0x54, 0x63, 0x70, 0x52, 0x03, 0x74, 0x63, 0x70, 0x12, 0x37, 0x0a, 0x04, 0x67, 0x72,
	0x70, 0x63, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x23, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e,
	0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x52, 0x65,
	0x74, 0x72, 0x79, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x2e, 0x47, 0x72, 0x70, 0x63, 0x52, 0x04, 0x67,
	0x72, 0x70, 0x63, 0x1a, 0x8d, 0x01, 0x0a, 0x07, 0x42, 0x61, 0x63, 0x6b, 0x4f, 0x66, 0x66, 0x12,
	0x44, 0x0a, 0x0d, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x76, 0x61, 0x6c,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x42, 0x04, 0x88, 0xb5, 0x18, 0x01, 0x52, 0x0c, 0x62, 0x61, 0x73, 0x65, 0x49, 0x6e, 0x74,
	0x65, 0x72, 0x76, 0x61, 0x6c, 0x12, 0x3c, 0x0a, 0x0c, 0x6d, 0x61, 0x78, 0x5f, 0x69, 0x6e, 0x74,
	0x65, 0x72, 0x76, 0x61, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x44, 0x75,
	0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x0b, 0x6d, 0x61, 0x78, 0x49, 0x6e, 0x74, 0x65, 0x72,
	0x76, 0x61, 0x6c, 0x1a, 0xc6, 0x03, 0x0a, 0x04, 0x48, 0x74, 0x74, 0x70, 0x12, 0x3d, 0x0a, 0x0b,
	0x6e, 0x75, 0x6d, 0x5f, 0x72, 0x65, 0x74, 0x72, 0x69, 0x65, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x1c, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x55, 0x49, 0x6e, 0x74, 0x33, 0x32, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52,
	0x0a, 0x6e, 0x75, 0x6d, 0x52, 0x65, 0x74, 0x72, 0x69, 0x65, 0x73, 0x12, 0x41, 0x0a, 0x0f, 0x70,
	0x65, 0x72, 0x5f, 0x74, 0x72, 0x79, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52,
	0x0d, 0x70, 0x65, 0x72, 0x54, 0x72, 0x79, 0x54, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x12, 0x41,
	0x0a, 0x08, 0x62, 0x61, 0x63, 0x6b, 0x5f, 0x6f, 0x66, 0x66, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x26, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x52, 0x65, 0x74, 0x72, 0x79, 0x2e, 0x43, 0x6f, 0x6e, 0x66,
	0x2e, 0x42, 0x61, 0x63, 0x6b, 0x4f, 0x66, 0x66, 0x52, 0x07, 0x62, 0x61, 0x63, 0x6b, 0x4f, 0x66,
	0x66, 0x12, 0x34, 0x0a, 0x16, 0x72, 0x65, 0x74, 0x72, 0x69, 0x61, 0x62, 0x6c, 0x65, 0x5f, 0x73,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x5f, 0x63, 0x6f, 0x64, 0x65, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28,
	0x0d, 0x52, 0x14, 0x72, 0x65, 0x74, 0x72, 0x69, 0x61, 0x62, 0x6c, 0x65, 0x53, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x43, 0x6f, 0x64, 0x65, 0x73, 0x12, 0x57, 0x0a, 0x11, 0x72, 0x65, 0x74, 0x72, 0x69,
	0x61, 0x62, 0x6c, 0x65, 0x5f, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x73, 0x18, 0x06, 0x20, 0x03,
	0x28, 0x0e, 0x32, 0x2a, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76,
	0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x52, 0x65, 0x74, 0x72, 0x79, 0x2e, 0x43, 0x6f,
	0x6e, 0x66, 0x2e, 0x48, 0x74, 0x74, 0x70, 0x2e, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x52, 0x10,
	0x72, 0x65, 0x74, 0x72, 0x69, 0x61, 0x62, 0x6c, 0x65, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x73,
	0x22, 0x6a, 0x0a, 0x06, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x12, 0x0b, 0x0a, 0x07, 0x43, 0x4f,
	0x4e, 0x4e, 0x45, 0x43, 0x54, 0x10, 0x00, 0x12, 0x0a, 0x0a, 0x06, 0x44, 0x45, 0x4c, 0x45, 0x54,
	0x45, 0x10, 0x01, 0x12, 0x07, 0x0a, 0x03, 0x47, 0x45, 0x54, 0x10, 0x02, 0x12, 0x08, 0x0a, 0x04,
	0x48, 0x45, 0x41, 0x44, 0x10, 0x03, 0x12, 0x0b, 0x0a, 0x07, 0x4f, 0x50, 0x54, 0x49, 0x4f, 0x4e,
	0x53, 0x10, 0x04, 0x12, 0x09, 0x0a, 0x05, 0x50, 0x41, 0x54, 0x43, 0x48, 0x10, 0x05, 0x12, 0x08,
	0x0a, 0x04, 0x50, 0x4f, 0x53, 0x54, 0x10, 0x06, 0x12, 0x07, 0x0a, 0x03, 0x50, 0x55, 0x54, 0x10,
	0x07, 0x12, 0x09, 0x0a, 0x05, 0x54, 0x52, 0x41, 0x43, 0x45, 0x10, 0x08, 0x1a, 0x37, 0x0a, 0x03,
	0x54, 0x63, 0x70, 0x12, 0x30, 0x0a, 0x14, 0x6d, 0x61, 0x78, 0x5f, 0x63, 0x6f, 0x6e, 0x6e, 0x65,
	0x63, 0x74, 0x5f, 0x61, 0x74, 0x74, 0x65, 0x6d, 0x70, 0x74, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0d, 0x52, 0x12, 0x6d, 0x61, 0x78, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x41, 0x74, 0x74,
	0x65, 0x6d, 0x70, 0x74, 0x73, 0x1a, 0xfb, 0x02, 0x0a, 0x04, 0x47, 0x72, 0x70, 0x63, 0x12, 0x46,
	0x0a, 0x08, 0x72, 0x65, 0x74, 0x72, 0x79, 0x5f, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0e,
	0x32, 0x2b, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x52, 0x65, 0x74, 0x72, 0x79, 0x2e, 0x43, 0x6f, 0x6e, 0x66,
	0x2e, 0x47, 0x72, 0x70, 0x63, 0x2e, 0x52, 0x65, 0x74, 0x72, 0x79, 0x4f, 0x6e, 0x52, 0x07, 0x72,
	0x65, 0x74, 0x72, 0x79, 0x4f, 0x6e, 0x12, 0x3d, 0x0a, 0x0b, 0x6e, 0x75, 0x6d, 0x5f, 0x72, 0x65,
	0x74, 0x72, 0x69, 0x65, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x55, 0x49,
	0x6e, 0x74, 0x33, 0x32, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x0a, 0x6e, 0x75, 0x6d, 0x52, 0x65,
	0x74, 0x72, 0x69, 0x65, 0x73, 0x12, 0x41, 0x0a, 0x0f, 0x70, 0x65, 0x72, 0x5f, 0x74, 0x72, 0x79,
	0x5f, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x0d, 0x70, 0x65, 0x72, 0x54, 0x72,
	0x79, 0x54, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x12, 0x41, 0x0a, 0x08, 0x62, 0x61, 0x63, 0x6b,
	0x5f, 0x6f, 0x66, 0x66, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x26, 0x2e, 0x6b, 0x75, 0x6d,
	0x61, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e,
	0x52, 0x65, 0x74, 0x72, 0x79, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x2e, 0x42, 0x61, 0x63, 0x6b, 0x4f,
	0x66, 0x66, 0x52, 0x07, 0x62, 0x61, 0x63, 0x6b, 0x4f, 0x66, 0x66, 0x22, 0x66, 0x0a, 0x07, 0x52,
	0x65, 0x74, 0x72, 0x79, 0x4f, 0x6e, 0x12, 0x0d, 0x0a, 0x09, 0x63, 0x61, 0x6e, 0x63, 0x65, 0x6c,
	0x6c, 0x65, 0x64, 0x10, 0x00, 0x12, 0x15, 0x0a, 0x11, 0x64, 0x65, 0x61, 0x64, 0x6c, 0x69, 0x6e,
	0x65, 0x5f, 0x65, 0x78, 0x63, 0x65, 0x65, 0x64, 0x65, 0x64, 0x10, 0x01, 0x12, 0x0c, 0x0a, 0x08,
	0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x10, 0x02, 0x12, 0x16, 0x0a, 0x12, 0x72, 0x65,
	0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f, 0x65, 0x78, 0x68, 0x61, 0x75, 0x73, 0x74, 0x65, 0x64,
	0x10, 0x03, 0x12, 0x0f, 0x0a, 0x0b, 0x75, 0x6e, 0x61, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x6c,
	0x65, 0x10, 0x04, 0x3a, 0x58, 0xaa, 0x8c, 0x89, 0xa6, 0x01, 0x0f, 0x0a, 0x0d, 0x52, 0x65, 0x74,
	0x72, 0x79, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0xaa, 0x8c, 0x89, 0xa6, 0x01, 0x07,
	0x12, 0x05, 0x52, 0x65, 0x74, 0x72, 0x79, 0xaa, 0x8c, 0x89, 0xa6, 0x01, 0x06, 0x22, 0x04, 0x6d,
	0x65, 0x73, 0x68, 0xaa, 0x8c, 0x89, 0xa6, 0x01, 0x04, 0x52, 0x02, 0x10, 0x01, 0xaa, 0x8c, 0x89,
	0xa6, 0x01, 0x09, 0x3a, 0x07, 0x0a, 0x05, 0x72, 0x65, 0x74, 0x72, 0x79, 0xaa, 0x8c, 0x89, 0xa6,
	0x01, 0x0b, 0x3a, 0x09, 0x12, 0x07, 0x72, 0x65, 0x74, 0x72, 0x69, 0x65, 0x73, 0x42, 0x40, 0x5a,
	0x28, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6b, 0x75, 0x6d, 0x61,
	0x68, 0x71, 0x2f, 0x6b, 0x75, 0x6d, 0x61, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x6d, 0x65, 0x73, 0x68,
	0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x8a, 0xb5, 0x18, 0x12, 0x50, 0x01, 0xa2,
	0x01, 0x05, 0x52, 0x65, 0x74, 0x72, 0x79, 0xf2, 0x01, 0x05, 0x72, 0x65, 0x74, 0x72, 0x79, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_mesh_v1alpha1_retry_proto_rawDescOnce sync.Once
	file_mesh_v1alpha1_retry_proto_rawDescData = file_mesh_v1alpha1_retry_proto_rawDesc
)

func file_mesh_v1alpha1_retry_proto_rawDescGZIP() []byte {
	file_mesh_v1alpha1_retry_proto_rawDescOnce.Do(func() {
		file_mesh_v1alpha1_retry_proto_rawDescData = protoimpl.X.CompressGZIP(file_mesh_v1alpha1_retry_proto_rawDescData)
	})
	return file_mesh_v1alpha1_retry_proto_rawDescData
}

var file_mesh_v1alpha1_retry_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_mesh_v1alpha1_retry_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_mesh_v1alpha1_retry_proto_goTypes = []interface{}{
	(Retry_Conf_Http_Method)(0),    // 0: kuma.mesh.v1alpha1.Retry.Conf.Http.Method
	(Retry_Conf_Grpc_RetryOn)(0),   // 1: kuma.mesh.v1alpha1.Retry.Conf.Grpc.RetryOn
	(*Retry)(nil),                  // 2: kuma.mesh.v1alpha1.Retry
	(*Retry_Conf)(nil),             // 3: kuma.mesh.v1alpha1.Retry.Conf
	(*Retry_Conf_BackOff)(nil),     // 4: kuma.mesh.v1alpha1.Retry.Conf.BackOff
	(*Retry_Conf_Http)(nil),        // 5: kuma.mesh.v1alpha1.Retry.Conf.Http
	(*Retry_Conf_Tcp)(nil),         // 6: kuma.mesh.v1alpha1.Retry.Conf.Tcp
	(*Retry_Conf_Grpc)(nil),        // 7: kuma.mesh.v1alpha1.Retry.Conf.Grpc
	(*Selector)(nil),               // 8: kuma.mesh.v1alpha1.Selector
	(*durationpb.Duration)(nil),    // 9: google.protobuf.Duration
	(*wrapperspb.UInt32Value)(nil), // 10: google.protobuf.UInt32Value
}
var file_mesh_v1alpha1_retry_proto_depIdxs = []int32{
	8,  // 0: kuma.mesh.v1alpha1.Retry.sources:type_name -> kuma.mesh.v1alpha1.Selector
	8,  // 1: kuma.mesh.v1alpha1.Retry.destinations:type_name -> kuma.mesh.v1alpha1.Selector
	3,  // 2: kuma.mesh.v1alpha1.Retry.conf:type_name -> kuma.mesh.v1alpha1.Retry.Conf
	5,  // 3: kuma.mesh.v1alpha1.Retry.Conf.http:type_name -> kuma.mesh.v1alpha1.Retry.Conf.Http
	6,  // 4: kuma.mesh.v1alpha1.Retry.Conf.tcp:type_name -> kuma.mesh.v1alpha1.Retry.Conf.Tcp
	7,  // 5: kuma.mesh.v1alpha1.Retry.Conf.grpc:type_name -> kuma.mesh.v1alpha1.Retry.Conf.Grpc
	9,  // 6: kuma.mesh.v1alpha1.Retry.Conf.BackOff.base_interval:type_name -> google.protobuf.Duration
	9,  // 7: kuma.mesh.v1alpha1.Retry.Conf.BackOff.max_interval:type_name -> google.protobuf.Duration
	10, // 8: kuma.mesh.v1alpha1.Retry.Conf.Http.num_retries:type_name -> google.protobuf.UInt32Value
	9,  // 9: kuma.mesh.v1alpha1.Retry.Conf.Http.per_try_timeout:type_name -> google.protobuf.Duration
	4,  // 10: kuma.mesh.v1alpha1.Retry.Conf.Http.back_off:type_name -> kuma.mesh.v1alpha1.Retry.Conf.BackOff
	0,  // 11: kuma.mesh.v1alpha1.Retry.Conf.Http.retriable_methods:type_name -> kuma.mesh.v1alpha1.Retry.Conf.Http.Method
	1,  // 12: kuma.mesh.v1alpha1.Retry.Conf.Grpc.retry_on:type_name -> kuma.mesh.v1alpha1.Retry.Conf.Grpc.RetryOn
	10, // 13: kuma.mesh.v1alpha1.Retry.Conf.Grpc.num_retries:type_name -> google.protobuf.UInt32Value
	9,  // 14: kuma.mesh.v1alpha1.Retry.Conf.Grpc.per_try_timeout:type_name -> google.protobuf.Duration
	4,  // 15: kuma.mesh.v1alpha1.Retry.Conf.Grpc.back_off:type_name -> kuma.mesh.v1alpha1.Retry.Conf.BackOff
	16, // [16:16] is the sub-list for method output_type
	16, // [16:16] is the sub-list for method input_type
	16, // [16:16] is the sub-list for extension type_name
	16, // [16:16] is the sub-list for extension extendee
	0,  // [0:16] is the sub-list for field type_name
}

func init() { file_mesh_v1alpha1_retry_proto_init() }
func file_mesh_v1alpha1_retry_proto_init() {
	if File_mesh_v1alpha1_retry_proto != nil {
		return
	}
	file_mesh_v1alpha1_selector_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_mesh_v1alpha1_retry_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Retry); i {
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
		file_mesh_v1alpha1_retry_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Retry_Conf); i {
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
		file_mesh_v1alpha1_retry_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Retry_Conf_BackOff); i {
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
		file_mesh_v1alpha1_retry_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Retry_Conf_Http); i {
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
		file_mesh_v1alpha1_retry_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Retry_Conf_Tcp); i {
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
		file_mesh_v1alpha1_retry_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Retry_Conf_Grpc); i {
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
			RawDescriptor: file_mesh_v1alpha1_retry_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_mesh_v1alpha1_retry_proto_goTypes,
		DependencyIndexes: file_mesh_v1alpha1_retry_proto_depIdxs,
		EnumInfos:         file_mesh_v1alpha1_retry_proto_enumTypes,
		MessageInfos:      file_mesh_v1alpha1_retry_proto_msgTypes,
	}.Build()
	File_mesh_v1alpha1_retry_proto = out.File
	file_mesh_v1alpha1_retry_proto_rawDesc = nil
	file_mesh_v1alpha1_retry_proto_goTypes = nil
	file_mesh_v1alpha1_retry_proto_depIdxs = nil
}
