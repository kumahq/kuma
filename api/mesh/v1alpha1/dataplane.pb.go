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
	Metrics              *Metrics `protobuf:"bytes,2,opt,name=metrics,proto3" json:"metrics,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
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

func (m *Dataplane) GetMetrics() *Metrics {
	if m != nil {
		return m.Metrics
	}
	return nil
}

// Networking describes inbound and outbound interfaces of a dataplane.
type Dataplane_Networking struct {
	// Public address on which the dataplane is accessible in the network.
	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	// Inbound describes a list of inbound interfaces of the dataplane.
	Inbound []*Dataplane_Networking_Inbound `protobuf:"bytes,2,rep,name=inbound,proto3" json:"inbound,omitempty"`
	// Outbound describes a list of outbound interfaces of the dataplane.
	Outbound []*Dataplane_Networking_Outbound `protobuf:"bytes,3,rep,name=outbound,proto3" json:"outbound,omitempty"`
	// Gateway describes configuration of gateway of the dataplane.
	Gateway *Dataplane_Networking_Gateway `protobuf:"bytes,4,opt,name=gateway,proto3" json:"gateway,omitempty"`
	// TransparentProxying describes configuration for transparent proxying.
	TransparentProxying  *Dataplane_Networking_TransparentProxying `protobuf:"bytes,5,opt,name=transparent_proxying,json=transparentProxying,proto3" json:"transparent_proxying,omitempty"`
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

func (m *Dataplane_Networking) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
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

func (m *Dataplane_Networking) GetGateway() *Dataplane_Networking_Gateway {
	if m != nil {
		return m.Gateway
	}
	return nil
}

func (m *Dataplane_Networking) GetTransparentProxying() *Dataplane_Networking_TransparentProxying {
	if m != nil {
		return m.TransparentProxying
	}
	return nil
}

// Inbound describes a service implemented by the dataplane.
type Dataplane_Networking_Inbound struct {
	// DEPRECATED: use networking.address, networking.inbound[].port and networking.inbound[].servicePort
	// Interface describes networking rules for incoming traffic.
	// The value is a string formatted as
	// <DATAPLANE_IP>:<DATAPLANE_PORT>:<WORKLOAD_PORT>, which means
	// that dataplane must listen on <DATAPLANE_IP>:<DATAPLANE_PORT>
	// and must dispatch to 127.0.0.1:<WORKLOAD_PORT>.
	//
	// E.g.,
	// "192.168.0.100:9090:8080" in case of IPv4 or
	// "[2001:db8::1]:7070:6060" in case of IPv6.
	Interface string `protobuf:"bytes,1,opt,name=interface,proto3" json:"interface,omitempty"`
	// Port of the inbound interface that will forward requests to the service.
	Port uint32 `protobuf:"varint,2,opt,name=port,proto3" json:"port,omitempty"`
	// Port of the service that requests will be forwarded to.
	ServicePort uint32 `protobuf:"varint,3,opt,name=servicePort,proto3" json:"servicePort,omitempty"`
	// Address on which inbound listener will be exposed. Defaults to networking.interface.
	Address string `protobuf:"bytes,4,opt,name=address,proto3" json:"address,omitempty"`
	// Tags associated with an application this dataplane is deployed next to,
	// e.g. service=web, version=1.0.
	// `service` tag is mandatory.
	Tags                 map[string]string `protobuf:"bytes,5,rep,name=tags,proto3" json:"tags,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *Dataplane_Networking_Inbound) Reset()         { *m = Dataplane_Networking_Inbound{} }
func (m *Dataplane_Networking_Inbound) String() string { return proto.CompactTextString(m) }
func (*Dataplane_Networking_Inbound) ProtoMessage()    {}
func (*Dataplane_Networking_Inbound) Descriptor() ([]byte, []int) {
	return fileDescriptor_7608682fd5ea84a4, []int{0, 0, 0}
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

func (m *Dataplane_Networking_Inbound) GetInterface() string {
	if m != nil {
		return m.Interface
	}
	return ""
}

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
	// DEPRECATED: use networking.interface and networking.outbound[].port
	// Interface describes networking rules for outgoing traffic.
	// The value is a string formatted as <DATAPLANE_IP>:<DATAPLANE_PORT>,
	// which means that dataplane must listen on
	// <DATAPLANE_IP>:<DATAPLANE_PORT> and must be dispatch to
	// <SERVICE>:<SERVICE_PORT>.
	//
	// E.g.,
	// "127.0.0.1:9090" in case of IPv4 or
	// "[::1]:8080" in case of IPv6 or
	// ":7070".
	Interface string `protobuf:"bytes,1,opt,name=interface,proto3" json:"interface,omitempty"`
	// Port on which the service will be available to this dataplane.
	Port uint32 `protobuf:"varint,2,opt,name=port,proto3" json:"port,omitempty"`
	// Service name.
	Service              string   `protobuf:"bytes,3,opt,name=service,proto3" json:"service,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Dataplane_Networking_Outbound) Reset()         { *m = Dataplane_Networking_Outbound{} }
func (m *Dataplane_Networking_Outbound) String() string { return proto.CompactTextString(m) }
func (*Dataplane_Networking_Outbound) ProtoMessage()    {}
func (*Dataplane_Networking_Outbound) Descriptor() ([]byte, []int) {
	return fileDescriptor_7608682fd5ea84a4, []int{0, 0, 1}
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

func (m *Dataplane_Networking_Outbound) GetInterface() string {
	if m != nil {
		return m.Interface
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
	return fileDescriptor_7608682fd5ea84a4, []int{0, 0, 2}
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
	// Port on which all traffic is being transparently redirected.
	RedirectPort         uint32   `protobuf:"varint,1,opt,name=redirect_port,json=redirectPort,proto3" json:"redirect_port,omitempty"`
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
	return fileDescriptor_7608682fd5ea84a4, []int{0, 0, 3}
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

func (m *Dataplane_Networking_TransparentProxying) GetRedirectPort() uint32 {
	if m != nil {
		return m.RedirectPort
	}
	return 0
}

func init() {
	proto.RegisterType((*Dataplane)(nil), "kuma.mesh.v1alpha1.Dataplane")
	proto.RegisterType((*Dataplane_Networking)(nil), "kuma.mesh.v1alpha1.Dataplane.Networking")
	proto.RegisterType((*Dataplane_Networking_Inbound)(nil), "kuma.mesh.v1alpha1.Dataplane.Networking.Inbound")
	proto.RegisterMapType((map[string]string)(nil), "kuma.mesh.v1alpha1.Dataplane.Networking.Inbound.TagsEntry")
	proto.RegisterType((*Dataplane_Networking_Outbound)(nil), "kuma.mesh.v1alpha1.Dataplane.Networking.Outbound")
	proto.RegisterType((*Dataplane_Networking_Gateway)(nil), "kuma.mesh.v1alpha1.Dataplane.Networking.Gateway")
	proto.RegisterMapType((map[string]string)(nil), "kuma.mesh.v1alpha1.Dataplane.Networking.Gateway.TagsEntry")
	proto.RegisterType((*Dataplane_Networking_TransparentProxying)(nil), "kuma.mesh.v1alpha1.Dataplane.Networking.TransparentProxying")
}

func init() { proto.RegisterFile("mesh/v1alpha1/dataplane.proto", fileDescriptor_7608682fd5ea84a4) }

var fileDescriptor_7608682fd5ea84a4 = []byte{
	// 512 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xb4, 0x94, 0xcf, 0x6e, 0xd3, 0x4c,
	0x14, 0xc5, 0xe5, 0x3f, 0xa9, 0xed, 0x9b, 0x2f, 0xd2, 0xa7, 0x69, 0x25, 0x2c, 0x17, 0xa4, 0x08,
	0x36, 0x11, 0x0b, 0xa7, 0x29, 0x42, 0xa0, 0x8a, 0x55, 0x04, 0x2a, 0x20, 0x15, 0xaa, 0x51, 0x57,
	0xdd, 0x54, 0xb7, 0xf1, 0x90, 0x58, 0x49, 0xc6, 0xd6, 0x78, 0x92, 0x92, 0x77, 0xe0, 0x09, 0x58,
	0xf0, 0x20, 0xac, 0x78, 0x0e, 0xde, 0xa0, 0x4f, 0x51, 0x34, 0xe3, 0x19, 0x27, 0x55, 0xbb, 0x48,
	0x91, 0xd8, 0x8d, 0xe7, 0xde, 0xf3, 0x9b, 0x7b, 0xcf, 0x89, 0x02, 0x4f, 0xe6, 0xac, 0x9a, 0xf4,
	0x97, 0x03, 0x9c, 0x95, 0x13, 0x1c, 0xf4, 0x33, 0x94, 0x58, 0xce, 0x90, 0xb3, 0xb4, 0x14, 0x85,
	0x2c, 0x08, 0x99, 0x2e, 0xe6, 0x98, 0xaa, 0x9e, 0xd4, 0xf6, 0x24, 0xfb, 0xb7, 0x25, 0x73, 0x26,
	0x45, 0x3e, 0xaa, 0x6a, 0x41, 0xf2, 0x68, 0x89, 0xb3, 0x3c, 0x43, 0xc9, 0xfa, 0xf6, 0x50, 0x17,
	0x9e, 0x5e, 0x87, 0x10, 0xbd, 0xb5, 0x74, 0xf2, 0x1e, 0x80, 0x33, 0x79, 0x55, 0x88, 0x69, 0xce,
	0xc7, 0xb1, 0xd3, 0x75, 0x7a, 0xed, 0xc3, 0x5e, 0x7a, 0xf7, 0xb1, 0xb4, 0x91, 0xa4, 0x9f, 0x9a,
	0x7e, 0xba, 0xa1, 0x25, 0x2f, 0x21, 0x30, 0x13, 0xc4, 0xae, 0xc6, 0xec, 0xdf, 0x87, 0x39, 0xa9,
	0x5b, 0xa8, 0xed, 0x4d, 0x7e, 0x07, 0x00, 0x6b, 0x22, 0x89, 0x21, 0xc0, 0x2c, 0x13, 0xac, 0xaa,
	0xf4, 0x30, 0x11, 0xb5, 0x9f, 0xe4, 0x23, 0x04, 0x39, 0xbf, 0x2c, 0x16, 0x3c, 0x8b, 0xdd, 0xae,
	0xd7, 0x6b, 0x1f, 0x1e, 0x6c, 0x3b, 0x66, 0xfa, 0xa1, 0xd6, 0x51, 0x0b, 0x20, 0x27, 0x10, 0x16,
	0x0b, 0x59, 0xc3, 0x3c, 0x0d, 0x1b, 0x6c, 0x0d, 0xfb, 0x6c, 0x84, 0xb4, 0x41, 0xa8, 0xd1, 0xc6,
	0x28, 0xd9, 0x15, 0xae, 0x62, 0x5f, 0xaf, 0xbe, 0xfd, 0x68, 0xc7, 0xb5, 0x8e, 0x5a, 0x00, 0x29,
	0x60, 0x4f, 0x0a, 0xe4, 0x55, 0x89, 0x82, 0x71, 0x79, 0x51, 0x8a, 0xe2, 0xeb, 0x4a, 0x45, 0xd3,
	0xd2, 0xe0, 0x37, 0x5b, 0x83, 0xcf, 0xd6, 0x90, 0x53, 0xc3, 0xa0, 0xbb, 0xf2, 0xee, 0x65, 0xf2,
	0xcd, 0x85, 0xc0, 0x18, 0x44, 0x1e, 0x43, 0x94, 0x73, 0xc9, 0xc4, 0x17, 0x1c, 0x31, 0xe3, 0xff,
	0xfa, 0x82, 0x10, 0xf0, 0xcb, 0x42, 0x48, 0x1d, 0x6f, 0x87, 0xea, 0x33, 0xe9, 0x42, 0xbb, 0x62,
	0x62, 0x99, 0x8f, 0xd8, 0xa9, 0x2a, 0x79, 0xba, 0xb4, 0x79, 0xb5, 0x99, 0xa8, 0x7f, 0x3b, 0xd1,
	0x73, 0xf0, 0x25, 0x8e, 0xab, 0xb8, 0xa5, 0x13, 0x38, 0x7a, 0x68, 0x9c, 0xe9, 0x19, 0x8e, 0xab,
	0x77, 0x5c, 0x8a, 0xd5, 0x10, 0x7e, 0x5e, 0xff, 0xf2, 0x5a, 0xdf, 0x1d, 0x37, 0x74, 0xa8, 0x66,
	0x26, 0xaf, 0x20, 0x6a, 0xca, 0xe4, 0x7f, 0xf0, 0xa6, 0x6c, 0x65, 0x16, 0x52, 0x47, 0xb2, 0x07,
	0xad, 0x25, 0xce, 0x16, 0x4c, 0xef, 0x12, 0xd1, 0xfa, 0xe3, 0xc8, 0x7d, 0xed, 0x24, 0x08, 0xa1,
	0x4d, 0xf8, 0x2f, 0xec, 0x78, 0x06, 0x81, 0xd9, 0x5d, 0x5b, 0x11, 0x0d, 0x23, 0x35, 0x99, 0x2f,
	0xdc, 0x89, 0x43, 0x6d, 0x25, 0xf9, 0xe1, 0x40, 0x60, 0x72, 0x6f, 0x3c, 0x70, 0x1e, 0xe8, 0x81,
	0xd1, 0xff, 0x1b, 0x0f, 0x8e, 0x61, 0xf7, 0x9e, 0x9f, 0x0f, 0x39, 0x80, 0x8e, 0x60, 0x59, 0x2e,
	0xd8, 0x48, 0x5e, 0xe8, 0xcd, 0x15, 0xac, 0x33, 0x6c, 0xab, 0x87, 0x77, 0x9e, 0xfb, 0xf1, 0xcd,
	0x8d, 0x47, 0xff, 0xb3, 0x1d, 0x2a, 0xfb, 0x21, 0x9c, 0x87, 0x76, 0x8d, 0xcb, 0x1d, 0xfd, 0xf7,
	0xf3, 0xe2, 0x4f, 0x00, 0x00, 0x00, 0xff, 0xff, 0x79, 0x38, 0x67, 0x18, 0xe9, 0x04, 0x00, 0x00,
}
