// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.20.0
// source: mesh/v1alpha1/gateway.proto

package v1alpha1

import (
	_ "github.com/envoyproxy/protoc-gen-validate/validate"
	_ "github.com/kumahq/kuma/api/mesh"
	v1alpha1 "github.com/kumahq/kuma/api/system/v1alpha1"
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

type MeshGateway_TLS_Mode int32

const (
	// NONE is not a valid TLS mode. Ether TERMINATE or PASSTHROUGH must
	// be explicitly configured.
	MeshGateway_TLS_NONE MeshGateway_TLS_Mode = 0
	// The TLS session between the downstream client and the MeshGateway
	// is terminated at the MeshGateway. This mode requires the certificate
	// field to be set.
	MeshGateway_TLS_TERMINATE MeshGateway_TLS_Mode = 1
)

// Enum value maps for MeshGateway_TLS_Mode.
var (
	MeshGateway_TLS_Mode_name = map[int32]string{
		0: "NONE",
		1: "TERMINATE",
	}
	MeshGateway_TLS_Mode_value = map[string]int32{
		"NONE":      0,
		"TERMINATE": 1,
	}
)

func (x MeshGateway_TLS_Mode) Enum() *MeshGateway_TLS_Mode {
	p := new(MeshGateway_TLS_Mode)
	*p = x
	return p
}

func (x MeshGateway_TLS_Mode) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (MeshGateway_TLS_Mode) Descriptor() protoreflect.EnumDescriptor {
	return file_mesh_v1alpha1_gateway_proto_enumTypes[0].Descriptor()
}

func (MeshGateway_TLS_Mode) Type() protoreflect.EnumType {
	return &file_mesh_v1alpha1_gateway_proto_enumTypes[0]
}

func (x MeshGateway_TLS_Mode) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use MeshGateway_TLS_Mode.Descriptor instead.
func (MeshGateway_TLS_Mode) EnumDescriptor() ([]byte, []int) {
	return file_mesh_v1alpha1_gateway_proto_rawDescGZIP(), []int{0, 0, 0}
}

type MeshGateway_Listener_Protocol int32

const (
	MeshGateway_Listener_NONE  MeshGateway_Listener_Protocol = 0
	MeshGateway_Listener_TCP   MeshGateway_Listener_Protocol = 1
	MeshGateway_Listener_HTTP  MeshGateway_Listener_Protocol = 4
	MeshGateway_Listener_HTTPS MeshGateway_Listener_Protocol = 5
)

// Enum value maps for MeshGateway_Listener_Protocol.
var (
	MeshGateway_Listener_Protocol_name = map[int32]string{
		0: "NONE",
		1: "TCP",
		4: "HTTP",
		5: "HTTPS",
	}
	MeshGateway_Listener_Protocol_value = map[string]int32{
		"NONE":  0,
		"TCP":   1,
		"HTTP":  4,
		"HTTPS": 5,
	}
)

func (x MeshGateway_Listener_Protocol) Enum() *MeshGateway_Listener_Protocol {
	p := new(MeshGateway_Listener_Protocol)
	*p = x
	return p
}

func (x MeshGateway_Listener_Protocol) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (MeshGateway_Listener_Protocol) Descriptor() protoreflect.EnumDescriptor {
	return file_mesh_v1alpha1_gateway_proto_enumTypes[1].Descriptor()
}

func (MeshGateway_Listener_Protocol) Type() protoreflect.EnumType {
	return &file_mesh_v1alpha1_gateway_proto_enumTypes[1]
}

func (x MeshGateway_Listener_Protocol) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use MeshGateway_Listener_Protocol.Descriptor instead.
func (MeshGateway_Listener_Protocol) EnumDescriptor() ([]byte, []int) {
	return file_mesh_v1alpha1_gateway_proto_rawDescGZIP(), []int{0, 1, 0}
}

// MeshGateway is a virtual proxy.
//
// Each MeshGateway is bound to a set of builtin gateway dataplanes.
// Each builtin dataplane instance can host exactly one Gateway
// proxy configuration.
//
// Gateway aligns with the Kubernetes Gateway API. See that
// spec for detailed documentation.
type MeshGateway struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Selectors is a list of selectors that are used to match builtin
	// gateway dataplanes that will receive this MeshGateway configuration.
	Selectors []*Selector `protobuf:"bytes,1,rep,name=selectors,proto3" json:"selectors,omitempty"`
	// Tags is the set of tags common to all of the gateway's listeners.
	//
	// This field must not include a `kuma.io/service` tag (the service is always
	// defined on the dataplanes).
	Tags map[string]string `protobuf:"bytes,2,rep,name=tags,proto3" json:"tags,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// The desired configuration of the MeshGateway.
	Conf *MeshGateway_Conf `protobuf:"bytes,3,opt,name=conf,proto3" json:"conf,omitempty"`
}

func (x *MeshGateway) Reset() {
	*x = MeshGateway{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mesh_v1alpha1_gateway_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MeshGateway) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MeshGateway) ProtoMessage() {}

func (x *MeshGateway) ProtoReflect() protoreflect.Message {
	mi := &file_mesh_v1alpha1_gateway_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MeshGateway.ProtoReflect.Descriptor instead.
func (*MeshGateway) Descriptor() ([]byte, []int) {
	return file_mesh_v1alpha1_gateway_proto_rawDescGZIP(), []int{0}
}

func (x *MeshGateway) GetSelectors() []*Selector {
	if x != nil {
		return x.Selectors
	}
	return nil
}

func (x *MeshGateway) GetTags() map[string]string {
	if x != nil {
		return x.Tags
	}
	return nil
}

func (x *MeshGateway) GetConf() *MeshGateway_Conf {
	if x != nil {
		return x.Conf
	}
	return nil
}

// TLSConfig describes a TLS configuration.
type MeshGateway_TLS struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *MeshGateway_TLS) Reset() {
	*x = MeshGateway_TLS{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mesh_v1alpha1_gateway_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MeshGateway_TLS) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MeshGateway_TLS) ProtoMessage() {}

func (x *MeshGateway_TLS) ProtoReflect() protoreflect.Message {
	mi := &file_mesh_v1alpha1_gateway_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MeshGateway_TLS.ProtoReflect.Descriptor instead.
func (*MeshGateway_TLS) Descriptor() ([]byte, []int) {
	return file_mesh_v1alpha1_gateway_proto_rawDescGZIP(), []int{0, 0}
}

type MeshGateway_Listener struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Hostname specifies the virtual hostname to match for protocol types that
	// define this concept. When unspecified, "", or `*`, all hostnames are
	// matched. This field can be omitted for protocols that don't require
	// hostname based matching.
	Hostname string `protobuf:"bytes,1,opt,name=hostname,proto3" json:"hostname,omitempty"`
	// Port is the network port. Multiple listeners may use the
	// same port, subject to the Listener compatibility rules.
	Port uint32 `protobuf:"varint,2,opt,name=port,proto3" json:"port,omitempty"`
	// Protocol specifies the network protocol this listener expects to receive.
	Protocol MeshGateway_Listener_Protocol `protobuf:"varint,3,opt,name=protocol,proto3,enum=kuma.mesh.v1alpha1.MeshGateway_Listener_Protocol" json:"protocol,omitempty"`
	// TLS is the TLS configuration for the Listener. This field
	// is required if the Protocol field is "HTTPS" or "TLS" and
	// ignored otherwise.
	Tls *MeshGateway_TLS_Conf `protobuf:"bytes,4,opt,name=tls,proto3" json:"tls,omitempty"`
	// Tags specifies a unique combination of tags that routes can use
	// to match themselves to this listener.
	//
	// When matching routes to listeners, the control plane constructs a
	// set of matching tags for each listener by forming the union of the
	// gateway tags and the listener tags. A route will be attached to the
	// listener if all of the route's tags are preset in the matching tags
	Tags map[string]string `protobuf:"bytes,5,rep,name=tags,proto3" json:"tags,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// CrossMesh enables traffic to flow to this listener only from other
	// meshes.
	CrossMesh bool `protobuf:"varint,6,opt,name=crossMesh,proto3" json:"crossMesh,omitempty"`
	// Resources is used to specify listener-specific resource settings.
	Resources *MeshGateway_Listener_Resources `protobuf:"bytes,7,opt,name=resources,proto3" json:"resources,omitempty"`
}

func (x *MeshGateway_Listener) Reset() {
	*x = MeshGateway_Listener{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mesh_v1alpha1_gateway_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MeshGateway_Listener) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MeshGateway_Listener) ProtoMessage() {}

func (x *MeshGateway_Listener) ProtoReflect() protoreflect.Message {
	mi := &file_mesh_v1alpha1_gateway_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MeshGateway_Listener.ProtoReflect.Descriptor instead.
func (*MeshGateway_Listener) Descriptor() ([]byte, []int) {
	return file_mesh_v1alpha1_gateway_proto_rawDescGZIP(), []int{0, 1}
}

func (x *MeshGateway_Listener) GetHostname() string {
	if x != nil {
		return x.Hostname
	}
	return ""
}

func (x *MeshGateway_Listener) GetPort() uint32 {
	if x != nil {
		return x.Port
	}
	return 0
}

func (x *MeshGateway_Listener) GetProtocol() MeshGateway_Listener_Protocol {
	if x != nil {
		return x.Protocol
	}
	return MeshGateway_Listener_NONE
}

func (x *MeshGateway_Listener) GetTls() *MeshGateway_TLS_Conf {
	if x != nil {
		return x.Tls
	}
	return nil
}

func (x *MeshGateway_Listener) GetTags() map[string]string {
	if x != nil {
		return x.Tags
	}
	return nil
}

func (x *MeshGateway_Listener) GetCrossMesh() bool {
	if x != nil {
		return x.CrossMesh
	}
	return false
}

func (x *MeshGateway_Listener) GetResources() *MeshGateway_Listener_Resources {
	if x != nil {
		return x.Resources
	}
	return nil
}

// Conf defines the desired state of MeshGateway.
//
// Aligns with MeshGatewaySpec.
type MeshGateway_Conf struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Listeners define logical endpoints that are bound on this MeshGateway's
	// address(es).
	Listeners []*MeshGateway_Listener `protobuf:"bytes,2,rep,name=listeners,proto3" json:"listeners,omitempty"`
}

func (x *MeshGateway_Conf) Reset() {
	*x = MeshGateway_Conf{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mesh_v1alpha1_gateway_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MeshGateway_Conf) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MeshGateway_Conf) ProtoMessage() {}

func (x *MeshGateway_Conf) ProtoReflect() protoreflect.Message {
	mi := &file_mesh_v1alpha1_gateway_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MeshGateway_Conf.ProtoReflect.Descriptor instead.
func (*MeshGateway_Conf) Descriptor() ([]byte, []int) {
	return file_mesh_v1alpha1_gateway_proto_rawDescGZIP(), []int{0, 2}
}

func (x *MeshGateway_Conf) GetListeners() []*MeshGateway_Listener {
	if x != nil {
		return x.Listeners
	}
	return nil
}

type MeshGateway_TLS_Options struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *MeshGateway_TLS_Options) Reset() {
	*x = MeshGateway_TLS_Options{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mesh_v1alpha1_gateway_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MeshGateway_TLS_Options) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MeshGateway_TLS_Options) ProtoMessage() {}

func (x *MeshGateway_TLS_Options) ProtoReflect() protoreflect.Message {
	mi := &file_mesh_v1alpha1_gateway_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MeshGateway_TLS_Options.ProtoReflect.Descriptor instead.
func (*MeshGateway_TLS_Options) Descriptor() ([]byte, []int) {
	return file_mesh_v1alpha1_gateway_proto_rawDescGZIP(), []int{0, 0, 0}
}

// Aligns with MeshGatewayTLSConfig.
type MeshGateway_TLS_Conf struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Mode defines the TLS behavior for the TLS session initiated
	// by the client.
	Mode MeshGateway_TLS_Mode `protobuf:"varint,1,opt,name=mode,proto3,enum=kuma.mesh.v1alpha1.MeshGateway_TLS_Mode" json:"mode,omitempty"`
	// Certificates is an array of datasources that contain TLS
	// certificates and private keys.  Each datasource must contain a
	// sequence of PEM-encoded objects. The server certificate and private
	// key are required, but additional certificates are allowed and will
	// be added to the certificate chain.  The server certificate must
	// be the first certificate in the datasource.
	//
	// When multiple certificate datasources are configured, they must have
	// different key types. In practice, this means that one datasource
	// should contain an RSA key and certificate, and the other an
	// ECDSA key and certificate.
	Certificates []*v1alpha1.DataSource `protobuf:"bytes,2,rep,name=certificates,proto3" json:"certificates,omitempty"`
	// Options should eventually configure how TLS is configured. This
	// is where cipher suite and version configuration can be specified,
	// client certificates enforced, and so on.
	Options *MeshGateway_TLS_Options `protobuf:"bytes,3,opt,name=options,proto3" json:"options,omitempty"`
}

func (x *MeshGateway_TLS_Conf) Reset() {
	*x = MeshGateway_TLS_Conf{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mesh_v1alpha1_gateway_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MeshGateway_TLS_Conf) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MeshGateway_TLS_Conf) ProtoMessage() {}

func (x *MeshGateway_TLS_Conf) ProtoReflect() protoreflect.Message {
	mi := &file_mesh_v1alpha1_gateway_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MeshGateway_TLS_Conf.ProtoReflect.Descriptor instead.
func (*MeshGateway_TLS_Conf) Descriptor() ([]byte, []int) {
	return file_mesh_v1alpha1_gateway_proto_rawDescGZIP(), []int{0, 0, 1}
}

func (x *MeshGateway_TLS_Conf) GetMode() MeshGateway_TLS_Mode {
	if x != nil {
		return x.Mode
	}
	return MeshGateway_TLS_NONE
}

func (x *MeshGateway_TLS_Conf) GetCertificates() []*v1alpha1.DataSource {
	if x != nil {
		return x.Certificates
	}
	return nil
}

func (x *MeshGateway_TLS_Conf) GetOptions() *MeshGateway_TLS_Options {
	if x != nil {
		return x.Options
	}
	return nil
}

type MeshGateway_Listener_Resources struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ConnectionLimit uint32 `protobuf:"varint,1,opt,name=connection_limit,json=connectionLimit,proto3" json:"connection_limit,omitempty"`
}

func (x *MeshGateway_Listener_Resources) Reset() {
	*x = MeshGateway_Listener_Resources{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mesh_v1alpha1_gateway_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MeshGateway_Listener_Resources) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MeshGateway_Listener_Resources) ProtoMessage() {}

func (x *MeshGateway_Listener_Resources) ProtoReflect() protoreflect.Message {
	mi := &file_mesh_v1alpha1_gateway_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MeshGateway_Listener_Resources.ProtoReflect.Descriptor instead.
func (*MeshGateway_Listener_Resources) Descriptor() ([]byte, []int) {
	return file_mesh_v1alpha1_gateway_proto_rawDescGZIP(), []int{0, 1, 0}
}

func (x *MeshGateway_Listener_Resources) GetConnectionLimit() uint32 {
	if x != nil {
		return x.ConnectionLimit
	}
	return 0
}

var File_mesh_v1alpha1_gateway_proto protoreflect.FileDescriptor

var file_mesh_v1alpha1_gateway_proto_rawDesc = []byte{
	0x0a, 0x1b, 0x6d, 0x65, 0x73, 0x68, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f,
	0x67, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x12, 0x6b,
	0x75, 0x6d, 0x61, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61,
	0x31, 0x1a, 0x15, 0x6b, 0x75, 0x6d, 0x61, 0x2d, 0x64, 0x6f, 0x63, 0x2f, 0x63, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x12, 0x6d, 0x65, 0x73, 0x68, 0x2f, 0x6f,
	0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x6d, 0x65,
	0x73, 0x68, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x73, 0x65, 0x6c, 0x65,
	0x63, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x20, 0x73, 0x79, 0x73, 0x74,
	0x65, 0x6d, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x64, 0x61, 0x74, 0x61,
	0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x17, 0x76, 0x61,
	0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x86, 0x0a, 0x0a, 0x0b, 0x4d, 0x65, 0x73, 0x68, 0x47, 0x61,
	0x74, 0x65, 0x77, 0x61, 0x79, 0x12, 0x44, 0x0a, 0x09, 0x73, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f,
	0x72, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e,
	0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x53, 0x65,
	0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x42, 0x08, 0xfa, 0x42, 0x05, 0x92, 0x01, 0x02, 0x08, 0x01,
	0x52, 0x09, 0x73, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x73, 0x12, 0x47, 0x0a, 0x04, 0x74,
	0x61, 0x67, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x29, 0x2e, 0x6b, 0x75, 0x6d, 0x61,
	0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x4d,
	0x65, 0x73, 0x68, 0x47, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x2e, 0x54, 0x61, 0x67, 0x73, 0x45,
	0x6e, 0x74, 0x72, 0x79, 0x42, 0x08, 0xfa, 0x42, 0x05, 0x92, 0x01, 0x02, 0x08, 0x01, 0x52, 0x04,
	0x74, 0x61, 0x67, 0x73, 0x12, 0x38, 0x0a, 0x04, 0x63, 0x6f, 0x6e, 0x66, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x24, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76,
	0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x4d, 0x65, 0x73, 0x68, 0x47, 0x61, 0x74, 0x65,
	0x77, 0x61, 0x79, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x52, 0x04, 0x63, 0x6f, 0x6e, 0x66, 0x1a, 0x91,
	0x02, 0x0a, 0x03, 0x54, 0x4c, 0x53, 0x1a, 0x09, 0x0a, 0x07, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x1a, 0xdd, 0x01, 0x0a, 0x04, 0x43, 0x6f, 0x6e, 0x66, 0x12, 0x3c, 0x0a, 0x04, 0x6d, 0x6f,
	0x64, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x28, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e,
	0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x4d, 0x65,
	0x73, 0x68, 0x47, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x2e, 0x54, 0x4c, 0x53, 0x2e, 0x4d, 0x6f,
	0x64, 0x65, 0x52, 0x04, 0x6d, 0x6f, 0x64, 0x65, 0x12, 0x50, 0x0a, 0x0c, 0x63, 0x65, 0x72, 0x74,
	0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x20,
	0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x2e, 0x76, 0x31, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x44, 0x61, 0x74, 0x61, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65,
	0x42, 0x0a, 0xfa, 0x42, 0x07, 0x92, 0x01, 0x04, 0x08, 0x01, 0x10, 0x02, 0x52, 0x0c, 0x63, 0x65,
	0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x73, 0x12, 0x45, 0x0a, 0x07, 0x6f, 0x70,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2b, 0x2e, 0x6b, 0x75,
	0x6d, 0x61, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31,
	0x2e, 0x4d, 0x65, 0x73, 0x68, 0x47, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x2e, 0x54, 0x4c, 0x53,
	0x2e, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x07, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x22, 0x1f, 0x0a, 0x04, 0x4d, 0x6f, 0x64, 0x65, 0x12, 0x08, 0x0a, 0x04, 0x4e, 0x4f, 0x4e,
	0x45, 0x10, 0x00, 0x12, 0x0d, 0x0a, 0x09, 0x54, 0x45, 0x52, 0x4d, 0x49, 0x4e, 0x41, 0x54, 0x45,
	0x10, 0x01, 0x1a, 0xa2, 0x04, 0x0a, 0x08, 0x4c, 0x69, 0x73, 0x74, 0x65, 0x6e, 0x65, 0x72, 0x12,
	0x1a, 0x0a, 0x08, 0x68, 0x6f, 0x73, 0x74, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x08, 0x68, 0x6f, 0x73, 0x74, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x70,
	0x6f, 0x72, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x12,
	0x4d, 0x0a, 0x08, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x0e, 0x32, 0x31, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76, 0x31,
	0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x4d, 0x65, 0x73, 0x68, 0x47, 0x61, 0x74, 0x65, 0x77,
	0x61, 0x79, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x65, 0x6e, 0x65, 0x72, 0x2e, 0x50, 0x72, 0x6f, 0x74,
	0x6f, 0x63, 0x6f, 0x6c, 0x52, 0x08, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x12, 0x3a,
	0x0a, 0x03, 0x74, 0x6c, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x28, 0x2e, 0x6b, 0x75,
	0x6d, 0x61, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31,
	0x2e, 0x4d, 0x65, 0x73, 0x68, 0x47, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x2e, 0x54, 0x4c, 0x53,
	0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x52, 0x03, 0x74, 0x6c, 0x73, 0x12, 0x46, 0x0a, 0x04, 0x74, 0x61,
	0x67, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x32, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e,
	0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x4d, 0x65,
	0x73, 0x68, 0x47, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x65, 0x6e,
	0x65, 0x72, 0x2e, 0x54, 0x61, 0x67, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x04, 0x74, 0x61,
	0x67, 0x73, 0x12, 0x1c, 0x0a, 0x09, 0x63, 0x72, 0x6f, 0x73, 0x73, 0x4d, 0x65, 0x73, 0x68, 0x18,
	0x06, 0x20, 0x01, 0x28, 0x08, 0x52, 0x09, 0x63, 0x72, 0x6f, 0x73, 0x73, 0x4d, 0x65, 0x73, 0x68,
	0x12, 0x50, 0x0a, 0x09, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x18, 0x07, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x32, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e,
	0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x4d, 0x65, 0x73, 0x68, 0x47, 0x61, 0x74,
	0x65, 0x77, 0x61, 0x79, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x65, 0x6e, 0x65, 0x72, 0x2e, 0x52, 0x65,
	0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x52, 0x09, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63,
	0x65, 0x73, 0x1a, 0x36, 0x0a, 0x09, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x12,
	0x29, 0x0a, 0x10, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x6c, 0x69,
	0x6d, 0x69, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x0f, 0x63, 0x6f, 0x6e, 0x6e, 0x65,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x4c, 0x69, 0x6d, 0x69, 0x74, 0x1a, 0x37, 0x0a, 0x09, 0x54, 0x61,
	0x67, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a,
	0x02, 0x38, 0x01, 0x22, 0x32, 0x0a, 0x08, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x12,
	0x08, 0x0a, 0x04, 0x4e, 0x4f, 0x4e, 0x45, 0x10, 0x00, 0x12, 0x07, 0x0a, 0x03, 0x54, 0x43, 0x50,
	0x10, 0x01, 0x12, 0x08, 0x0a, 0x04, 0x48, 0x54, 0x54, 0x50, 0x10, 0x04, 0x12, 0x09, 0x0a, 0x05,
	0x48, 0x54, 0x54, 0x50, 0x53, 0x10, 0x05, 0x1a, 0x58, 0x0a, 0x04, 0x43, 0x6f, 0x6e, 0x66, 0x12,
	0x50, 0x0a, 0x09, 0x6c, 0x69, 0x73, 0x74, 0x65, 0x6e, 0x65, 0x72, 0x73, 0x18, 0x02, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x28, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76,
	0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x4d, 0x65, 0x73, 0x68, 0x47, 0x61, 0x74, 0x65,
	0x77, 0x61, 0x79, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x65, 0x6e, 0x65, 0x72, 0x42, 0x08, 0xfa, 0x42,
	0x05, 0x92, 0x01, 0x02, 0x08, 0x01, 0x52, 0x09, 0x6c, 0x69, 0x73, 0x74, 0x65, 0x6e, 0x65, 0x72,
	0x73, 0x1a, 0x37, 0x0a, 0x09, 0x54, 0x61, 0x67, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10,
	0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79,
	0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x3a, 0x62, 0xaa, 0x8c, 0x89, 0xa6,
	0x01, 0x15, 0x0a, 0x13, 0x4d, 0x65, 0x73, 0x68, 0x47, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x52,
	0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0xaa, 0x8c, 0x89, 0xa6, 0x01, 0x0d, 0x12, 0x0b, 0x4d,
	0x65, 0x73, 0x68, 0x47, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0xaa, 0x8c, 0x89, 0xa6, 0x01, 0x06,
	0x22, 0x04, 0x6d, 0x65, 0x73, 0x68, 0xaa, 0x8c, 0x89, 0xa6, 0x01, 0x03, 0x80, 0x01, 0x01, 0xaa,
	0x8c, 0x89, 0xa6, 0x01, 0x04, 0x52, 0x02, 0x10, 0x01, 0xaa, 0x8c, 0x89, 0xa6, 0x01, 0x0f, 0x3a,
	0x0d, 0x0a, 0x0b, 0x6d, 0x65, 0x73, 0x68, 0x67, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x42, 0x4c,
	0x5a, 0x28, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6b, 0x75, 0x6d,
	0x61, 0x68, 0x71, 0x2f, 0x6b, 0x75, 0x6d, 0x61, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x6d, 0x65, 0x73,
	0x68, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x8a, 0xb5, 0x18, 0x1e, 0x50, 0x01,
	0xa2, 0x01, 0x0b, 0x4d, 0x65, 0x73, 0x68, 0x47, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0xf2, 0x01,
	0x0b, 0x6d, 0x65, 0x73, 0x68, 0x67, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_mesh_v1alpha1_gateway_proto_rawDescOnce sync.Once
	file_mesh_v1alpha1_gateway_proto_rawDescData = file_mesh_v1alpha1_gateway_proto_rawDesc
)

func file_mesh_v1alpha1_gateway_proto_rawDescGZIP() []byte {
	file_mesh_v1alpha1_gateway_proto_rawDescOnce.Do(func() {
		file_mesh_v1alpha1_gateway_proto_rawDescData = protoimpl.X.CompressGZIP(file_mesh_v1alpha1_gateway_proto_rawDescData)
	})
	return file_mesh_v1alpha1_gateway_proto_rawDescData
}

var file_mesh_v1alpha1_gateway_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_mesh_v1alpha1_gateway_proto_msgTypes = make([]protoimpl.MessageInfo, 9)
var file_mesh_v1alpha1_gateway_proto_goTypes = []interface{}{
	(MeshGateway_TLS_Mode)(0),              // 0: kuma.mesh.v1alpha1.MeshGateway.TLS.Mode
	(MeshGateway_Listener_Protocol)(0),     // 1: kuma.mesh.v1alpha1.MeshGateway.Listener.Protocol
	(*MeshGateway)(nil),                    // 2: kuma.mesh.v1alpha1.MeshGateway
	(*MeshGateway_TLS)(nil),                // 3: kuma.mesh.v1alpha1.MeshGateway.TLS
	(*MeshGateway_Listener)(nil),           // 4: kuma.mesh.v1alpha1.MeshGateway.Listener
	(*MeshGateway_Conf)(nil),               // 5: kuma.mesh.v1alpha1.MeshGateway.Conf
	nil,                                    // 6: kuma.mesh.v1alpha1.MeshGateway.TagsEntry
	(*MeshGateway_TLS_Options)(nil),        // 7: kuma.mesh.v1alpha1.MeshGateway.TLS.Options
	(*MeshGateway_TLS_Conf)(nil),           // 8: kuma.mesh.v1alpha1.MeshGateway.TLS.Conf
	(*MeshGateway_Listener_Resources)(nil), // 9: kuma.mesh.v1alpha1.MeshGateway.Listener.Resources
	nil,                                    // 10: kuma.mesh.v1alpha1.MeshGateway.Listener.TagsEntry
	(*Selector)(nil),                       // 11: kuma.mesh.v1alpha1.Selector
	(*v1alpha1.DataSource)(nil),            // 12: kuma.system.v1alpha1.DataSource
}
var file_mesh_v1alpha1_gateway_proto_depIdxs = []int32{
	11, // 0: kuma.mesh.v1alpha1.MeshGateway.selectors:type_name -> kuma.mesh.v1alpha1.Selector
	6,  // 1: kuma.mesh.v1alpha1.MeshGateway.tags:type_name -> kuma.mesh.v1alpha1.MeshGateway.TagsEntry
	5,  // 2: kuma.mesh.v1alpha1.MeshGateway.conf:type_name -> kuma.mesh.v1alpha1.MeshGateway.Conf
	1,  // 3: kuma.mesh.v1alpha1.MeshGateway.Listener.protocol:type_name -> kuma.mesh.v1alpha1.MeshGateway.Listener.Protocol
	8,  // 4: kuma.mesh.v1alpha1.MeshGateway.Listener.tls:type_name -> kuma.mesh.v1alpha1.MeshGateway.TLS.Conf
	10, // 5: kuma.mesh.v1alpha1.MeshGateway.Listener.tags:type_name -> kuma.mesh.v1alpha1.MeshGateway.Listener.TagsEntry
	9,  // 6: kuma.mesh.v1alpha1.MeshGateway.Listener.resources:type_name -> kuma.mesh.v1alpha1.MeshGateway.Listener.Resources
	4,  // 7: kuma.mesh.v1alpha1.MeshGateway.Conf.listeners:type_name -> kuma.mesh.v1alpha1.MeshGateway.Listener
	0,  // 8: kuma.mesh.v1alpha1.MeshGateway.TLS.Conf.mode:type_name -> kuma.mesh.v1alpha1.MeshGateway.TLS.Mode
	12, // 9: kuma.mesh.v1alpha1.MeshGateway.TLS.Conf.certificates:type_name -> kuma.system.v1alpha1.DataSource
	7,  // 10: kuma.mesh.v1alpha1.MeshGateway.TLS.Conf.options:type_name -> kuma.mesh.v1alpha1.MeshGateway.TLS.Options
	11, // [11:11] is the sub-list for method output_type
	11, // [11:11] is the sub-list for method input_type
	11, // [11:11] is the sub-list for extension type_name
	11, // [11:11] is the sub-list for extension extendee
	0,  // [0:11] is the sub-list for field type_name
}

func init() { file_mesh_v1alpha1_gateway_proto_init() }
func file_mesh_v1alpha1_gateway_proto_init() {
	if File_mesh_v1alpha1_gateway_proto != nil {
		return
	}
	file_mesh_v1alpha1_selector_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_mesh_v1alpha1_gateway_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MeshGateway); i {
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
		file_mesh_v1alpha1_gateway_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MeshGateway_TLS); i {
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
		file_mesh_v1alpha1_gateway_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MeshGateway_Listener); i {
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
		file_mesh_v1alpha1_gateway_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MeshGateway_Conf); i {
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
		file_mesh_v1alpha1_gateway_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MeshGateway_TLS_Options); i {
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
		file_mesh_v1alpha1_gateway_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MeshGateway_TLS_Conf); i {
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
		file_mesh_v1alpha1_gateway_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MeshGateway_Listener_Resources); i {
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
			RawDescriptor: file_mesh_v1alpha1_gateway_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   9,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_mesh_v1alpha1_gateway_proto_goTypes,
		DependencyIndexes: file_mesh_v1alpha1_gateway_proto_depIdxs,
		EnumInfos:         file_mesh_v1alpha1_gateway_proto_enumTypes,
		MessageInfos:      file_mesh_v1alpha1_gateway_proto_msgTypes,
	}.Build()
	File_mesh_v1alpha1_gateway_proto = out.File
	file_mesh_v1alpha1_gateway_proto_rawDesc = nil
	file_mesh_v1alpha1_gateway_proto_goTypes = nil
	file_mesh_v1alpha1_gateway_proto_depIdxs = nil
}
