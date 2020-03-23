// Code generated by protoc-gen-go. DO NOT EDIT.
// source: mesh/v1alpha1/fault_injection.proto

package v1alpha1

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	duration "github.com/golang/protobuf/ptypes/duration"
	wrappers "github.com/golang/protobuf/ptypes/wrappers"
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

type FaultInjection struct {
	Sources              []*Selector          `protobuf:"bytes,1,rep,name=sources,proto3" json:"sources,omitempty"`
	Destinations         []*Selector          `protobuf:"bytes,2,rep,name=destinations,proto3" json:"destinations,omitempty"`
	Conf                 *FaultInjection_Conf `protobuf:"bytes,3,opt,name=conf,proto3" json:"conf,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *FaultInjection) Reset()         { *m = FaultInjection{} }
func (m *FaultInjection) String() string { return proto.CompactTextString(m) }
func (*FaultInjection) ProtoMessage()    {}
func (*FaultInjection) Descriptor() ([]byte, []int) {
	return fileDescriptor_ff4d722195e1e7eb, []int{0}
}

func (m *FaultInjection) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FaultInjection.Unmarshal(m, b)
}
func (m *FaultInjection) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FaultInjection.Marshal(b, m, deterministic)
}
func (m *FaultInjection) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FaultInjection.Merge(m, src)
}
func (m *FaultInjection) XXX_Size() int {
	return xxx_messageInfo_FaultInjection.Size(m)
}
func (m *FaultInjection) XXX_DiscardUnknown() {
	xxx_messageInfo_FaultInjection.DiscardUnknown(m)
}

var xxx_messageInfo_FaultInjection proto.InternalMessageInfo

func (m *FaultInjection) GetSources() []*Selector {
	if m != nil {
		return m.Sources
	}
	return nil
}

func (m *FaultInjection) GetDestinations() []*Selector {
	if m != nil {
		return m.Destinations
	}
	return nil
}

func (m *FaultInjection) GetConf() *FaultInjection_Conf {
	if m != nil {
		return m.Conf
	}
	return nil
}

type FaultInjection_Conf struct {
	Delay                *FaultInjection_Conf_Delay             `protobuf:"bytes,1,opt,name=delay,proto3" json:"delay,omitempty"`
	Abort                *FaultInjection_Conf_Abort             `protobuf:"bytes,2,opt,name=abort,proto3" json:"abort,omitempty"`
	ResponseBandwidth    *FaultInjection_Conf_ResponseBandwidth `protobuf:"bytes,3,opt,name=response_bandwidth,json=responseBandwidth,proto3" json:"response_bandwidth,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                               `json:"-"`
	XXX_unrecognized     []byte                                 `json:"-"`
	XXX_sizecache        int32                                  `json:"-"`
}

func (m *FaultInjection_Conf) Reset()         { *m = FaultInjection_Conf{} }
func (m *FaultInjection_Conf) String() string { return proto.CompactTextString(m) }
func (*FaultInjection_Conf) ProtoMessage()    {}
func (*FaultInjection_Conf) Descriptor() ([]byte, []int) {
	return fileDescriptor_ff4d722195e1e7eb, []int{0, 0}
}

func (m *FaultInjection_Conf) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FaultInjection_Conf.Unmarshal(m, b)
}
func (m *FaultInjection_Conf) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FaultInjection_Conf.Marshal(b, m, deterministic)
}
func (m *FaultInjection_Conf) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FaultInjection_Conf.Merge(m, src)
}
func (m *FaultInjection_Conf) XXX_Size() int {
	return xxx_messageInfo_FaultInjection_Conf.Size(m)
}
func (m *FaultInjection_Conf) XXX_DiscardUnknown() {
	xxx_messageInfo_FaultInjection_Conf.DiscardUnknown(m)
}

var xxx_messageInfo_FaultInjection_Conf proto.InternalMessageInfo

func (m *FaultInjection_Conf) GetDelay() *FaultInjection_Conf_Delay {
	if m != nil {
		return m.Delay
	}
	return nil
}

func (m *FaultInjection_Conf) GetAbort() *FaultInjection_Conf_Abort {
	if m != nil {
		return m.Abort
	}
	return nil
}

func (m *FaultInjection_Conf) GetResponseBandwidth() *FaultInjection_Conf_ResponseBandwidth {
	if m != nil {
		return m.ResponseBandwidth
	}
	return nil
}

type FaultInjection_Conf_Delay struct {
	Percentage           *wrappers.DoubleValue `protobuf:"bytes,1,opt,name=percentage,proto3" json:"percentage,omitempty"`
	Value                *duration.Duration    `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	XXX_NoUnkeyedLiteral struct{}              `json:"-"`
	XXX_unrecognized     []byte                `json:"-"`
	XXX_sizecache        int32                 `json:"-"`
}

func (m *FaultInjection_Conf_Delay) Reset()         { *m = FaultInjection_Conf_Delay{} }
func (m *FaultInjection_Conf_Delay) String() string { return proto.CompactTextString(m) }
func (*FaultInjection_Conf_Delay) ProtoMessage()    {}
func (*FaultInjection_Conf_Delay) Descriptor() ([]byte, []int) {
	return fileDescriptor_ff4d722195e1e7eb, []int{0, 0, 0}
}

func (m *FaultInjection_Conf_Delay) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FaultInjection_Conf_Delay.Unmarshal(m, b)
}
func (m *FaultInjection_Conf_Delay) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FaultInjection_Conf_Delay.Marshal(b, m, deterministic)
}
func (m *FaultInjection_Conf_Delay) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FaultInjection_Conf_Delay.Merge(m, src)
}
func (m *FaultInjection_Conf_Delay) XXX_Size() int {
	return xxx_messageInfo_FaultInjection_Conf_Delay.Size(m)
}
func (m *FaultInjection_Conf_Delay) XXX_DiscardUnknown() {
	xxx_messageInfo_FaultInjection_Conf_Delay.DiscardUnknown(m)
}

var xxx_messageInfo_FaultInjection_Conf_Delay proto.InternalMessageInfo

func (m *FaultInjection_Conf_Delay) GetPercentage() *wrappers.DoubleValue {
	if m != nil {
		return m.Percentage
	}
	return nil
}

func (m *FaultInjection_Conf_Delay) GetValue() *duration.Duration {
	if m != nil {
		return m.Value
	}
	return nil
}

type FaultInjection_Conf_Abort struct {
	Percentage           *wrappers.DoubleValue `protobuf:"bytes,1,opt,name=percentage,proto3" json:"percentage,omitempty"`
	HttpStatus           *wrappers.UInt32Value `protobuf:"bytes,2,opt,name=httpStatus,proto3" json:"httpStatus,omitempty"`
	XXX_NoUnkeyedLiteral struct{}              `json:"-"`
	XXX_unrecognized     []byte                `json:"-"`
	XXX_sizecache        int32                 `json:"-"`
}

func (m *FaultInjection_Conf_Abort) Reset()         { *m = FaultInjection_Conf_Abort{} }
func (m *FaultInjection_Conf_Abort) String() string { return proto.CompactTextString(m) }
func (*FaultInjection_Conf_Abort) ProtoMessage()    {}
func (*FaultInjection_Conf_Abort) Descriptor() ([]byte, []int) {
	return fileDescriptor_ff4d722195e1e7eb, []int{0, 0, 1}
}

func (m *FaultInjection_Conf_Abort) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FaultInjection_Conf_Abort.Unmarshal(m, b)
}
func (m *FaultInjection_Conf_Abort) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FaultInjection_Conf_Abort.Marshal(b, m, deterministic)
}
func (m *FaultInjection_Conf_Abort) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FaultInjection_Conf_Abort.Merge(m, src)
}
func (m *FaultInjection_Conf_Abort) XXX_Size() int {
	return xxx_messageInfo_FaultInjection_Conf_Abort.Size(m)
}
func (m *FaultInjection_Conf_Abort) XXX_DiscardUnknown() {
	xxx_messageInfo_FaultInjection_Conf_Abort.DiscardUnknown(m)
}

var xxx_messageInfo_FaultInjection_Conf_Abort proto.InternalMessageInfo

func (m *FaultInjection_Conf_Abort) GetPercentage() *wrappers.DoubleValue {
	if m != nil {
		return m.Percentage
	}
	return nil
}

func (m *FaultInjection_Conf_Abort) GetHttpStatus() *wrappers.UInt32Value {
	if m != nil {
		return m.HttpStatus
	}
	return nil
}

type FaultInjection_Conf_ResponseBandwidth struct {
	Percentage           *wrappers.DoubleValue `protobuf:"bytes,1,opt,name=percentage,proto3" json:"percentage,omitempty"`
	Limit                *wrappers.StringValue `protobuf:"bytes,2,opt,name=limit,proto3" json:"limit,omitempty"`
	XXX_NoUnkeyedLiteral struct{}              `json:"-"`
	XXX_unrecognized     []byte                `json:"-"`
	XXX_sizecache        int32                 `json:"-"`
}

func (m *FaultInjection_Conf_ResponseBandwidth) Reset()         { *m = FaultInjection_Conf_ResponseBandwidth{} }
func (m *FaultInjection_Conf_ResponseBandwidth) String() string { return proto.CompactTextString(m) }
func (*FaultInjection_Conf_ResponseBandwidth) ProtoMessage()    {}
func (*FaultInjection_Conf_ResponseBandwidth) Descriptor() ([]byte, []int) {
	return fileDescriptor_ff4d722195e1e7eb, []int{0, 0, 2}
}

func (m *FaultInjection_Conf_ResponseBandwidth) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FaultInjection_Conf_ResponseBandwidth.Unmarshal(m, b)
}
func (m *FaultInjection_Conf_ResponseBandwidth) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FaultInjection_Conf_ResponseBandwidth.Marshal(b, m, deterministic)
}
func (m *FaultInjection_Conf_ResponseBandwidth) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FaultInjection_Conf_ResponseBandwidth.Merge(m, src)
}
func (m *FaultInjection_Conf_ResponseBandwidth) XXX_Size() int {
	return xxx_messageInfo_FaultInjection_Conf_ResponseBandwidth.Size(m)
}
func (m *FaultInjection_Conf_ResponseBandwidth) XXX_DiscardUnknown() {
	xxx_messageInfo_FaultInjection_Conf_ResponseBandwidth.DiscardUnknown(m)
}

var xxx_messageInfo_FaultInjection_Conf_ResponseBandwidth proto.InternalMessageInfo

func (m *FaultInjection_Conf_ResponseBandwidth) GetPercentage() *wrappers.DoubleValue {
	if m != nil {
		return m.Percentage
	}
	return nil
}

func (m *FaultInjection_Conf_ResponseBandwidth) GetLimit() *wrappers.StringValue {
	if m != nil {
		return m.Limit
	}
	return nil
}

func init() {
	proto.RegisterType((*FaultInjection)(nil), "kuma.mesh.v1alpha1.FaultInjection")
	proto.RegisterType((*FaultInjection_Conf)(nil), "kuma.mesh.v1alpha1.FaultInjection.Conf")
	proto.RegisterType((*FaultInjection_Conf_Delay)(nil), "kuma.mesh.v1alpha1.FaultInjection.Conf.Delay")
	proto.RegisterType((*FaultInjection_Conf_Abort)(nil), "kuma.mesh.v1alpha1.FaultInjection.Conf.Abort")
	proto.RegisterType((*FaultInjection_Conf_ResponseBandwidth)(nil), "kuma.mesh.v1alpha1.FaultInjection.Conf.ResponseBandwidth")
}

func init() {
	proto.RegisterFile("mesh/v1alpha1/fault_injection.proto", fileDescriptor_ff4d722195e1e7eb)
}

var fileDescriptor_ff4d722195e1e7eb = []byte{
	// 412 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xa4, 0x92, 0x4f, 0x8b, 0xd4, 0x40,
	0x10, 0xc5, 0xc9, 0xee, 0x66, 0x95, 0x5a, 0x11, 0xb6, 0x4f, 0x31, 0x0c, 0xb2, 0xe8, 0xc1, 0xbd,
	0xd8, 0x61, 0xb3, 0x20, 0x88, 0x7b, 0xd0, 0xdd, 0x45, 0x98, 0x6b, 0x0f, 0x7a, 0xf0, 0x32, 0x74,
	0x92, 0xca, 0x1f, 0xed, 0xe9, 0x0e, 0xfd, 0x67, 0x06, 0x6f, 0x82, 0xf8, 0xe5, 0xfc, 0x54, 0x92,
	0xa4, 0xc3, 0x38, 0x46, 0x61, 0x64, 0x8e, 0xe9, 0x7a, 0xbf, 0x57, 0xaf, 0x2a, 0x05, 0xcf, 0x57,
	0x68, 0xea, 0x64, 0x7d, 0xc5, 0x45, 0x5b, 0xf3, 0xab, 0xa4, 0xe4, 0x4e, 0xd8, 0x65, 0x23, 0x3f,
	0x63, 0x6e, 0x1b, 0x25, 0x69, 0xab, 0x95, 0x55, 0x84, 0x7c, 0x71, 0x2b, 0x4e, 0x3b, 0x25, 0x1d,
	0x95, 0xf1, 0xd3, 0x4a, 0xa9, 0x4a, 0x60, 0xd2, 0x2b, 0x32, 0x57, 0x26, 0x85, 0xd3, 0x7c, 0xcb,
	0xc4, 0xb3, 0x5d, 0x63, 0x83, 0x02, 0x73, 0xab, 0xb4, 0xaf, 0x4e, 0xe8, 0x8d, 0xe6, 0x6d, 0x8b,
	0xda, 0x0c, 0xf5, 0x67, 0x3f, 0x4f, 0xe1, 0xf1, 0xfb, 0x2e, 0xcb, 0x7c, 0x8c, 0x42, 0x5e, 0xc1,
	0x03, 0xa3, 0x9c, 0xce, 0xd1, 0x44, 0xc1, 0xc5, 0xf1, 0xe5, 0x59, 0x3a, 0xa3, 0xd3, 0x58, 0x74,
	0xe1, 0xfb, 0xb0, 0x51, 0x4c, 0xde, 0xc2, 0xa3, 0x02, 0x8d, 0x6d, 0x64, 0x9f, 0xce, 0x44, 0x47,
	0x7b, 0xc0, 0x3b, 0x04, 0x79, 0x03, 0x27, 0xb9, 0x92, 0x65, 0x74, 0x7c, 0x11, 0x5c, 0x9e, 0xa5,
	0x2f, 0xfe, 0x46, 0xee, 0x66, 0xa5, 0x77, 0x4a, 0x96, 0xac, 0x87, 0xe2, 0x6f, 0x21, 0x9c, 0x74,
	0x9f, 0xe4, 0x0e, 0xc2, 0x02, 0x05, 0xff, 0x1a, 0x05, 0xbd, 0xcd, 0xcb, 0x3d, 0x6d, 0xe8, 0x7d,
	0x07, 0xb1, 0x81, 0xed, 0x4c, 0x78, 0xa6, 0xb4, 0x8d, 0x8e, 0xfe, 0xcf, 0xe4, 0x5d, 0x07, 0xb1,
	0x81, 0x25, 0x35, 0x10, 0x8d, 0xa6, 0x55, 0xd2, 0xe0, 0x32, 0xe3, 0xb2, 0xd8, 0x34, 0x85, 0xad,
	0xfd, 0x74, 0xaf, 0xf7, 0x75, 0x64, 0xde, 0xe1, 0x76, 0x34, 0x60, 0xe7, 0xfa, 0xcf, 0xa7, 0x78,
	0x0d, 0x61, 0x1f, 0x9f, 0xdc, 0x00, 0xb4, 0xa8, 0x73, 0x94, 0x96, 0x57, 0xe8, 0x37, 0x30, 0xa3,
	0xc3, 0x11, 0xd0, 0xf1, 0x08, 0xe8, 0xbd, 0x72, 0x99, 0xc0, 0x8f, 0x5c, 0x38, 0x64, 0xbf, 0xe9,
	0x49, 0x02, 0xe1, 0xba, 0x7b, 0xf4, 0x53, 0x3f, 0x99, 0x82, 0xfe, 0xf6, 0xd8, 0xa0, 0x8b, 0xbf,
	0x07, 0x10, 0xf6, 0x23, 0x1f, 0xd8, 0xf8, 0x06, 0xa0, 0xb6, 0xb6, 0x5d, 0x58, 0x6e, 0x9d, 0xf1,
	0xdd, 0xa7, 0xf4, 0x87, 0xb9, 0xb4, 0xd7, 0xa9, 0xa7, 0xb7, 0xfa, 0xf8, 0x47, 0x00, 0xe7, 0x93,
	0x35, 0x1d, 0x98, 0x28, 0x85, 0x50, 0x34, 0xab, 0xc6, 0xfe, 0x33, 0xcc, 0xc2, 0xea, 0x46, 0x56,
	0x03, 0x38, 0x48, 0x6f, 0xe1, 0xd3, 0xc3, 0xf1, 0x57, 0x66, 0xa7, 0xbd, 0xf0, 0xfa, 0x57, 0x00,
	0x00, 0x00, 0xff, 0xff, 0xeb, 0x09, 0x7e, 0x32, 0xf8, 0x03, 0x00, 0x00,
}
