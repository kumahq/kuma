// Code generated by protoc-gen-go. DO NOT EDIT.
// source: mesh/v1alpha1/dataplane.proto

package v1alpha1

import (
	fmt "fmt"
	_ "github.com/envoyproxy/protoc-gen-validate/validate"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// Dataplane defines configuration of a side-car proxy.
type Dataplane struct {
	// Networking describes inbound and outbound interfaces of the dataplane.
	Networking *Dataplane_Networking `protobuf:"bytes,1,opt,name=networking,proto3" json:"networking,omitempty"`
	// Configuration for metrics that should be collected and exposed by the
	// dataplane.
	//
	// Settings defined here will override their respective defaults
	// defined at a Mesh level.
	Metrics              *MetricsBackend `protobuf:"bytes,2,opt,name=metrics,proto3" json:"metrics,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *Dataplane) Reset()         { *m = Dataplane{} }
func (m *Dataplane) String() string { return proto.CompactTextString(m) }
func (*Dataplane) ProtoMessage()    {}
func (*Dataplane) Descriptor() ([]byte, []int) {
	return fileDescriptor_7608682fd5ea84a4, []int{0}
}

func (m *Dataplane) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Dataplane.Unmarshal(m, b)
}
func (m *Dataplane) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Dataplane.Marshal(b, m, deterministic)
}
func (m *Dataplane) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Dataplane.Merge(m, src)
}
func (m *Dataplane) XXX_Size() int {
	return xxx_messageInfo_Dataplane.Size(m)
}
func (m *Dataplane) XXX_DiscardUnknown() {
	xxx_messageInfo_Dataplane.DiscardUnknown(m)
}

var xxx_messageInfo_Dataplane proto.InternalMessageInfo

func (m *Dataplane) GetNetworking() *Dataplane_Networking {
	if m != nil {
		return m.Networking
	}
	return nil
}

func (m *Dataplane) GetMetrics() *MetricsBackend {
	if m != nil {
		return m.Metrics
	}
	return nil
}

// Networking describes inbound and outbound interfaces of a dataplane.
type Dataplane_Networking struct {
	// Ingress if not nil, dataplane will be work in the Ingress mode
	Ingress *Dataplane_Networking_Ingress `protobuf:"bytes,6,opt,name=ingress,proto3" json:"ingress,omitempty"`
	// Public IP on which the dataplane is accessible in the network.
	// Host names and DNS are not allowed.
	Address string `protobuf:"bytes,5,opt,name=address,proto3" json:"address,omitempty"`
	// Gateway describes configuration of gateway of the dataplane.
	Gateway *Dataplane_Networking_Gateway `protobuf:"bytes,3,opt,name=gateway,proto3" json:"gateway,omitempty"`
	// Inbound describes a list of inbound interfaces of the dataplane.
	Inbound []*Dataplane_Networking_Inbound `protobuf:"bytes,1,rep,name=inbound,proto3" json:"inbound,omitempty"`
	// Outbound describes a list of outbound interfaces of the dataplane.
	Outbound []*Dataplane_Networking_Outbound `protobuf:"bytes,2,rep,name=outbound,proto3" json:"outbound,omitempty"`
	// TransparentProxying describes configuration for transparent proxying.
	TransparentProxying  *Dataplane_Networking_TransparentProxying `protobuf:"bytes,4,opt,name=transparent_proxying,json=transparentProxying,proto3" json:"transparent_proxying,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                                  `json:"-"`
	XXX_unrecognized     []byte                                    `json:"-"`
	XXX_sizecache        int32                                     `json:"-"`
}

func (m *Dataplane_Networking) Reset()         { *m = Dataplane_Networking{} }
func (m *Dataplane_Networking) String() string { return proto.CompactTextString(m) }
func (*Dataplane_Networking) ProtoMessage()    {}
func (*Dataplane_Networking) Descriptor() ([]byte, []int) {
	return fileDescriptor_7608682fd5ea84a4, []int{0, 0}
}

func (m *Dataplane_Networking) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Dataplane_Networking.Unmarshal(m, b)
}
func (m *Dataplane_Networking) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Dataplane_Networking.Marshal(b, m, deterministic)
}
func (m *Dataplane_Networking) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Dataplane_Networking.Merge(m, src)
}
func (m *Dataplane_Networking) XXX_Size() int {
	return xxx_messageInfo_Dataplane_Networking.Size(m)
}
func (m *Dataplane_Networking) XXX_DiscardUnknown() {
	xxx_messageInfo_Dataplane_Networking.DiscardUnknown(m)
}

var xxx_messageInfo_Dataplane_Networking proto.InternalMessageInfo

func (m *Dataplane_Networking) GetIngress() *Dataplane_Networking_Ingress {
	if m != nil {
		return m.Ingress
	}
	return nil
}

func (m *Dataplane_Networking) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func (m *Dataplane_Networking) GetGateway() *Dataplane_Networking_Gateway {
	if m != nil {
		return m.Gateway
	}
	return nil
}

func (m *Dataplane_Networking) GetInbound() []*Dataplane_Networking_Inbound {
	if m != nil {
		return m.Inbound
	}
	return nil
}

func (m *Dataplane_Networking) GetOutbound() []*Dataplane_Networking_Outbound {
	if m != nil {
		return m.Outbound
	}
	return nil
}

func (m *Dataplane_Networking) GetTransparentProxying() *Dataplane_Networking_TransparentProxying {
	if m != nil {
		return m.TransparentProxying
	}
	return nil
}

// Ingress allows us to configure dataplane in the Ingress mode. In this
// mode, dataplane has only inbound interfaces (outbound and gateway
// prohibited). Every inbound interface matches with services that reside in
// that cluster.
type Dataplane_Networking_Ingress struct {
	AvailableServices    []*Dataplane_Networking_Ingress_AvailableService `protobuf:"bytes,1,rep,name=availableServices,proto3" json:"availableServices,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                                         `json:"-"`
	XXX_unrecognized     []byte                                           `json:"-"`
	XXX_sizecache        int32                                            `json:"-"`
}

func (m *Dataplane_Networking_Ingress) Reset()         { *m = Dataplane_Networking_Ingress{} }
func (m *Dataplane_Networking_Ingress) String() string { return proto.CompactTextString(m) }
func (*Dataplane_Networking_Ingress) ProtoMessage()    {}
func (*Dataplane_Networking_Ingress) Descriptor() ([]byte, []int) {
	return fileDescriptor_7608682fd5ea84a4, []int{0, 0, 0}
}

func (m *Dataplane_Networking_Ingress) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Dataplane_Networking_Ingress.Unmarshal(m, b)
}
func (m *Dataplane_Networking_Ingress) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Dataplane_Networking_Ingress.Marshal(b, m, deterministic)
}
func (m *Dataplane_Networking_Ingress) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Dataplane_Networking_Ingress.Merge(m, src)
}
func (m *Dataplane_Networking_Ingress) XXX_Size() int {
	return xxx_messageInfo_Dataplane_Networking_Ingress.Size(m)
}
func (m *Dataplane_Networking_Ingress) XXX_DiscardUnknown() {
	xxx_messageInfo_Dataplane_Networking_Ingress.DiscardUnknown(m)
}

var xxx_messageInfo_Dataplane_Networking_Ingress proto.InternalMessageInfo

func (m *Dataplane_Networking_Ingress) GetAvailableServices() []*Dataplane_Networking_Ingress_AvailableService {
	if m != nil {
		return m.AvailableServices
	}
	return nil
}

// AvailableService contains tags that represent unique subset of
// endpoints
type Dataplane_Networking_Ingress_AvailableService struct {
	// tags of the service
	Tags map[string]string `protobuf:"bytes,1,rep,name=tags,proto3" json:"tags,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// number of instances available for given tags
	Instances uint32 `protobuf:"varint,2,opt,name=instances,proto3" json:"instances,omitempty"`
	// mesh of the instances available for given tags
	Mesh                 string   `protobuf:"bytes,3,opt,name=mesh,proto3" json:"mesh,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Dataplane_Networking_Ingress_AvailableService) Reset() {
	*m = Dataplane_Networking_Ingress_AvailableService{}
}
func (m *Dataplane_Networking_Ingress_AvailableService) String() string {
	return proto.CompactTextString(m)
}
func (*Dataplane_Networking_Ingress_AvailableService) ProtoMessage() {}
func (*Dataplane_Networking_Ingress_AvailableService) Descriptor() ([]byte, []int) {
	return fileDescriptor_7608682fd5ea84a4, []int{0, 0, 0, 0}
}

func (m *Dataplane_Networking_Ingress_AvailableService) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Dataplane_Networking_Ingress_AvailableService.Unmarshal(m, b)
}
func (m *Dataplane_Networking_Ingress_AvailableService) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Dataplane_Networking_Ingress_AvailableService.Marshal(b, m, deterministic)
}
func (m *Dataplane_Networking_Ingress_AvailableService) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Dataplane_Networking_Ingress_AvailableService.Merge(m, src)
}
func (m *Dataplane_Networking_Ingress_AvailableService) XXX_Size() int {
	return xxx_messageInfo_Dataplane_Networking_Ingress_AvailableService.Size(m)
}
func (m *Dataplane_Networking_Ingress_AvailableService) XXX_DiscardUnknown() {
	xxx_messageInfo_Dataplane_Networking_Ingress_AvailableService.DiscardUnknown(m)
}

var xxx_messageInfo_Dataplane_Networking_Ingress_AvailableService proto.InternalMessageInfo

func (m *Dataplane_Networking_Ingress_AvailableService) GetTags() map[string]string {
	if m != nil {
		return m.Tags
	}
	return nil
}

func (m *Dataplane_Networking_Ingress_AvailableService) GetInstances() uint32 {
	if m != nil {
		return m.Instances
	}
	return 0
}

func (m *Dataplane_Networking_Ingress_AvailableService) GetMesh() string {
	if m != nil {
		return m.Mesh
	}
	return ""
}

// Inbound describes a service implemented by the dataplane.
type Dataplane_Networking_Inbound struct {
	// Port of the inbound interface that will forward requests to the
	// service.
	Port uint32 `protobuf:"varint,3,opt,name=port,proto3" json:"port,omitempty"`
	// Port of the service that requests will be forwarded to.
	ServicePort uint32 `protobuf:"varint,4,opt,name=servicePort,proto3" json:"servicePort,omitempty"`
	// Address on which inbound listener will be exposed. Defaults to
	// networking.address.
	Address string `protobuf:"bytes,5,opt,name=address,proto3" json:"address,omitempty"`
	// Tags associated with an application this dataplane is deployed next to,
	// e.g. service=web, version=1.0.
	// `service` tag is mandatory.
	Tags                 map[string]string `protobuf:"bytes,2,rep,name=tags,proto3" json:"tags,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *Dataplane_Networking_Inbound) Reset()         { *m = Dataplane_Networking_Inbound{} }
func (m *Dataplane_Networking_Inbound) String() string { return proto.CompactTextString(m) }
func (*Dataplane_Networking_Inbound) ProtoMessage()    {}
func (*Dataplane_Networking_Inbound) Descriptor() ([]byte, []int) {
	return fileDescriptor_7608682fd5ea84a4, []int{0, 0, 1}
}

func (m *Dataplane_Networking_Inbound) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Dataplane_Networking_Inbound.Unmarshal(m, b)
}
func (m *Dataplane_Networking_Inbound) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Dataplane_Networking_Inbound.Marshal(b, m, deterministic)
}
func (m *Dataplane_Networking_Inbound) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Dataplane_Networking_Inbound.Merge(m, src)
}
func (m *Dataplane_Networking_Inbound) XXX_Size() int {
	return xxx_messageInfo_Dataplane_Networking_Inbound.Size(m)
}
func (m *Dataplane_Networking_Inbound) XXX_DiscardUnknown() {
	xxx_messageInfo_Dataplane_Networking_Inbound.DiscardUnknown(m)
}

var xxx_messageInfo_Dataplane_Networking_Inbound proto.InternalMessageInfo

func (m *Dataplane_Networking_Inbound) GetPort() uint32 {
	if m != nil {
		return m.Port
	}
	return 0
}

func (m *Dataplane_Networking_Inbound) GetServicePort() uint32 {
	if m != nil {
		return m.ServicePort
	}
	return 0
}

func (m *Dataplane_Networking_Inbound) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func (m *Dataplane_Networking_Inbound) GetTags() map[string]string {
	if m != nil {
		return m.Tags
	}
	return nil
}

// Outbound describes a service consumed by the dataplane.
type Dataplane_Networking_Outbound struct {
	// Address on which the service will be available to this dataplane.
	// Defaults to 127.0.0.1
	Address string `protobuf:"bytes,3,opt,name=address,proto3" json:"address,omitempty"`
	// Port on which the service will be available to this dataplane.
	Port uint32 `protobuf:"varint,4,opt,name=port,proto3" json:"port,omitempty"`
	// DEPRECATED: use networking.outbound[].tags
	// Service name.
	Service string `protobuf:"bytes,2,opt,name=service,proto3" json:"service,omitempty"`
	// Tags
	Tags                 map[string]string `protobuf:"bytes,5,rep,name=tags,proto3" json:"tags,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *Dataplane_Networking_Outbound) Reset()         { *m = Dataplane_Networking_Outbound{} }
func (m *Dataplane_Networking_Outbound) String() string { return proto.CompactTextString(m) }
func (*Dataplane_Networking_Outbound) ProtoMessage()    {}
func (*Dataplane_Networking_Outbound) Descriptor() ([]byte, []int) {
	return fileDescriptor_7608682fd5ea84a4, []int{0, 0, 2}
}

func (m *Dataplane_Networking_Outbound) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Dataplane_Networking_Outbound.Unmarshal(m, b)
}
func (m *Dataplane_Networking_Outbound) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Dataplane_Networking_Outbound.Marshal(b, m, deterministic)
}
func (m *Dataplane_Networking_Outbound) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Dataplane_Networking_Outbound.Merge(m, src)
}
func (m *Dataplane_Networking_Outbound) XXX_Size() int {
	return xxx_messageInfo_Dataplane_Networking_Outbound.Size(m)
}
func (m *Dataplane_Networking_Outbound) XXX_DiscardUnknown() {
	xxx_messageInfo_Dataplane_Networking_Outbound.DiscardUnknown(m)
}

var xxx_messageInfo_Dataplane_Networking_Outbound proto.InternalMessageInfo

func (m *Dataplane_Networking_Outbound) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func (m *Dataplane_Networking_Outbound) GetPort() uint32 {
	if m != nil {
		return m.Port
	}
	return 0
}

func (m *Dataplane_Networking_Outbound) GetService() string {
	if m != nil {
		return m.Service
	}
	return ""
}

func (m *Dataplane_Networking_Outbound) GetTags() map[string]string {
	if m != nil {
		return m.Tags
	}
	return nil
}

// Gateway describes a service that ingress should not be proxied.
type Dataplane_Networking_Gateway struct {
	// Tags associated with a gateway (e.g., Kong, Contour, etc) this
	// dataplane is deployed next to, e.g. service=gateway, env=prod.
	// `service` tag is mandatory.
	Tags                 map[string]string `protobuf:"bytes,1,rep,name=tags,proto3" json:"tags,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *Dataplane_Networking_Gateway) Reset()         { *m = Dataplane_Networking_Gateway{} }
func (m *Dataplane_Networking_Gateway) String() string { return proto.CompactTextString(m) }
func (*Dataplane_Networking_Gateway) ProtoMessage()    {}
func (*Dataplane_Networking_Gateway) Descriptor() ([]byte, []int) {
	return fileDescriptor_7608682fd5ea84a4, []int{0, 0, 3}
}

func (m *Dataplane_Networking_Gateway) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Dataplane_Networking_Gateway.Unmarshal(m, b)
}
func (m *Dataplane_Networking_Gateway) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Dataplane_Networking_Gateway.Marshal(b, m, deterministic)
}
func (m *Dataplane_Networking_Gateway) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Dataplane_Networking_Gateway.Merge(m, src)
}
func (m *Dataplane_Networking_Gateway) XXX_Size() int {
	return xxx_messageInfo_Dataplane_Networking_Gateway.Size(m)
}
func (m *Dataplane_Networking_Gateway) XXX_DiscardUnknown() {
	xxx_messageInfo_Dataplane_Networking_Gateway.DiscardUnknown(m)
}

var xxx_messageInfo_Dataplane_Networking_Gateway proto.InternalMessageInfo

func (m *Dataplane_Networking_Gateway) GetTags() map[string]string {
	if m != nil {
		return m.Tags
	}
	return nil
}

// TransparentProxying describes configuration for transparent proxying.
type Dataplane_Networking_TransparentProxying struct {
	// Port on which all inbound traffic is being transparently redirected.
	RedirectPortInbound uint32 `protobuf:"varint,1,opt,name=redirect_port_inbound,json=redirectPortInbound,proto3" json:"redirect_port_inbound,omitempty"`
	// Port on which all outbound traffic is being transparently redirected.
	RedirectPortOutbound uint32 `protobuf:"varint,2,opt,name=redirect_port_outbound,json=redirectPortOutbound,proto3" json:"redirect_port_outbound,omitempty"`
	// List of services that will be access directly via IP:PORT
	DirectAccessServices []string `protobuf:"bytes,3,rep,name=direct_access_services,json=directAccessServices,proto3" json:"direct_access_services,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Dataplane_Networking_TransparentProxying) Reset() {
	*m = Dataplane_Networking_TransparentProxying{}
}
func (m *Dataplane_Networking_TransparentProxying) String() string { return proto.CompactTextString(m) }
func (*Dataplane_Networking_TransparentProxying) ProtoMessage()    {}
func (*Dataplane_Networking_TransparentProxying) Descriptor() ([]byte, []int) {
	return fileDescriptor_7608682fd5ea84a4, []int{0, 0, 4}
}

func (m *Dataplane_Networking_TransparentProxying) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Dataplane_Networking_TransparentProxying.Unmarshal(m, b)
}
func (m *Dataplane_Networking_TransparentProxying) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Dataplane_Networking_TransparentProxying.Marshal(b, m, deterministic)
}
func (m *Dataplane_Networking_TransparentProxying) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Dataplane_Networking_TransparentProxying.Merge(m, src)
}
func (m *Dataplane_Networking_TransparentProxying) XXX_Size() int {
	return xxx_messageInfo_Dataplane_Networking_TransparentProxying.Size(m)
}
func (m *Dataplane_Networking_TransparentProxying) XXX_DiscardUnknown() {
	xxx_messageInfo_Dataplane_Networking_TransparentProxying.DiscardUnknown(m)
}

var xxx_messageInfo_Dataplane_Networking_TransparentProxying proto.InternalMessageInfo

func (m *Dataplane_Networking_TransparentProxying) GetRedirectPortInbound() uint32 {
	if m != nil {
		return m.RedirectPortInbound
	}
	return 0
}

func (m *Dataplane_Networking_TransparentProxying) GetRedirectPortOutbound() uint32 {
	if m != nil {
		return m.RedirectPortOutbound
	}
	return 0
}

func (m *Dataplane_Networking_TransparentProxying) GetDirectAccessServices() []string {
	if m != nil {
		return m.DirectAccessServices
	}
	return nil
}

func init() {
	proto.RegisterType((*Dataplane)(nil), "kuma.mesh.v1alpha1.Dataplane")
	proto.RegisterType((*Dataplane_Networking)(nil), "kuma.mesh.v1alpha1.Dataplane.Networking")
	proto.RegisterType((*Dataplane_Networking_Ingress)(nil), "kuma.mesh.v1alpha1.Dataplane.Networking.Ingress")
	proto.RegisterType((*Dataplane_Networking_Ingress_AvailableService)(nil), "kuma.mesh.v1alpha1.Dataplane.Networking.Ingress.AvailableService")
	proto.RegisterMapType((map[string]string)(nil), "kuma.mesh.v1alpha1.Dataplane.Networking.Ingress.AvailableService.TagsEntry")
	proto.RegisterType((*Dataplane_Networking_Inbound)(nil), "kuma.mesh.v1alpha1.Dataplane.Networking.Inbound")
	proto.RegisterMapType((map[string]string)(nil), "kuma.mesh.v1alpha1.Dataplane.Networking.Inbound.TagsEntry")
	proto.RegisterType((*Dataplane_Networking_Outbound)(nil), "kuma.mesh.v1alpha1.Dataplane.Networking.Outbound")
	proto.RegisterMapType((map[string]string)(nil), "kuma.mesh.v1alpha1.Dataplane.Networking.Outbound.TagsEntry")
	proto.RegisterType((*Dataplane_Networking_Gateway)(nil), "kuma.mesh.v1alpha1.Dataplane.Networking.Gateway")
	proto.RegisterMapType((map[string]string)(nil), "kuma.mesh.v1alpha1.Dataplane.Networking.Gateway.TagsEntry")
	proto.RegisterType((*Dataplane_Networking_TransparentProxying)(nil), "kuma.mesh.v1alpha1.Dataplane.Networking.TransparentProxying")
}

func init() { proto.RegisterFile("mesh/v1alpha1/dataplane.proto", fileDescriptor_7608682fd5ea84a4) }

var fileDescriptor_7608682fd5ea84a4 = []byte{
	// 657 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xb4, 0x95, 0x4d, 0x6f, 0xd3, 0x30,
	0x18, 0xc7, 0xe5, 0x26, 0x5d, 0x9a, 0x67, 0x0c, 0x0d, 0xaf, 0x8c, 0x28, 0x80, 0x54, 0x76, 0xaa,
	0x38, 0x64, 0x0c, 0x90, 0x40, 0x63, 0x08, 0x2d, 0x02, 0xf1, 0xa6, 0xb1, 0xc9, 0xec, 0x80, 0xb8,
	0x54, 0x5e, 0x63, 0x75, 0x51, 0xbb, 0xa4, 0xb2, 0xdd, 0x8e, 0x5e, 0xf9, 0x18, 0x88, 0x6f, 0xc4,
	0x47, 0x40, 0xe2, 0xc6, 0x6d, 0x9f, 0x60, 0x97, 0xa1, 0xd8, 0x71, 0x9a, 0x75, 0x13, 0x6a, 0x86,
	0xb8, 0xb9, 0xf6, 0xf3, 0xff, 0x3d, 0xaf, 0x79, 0x0a, 0x77, 0x8f, 0x98, 0x38, 0x5c, 0x1f, 0x6f,
	0xd0, 0xc1, 0xf0, 0x90, 0x6e, 0xac, 0x47, 0x54, 0xd2, 0xe1, 0x80, 0x26, 0x2c, 0x18, 0xf2, 0x54,
	0xa6, 0x18, 0xf7, 0x47, 0x47, 0x34, 0xc8, 0x6c, 0x02, 0x63, 0xe3, 0xdf, 0x3e, 0x2f, 0x39, 0x62,
	0x92, 0xc7, 0x5d, 0xa1, 0x05, 0xfe, 0xad, 0x31, 0x1d, 0xc4, 0x11, 0x95, 0x6c, 0xdd, 0x1c, 0xf4,
	0xc3, 0xda, 0xd7, 0xeb, 0xe0, 0xbe, 0x34, 0x74, 0xfc, 0x06, 0x20, 0x61, 0xf2, 0x38, 0xe5, 0xfd,
	0x38, 0xe9, 0x79, 0xa8, 0x85, 0xda, 0x8b, 0x0f, 0xdb, 0xc1, 0x45, 0x67, 0x41, 0x21, 0x09, 0x3e,
	0x14, 0xf6, 0xa4, 0xa4, 0xc5, 0x5b, 0xe0, 0xe4, 0x11, 0x78, 0x35, 0x85, 0x59, 0xbb, 0x0c, 0xb3,
	0xa3, 0x4d, 0x42, 0xda, 0xed, 0xb3, 0x24, 0x22, 0x46, 0xe2, 0x9f, 0x5c, 0x03, 0x98, 0x82, 0xf1,
	0x3b, 0x70, 0xe2, 0xa4, 0xc7, 0x99, 0x10, 0xde, 0x82, 0x82, 0x3d, 0x98, 0x37, 0xa6, 0xe0, 0xad,
	0xd6, 0x11, 0x03, 0xc0, 0x1e, 0x38, 0x34, 0x8a, 0x14, 0xab, 0xde, 0x42, 0x6d, 0x97, 0x98, 0x9f,
	0x99, 0x97, 0x1e, 0x95, 0xec, 0x98, 0x4e, 0x3c, 0xab, 0xa2, 0x97, 0xd7, 0x5a, 0x47, 0x0c, 0x40,
	0x47, 0x7c, 0x90, 0x8e, 0x92, 0xc8, 0x43, 0x2d, 0xab, 0x62, 0xc4, 0x4a, 0x47, 0x0c, 0x00, 0xef,
	0x40, 0x23, 0x1d, 0x49, 0x0d, 0xab, 0x29, 0xd8, 0xc6, 0xdc, 0xb0, 0xdd, 0x5c, 0x48, 0x0a, 0x04,
	0x4e, 0xa1, 0x29, 0x39, 0x4d, 0xc4, 0x90, 0x72, 0x96, 0xc8, 0xce, 0x90, 0xa7, 0x5f, 0x26, 0x59,
	0xb7, 0x6d, 0x95, 0xf3, 0xd6, 0xdc, 0xe8, 0xfd, 0x29, 0x64, 0x2f, 0x67, 0x90, 0x15, 0x79, 0xf1,
	0xd2, 0xff, 0x59, 0x03, 0x27, 0x6f, 0x03, 0x4e, 0xe1, 0x06, 0x1d, 0xd3, 0x78, 0x40, 0x0f, 0x06,
	0xec, 0x23, 0xe3, 0xe3, 0xb8, 0xcb, 0x44, 0x5e, 0xa1, 0xed, 0xaa, 0x3d, 0x0d, 0xb6, 0x67, 0x48,
	0xe4, 0x22, 0xdb, 0xff, 0x85, 0x60, 0x79, 0xd6, 0x0e, 0x77, 0xc0, 0x96, 0xb4, 0x67, 0x1c, 0xbf,
	0xff, 0x67, 0xc7, 0xc1, 0x3e, 0xed, 0x89, 0x57, 0x89, 0xe4, 0x13, 0xa2, 0xc0, 0xf8, 0x0e, 0xb8,
	0x71, 0x22, 0x24, 0x4d, 0xb2, 0xf4, 0xb2, 0xf9, 0x5f, 0x22, 0xd3, 0x0b, 0x8c, 0xc1, 0xce, 0x9c,
	0xa9, 0x29, 0x73, 0x89, 0x3a, 0xfb, 0x4f, 0xc0, 0x2d, 0x20, 0x78, 0x19, 0xac, 0x3e, 0x9b, 0xa8,
	0xef, 0xcf, 0x25, 0xd9, 0x11, 0x37, 0xa1, 0x3e, 0xa6, 0x83, 0x11, 0x53, 0x30, 0x97, 0xe8, 0x1f,
	0x9b, 0xb5, 0xa7, 0xc8, 0x3f, 0x41, 0x59, 0x75, 0x75, 0x6b, 0x31, 0xd8, 0xc3, 0x94, 0x4b, 0x05,
	0x5e, 0x22, 0xea, 0x8c, 0x5b, 0xb0, 0x28, 0x74, 0x94, 0x7b, 0xd9, 0x93, 0xad, 0x9e, 0xca, 0x57,
	0x7f, 0xf9, 0x22, 0x3e, 0xe5, 0x75, 0xd2, 0x53, 0xb7, 0x59, 0x75, 0x84, 0xa7, 0x65, 0x09, 0x1b,
	0xa7, 0x61, 0xfd, 0x1b, 0xaa, 0x35, 0x90, 0x2e, 0xd0, 0xd5, 0xd3, 0xfd, 0x8d, 0xa0, 0x61, 0x86,
	0xba, 0x1c, 0xb9, 0x75, 0x3e, 0x72, 0x53, 0x09, 0xbb, 0x54, 0x89, 0x7b, 0xe0, 0xe4, 0x69, 0x6b,
	0x6c, 0xe8, 0x9c, 0x86, 0x36, 0xaf, 0x1d, 0x22, 0x62, 0xee, 0xf1, 0x6e, 0x9e, 0x70, 0x5d, 0x25,
	0xfc, 0xac, 0xf2, 0x67, 0x36, 0x3b, 0x08, 0x57, 0xcf, 0xf3, 0x3b, 0x02, 0x27, 0xdf, 0x2a, 0x45,
	0x1b, 0x50, 0xc5, 0x36, 0xe4, 0xfa, 0xff, 0xd1, 0x86, 0x1f, 0x08, 0x56, 0x2e, 0x59, 0x00, 0xf8,
	0x39, 0xdc, 0xe4, 0x2c, 0x8a, 0x39, 0xeb, 0xca, 0x4e, 0x56, 0xf4, 0xce, 0x74, 0x0b, 0xa2, 0xf6,
	0x52, 0xe8, 0x9e, 0x86, 0x0b, 0xf7, 0x6d, 0xef, 0xec, 0xcc, 0x22, 0x2b, 0xc6, 0x2e, 0x9b, 0x41,
	0x33, 0xc0, 0x2f, 0x60, 0xf5, 0xbc, 0xbc, 0xb4, 0xf8, 0x66, 0xf4, 0xcd, 0xb2, 0xbe, 0x98, 0x88,
	0xc7, 0xb0, 0x9a, 0xcb, 0x69, 0xb7, 0xcb, 0x84, 0xe8, 0x08, 0xb3, 0x64, 0xac, 0x96, 0xd5, 0x76,
	0x49, 0x53, 0xbf, 0x6e, 0xab, 0x47, 0xb3, 0x24, 0x42, 0xf8, 0xdc, 0x30, 0x95, 0x3c, 0x58, 0x50,
	0xff, 0x8b, 0x8f, 0xfe, 0x04, 0x00, 0x00, 0xff, 0xff, 0x33, 0xd8, 0xc6, 0xbe, 0x82, 0x07, 0x00,
	0x00,
}
