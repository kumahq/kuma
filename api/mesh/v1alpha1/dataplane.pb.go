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
	Tags                 map[string]string                    `protobuf:"bytes,2,rep,name=tags,proto3" json:"tags,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Health               *Dataplane_Networking_Inbound_Health `protobuf:"bytes,7,opt,name=health,proto3" json:"health,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                             `json:"-"`
	XXX_unrecognized     []byte                               `json:"-"`
	XXX_sizecache        int32                                `json:"-"`
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

func (m *Dataplane_Networking_Inbound) GetHealth() *Dataplane_Networking_Inbound_Health {
	if m != nil {
		return m.Health
	}
	return nil
}

type Dataplane_Networking_Inbound_Health struct {
	Ready                bool     `protobuf:"varint,1,opt,name=ready,proto3" json:"ready,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Dataplane_Networking_Inbound_Health) Reset()         { *m = Dataplane_Networking_Inbound_Health{} }
func (m *Dataplane_Networking_Inbound_Health) String() string { return proto.CompactTextString(m) }
func (*Dataplane_Networking_Inbound_Health) ProtoMessage()    {}
func (*Dataplane_Networking_Inbound_Health) Descriptor() ([]byte, []int) {
	return fileDescriptor_7608682fd5ea84a4, []int{0, 0, 1, 1}
}

func (m *Dataplane_Networking_Inbound_Health) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Dataplane_Networking_Inbound_Health.Unmarshal(m, b)
}
func (m *Dataplane_Networking_Inbound_Health) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Dataplane_Networking_Inbound_Health.Marshal(b, m, deterministic)
}
func (m *Dataplane_Networking_Inbound_Health) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Dataplane_Networking_Inbound_Health.Merge(m, src)
}
func (m *Dataplane_Networking_Inbound_Health) XXX_Size() int {
	return xxx_messageInfo_Dataplane_Networking_Inbound_Health.Size(m)
}
func (m *Dataplane_Networking_Inbound_Health) XXX_DiscardUnknown() {
	xxx_messageInfo_Dataplane_Networking_Inbound_Health.DiscardUnknown(m)
}

var xxx_messageInfo_Dataplane_Networking_Inbound_Health proto.InternalMessageInfo

func (m *Dataplane_Networking_Inbound_Health) GetReady() bool {
	if m != nil {
		return m.Ready
	}
	return false
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
	proto.RegisterType((*Dataplane_Networking_Inbound_Health)(nil), "kuma.mesh.v1alpha1.Dataplane.Networking.Inbound.Health")
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
	// 844 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xa4, 0x56, 0xdd, 0x8e, 0xdb, 0x44,
	0x14, 0xd6, 0xc4, 0x5e, 0x3b, 0x3e, 0xcb, 0x56, 0xed, 0xec, 0x52, 0x2c, 0x03, 0x55, 0x5a, 0x55,
	0x28, 0xaa, 0x84, 0xc3, 0x02, 0x52, 0x51, 0x59, 0x84, 0xd6, 0xa2, 0xa2, 0x50, 0x95, 0x8d, 0x86,
	0x5e, 0x20, 0x6e, 0xa2, 0x89, 0x3d, 0x8a, 0xad, 0x38, 0xb6, 0xb1, 0x27, 0x29, 0x79, 0x00, 0x5e,
	0x02, 0x71, 0xc3, 0xf3, 0xf4, 0x82, 0x37, 0xe0, 0x8e, 0x07, 0xe0, 0x76, 0x6f, 0x8a, 0x3c, 0x3f,
	0xb6, 0x93, 0x2d, 0x55, 0xb2, 0xbd, 0xf2, 0xcc, 0x99, 0xf3, 0x7d, 0xe7, 0xf8, 0x3b, 0x67, 0x8e,
	0x0d, 0x1f, 0x2e, 0x58, 0x15, 0x8f, 0x56, 0xa7, 0x34, 0x2d, 0x62, 0x7a, 0x3a, 0x8a, 0x28, 0xa7,
	0x45, 0x4a, 0x33, 0xe6, 0x17, 0x65, 0xce, 0x73, 0x8c, 0xe7, 0xcb, 0x05, 0xf5, 0x6b, 0x1f, 0x5f,
	0xfb, 0x78, 0xef, 0x6f, 0x42, 0x16, 0x8c, 0x97, 0x49, 0x58, 0x49, 0x80, 0xf7, 0xde, 0x8a, 0xa6,
	0x49, 0x44, 0x39, 0x1b, 0xe9, 0x85, 0x3c, 0xb8, 0xf7, 0xef, 0x2d, 0x70, 0xbe, 0xd1, 0xec, 0xf8,
	0x09, 0x40, 0xc6, 0xf8, 0x8b, 0xbc, 0x9c, 0x27, 0xd9, 0xcc, 0x45, 0x03, 0x34, 0x3c, 0xfc, 0x74,
	0xe8, 0x5f, 0x0d, 0xe6, 0x37, 0x10, 0xff, 0x87, 0xc6, 0x9f, 0x74, 0xb0, 0xf8, 0x0c, 0x6c, 0x95,
	0x81, 0xdb, 0x13, 0x34, 0xf7, 0x5e, 0x47, 0xf3, 0x4c, 0xba, 0x04, 0x34, 0x9c, 0xb3, 0x2c, 0x22,
	0x1a, 0x82, 0xcf, 0xc0, 0x2a, 0xca, 0x7c, 0xca, 0x2a, 0xd7, 0x10, 0xe0, 0xfb, 0x6f, 0xce, 0x61,
	0x2c, 0x7c, 0x89, 0xc2, 0x78, 0x2f, 0x6f, 0x00, 0xb4, 0x69, 0xe1, 0xef, 0xc1, 0x4e, 0xb2, 0x59,
	0xc9, 0xaa, 0xca, 0xb5, 0x04, 0xdb, 0x27, 0xbb, 0xbe, 0x91, 0xff, 0x9d, 0xc4, 0x11, 0x4d, 0x80,
	0x5d, 0xb0, 0x69, 0x14, 0x09, 0xae, 0x83, 0x01, 0x1a, 0x3a, 0x44, 0x6f, 0xeb, 0x28, 0x33, 0xca,
	0xd9, 0x0b, 0xba, 0x56, 0x39, 0xef, 0x1e, 0xe5, 0x5b, 0x89, 0x23, 0x9a, 0x40, 0x66, 0x3c, 0xcd,
	0x97, 0x59, 0xe4, 0xa2, 0x81, 0xb1, 0x67, 0xc6, 0x02, 0x47, 0x34, 0x01, 0x7e, 0x06, 0xfd, 0x7c,
	0xc9, 0x25, 0x59, 0x4f, 0x90, 0x9d, 0xee, 0x4c, 0x76, 0xa1, 0x80, 0xa4, 0xa1, 0xc0, 0x39, 0x9c,
	0xf0, 0x92, 0x66, 0x55, 0x41, 0x4b, 0x96, 0xf1, 0x49, 0x51, 0xe6, 0xbf, 0xae, 0xeb, 0x5e, 0x31,
	0xc5, 0x3b, 0x9f, 0xed, 0x4c, 0xfd, 0xbc, 0x25, 0x19, 0x2b, 0x0e, 0x72, 0xcc, 0xaf, 0x1a, 0xbd,
	0x3f, 0x0d, 0xb0, 0x55, 0x19, 0x70, 0x0e, 0xb7, 0xe8, 0x8a, 0x26, 0x29, 0x9d, 0xa6, 0xec, 0x47,
	0x56, 0xae, 0x92, 0x90, 0x55, 0x4a, 0xa1, 0xf3, 0x7d, 0x6b, 0xea, 0x9f, 0x6f, 0x31, 0x91, 0xab,
	0xdc, 0xf8, 0x3e, 0x1c, 0x15, 0xcb, 0x69, 0x9a, 0x84, 0xe7, 0xaa, 0xe8, 0x3d, 0x51, 0xf4, 0x4d,
	0x23, 0xbe, 0x03, 0x20, 0x0d, 0xe3, 0xbc, 0xe4, 0xa2, 0xfa, 0x47, 0xa4, 0x63, 0xf1, 0xfe, 0x46,
	0x70, 0x73, 0x3b, 0x1a, 0x9e, 0x80, 0xc9, 0xe9, 0x4c, 0xa7, 0xff, 0xf4, 0xad, 0xd3, 0xf7, 0x9f,
	0xd3, 0x59, 0xf5, 0x38, 0xe3, 0xe5, 0x9a, 0x08, 0x62, 0xfc, 0x01, 0x38, 0x49, 0x56, 0x71, 0x9a,
	0xd5, 0x22, 0xf5, 0x44, 0x52, 0xad, 0x01, 0x63, 0x30, 0xeb, 0x60, 0x22, 0x5b, 0x87, 0x88, 0xb5,
	0xf7, 0x10, 0x9c, 0x86, 0x04, 0xdf, 0x04, 0x63, 0xce, 0xd6, 0x62, 0x06, 0x38, 0xa4, 0x5e, 0xe2,
	0x13, 0x38, 0x58, 0xd1, 0x74, 0xc9, 0x94, 0x08, 0x72, 0xf3, 0xa8, 0xf7, 0x05, 0xf2, 0x7e, 0x13,
	0x35, 0x92, 0x0d, 0x82, 0xc1, 0x2c, 0x5a, 0x19, 0xc4, 0x1a, 0x0f, 0xe0, 0xb0, 0x92, 0x59, 0x0a,
	0x85, 0x4c, 0x71, 0xd4, 0x35, 0xe1, 0x8f, 0xe0, 0x86, 0xda, 0x6a, 0xa5, 0x2d, 0x11, 0x64, 0xcb,
	0xfa, 0x86, 0xfb, 0xf7, 0x93, 0xd2, 0x53, 0xf6, 0xf8, 0xa3, 0x7d, 0x2f, 0x4c, 0x2b, 0x5f, 0xd0,
	0xbf, 0x0c, 0x0e, 0x7e, 0x47, 0xbd, 0x3e, 0x52, 0x42, 0x5e, 0x80, 0x15, 0x33, 0x9a, 0xf2, 0xd8,
	0xb5, 0x45, 0x93, 0x3f, 0xdc, 0x9b, 0xfb, 0x89, 0x80, 0x13, 0x45, 0x73, 0x7d, 0x9d, 0xef, 0x80,
	0x25, 0xa9, 0x6a, 0x9f, 0x92, 0xd1, 0x48, 0xe2, 0xfa, 0x44, 0x6e, 0xbc, 0x7f, 0x10, 0xf4, 0xf5,
	0x9d, 0xed, 0x4a, 0x65, 0x6c, 0x4a, 0xa5, 0x4b, 0x64, 0x76, 0x4a, 0x74, 0x17, 0x6c, 0x25, 0xb5,
	0x0c, 0x1b, 0xd8, 0x97, 0x81, 0x59, 0xf6, 0x62, 0x44, 0xb4, 0x1d, 0x5f, 0x28, 0x85, 0x0f, 0x84,
	0xc2, 0x5f, 0xee, 0x3d, 0x45, 0xb6, 0x3b, 0xf4, 0xfa, 0x3a, 0xfc, 0x81, 0xc0, 0x56, 0x43, 0xb3,
	0xa9, 0x3b, 0xda, 0xb3, 0xee, 0x0a, 0xff, 0xff, 0x75, 0xbf, 0x7e, 0x7a, 0x2f, 0x11, 0x1c, 0xbf,
	0x66, 0xbe, 0xe1, 0xaf, 0xe0, 0xdd, 0x92, 0x45, 0x49, 0xc9, 0x42, 0x3e, 0xa9, 0x45, 0x9f, 0xb4,
	0x43, 0x1e, 0x0d, 0x8f, 0x02, 0xe7, 0x32, 0xb0, 0x1e, 0x98, 0xee, 0xab, 0x57, 0x06, 0x39, 0xd6,
	0x7e, 0xf5, 0xe5, 0xd0, 0x37, 0xeb, 0x6b, 0xb8, 0xbd, 0x09, 0xef, 0xcc, 0xf5, 0x2d, 0xfc, 0x49,
	0x17, 0xdf, 0x74, 0xc4, 0xe7, 0x70, 0x5b, 0xc1, 0x69, 0x18, 0xb2, 0xaa, 0x9a, 0x54, 0x7a, 0x86,
	0x1a, 0x03, 0x63, 0xe8, 0x90, 0x13, 0x79, 0x7a, 0x2e, 0x0e, 0xf5, 0x0c, 0xf4, 0xfe, 0x42, 0x60,
	0xc9, 0x0f, 0x6c, 0xd3, 0x38, 0xa8, 0xd3, 0x38, 0x4f, 0xc1, 0x61, 0x59, 0x54, 0xe4, 0x49, 0xc6,
	0xf5, 0xe5, 0xfb, 0x78, 0x97, 0xaf, 0xb5, 0xff, 0x58, 0xa1, 0x48, 0x8b, 0xf7, 0x22, 0xe8, 0x6b,
	0x33, 0xbe, 0x0b, 0xef, 0x28, 0x7d, 0x26, 0x9d, 0xa0, 0x87, 0xca, 0x36, 0x96, 0x4d, 0xdb, 0xba,
	0x50, 0x1e, 0xab, 0x4a, 0x34, 0x2e, 0x94, 0xc7, 0x22, 0xe5, 0xfa, 0x48, 0xcd, 0xb9, 0x7a, 0x1d,
	0x3c, 0xf8, 0x79, 0x38, 0x4b, 0x78, 0xbc, 0x9c, 0xfa, 0x61, 0xbe, 0x18, 0xd5, 0xb9, 0xc6, 0xbf,
	0x88, 0xc7, 0x88, 0x16, 0xc9, 0x68, 0xe3, 0x37, 0x6a, 0x6a, 0x89, 0xdf, 0xa4, 0xcf, 0xfe, 0x0b,
	0x00, 0x00, 0xff, 0xff, 0x44, 0x09, 0xee, 0xdd, 0x91, 0x09, 0x00, 0x00,
}
