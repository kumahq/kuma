// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.14.0
// source: mesh/v1alpha1/fault_injection.proto

package v1alpha1

import (
	proto "github.com/golang/protobuf/proto"
	duration "github.com/golang/protobuf/ptypes/duration"
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

// FaultInjection defines the configuration of faults between dataplanes.
type FaultInjection struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// List of selectors to match dataplanes that are sources of traffic.
	Sources []*Selector `protobuf:"bytes,1,rep,name=sources,proto3" json:"sources,omitempty"`
	// List of selectors to match services that are destinations of traffic.
	Destinations []*Selector `protobuf:"bytes,2,rep,name=destinations,proto3" json:"destinations,omitempty"`
	// Configuration of FaultInjection
	Conf *FaultInjection_Conf `protobuf:"bytes,3,opt,name=conf,proto3" json:"conf,omitempty"`
}

func (x *FaultInjection) Reset() {
	*x = FaultInjection{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mesh_v1alpha1_fault_injection_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FaultInjection) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FaultInjection) ProtoMessage() {}

func (x *FaultInjection) ProtoReflect() protoreflect.Message {
	mi := &file_mesh_v1alpha1_fault_injection_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FaultInjection.ProtoReflect.Descriptor instead.
func (*FaultInjection) Descriptor() ([]byte, []int) {
	return file_mesh_v1alpha1_fault_injection_proto_rawDescGZIP(), []int{0}
}

func (x *FaultInjection) GetSources() []*Selector {
	if x != nil {
		return x.Sources
	}
	return nil
}

func (x *FaultInjection) GetDestinations() []*Selector {
	if x != nil {
		return x.Destinations
	}
	return nil
}

func (x *FaultInjection) GetConf() *FaultInjection_Conf {
	if x != nil {
		return x.Conf
	}
	return nil
}

// Conf defines several types of faults, at least one fault should be
// specified
type FaultInjection_Conf struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Delay if specified then response from the destination will be delivered
	// with a delay
	Delay *FaultInjection_Conf_Delay `protobuf:"bytes,1,opt,name=delay,proto3" json:"delay,omitempty"`
	// Abort if specified makes source side to receive specified httpStatus code
	Abort *FaultInjection_Conf_Abort `protobuf:"bytes,2,opt,name=abort,proto3" json:"abort,omitempty"`
	// ResponseBandwidth if specified limits the speed of sending response body
	ResponseBandwidth *FaultInjection_Conf_ResponseBandwidth `protobuf:"bytes,3,opt,name=response_bandwidth,json=responseBandwidth,proto3" json:"response_bandwidth,omitempty"`
}

func (x *FaultInjection_Conf) Reset() {
	*x = FaultInjection_Conf{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mesh_v1alpha1_fault_injection_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FaultInjection_Conf) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FaultInjection_Conf) ProtoMessage() {}

func (x *FaultInjection_Conf) ProtoReflect() protoreflect.Message {
	mi := &file_mesh_v1alpha1_fault_injection_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FaultInjection_Conf.ProtoReflect.Descriptor instead.
func (*FaultInjection_Conf) Descriptor() ([]byte, []int) {
	return file_mesh_v1alpha1_fault_injection_proto_rawDescGZIP(), []int{0, 0}
}

func (x *FaultInjection_Conf) GetDelay() *FaultInjection_Conf_Delay {
	if x != nil {
		return x.Delay
	}
	return nil
}

func (x *FaultInjection_Conf) GetAbort() *FaultInjection_Conf_Abort {
	if x != nil {
		return x.Abort
	}
	return nil
}

func (x *FaultInjection_Conf) GetResponseBandwidth() *FaultInjection_Conf_ResponseBandwidth {
	if x != nil {
		return x.ResponseBandwidth
	}
	return nil
}

// Delay defines configuration of delaying a response from a destination
type FaultInjection_Conf_Delay struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Percentage of requests on which delay will be injected, has to be in
	// [0.0 - 100.0] range
	Percentage *wrappers.DoubleValue `protobuf:"bytes,1,opt,name=percentage,proto3" json:"percentage,omitempty"`
	// The duration during which the response will be delayed
	Value *duration.Duration `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *FaultInjection_Conf_Delay) Reset() {
	*x = FaultInjection_Conf_Delay{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mesh_v1alpha1_fault_injection_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FaultInjection_Conf_Delay) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FaultInjection_Conf_Delay) ProtoMessage() {}

func (x *FaultInjection_Conf_Delay) ProtoReflect() protoreflect.Message {
	mi := &file_mesh_v1alpha1_fault_injection_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FaultInjection_Conf_Delay.ProtoReflect.Descriptor instead.
func (*FaultInjection_Conf_Delay) Descriptor() ([]byte, []int) {
	return file_mesh_v1alpha1_fault_injection_proto_rawDescGZIP(), []int{0, 0, 0}
}

func (x *FaultInjection_Conf_Delay) GetPercentage() *wrappers.DoubleValue {
	if x != nil {
		return x.Percentage
	}
	return nil
}

func (x *FaultInjection_Conf_Delay) GetValue() *duration.Duration {
	if x != nil {
		return x.Value
	}
	return nil
}

// Abort defines a configuration of not delivering requests to destination
// service and replacing the responses from destination dataplane by
// predefined status code
type FaultInjection_Conf_Abort struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Percentage of requests on which abort will be injected, has to be in
	// [0.0 - 100.0] range
	Percentage *wrappers.DoubleValue `protobuf:"bytes,1,opt,name=percentage,proto3" json:"percentage,omitempty"`
	// HTTP status code which will be returned to source side
	HttpStatus *wrappers.UInt32Value `protobuf:"bytes,2,opt,name=httpStatus,proto3" json:"httpStatus,omitempty"`
}

func (x *FaultInjection_Conf_Abort) Reset() {
	*x = FaultInjection_Conf_Abort{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mesh_v1alpha1_fault_injection_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FaultInjection_Conf_Abort) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FaultInjection_Conf_Abort) ProtoMessage() {}

func (x *FaultInjection_Conf_Abort) ProtoReflect() protoreflect.Message {
	mi := &file_mesh_v1alpha1_fault_injection_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FaultInjection_Conf_Abort.ProtoReflect.Descriptor instead.
func (*FaultInjection_Conf_Abort) Descriptor() ([]byte, []int) {
	return file_mesh_v1alpha1_fault_injection_proto_rawDescGZIP(), []int{0, 0, 1}
}

func (x *FaultInjection_Conf_Abort) GetPercentage() *wrappers.DoubleValue {
	if x != nil {
		return x.Percentage
	}
	return nil
}

func (x *FaultInjection_Conf_Abort) GetHttpStatus() *wrappers.UInt32Value {
	if x != nil {
		return x.HttpStatus
	}
	return nil
}

// ResponseBandwidth defines a configuration to limit the speed of
// responding to the requests
type FaultInjection_Conf_ResponseBandwidth struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Percentage of requests on which response bandwidth limit will be
	// injected, has to be in [0.0 - 100.0] range
	Percentage *wrappers.DoubleValue `protobuf:"bytes,1,opt,name=percentage,proto3" json:"percentage,omitempty"`
	// Limit is represented by value measure in gbps, mbps, kbps or bps, e.g.
	// 10kbps
	Limit *wrappers.StringValue `protobuf:"bytes,2,opt,name=limit,proto3" json:"limit,omitempty"`
}

func (x *FaultInjection_Conf_ResponseBandwidth) Reset() {
	*x = FaultInjection_Conf_ResponseBandwidth{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mesh_v1alpha1_fault_injection_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FaultInjection_Conf_ResponseBandwidth) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FaultInjection_Conf_ResponseBandwidth) ProtoMessage() {}

func (x *FaultInjection_Conf_ResponseBandwidth) ProtoReflect() protoreflect.Message {
	mi := &file_mesh_v1alpha1_fault_injection_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FaultInjection_Conf_ResponseBandwidth.ProtoReflect.Descriptor instead.
func (*FaultInjection_Conf_ResponseBandwidth) Descriptor() ([]byte, []int) {
	return file_mesh_v1alpha1_fault_injection_proto_rawDescGZIP(), []int{0, 0, 2}
}

func (x *FaultInjection_Conf_ResponseBandwidth) GetPercentage() *wrappers.DoubleValue {
	if x != nil {
		return x.Percentage
	}
	return nil
}

func (x *FaultInjection_Conf_ResponseBandwidth) GetLimit() *wrappers.StringValue {
	if x != nil {
		return x.Limit
	}
	return nil
}

var File_mesh_v1alpha1_fault_injection_proto protoreflect.FileDescriptor

var file_mesh_v1alpha1_fault_injection_proto_rawDesc = []byte{
	0x0a, 0x23, 0x6d, 0x65, 0x73, 0x68, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f,
	0x66, 0x61, 0x75, 0x6c, 0x74, 0x5f, 0x69, 0x6e, 0x6a, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x12, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x6d, 0x65, 0x73, 0x68,
	0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x1a, 0x1e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x75, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x6d, 0x65, 0x73, 0x68, 0x2f,
	0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x73, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f,
	0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x77, 0x72, 0x61, 0x70, 0x70, 0x65, 0x72,
	0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xca, 0x06, 0x0a, 0x0e, 0x46, 0x61, 0x75, 0x6c,
	0x74, 0x49, 0x6e, 0x6a, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x36, 0x0a, 0x07, 0x73, 0x6f,
	0x75, 0x72, 0x63, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x6b, 0x75,
	0x6d, 0x61, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31,
	0x2e, 0x53, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x52, 0x07, 0x73, 0x6f, 0x75, 0x72, 0x63,
	0x65, 0x73, 0x12, 0x40, 0x0a, 0x0c, 0x64, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e,
	0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x53, 0x65,
	0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x52, 0x0c, 0x64, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x12, 0x3b, 0x0a, 0x04, 0x63, 0x6f, 0x6e, 0x66, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x27, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76,
	0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x46, 0x61, 0x75, 0x6c, 0x74, 0x49, 0x6e, 0x6a,
	0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x52, 0x04, 0x63, 0x6f, 0x6e,
	0x66, 0x1a, 0x80, 0x05, 0x0a, 0x04, 0x43, 0x6f, 0x6e, 0x66, 0x12, 0x43, 0x0a, 0x05, 0x64, 0x65,
	0x6c, 0x61, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2d, 0x2e, 0x6b, 0x75, 0x6d, 0x61,
	0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x46,
	0x61, 0x75, 0x6c, 0x74, 0x49, 0x6e, 0x6a, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x43, 0x6f,
	0x6e, 0x66, 0x2e, 0x44, 0x65, 0x6c, 0x61, 0x79, 0x52, 0x05, 0x64, 0x65, 0x6c, 0x61, 0x79, 0x12,
	0x43, 0x0a, 0x05, 0x61, 0x62, 0x6f, 0x72, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2d,
	0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70,
	0x68, 0x61, 0x31, 0x2e, 0x46, 0x61, 0x75, 0x6c, 0x74, 0x49, 0x6e, 0x6a, 0x65, 0x63, 0x74, 0x69,
	0x6f, 0x6e, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x2e, 0x41, 0x62, 0x6f, 0x72, 0x74, 0x52, 0x05, 0x61,
	0x62, 0x6f, 0x72, 0x74, 0x12, 0x68, 0x0a, 0x12, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x5f, 0x62, 0x61, 0x6e, 0x64, 0x77, 0x69, 0x64, 0x74, 0x68, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x39, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x46, 0x61, 0x75, 0x6c, 0x74, 0x49, 0x6e, 0x6a, 0x65, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x2e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x42, 0x61, 0x6e, 0x64, 0x77, 0x69, 0x64, 0x74, 0x68, 0x52, 0x11, 0x72, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x61, 0x6e, 0x64, 0x77, 0x69, 0x64, 0x74, 0x68, 0x1a, 0x76,
	0x0a, 0x05, 0x44, 0x65, 0x6c, 0x61, 0x79, 0x12, 0x3c, 0x0a, 0x0a, 0x70, 0x65, 0x72, 0x63, 0x65,
	0x6e, 0x74, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x44, 0x6f,
	0x75, 0x62, 0x6c, 0x65, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x0a, 0x70, 0x65, 0x72, 0x63, 0x65,
	0x6e, 0x74, 0x61, 0x67, 0x65, 0x12, 0x2f, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52,
	0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x1a, 0x83, 0x01, 0x0a, 0x05, 0x41, 0x62, 0x6f, 0x72, 0x74,
	0x12, 0x3c, 0x0a, 0x0a, 0x70, 0x65, 0x72, 0x63, 0x65, 0x6e, 0x74, 0x61, 0x67, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x44, 0x6f, 0x75, 0x62, 0x6c, 0x65, 0x56, 0x61, 0x6c,
	0x75, 0x65, 0x52, 0x0a, 0x70, 0x65, 0x72, 0x63, 0x65, 0x6e, 0x74, 0x61, 0x67, 0x65, 0x12, 0x3c,
	0x0a, 0x0a, 0x68, 0x74, 0x74, 0x70, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2e, 0x55, 0x49, 0x6e, 0x74, 0x33, 0x32, 0x56, 0x61, 0x6c, 0x75, 0x65,
	0x52, 0x0a, 0x68, 0x74, 0x74, 0x70, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x1a, 0x85, 0x01, 0x0a,
	0x11, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x61, 0x6e, 0x64, 0x77, 0x69, 0x64,
	0x74, 0x68, 0x12, 0x3c, 0x0a, 0x0a, 0x70, 0x65, 0x72, 0x63, 0x65, 0x6e, 0x74, 0x61, 0x67, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x44, 0x6f, 0x75, 0x62, 0x6c, 0x65, 0x56,
	0x61, 0x6c, 0x75, 0x65, 0x52, 0x0a, 0x70, 0x65, 0x72, 0x63, 0x65, 0x6e, 0x74, 0x61, 0x67, 0x65,
	0x12, 0x32, 0x0a, 0x05, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x1c, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2e, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x05, 0x6c,
	0x69, 0x6d, 0x69, 0x74, 0x42, 0x2a, 0x5a, 0x28, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x6b, 0x75, 0x6d, 0x61, 0x68, 0x71, 0x2f, 0x6b, 0x75, 0x6d, 0x61, 0x2f, 0x61,
	0x70, 0x69, 0x2f, 0x6d, 0x65, 0x73, 0x68, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_mesh_v1alpha1_fault_injection_proto_rawDescOnce sync.Once
	file_mesh_v1alpha1_fault_injection_proto_rawDescData = file_mesh_v1alpha1_fault_injection_proto_rawDesc
)

func file_mesh_v1alpha1_fault_injection_proto_rawDescGZIP() []byte {
	file_mesh_v1alpha1_fault_injection_proto_rawDescOnce.Do(func() {
		file_mesh_v1alpha1_fault_injection_proto_rawDescData = protoimpl.X.CompressGZIP(file_mesh_v1alpha1_fault_injection_proto_rawDescData)
	})
	return file_mesh_v1alpha1_fault_injection_proto_rawDescData
}

var file_mesh_v1alpha1_fault_injection_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_mesh_v1alpha1_fault_injection_proto_goTypes = []interface{}{
	(*FaultInjection)(nil),                        // 0: kuma.mesh.v1alpha1.FaultInjection
	(*FaultInjection_Conf)(nil),                   // 1: kuma.mesh.v1alpha1.FaultInjection.Conf
	(*FaultInjection_Conf_Delay)(nil),             // 2: kuma.mesh.v1alpha1.FaultInjection.Conf.Delay
	(*FaultInjection_Conf_Abort)(nil),             // 3: kuma.mesh.v1alpha1.FaultInjection.Conf.Abort
	(*FaultInjection_Conf_ResponseBandwidth)(nil), // 4: kuma.mesh.v1alpha1.FaultInjection.Conf.ResponseBandwidth
	(*Selector)(nil),                              // 5: kuma.mesh.v1alpha1.Selector
	(*wrappers.DoubleValue)(nil),                  // 6: google.protobuf.DoubleValue
	(*duration.Duration)(nil),                     // 7: google.protobuf.Duration
	(*wrappers.UInt32Value)(nil),                  // 8: google.protobuf.UInt32Value
	(*wrappers.StringValue)(nil),                  // 9: google.protobuf.StringValue
}
var file_mesh_v1alpha1_fault_injection_proto_depIdxs = []int32{
	5,  // 0: kuma.mesh.v1alpha1.FaultInjection.sources:type_name -> kuma.mesh.v1alpha1.Selector
	5,  // 1: kuma.mesh.v1alpha1.FaultInjection.destinations:type_name -> kuma.mesh.v1alpha1.Selector
	1,  // 2: kuma.mesh.v1alpha1.FaultInjection.conf:type_name -> kuma.mesh.v1alpha1.FaultInjection.Conf
	2,  // 3: kuma.mesh.v1alpha1.FaultInjection.Conf.delay:type_name -> kuma.mesh.v1alpha1.FaultInjection.Conf.Delay
	3,  // 4: kuma.mesh.v1alpha1.FaultInjection.Conf.abort:type_name -> kuma.mesh.v1alpha1.FaultInjection.Conf.Abort
	4,  // 5: kuma.mesh.v1alpha1.FaultInjection.Conf.response_bandwidth:type_name -> kuma.mesh.v1alpha1.FaultInjection.Conf.ResponseBandwidth
	6,  // 6: kuma.mesh.v1alpha1.FaultInjection.Conf.Delay.percentage:type_name -> google.protobuf.DoubleValue
	7,  // 7: kuma.mesh.v1alpha1.FaultInjection.Conf.Delay.value:type_name -> google.protobuf.Duration
	6,  // 8: kuma.mesh.v1alpha1.FaultInjection.Conf.Abort.percentage:type_name -> google.protobuf.DoubleValue
	8,  // 9: kuma.mesh.v1alpha1.FaultInjection.Conf.Abort.httpStatus:type_name -> google.protobuf.UInt32Value
	6,  // 10: kuma.mesh.v1alpha1.FaultInjection.Conf.ResponseBandwidth.percentage:type_name -> google.protobuf.DoubleValue
	9,  // 11: kuma.mesh.v1alpha1.FaultInjection.Conf.ResponseBandwidth.limit:type_name -> google.protobuf.StringValue
	12, // [12:12] is the sub-list for method output_type
	12, // [12:12] is the sub-list for method input_type
	12, // [12:12] is the sub-list for extension type_name
	12, // [12:12] is the sub-list for extension extendee
	0,  // [0:12] is the sub-list for field type_name
}

func init() { file_mesh_v1alpha1_fault_injection_proto_init() }
func file_mesh_v1alpha1_fault_injection_proto_init() {
	if File_mesh_v1alpha1_fault_injection_proto != nil {
		return
	}
	file_mesh_v1alpha1_selector_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_mesh_v1alpha1_fault_injection_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FaultInjection); i {
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
		file_mesh_v1alpha1_fault_injection_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FaultInjection_Conf); i {
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
		file_mesh_v1alpha1_fault_injection_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FaultInjection_Conf_Delay); i {
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
		file_mesh_v1alpha1_fault_injection_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FaultInjection_Conf_Abort); i {
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
		file_mesh_v1alpha1_fault_injection_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FaultInjection_Conf_ResponseBandwidth); i {
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
			RawDescriptor: file_mesh_v1alpha1_fault_injection_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_mesh_v1alpha1_fault_injection_proto_goTypes,
		DependencyIndexes: file_mesh_v1alpha1_fault_injection_proto_depIdxs,
		MessageInfos:      file_mesh_v1alpha1_fault_injection_proto_msgTypes,
	}.Build()
	File_mesh_v1alpha1_fault_injection_proto = out.File
	file_mesh_v1alpha1_fault_injection_proto_rawDesc = nil
	file_mesh_v1alpha1_fault_injection_proto_goTypes = nil
	file_mesh_v1alpha1_fault_injection_proto_depIdxs = nil
}
