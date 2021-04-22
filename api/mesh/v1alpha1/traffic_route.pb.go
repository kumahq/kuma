// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.23.0
// 	protoc        v3.14.0
// source: mesh/v1alpha1/traffic_route.proto

package v1alpha1

import (
	_ "github.com/envoyproxy/protoc-gen-validate/validate"
	proto "github.com/golang/protobuf/proto"
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

// TrafficRoute defines routing rules for L4 traffic.
type TrafficRoute struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// List of selectors to match dataplanes that are sources of traffic.
	Sources []*Selector `protobuf:"bytes,1,rep,name=sources,proto3" json:"sources,omitempty"`
	// List of selectors to match services that are destinations of traffic.
	//
	// Notice the difference between sources and destinations.
	// While the source of traffic is always a dataplane within a mesh,
	// the destination is a service that could be either within or outside
	// of a mesh.
	Destinations []*Selector `protobuf:"bytes,2,rep,name=destinations,proto3" json:"destinations,omitempty"`
	// Configuration for the route
	Conf *TrafficRoute_Conf `protobuf:"bytes,3,opt,name=conf,proto3" json:"conf,omitempty"`
}

func (x *TrafficRoute) Reset() {
	*x = TrafficRoute{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mesh_v1alpha1_traffic_route_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TrafficRoute) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TrafficRoute) ProtoMessage() {}

func (x *TrafficRoute) ProtoReflect() protoreflect.Message {
	mi := &file_mesh_v1alpha1_traffic_route_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TrafficRoute.ProtoReflect.Descriptor instead.
func (*TrafficRoute) Descriptor() ([]byte, []int) {
	return file_mesh_v1alpha1_traffic_route_proto_rawDescGZIP(), []int{0}
}

func (x *TrafficRoute) GetSources() []*Selector {
	if x != nil {
		return x.Sources
	}
	return nil
}

func (x *TrafficRoute) GetDestinations() []*Selector {
	if x != nil {
		return x.Destinations
	}
	return nil
}

func (x *TrafficRoute) GetConf() *TrafficRoute_Conf {
	if x != nil {
		return x.Conf
	}
	return nil
}

// Split defines a destination with a weight assigned to it.
type TrafficRoute_Split struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Weight assigned to that destination.
	Weight uint32 `protobuf:"varint,1,opt,name=weight,proto3" json:"weight,omitempty"`
	// Selector to match individual endpoints that comprise that destination.
	//
	// Notice that an endpoint can be either inside or outside the mesh.
	// In the former case an endpoint corresponds to a dataplane,
	// in the latter case an endpoint is a black box.
	Destination map[string]string `protobuf:"bytes,2,rep,name=destination,proto3" json:"destination,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *TrafficRoute_Split) Reset() {
	*x = TrafficRoute_Split{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mesh_v1alpha1_traffic_route_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TrafficRoute_Split) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TrafficRoute_Split) ProtoMessage() {}

func (x *TrafficRoute_Split) ProtoReflect() protoreflect.Message {
	mi := &file_mesh_v1alpha1_traffic_route_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TrafficRoute_Split.ProtoReflect.Descriptor instead.
func (*TrafficRoute_Split) Descriptor() ([]byte, []int) {
	return file_mesh_v1alpha1_traffic_route_proto_rawDescGZIP(), []int{0, 0}
}

func (x *TrafficRoute_Split) GetWeight() uint32 {
	if x != nil {
		return x.Weight
	}
	return 0
}

func (x *TrafficRoute_Split) GetDestination() map[string]string {
	if x != nil {
		return x.Destination
	}
	return nil
}

// LoadBalancer defines the load balancing policy and configuration
type TrafficRoute_LoadBalancer struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to LbType:
	//	*TrafficRoute_LoadBalancer_RoundRobin_
	//	*TrafficRoute_LoadBalancer_LeastRequest_
	//	*TrafficRoute_LoadBalancer_RingHash_
	//	*TrafficRoute_LoadBalancer_Random_
	//	*TrafficRoute_LoadBalancer_Maglev_
	LbType isTrafficRoute_LoadBalancer_LbType `protobuf_oneof:"lb_type"`
}

func (x *TrafficRoute_LoadBalancer) Reset() {
	*x = TrafficRoute_LoadBalancer{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mesh_v1alpha1_traffic_route_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TrafficRoute_LoadBalancer) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TrafficRoute_LoadBalancer) ProtoMessage() {}

func (x *TrafficRoute_LoadBalancer) ProtoReflect() protoreflect.Message {
	mi := &file_mesh_v1alpha1_traffic_route_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TrafficRoute_LoadBalancer.ProtoReflect.Descriptor instead.
func (*TrafficRoute_LoadBalancer) Descriptor() ([]byte, []int) {
	return file_mesh_v1alpha1_traffic_route_proto_rawDescGZIP(), []int{0, 1}
}

func (m *TrafficRoute_LoadBalancer) GetLbType() isTrafficRoute_LoadBalancer_LbType {
	if m != nil {
		return m.LbType
	}
	return nil
}

func (x *TrafficRoute_LoadBalancer) GetRoundRobin() *TrafficRoute_LoadBalancer_RoundRobin {
	if x, ok := x.GetLbType().(*TrafficRoute_LoadBalancer_RoundRobin_); ok {
		return x.RoundRobin
	}
	return nil
}

func (x *TrafficRoute_LoadBalancer) GetLeastRequest() *TrafficRoute_LoadBalancer_LeastRequest {
	if x, ok := x.GetLbType().(*TrafficRoute_LoadBalancer_LeastRequest_); ok {
		return x.LeastRequest
	}
	return nil
}

func (x *TrafficRoute_LoadBalancer) GetRingHash() *TrafficRoute_LoadBalancer_RingHash {
	if x, ok := x.GetLbType().(*TrafficRoute_LoadBalancer_RingHash_); ok {
		return x.RingHash
	}
	return nil
}

func (x *TrafficRoute_LoadBalancer) GetRandom() *TrafficRoute_LoadBalancer_Random {
	if x, ok := x.GetLbType().(*TrafficRoute_LoadBalancer_Random_); ok {
		return x.Random
	}
	return nil
}

func (x *TrafficRoute_LoadBalancer) GetMaglev() *TrafficRoute_LoadBalancer_Maglev {
	if x, ok := x.GetLbType().(*TrafficRoute_LoadBalancer_Maglev_); ok {
		return x.Maglev
	}
	return nil
}

type isTrafficRoute_LoadBalancer_LbType interface {
	isTrafficRoute_LoadBalancer_LbType()
}

type TrafficRoute_LoadBalancer_RoundRobin_ struct {
	RoundRobin *TrafficRoute_LoadBalancer_RoundRobin `protobuf:"bytes,1,opt,name=round_robin,json=roundRobin,proto3,oneof"`
}

type TrafficRoute_LoadBalancer_LeastRequest_ struct {
	LeastRequest *TrafficRoute_LoadBalancer_LeastRequest `protobuf:"bytes,2,opt,name=least_request,json=leastRequest,proto3,oneof"`
}

type TrafficRoute_LoadBalancer_RingHash_ struct {
	RingHash *TrafficRoute_LoadBalancer_RingHash `protobuf:"bytes,3,opt,name=ring_hash,json=ringHash,proto3,oneof"`
}

type TrafficRoute_LoadBalancer_Random_ struct {
	Random *TrafficRoute_LoadBalancer_Random `protobuf:"bytes,4,opt,name=random,proto3,oneof"`
}

type TrafficRoute_LoadBalancer_Maglev_ struct {
	Maglev *TrafficRoute_LoadBalancer_Maglev `protobuf:"bytes,5,opt,name=maglev,proto3,oneof"`
}

func (*TrafficRoute_LoadBalancer_RoundRobin_) isTrafficRoute_LoadBalancer_LbType() {}

func (*TrafficRoute_LoadBalancer_LeastRequest_) isTrafficRoute_LoadBalancer_LbType() {}

func (*TrafficRoute_LoadBalancer_RingHash_) isTrafficRoute_LoadBalancer_LbType() {}

func (*TrafficRoute_LoadBalancer_Random_) isTrafficRoute_LoadBalancer_LbType() {}

func (*TrafficRoute_LoadBalancer_Maglev_) isTrafficRoute_LoadBalancer_LbType() {}

// Conf defines the destination configuration
type TrafficRoute_Conf struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// List of destinations with weights assigned to them.
	Split        []*TrafficRoute_Split      `protobuf:"bytes,1,rep,name=split,proto3" json:"split,omitempty"`
	LoadBalancer *TrafficRoute_LoadBalancer `protobuf:"bytes,2,opt,name=load_balancer,json=loadBalancer,proto3" json:"load_balancer,omitempty"`
}

func (x *TrafficRoute_Conf) Reset() {
	*x = TrafficRoute_Conf{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mesh_v1alpha1_traffic_route_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TrafficRoute_Conf) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TrafficRoute_Conf) ProtoMessage() {}

func (x *TrafficRoute_Conf) ProtoReflect() protoreflect.Message {
	mi := &file_mesh_v1alpha1_traffic_route_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TrafficRoute_Conf.ProtoReflect.Descriptor instead.
func (*TrafficRoute_Conf) Descriptor() ([]byte, []int) {
	return file_mesh_v1alpha1_traffic_route_proto_rawDescGZIP(), []int{0, 2}
}

func (x *TrafficRoute_Conf) GetSplit() []*TrafficRoute_Split {
	if x != nil {
		return x.Split
	}
	return nil
}

func (x *TrafficRoute_Conf) GetLoadBalancer() *TrafficRoute_LoadBalancer {
	if x != nil {
		return x.LoadBalancer
	}
	return nil
}

// RoundRobin is a simple policy in which each available upstream host is
// selected in round robin order
type TrafficRoute_LoadBalancer_RoundRobin struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *TrafficRoute_LoadBalancer_RoundRobin) Reset() {
	*x = TrafficRoute_LoadBalancer_RoundRobin{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mesh_v1alpha1_traffic_route_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TrafficRoute_LoadBalancer_RoundRobin) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TrafficRoute_LoadBalancer_RoundRobin) ProtoMessage() {}

func (x *TrafficRoute_LoadBalancer_RoundRobin) ProtoReflect() protoreflect.Message {
	mi := &file_mesh_v1alpha1_traffic_route_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TrafficRoute_LoadBalancer_RoundRobin.ProtoReflect.Descriptor instead.
func (*TrafficRoute_LoadBalancer_RoundRobin) Descriptor() ([]byte, []int) {
	return file_mesh_v1alpha1_traffic_route_proto_rawDescGZIP(), []int{0, 1, 0}
}

// LeastRequest uses different algorithms depending on whether hosts have
// the same or different weights
type TrafficRoute_LoadBalancer_LeastRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The number of random healthy hosts from which the host with the fewest
	// active requests will be chosen. Defaults to 2 so that we perform
	// two-choice selection if the field is not set.
	ChoiceCount uint32 `protobuf:"varint,1,opt,name=choice_count,json=choiceCount,proto3" json:"choice_count,omitempty"`
}

func (x *TrafficRoute_LoadBalancer_LeastRequest) Reset() {
	*x = TrafficRoute_LoadBalancer_LeastRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mesh_v1alpha1_traffic_route_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TrafficRoute_LoadBalancer_LeastRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TrafficRoute_LoadBalancer_LeastRequest) ProtoMessage() {}

func (x *TrafficRoute_LoadBalancer_LeastRequest) ProtoReflect() protoreflect.Message {
	mi := &file_mesh_v1alpha1_traffic_route_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TrafficRoute_LoadBalancer_LeastRequest.ProtoReflect.Descriptor instead.
func (*TrafficRoute_LoadBalancer_LeastRequest) Descriptor() ([]byte, []int) {
	return file_mesh_v1alpha1_traffic_route_proto_rawDescGZIP(), []int{0, 1, 1}
}

func (x *TrafficRoute_LoadBalancer_LeastRequest) GetChoiceCount() uint32 {
	if x != nil {
		return x.ChoiceCount
	}
	return 0
}

// RingHash implements consistent hashing to upstream hosts
type TrafficRoute_LoadBalancer_RingHash struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The hash function used to hash hosts onto the ketama ring. The value
	// defaults to 'XX_HASH'
	HashFunction string `protobuf:"bytes,1,opt,name=hash_function,json=hashFunction,proto3" json:"hash_function,omitempty"`
	// Minimum hash ring size
	MinRingSize uint64 `protobuf:"varint,2,opt,name=min_ring_size,json=minRingSize,proto3" json:"min_ring_size,omitempty"`
	// Maximum hash ring size.
	MaxRingSize uint64 `protobuf:"varint,3,opt,name=max_ring_size,json=maxRingSize,proto3" json:"max_ring_size,omitempty"`
}

func (x *TrafficRoute_LoadBalancer_RingHash) Reset() {
	*x = TrafficRoute_LoadBalancer_RingHash{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mesh_v1alpha1_traffic_route_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TrafficRoute_LoadBalancer_RingHash) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TrafficRoute_LoadBalancer_RingHash) ProtoMessage() {}

func (x *TrafficRoute_LoadBalancer_RingHash) ProtoReflect() protoreflect.Message {
	mi := &file_mesh_v1alpha1_traffic_route_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TrafficRoute_LoadBalancer_RingHash.ProtoReflect.Descriptor instead.
func (*TrafficRoute_LoadBalancer_RingHash) Descriptor() ([]byte, []int) {
	return file_mesh_v1alpha1_traffic_route_proto_rawDescGZIP(), []int{0, 1, 2}
}

func (x *TrafficRoute_LoadBalancer_RingHash) GetHashFunction() string {
	if x != nil {
		return x.HashFunction
	}
	return ""
}

func (x *TrafficRoute_LoadBalancer_RingHash) GetMinRingSize() uint64 {
	if x != nil {
		return x.MinRingSize
	}
	return 0
}

func (x *TrafficRoute_LoadBalancer_RingHash) GetMaxRingSize() uint64 {
	if x != nil {
		return x.MaxRingSize
	}
	return 0
}

// Random selects a random available host
type TrafficRoute_LoadBalancer_Random struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *TrafficRoute_LoadBalancer_Random) Reset() {
	*x = TrafficRoute_LoadBalancer_Random{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mesh_v1alpha1_traffic_route_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TrafficRoute_LoadBalancer_Random) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TrafficRoute_LoadBalancer_Random) ProtoMessage() {}

func (x *TrafficRoute_LoadBalancer_Random) ProtoReflect() protoreflect.Message {
	mi := &file_mesh_v1alpha1_traffic_route_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TrafficRoute_LoadBalancer_Random.ProtoReflect.Descriptor instead.
func (*TrafficRoute_LoadBalancer_Random) Descriptor() ([]byte, []int) {
	return file_mesh_v1alpha1_traffic_route_proto_rawDescGZIP(), []int{0, 1, 3}
}

// Maglev implements consistent hashing to upstream hosts
type TrafficRoute_LoadBalancer_Maglev struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *TrafficRoute_LoadBalancer_Maglev) Reset() {
	*x = TrafficRoute_LoadBalancer_Maglev{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mesh_v1alpha1_traffic_route_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TrafficRoute_LoadBalancer_Maglev) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TrafficRoute_LoadBalancer_Maglev) ProtoMessage() {}

func (x *TrafficRoute_LoadBalancer_Maglev) ProtoReflect() protoreflect.Message {
	mi := &file_mesh_v1alpha1_traffic_route_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TrafficRoute_LoadBalancer_Maglev.ProtoReflect.Descriptor instead.
func (*TrafficRoute_LoadBalancer_Maglev) Descriptor() ([]byte, []int) {
	return file_mesh_v1alpha1_traffic_route_proto_rawDescGZIP(), []int{0, 1, 4}
}

var File_mesh_v1alpha1_traffic_route_proto protoreflect.FileDescriptor

var file_mesh_v1alpha1_traffic_route_proto_rawDesc = []byte{
	0x0a, 0x21, 0x6d, 0x65, 0x73, 0x68, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f,
	0x74, 0x72, 0x61, 0x66, 0x66, 0x69, 0x63, 0x5f, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x12, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76,
	0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x1a, 0x1c, 0x6d, 0x65, 0x73, 0x68, 0x2f, 0x76, 0x31,
	0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x73, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x17, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f,
	0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x83,
	0x0a, 0x0a, 0x0c, 0x54, 0x72, 0x61, 0x66, 0x66, 0x69, 0x63, 0x52, 0x6f, 0x75, 0x74, 0x65, 0x12,
	0x40, 0x0a, 0x07, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x1c, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x53, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x42, 0x08,
	0xfa, 0x42, 0x05, 0x92, 0x01, 0x02, 0x08, 0x01, 0x52, 0x07, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65,
	0x73, 0x12, 0x4a, 0x0a, 0x0c, 0x64, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x6d,
	0x65, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x53, 0x65, 0x6c,
	0x65, 0x63, 0x74, 0x6f, 0x72, 0x42, 0x08, 0xfa, 0x42, 0x05, 0x92, 0x01, 0x02, 0x08, 0x01, 0x52,
	0x0c, 0x64, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x43, 0x0a,
	0x04, 0x63, 0x6f, 0x6e, 0x66, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x25, 0x2e, 0x6b, 0x75,
	0x6d, 0x61, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31,
	0x2e, 0x54, 0x72, 0x61, 0x66, 0x66, 0x69, 0x63, 0x52, 0x6f, 0x75, 0x74, 0x65, 0x2e, 0x43, 0x6f,
	0x6e, 0x66, 0x42, 0x08, 0xfa, 0x42, 0x05, 0x8a, 0x01, 0x02, 0x10, 0x01, 0x52, 0x04, 0x63, 0x6f,
	0x6e, 0x66, 0x1a, 0xd9, 0x01, 0x0a, 0x05, 0x53, 0x70, 0x6c, 0x69, 0x74, 0x12, 0x1f, 0x0a, 0x06,
	0x77, 0x65, 0x69, 0x67, 0x68, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x42, 0x07, 0xfa, 0x42,
	0x04, 0x2a, 0x02, 0x28, 0x00, 0x52, 0x06, 0x77, 0x65, 0x69, 0x67, 0x68, 0x74, 0x12, 0x6f, 0x0a,
	0x0b, 0x64, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x37, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76,
	0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x54, 0x72, 0x61, 0x66, 0x66, 0x69, 0x63, 0x52,
	0x6f, 0x75, 0x74, 0x65, 0x2e, 0x53, 0x70, 0x6c, 0x69, 0x74, 0x2e, 0x44, 0x65, 0x73, 0x74, 0x69,
	0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x42, 0x14, 0xfa, 0x42, 0x11,
	0x9a, 0x01, 0x0e, 0x08, 0x01, 0x22, 0x04, 0x72, 0x02, 0x10, 0x01, 0x2a, 0x04, 0x72, 0x02, 0x10,
	0x01, 0x52, 0x0b, 0x64, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x1a, 0x3e,
	0x0a, 0x10, 0x44, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x45, 0x6e, 0x74,
	0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x1a, 0x9e,
	0x05, 0x0a, 0x0c, 0x4c, 0x6f, 0x61, 0x64, 0x42, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65, 0x72, 0x12,
	0x5b, 0x0a, 0x0b, 0x72, 0x6f, 0x75, 0x6e, 0x64, 0x5f, 0x72, 0x6f, 0x62, 0x69, 0x6e, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x38, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x6d, 0x65, 0x73, 0x68,
	0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x54, 0x72, 0x61, 0x66, 0x66, 0x69,
	0x63, 0x52, 0x6f, 0x75, 0x74, 0x65, 0x2e, 0x4c, 0x6f, 0x61, 0x64, 0x42, 0x61, 0x6c, 0x61, 0x6e,
	0x63, 0x65, 0x72, 0x2e, 0x52, 0x6f, 0x75, 0x6e, 0x64, 0x52, 0x6f, 0x62, 0x69, 0x6e, 0x48, 0x00,
	0x52, 0x0a, 0x72, 0x6f, 0x75, 0x6e, 0x64, 0x52, 0x6f, 0x62, 0x69, 0x6e, 0x12, 0x61, 0x0a, 0x0d,
	0x6c, 0x65, 0x61, 0x73, 0x74, 0x5f, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x3a, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e,
	0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x54, 0x72, 0x61, 0x66, 0x66, 0x69, 0x63,
	0x52, 0x6f, 0x75, 0x74, 0x65, 0x2e, 0x4c, 0x6f, 0x61, 0x64, 0x42, 0x61, 0x6c, 0x61, 0x6e, 0x63,
	0x65, 0x72, 0x2e, 0x4c, 0x65, 0x61, 0x73, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x48,
	0x00, 0x52, 0x0c, 0x6c, 0x65, 0x61, 0x73, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x55, 0x0a, 0x09, 0x72, 0x69, 0x6e, 0x67, 0x5f, 0x68, 0x61, 0x73, 0x68, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x36, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76,
	0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x54, 0x72, 0x61, 0x66, 0x66, 0x69, 0x63, 0x52,
	0x6f, 0x75, 0x74, 0x65, 0x2e, 0x4c, 0x6f, 0x61, 0x64, 0x42, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65,
	0x72, 0x2e, 0x52, 0x69, 0x6e, 0x67, 0x48, 0x61, 0x73, 0x68, 0x48, 0x00, 0x52, 0x08, 0x72, 0x69,
	0x6e, 0x67, 0x48, 0x61, 0x73, 0x68, 0x12, 0x4e, 0x0a, 0x06, 0x72, 0x61, 0x6e, 0x64, 0x6f, 0x6d,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x34, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x6d, 0x65,
	0x73, 0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x54, 0x72, 0x61, 0x66,
	0x66, 0x69, 0x63, 0x52, 0x6f, 0x75, 0x74, 0x65, 0x2e, 0x4c, 0x6f, 0x61, 0x64, 0x42, 0x61, 0x6c,
	0x61, 0x6e, 0x63, 0x65, 0x72, 0x2e, 0x52, 0x61, 0x6e, 0x64, 0x6f, 0x6d, 0x48, 0x00, 0x52, 0x06,
	0x72, 0x61, 0x6e, 0x64, 0x6f, 0x6d, 0x12, 0x4e, 0x0a, 0x06, 0x6d, 0x61, 0x67, 0x6c, 0x65, 0x76,
	0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x34, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x6d, 0x65,
	0x73, 0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x54, 0x72, 0x61, 0x66,
	0x66, 0x69, 0x63, 0x52, 0x6f, 0x75, 0x74, 0x65, 0x2e, 0x4c, 0x6f, 0x61, 0x64, 0x42, 0x61, 0x6c,
	0x61, 0x6e, 0x63, 0x65, 0x72, 0x2e, 0x4d, 0x61, 0x67, 0x6c, 0x65, 0x76, 0x48, 0x00, 0x52, 0x06,
	0x6d, 0x61, 0x67, 0x6c, 0x65, 0x76, 0x1a, 0x0c, 0x0a, 0x0a, 0x52, 0x6f, 0x75, 0x6e, 0x64, 0x52,
	0x6f, 0x62, 0x69, 0x6e, 0x1a, 0x31, 0x0a, 0x0c, 0x4c, 0x65, 0x61, 0x73, 0x74, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x21, 0x0a, 0x0c, 0x63, 0x68, 0x6f, 0x69, 0x63, 0x65, 0x5f, 0x63,
	0x6f, 0x75, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x0b, 0x63, 0x68, 0x6f, 0x69,
	0x63, 0x65, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x1a, 0x77, 0x0a, 0x08, 0x52, 0x69, 0x6e, 0x67, 0x48,
	0x61, 0x73, 0x68, 0x12, 0x23, 0x0a, 0x0d, 0x68, 0x61, 0x73, 0x68, 0x5f, 0x66, 0x75, 0x6e, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x68, 0x61, 0x73, 0x68,
	0x46, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x22, 0x0a, 0x0d, 0x6d, 0x69, 0x6e, 0x5f,
	0x72, 0x69, 0x6e, 0x67, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52,
	0x0b, 0x6d, 0x69, 0x6e, 0x52, 0x69, 0x6e, 0x67, 0x53, 0x69, 0x7a, 0x65, 0x12, 0x22, 0x0a, 0x0d,
	0x6d, 0x61, 0x78, 0x5f, 0x72, 0x69, 0x6e, 0x67, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x04, 0x52, 0x0b, 0x6d, 0x61, 0x78, 0x52, 0x69, 0x6e, 0x67, 0x53, 0x69, 0x7a, 0x65,
	0x1a, 0x08, 0x0a, 0x06, 0x52, 0x61, 0x6e, 0x64, 0x6f, 0x6d, 0x1a, 0x08, 0x0a, 0x06, 0x4d, 0x61,
	0x67, 0x6c, 0x65, 0x76, 0x42, 0x09, 0x0a, 0x07, 0x6c, 0x62, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x1a,
	0xa2, 0x01, 0x0a, 0x04, 0x43, 0x6f, 0x6e, 0x66, 0x12, 0x46, 0x0a, 0x05, 0x73, 0x70, 0x6c, 0x69,
	0x74, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x26, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x6d,
	0x65, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x54, 0x72, 0x61,
	0x66, 0x66, 0x69, 0x63, 0x52, 0x6f, 0x75, 0x74, 0x65, 0x2e, 0x53, 0x70, 0x6c, 0x69, 0x74, 0x42,
	0x08, 0xfa, 0x42, 0x05, 0x92, 0x01, 0x02, 0x08, 0x01, 0x52, 0x05, 0x73, 0x70, 0x6c, 0x69, 0x74,
	0x12, 0x52, 0x0a, 0x0d, 0x6c, 0x6f, 0x61, 0x64, 0x5f, 0x62, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65,
	0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2d, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x6d,
	0x65, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x54, 0x72, 0x61,
	0x66, 0x66, 0x69, 0x63, 0x52, 0x6f, 0x75, 0x74, 0x65, 0x2e, 0x4c, 0x6f, 0x61, 0x64, 0x42, 0x61,
	0x6c, 0x61, 0x6e, 0x63, 0x65, 0x72, 0x52, 0x0c, 0x6c, 0x6f, 0x61, 0x64, 0x42, 0x61, 0x6c, 0x61,
	0x6e, 0x63, 0x65, 0x72, 0x42, 0x2a, 0x5a, 0x28, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x6b, 0x75, 0x6d, 0x61, 0x68, 0x71, 0x2f, 0x6b, 0x75, 0x6d, 0x61, 0x2f, 0x61,
	0x70, 0x69, 0x2f, 0x6d, 0x65, 0x73, 0x68, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_mesh_v1alpha1_traffic_route_proto_rawDescOnce sync.Once
	file_mesh_v1alpha1_traffic_route_proto_rawDescData = file_mesh_v1alpha1_traffic_route_proto_rawDesc
)

func file_mesh_v1alpha1_traffic_route_proto_rawDescGZIP() []byte {
	file_mesh_v1alpha1_traffic_route_proto_rawDescOnce.Do(func() {
		file_mesh_v1alpha1_traffic_route_proto_rawDescData = protoimpl.X.CompressGZIP(file_mesh_v1alpha1_traffic_route_proto_rawDescData)
	})
	return file_mesh_v1alpha1_traffic_route_proto_rawDescData
}

var file_mesh_v1alpha1_traffic_route_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_mesh_v1alpha1_traffic_route_proto_goTypes = []interface{}{
	(*TrafficRoute)(nil),              // 0: kuma.mesh.v1alpha1.TrafficRoute
	(*TrafficRoute_Split)(nil),        // 1: kuma.mesh.v1alpha1.TrafficRoute.Split
	(*TrafficRoute_LoadBalancer)(nil), // 2: kuma.mesh.v1alpha1.TrafficRoute.LoadBalancer
	(*TrafficRoute_Conf)(nil),         // 3: kuma.mesh.v1alpha1.TrafficRoute.Conf
	nil,                               // 4: kuma.mesh.v1alpha1.TrafficRoute.Split.DestinationEntry
	(*TrafficRoute_LoadBalancer_RoundRobin)(nil),   // 5: kuma.mesh.v1alpha1.TrafficRoute.LoadBalancer.RoundRobin
	(*TrafficRoute_LoadBalancer_LeastRequest)(nil), // 6: kuma.mesh.v1alpha1.TrafficRoute.LoadBalancer.LeastRequest
	(*TrafficRoute_LoadBalancer_RingHash)(nil),     // 7: kuma.mesh.v1alpha1.TrafficRoute.LoadBalancer.RingHash
	(*TrafficRoute_LoadBalancer_Random)(nil),       // 8: kuma.mesh.v1alpha1.TrafficRoute.LoadBalancer.Random
	(*TrafficRoute_LoadBalancer_Maglev)(nil),       // 9: kuma.mesh.v1alpha1.TrafficRoute.LoadBalancer.Maglev
	(*Selector)(nil),                               // 10: kuma.mesh.v1alpha1.Selector
}
var file_mesh_v1alpha1_traffic_route_proto_depIdxs = []int32{
	10, // 0: kuma.mesh.v1alpha1.TrafficRoute.sources:type_name -> kuma.mesh.v1alpha1.Selector
	10, // 1: kuma.mesh.v1alpha1.TrafficRoute.destinations:type_name -> kuma.mesh.v1alpha1.Selector
	3,  // 2: kuma.mesh.v1alpha1.TrafficRoute.conf:type_name -> kuma.mesh.v1alpha1.TrafficRoute.Conf
	4,  // 3: kuma.mesh.v1alpha1.TrafficRoute.Split.destination:type_name -> kuma.mesh.v1alpha1.TrafficRoute.Split.DestinationEntry
	5,  // 4: kuma.mesh.v1alpha1.TrafficRoute.LoadBalancer.round_robin:type_name -> kuma.mesh.v1alpha1.TrafficRoute.LoadBalancer.RoundRobin
	6,  // 5: kuma.mesh.v1alpha1.TrafficRoute.LoadBalancer.least_request:type_name -> kuma.mesh.v1alpha1.TrafficRoute.LoadBalancer.LeastRequest
	7,  // 6: kuma.mesh.v1alpha1.TrafficRoute.LoadBalancer.ring_hash:type_name -> kuma.mesh.v1alpha1.TrafficRoute.LoadBalancer.RingHash
	8,  // 7: kuma.mesh.v1alpha1.TrafficRoute.LoadBalancer.random:type_name -> kuma.mesh.v1alpha1.TrafficRoute.LoadBalancer.Random
	9,  // 8: kuma.mesh.v1alpha1.TrafficRoute.LoadBalancer.maglev:type_name -> kuma.mesh.v1alpha1.TrafficRoute.LoadBalancer.Maglev
	1,  // 9: kuma.mesh.v1alpha1.TrafficRoute.Conf.split:type_name -> kuma.mesh.v1alpha1.TrafficRoute.Split
	2,  // 10: kuma.mesh.v1alpha1.TrafficRoute.Conf.load_balancer:type_name -> kuma.mesh.v1alpha1.TrafficRoute.LoadBalancer
	11, // [11:11] is the sub-list for method output_type
	11, // [11:11] is the sub-list for method input_type
	11, // [11:11] is the sub-list for extension type_name
	11, // [11:11] is the sub-list for extension extendee
	0,  // [0:11] is the sub-list for field type_name
}

func init() { file_mesh_v1alpha1_traffic_route_proto_init() }
func file_mesh_v1alpha1_traffic_route_proto_init() {
	if File_mesh_v1alpha1_traffic_route_proto != nil {
		return
	}
	file_mesh_v1alpha1_selector_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_mesh_v1alpha1_traffic_route_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TrafficRoute); i {
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
		file_mesh_v1alpha1_traffic_route_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TrafficRoute_Split); i {
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
		file_mesh_v1alpha1_traffic_route_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TrafficRoute_LoadBalancer); i {
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
		file_mesh_v1alpha1_traffic_route_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TrafficRoute_Conf); i {
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
		file_mesh_v1alpha1_traffic_route_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TrafficRoute_LoadBalancer_RoundRobin); i {
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
		file_mesh_v1alpha1_traffic_route_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TrafficRoute_LoadBalancer_LeastRequest); i {
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
		file_mesh_v1alpha1_traffic_route_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TrafficRoute_LoadBalancer_RingHash); i {
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
		file_mesh_v1alpha1_traffic_route_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TrafficRoute_LoadBalancer_Random); i {
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
		file_mesh_v1alpha1_traffic_route_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TrafficRoute_LoadBalancer_Maglev); i {
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
	file_mesh_v1alpha1_traffic_route_proto_msgTypes[2].OneofWrappers = []interface{}{
		(*TrafficRoute_LoadBalancer_RoundRobin_)(nil),
		(*TrafficRoute_LoadBalancer_LeastRequest_)(nil),
		(*TrafficRoute_LoadBalancer_RingHash_)(nil),
		(*TrafficRoute_LoadBalancer_Random_)(nil),
		(*TrafficRoute_LoadBalancer_Maglev_)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_mesh_v1alpha1_traffic_route_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_mesh_v1alpha1_traffic_route_proto_goTypes,
		DependencyIndexes: file_mesh_v1alpha1_traffic_route_proto_depIdxs,
		MessageInfos:      file_mesh_v1alpha1_traffic_route_proto_msgTypes,
	}.Build()
	File_mesh_v1alpha1_traffic_route_proto = out.File
	file_mesh_v1alpha1_traffic_route_proto_rawDesc = nil
	file_mesh_v1alpha1_traffic_route_proto_goTypes = nil
	file_mesh_v1alpha1_traffic_route_proto_depIdxs = nil
}
