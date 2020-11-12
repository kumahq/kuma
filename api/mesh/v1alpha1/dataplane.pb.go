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
	Metrics *MetricsBackend `protobuf:"bytes,2,opt,name=metrics,proto3" json:"metrics,omitempty"`
	// Probes describes list of endpoints which will redirect traffic from
	// insecure port to localhost path
	Probes               *Dataplane_Probes `protobuf:"bytes,3,opt,name=probes,proto3" json:"probes,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
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

func (m *Dataplane) GetProbes() *Dataplane_Probes {
	if m != nil {
		return m.Probes
	}
	return nil
}

// Networking describes inbound and outbound interfaces of a dataplane.
type Dataplane_Networking struct {
	// Ingress if not nil, dataplane will be work in the Ingress mode
	Ingress *Dataplane_Networking_Ingress `protobuf:"bytes,6,opt,name=ingress,proto3" json:"ingress,omitempty"`
	// Public IP on which the dataplane is accessible in the network.
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
	AvailableServices []*Dataplane_Networking_Ingress_AvailableService `protobuf:"bytes,1,rep,name=availableServices,proto3" json:"availableServices,omitempty"`
	// PublicAddress defines IP or DNS name on which Ingress is accessible to
	// other Kuma clusters.
	PublicAddress string `protobuf:"bytes,2,opt,name=publicAddress,proto3" json:"publicAddress,omitempty"`
	// PublicPort defines port on which Ingress is accessible to other Kuma
	// clusters.
	PublicPort           uint32   `protobuf:"varint,3,opt,name=publicPort,proto3" json:"publicPort,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
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

func (m *Dataplane_Networking_Ingress) GetPublicAddress() string {
	if m != nil {
		return m.PublicAddress
	}
	return ""
}

func (m *Dataplane_Networking_Ingress) GetPublicPort() uint32 {
	if m != nil {
		return m.PublicPort
	}
	return 0
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
	// Address of the service that requests will be forwarded to.
	// Empty value defaults to '127.0.0.1', since Kuma DP should be deployed
	// next to service.
	ServiceAddress string `protobuf:"bytes,6,opt,name=serviceAddress,proto3" json:"serviceAddress,omitempty"`
	// Address on which inbound listener will be exposed. Defaults to
	// networking.address.
	Address string `protobuf:"bytes,5,opt,name=address,proto3" json:"address,omitempty"`
	// Tags associated with an application this dataplane is deployed next to,
	// e.g. kuma.io/service=web, version=1.0.
	// `kuma.io/service` tag is mandatory.
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

func (m *Dataplane_Networking_Inbound) GetServiceAddress() string {
	if m != nil {
		return m.ServiceAddress
	}
	return ""
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

type Dataplane_Probes struct {
	Port                 uint32                       `protobuf:"varint,1,opt,name=port,proto3" json:"port,omitempty"`
	Endpoints            []*Dataplane_Probes_Endpoint `protobuf:"bytes,2,rep,name=endpoints,proto3" json:"endpoints,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                     `json:"-"`
	XXX_unrecognized     []byte                       `json:"-"`
	XXX_sizecache        int32                        `json:"-"`
}

func (m *Dataplane_Probes) Reset()         { *m = Dataplane_Probes{} }
func (m *Dataplane_Probes) String() string { return proto.CompactTextString(m) }
func (*Dataplane_Probes) ProtoMessage()    {}
func (*Dataplane_Probes) Descriptor() ([]byte, []int) {
	return fileDescriptor_7608682fd5ea84a4, []int{0, 1}
}

func (m *Dataplane_Probes) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Dataplane_Probes.Unmarshal(m, b)
}
func (m *Dataplane_Probes) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Dataplane_Probes.Marshal(b, m, deterministic)
}
func (m *Dataplane_Probes) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Dataplane_Probes.Merge(m, src)
}
func (m *Dataplane_Probes) XXX_Size() int {
	return xxx_messageInfo_Dataplane_Probes.Size(m)
}
func (m *Dataplane_Probes) XXX_DiscardUnknown() {
	xxx_messageInfo_Dataplane_Probes.DiscardUnknown(m)
}

var xxx_messageInfo_Dataplane_Probes proto.InternalMessageInfo

func (m *Dataplane_Probes) GetPort() uint32 {
	if m != nil {
		return m.Port
	}
	return 0
}

func (m *Dataplane_Probes) GetEndpoints() []*Dataplane_Probes_Endpoint {
	if m != nil {
		return m.Endpoints
	}
	return nil
}

type Dataplane_Probes_Endpoint struct {
	InboundPort          uint32   `protobuf:"varint,1,opt,name=inbound_port,json=inboundPort,proto3" json:"inbound_port,omitempty"`
	InboundPath          string   `protobuf:"bytes,2,opt,name=inbound_path,json=inboundPath,proto3" json:"inbound_path,omitempty"`
	Path                 string   `protobuf:"bytes,3,opt,name=path,proto3" json:"path,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Dataplane_Probes_Endpoint) Reset()         { *m = Dataplane_Probes_Endpoint{} }
func (m *Dataplane_Probes_Endpoint) String() string { return proto.CompactTextString(m) }
func (*Dataplane_Probes_Endpoint) ProtoMessage()    {}
func (*Dataplane_Probes_Endpoint) Descriptor() ([]byte, []int) {
	return fileDescriptor_7608682fd5ea84a4, []int{0, 1, 0}
}

func (m *Dataplane_Probes_Endpoint) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Dataplane_Probes_Endpoint.Unmarshal(m, b)
}
func (m *Dataplane_Probes_Endpoint) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Dataplane_Probes_Endpoint.Marshal(b, m, deterministic)
}
func (m *Dataplane_Probes_Endpoint) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Dataplane_Probes_Endpoint.Merge(m, src)
}
func (m *Dataplane_Probes_Endpoint) XXX_Size() int {
	return xxx_messageInfo_Dataplane_Probes_Endpoint.Size(m)
}
func (m *Dataplane_Probes_Endpoint) XXX_DiscardUnknown() {
	xxx_messageInfo_Dataplane_Probes_Endpoint.DiscardUnknown(m)
}

var xxx_messageInfo_Dataplane_Probes_Endpoint proto.InternalMessageInfo

func (m *Dataplane_Probes_Endpoint) GetInboundPort() uint32 {
	if m != nil {
		return m.InboundPort
	}
	return 0
}

func (m *Dataplane_Probes_Endpoint) GetInboundPath() string {
	if m != nil {
		return m.InboundPath
	}
	return ""
}

func (m *Dataplane_Probes_Endpoint) GetPath() string {
	if m != nil {
		return m.Path
	}
	return ""
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
	proto.RegisterType((*Dataplane_Probes)(nil), "kuma.mesh.v1alpha1.Dataplane.Probes")
	proto.RegisterType((*Dataplane_Probes_Endpoint)(nil), "kuma.mesh.v1alpha1.Dataplane.Probes.Endpoint")
}

func init() { proto.RegisterFile("mesh/v1alpha1/dataplane.proto", fileDescriptor_7608682fd5ea84a4) }

var fileDescriptor_7608682fd5ea84a4 = []byte{
	// 789 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xb4, 0x56, 0x4b, 0x6b, 0xeb, 0x46,
	0x14, 0x66, 0x24, 0xf9, 0xa1, 0x93, 0xba, 0x24, 0x13, 0x37, 0x35, 0xea, 0x03, 0x27, 0x84, 0x62,
	0x0a, 0x55, 0x9a, 0xb6, 0xd0, 0x92, 0xa6, 0x14, 0x8b, 0x86, 0x3e, 0x42, 0x1a, 0x33, 0xcd, 0xa2,
	0x74, 0x63, 0xc6, 0xd6, 0x60, 0x0b, 0x3b, 0x92, 0x90, 0xc6, 0x4e, 0xbd, 0xeb, 0x8f, 0xe8, 0xaa,
	0x74, 0x73, 0x7f, 0xcf, 0x5d, 0xdc, 0xfd, 0x5d, 0xdc, 0xdd, 0xfd, 0x13, 0xd9, 0xe4, 0xa2, 0x79,
	0x48, 0xb2, 0x13, 0x82, 0x9d, 0xcb, 0xdd, 0x8d, 0xce, 0x9c, 0xef, 0x9b, 0x6f, 0xbe, 0x73, 0xe6,
	0xd8, 0xf0, 0xc9, 0x35, 0x4b, 0xc7, 0x47, 0xf3, 0x63, 0x3a, 0x8d, 0xc7, 0xf4, 0xf8, 0xc8, 0xa7,
	0x9c, 0xc6, 0x53, 0x1a, 0x32, 0x37, 0x4e, 0x22, 0x1e, 0x61, 0x3c, 0x99, 0x5d, 0x53, 0x37, 0xcb,
	0x71, 0x75, 0x8e, 0xf3, 0xd1, 0x32, 0xe4, 0x9a, 0xf1, 0x24, 0x18, 0xa6, 0x12, 0xe0, 0x7c, 0x38,
	0xa7, 0xd3, 0xc0, 0xa7, 0x9c, 0x1d, 0xe9, 0x85, 0xdc, 0x38, 0xf8, 0x67, 0x07, 0xec, 0x9f, 0x34,
	0x3b, 0xfe, 0x05, 0x20, 0x64, 0xfc, 0x26, 0x4a, 0x26, 0x41, 0x38, 0x6a, 0xa1, 0x36, 0xea, 0x6c,
	0x7d, 0xd5, 0x71, 0xef, 0x1f, 0xe6, 0xe6, 0x10, 0xf7, 0xf7, 0x3c, 0x9f, 0x94, 0xb0, 0xf8, 0x14,
	0x6a, 0x4a, 0x41, 0xcb, 0x10, 0x34, 0x07, 0x0f, 0xd1, 0x5c, 0xc8, 0x14, 0x8f, 0x0e, 0x27, 0x2c,
	0xf4, 0x89, 0x86, 0xe0, 0x53, 0xa8, 0xc6, 0x49, 0x34, 0x60, 0x69, 0xcb, 0x14, 0xe0, 0xc3, 0xc7,
	0x35, 0xf4, 0x44, 0x2e, 0x51, 0x18, 0xe7, 0x65, 0x03, 0xa0, 0x90, 0x85, 0x7f, 0x83, 0x5a, 0x10,
	0x8e, 0x12, 0x96, 0xa6, 0xad, 0xaa, 0x60, 0xfb, 0x72, 0xdd, 0x1b, 0xb9, 0xbf, 0x4a, 0x1c, 0xd1,
	0x04, 0xb8, 0x05, 0x35, 0xea, 0xfb, 0x82, 0xab, 0xd2, 0x46, 0x1d, 0x9b, 0xe8, 0xcf, 0xec, 0x94,
	0x11, 0xe5, 0xec, 0x86, 0x2e, 0x94, 0xe6, 0xf5, 0x4f, 0xf9, 0x59, 0xe2, 0x88, 0x26, 0x90, 0x8a,
	0x07, 0xd1, 0x2c, 0xf4, 0x5b, 0xa8, 0x6d, 0x6e, 0xa8, 0x58, 0xe0, 0x88, 0x26, 0xc0, 0x17, 0x50,
	0x8f, 0x66, 0x5c, 0x92, 0x19, 0x82, 0xec, 0x78, 0x6d, 0xb2, 0x4b, 0x05, 0x24, 0x39, 0x05, 0x8e,
	0xa0, 0xc9, 0x13, 0x1a, 0xa6, 0x31, 0x4d, 0x58, 0xc8, 0xfb, 0x71, 0x12, 0xfd, 0xbd, 0xc8, 0x7a,
	0xc5, 0x12, 0x77, 0x3e, 0x5d, 0x9b, 0xfa, 0xaa, 0x20, 0xe9, 0x29, 0x0e, 0xb2, 0xcb, 0xef, 0x07,
	0x9d, 0x67, 0x26, 0xd4, 0x54, 0x19, 0x70, 0x04, 0x3b, 0x74, 0x4e, 0x83, 0x29, 0x1d, 0x4c, 0xd9,
	0x1f, 0x2c, 0x99, 0x07, 0x43, 0x96, 0x2a, 0x87, 0xba, 0x9b, 0xd6, 0xd4, 0xed, 0xae, 0x30, 0x91,
	0xfb, 0xdc, 0xf8, 0x10, 0x1a, 0xf1, 0x6c, 0x30, 0x0d, 0x86, 0x5d, 0x55, 0x74, 0x43, 0x14, 0x7d,
	0x39, 0x88, 0x3f, 0x05, 0x90, 0x81, 0x5e, 0x94, 0x70, 0x51, 0xfd, 0x06, 0x29, 0x45, 0x9c, 0x57,
	0x08, 0xb6, 0x57, 0x4f, 0xc3, 0x7d, 0xb0, 0x38, 0x1d, 0x69, 0xf9, 0xe7, 0x6f, 0x2d, 0xdf, 0xbd,
	0xa2, 0xa3, 0xf4, 0x2c, 0xe4, 0xc9, 0x82, 0x08, 0x62, 0xfc, 0x31, 0xd8, 0x41, 0x98, 0x72, 0x1a,
	0x66, 0x26, 0x19, 0x42, 0x54, 0x11, 0xc0, 0x18, 0xac, 0xec, 0x30, 0xa1, 0xd6, 0x26, 0x62, 0xed,
	0x7c, 0x0b, 0x76, 0x4e, 0x82, 0xb7, 0xc1, 0x9c, 0xb0, 0x85, 0x98, 0x01, 0x36, 0xc9, 0x96, 0xb8,
	0x09, 0x95, 0x39, 0x9d, 0xce, 0x98, 0x32, 0x41, 0x7e, 0x9c, 0x18, 0xdf, 0x21, 0xe7, 0x5f, 0x23,
	0xab, 0x91, 0x6c, 0x10, 0x0c, 0x56, 0x5c, 0xd8, 0x20, 0xd6, 0xb8, 0x0d, 0x5b, 0xa9, 0x54, 0x29,
	0x1c, 0xb2, 0xc4, 0x56, 0x39, 0x84, 0x3f, 0x83, 0xf7, 0xd5, 0xa7, 0x76, 0xba, 0x2a, 0x0e, 0x59,
	0x89, 0x3e, 0xf2, 0xfe, 0xfe, 0x54, 0x7e, 0xca, 0x1e, 0x3f, 0xd9, 0xf4, 0xc1, 0x14, 0xf6, 0x79,
	0xf5, 0x5b, 0xaf, 0xf2, 0x1f, 0x32, 0xea, 0x48, 0x1a, 0xf9, 0x74, 0x5b, 0x5e, 0x23, 0xa8, 0xeb,
	0x27, 0x54, 0x56, 0x6e, 0x2e, 0x2b, 0xd7, 0x8e, 0x59, 0x25, 0xc7, 0xf6, 0xa1, 0xa6, 0x6e, 0x2e,
	0x69, 0xbd, 0xda, 0xad, 0x67, 0x25, 0xc6, 0x18, 0x11, 0x1d, 0xc7, 0x97, 0xea, 0xc2, 0x15, 0x71,
	0xe1, 0xef, 0x37, 0x7e, 0xd4, 0xab, 0x0d, 0xf3, 0xf4, 0x7b, 0xfe, 0x8f, 0xa0, 0xa6, 0x66, 0x58,
	0x5e, 0x06, 0xb4, 0x61, 0x19, 0x14, 0xfe, 0x5d, 0x94, 0xe1, 0x39, 0x82, 0xdd, 0x07, 0xc6, 0x0d,
	0xfe, 0x01, 0x3e, 0x48, 0x98, 0x1f, 0x24, 0x6c, 0xc8, 0xfb, 0x99, 0xe9, 0xfd, 0x62, 0xe6, 0xa2,
	0x4e, 0xc3, 0xb3, 0x6f, 0xbd, 0xea, 0xe7, 0x56, 0xeb, 0xee, 0xce, 0x24, 0xbb, 0x3a, 0x2f, 0xeb,
	0x55, 0xdd, 0xe8, 0x3f, 0xc2, 0xde, 0x32, 0xbc, 0x34, 0x66, 0x57, 0xf0, 0xcd, 0x32, 0x3e, 0xef,
	0x88, 0x6f, 0x60, 0x4f, 0xc1, 0xe9, 0x70, 0xc8, 0xd2, 0xb4, 0x9f, 0xea, 0x91, 0x66, 0xb6, 0xcd,
	0x8e, 0x4d, 0x9a, 0x72, 0xb7, 0x2b, 0x36, 0xf5, 0x48, 0x72, 0x5e, 0x20, 0xa8, 0xca, 0xdf, 0xbb,
	0xbc, 0x71, 0x50, 0xa9, 0x71, 0xce, 0xc1, 0x66, 0xa1, 0x1f, 0x47, 0x41, 0xc8, 0xf5, 0x5b, 0xf8,
	0x62, 0x9d, 0x1f, 0x4f, 0xf7, 0x4c, 0xa1, 0x48, 0x81, 0x77, 0x7c, 0xa8, 0xeb, 0x30, 0xde, 0x87,
	0xf7, 0x94, 0x3f, 0xfd, 0xd2, 0xa1, 0x5b, 0x2a, 0xd6, 0x93, 0x4d, 0x5b, 0xa4, 0x50, 0x3e, 0x56,
	0x95, 0xc8, 0x53, 0x28, 0x1f, 0x0b, 0xc9, 0xd9, 0x96, 0x1a, 0x3b, 0xd9, 0xda, 0x83, 0xbf, 0xea,
	0x5a, 0xd6, 0xa0, 0x2a, 0xfe, 0x95, 0x7c, 0xfd, 0x26, 0x00, 0x00, 0xff, 0xff, 0x35, 0x1f, 0x84,
	0x1b, 0x00, 0x09, 0x00, 0x00,
}
