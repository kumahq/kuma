// Code generated by protoc-gen-go. DO NOT EDIT.
// source: mesh/v1alpha1/traffic_log.proto

package v1alpha1

import (
	fmt "fmt"
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

// TrafficLog defines log for traffic between dataplanes.
type TrafficLog struct {
	// List of selectors to match dataplanes that are sources of traffic.
	Sources []*Selector `protobuf:"bytes,1,rep,name=sources,proto3" json:"sources,omitempty"`
	// List of selectors to match services that are destinations of traffic.
	Destinations []*Selector `protobuf:"bytes,2,rep,name=destinations,proto3" json:"destinations,omitempty"`
	// Configuration of the logging.
	Conf                 *TrafficLog_Conf `protobuf:"bytes,3,opt,name=conf,proto3" json:"conf,omitempty"`
	XXX_NoUnkeyedLiteral struct{}         `json:"-"`
	XXX_unrecognized     []byte           `json:"-"`
	XXX_sizecache        int32            `json:"-"`
}

func (m *TrafficLog) Reset()         { *m = TrafficLog{} }
func (m *TrafficLog) String() string { return proto.CompactTextString(m) }
func (*TrafficLog) ProtoMessage()    {}
func (*TrafficLog) Descriptor() ([]byte, []int) {
	return fileDescriptor_47c4f4c9c894eeed, []int{0}
}

func (m *TrafficLog) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TrafficLog.Unmarshal(m, b)
}
func (m *TrafficLog) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TrafficLog.Marshal(b, m, deterministic)
}
func (m *TrafficLog) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TrafficLog.Merge(m, src)
}
func (m *TrafficLog) XXX_Size() int {
	return xxx_messageInfo_TrafficLog.Size(m)
}
func (m *TrafficLog) XXX_DiscardUnknown() {
	xxx_messageInfo_TrafficLog.DiscardUnknown(m)
}

var xxx_messageInfo_TrafficLog proto.InternalMessageInfo

func (m *TrafficLog) GetSources() []*Selector {
	if m != nil {
		return m.Sources
	}
	return nil
}

func (m *TrafficLog) GetDestinations() []*Selector {
	if m != nil {
		return m.Destinations
	}
	return nil
}

func (m *TrafficLog) GetConf() *TrafficLog_Conf {
	if m != nil {
		return m.Conf
	}
	return nil
}

// Configuration defines settings of the logging.
type TrafficLog_Conf struct {
	// Backend defined in the Mesh entity.
	Backend              string   `protobuf:"bytes,1,opt,name=backend,proto3" json:"backend,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *TrafficLog_Conf) Reset()         { *m = TrafficLog_Conf{} }
func (m *TrafficLog_Conf) String() string { return proto.CompactTextString(m) }
func (*TrafficLog_Conf) ProtoMessage()    {}
func (*TrafficLog_Conf) Descriptor() ([]byte, []int) {
	return fileDescriptor_47c4f4c9c894eeed, []int{0, 0}
}

func (m *TrafficLog_Conf) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TrafficLog_Conf.Unmarshal(m, b)
}
func (m *TrafficLog_Conf) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TrafficLog_Conf.Marshal(b, m, deterministic)
}
func (m *TrafficLog_Conf) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TrafficLog_Conf.Merge(m, src)
}
func (m *TrafficLog_Conf) XXX_Size() int {
	return xxx_messageInfo_TrafficLog_Conf.Size(m)
}
func (m *TrafficLog_Conf) XXX_DiscardUnknown() {
	xxx_messageInfo_TrafficLog_Conf.DiscardUnknown(m)
}

var xxx_messageInfo_TrafficLog_Conf proto.InternalMessageInfo

func (m *TrafficLog_Conf) GetBackend() string {
	if m != nil {
		return m.Backend
	}
	return ""
}

func init() {
	proto.RegisterType((*TrafficLog)(nil), "kuma.mesh.v1alpha1.TrafficLog")
	proto.RegisterType((*TrafficLog_Conf)(nil), "kuma.mesh.v1alpha1.TrafficLog.Conf")
}

func init() { proto.RegisterFile("mesh/v1alpha1/traffic_log.proto", fileDescriptor_47c4f4c9c894eeed) }

var fileDescriptor_47c4f4c9c894eeed = []byte{
	// 233 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x90, 0xb1, 0x4b, 0xc4, 0x30,
	0x14, 0xc6, 0x89, 0x77, 0x78, 0xf8, 0xce, 0x29, 0x53, 0x38, 0x0e, 0x2c, 0xba, 0x14, 0x87, 0x84,
	0x3b, 0x41, 0x57, 0xd1, 0xd5, 0xa9, 0x3a, 0xb9, 0x48, 0x9a, 0x4b, 0x9a, 0x70, 0x6d, 0x5e, 0x4d,
	0x52, 0xff, 0x5f, 0xff, 0x13, 0x69, 0x6b, 0x91, 0xa2, 0x83, 0xd3, 0xe3, 0x83, 0xef, 0xf7, 0xe3,
	0xf1, 0xc1, 0x45, 0xa3, 0xa3, 0x15, 0x1f, 0x3b, 0x59, 0xb7, 0x56, 0xee, 0x44, 0x0a, 0xd2, 0x18,
	0xa7, 0xde, 0x6a, 0xac, 0x78, 0x1b, 0x30, 0x21, 0xa5, 0xc7, 0xae, 0x91, 0xbc, 0x6f, 0xf1, 0xa9,
	0xb5, 0xd9, 0xce, 0xa1, 0xa8, 0x6b, 0xad, 0x12, 0x86, 0x91, 0xb8, 0xfc, 0x24, 0x00, 0x2f, 0xa3,
	0xe7, 0x09, 0x2b, 0x7a, 0x0b, 0xab, 0x88, 0x5d, 0x50, 0x3a, 0x32, 0x92, 0x2d, 0xf2, 0xf5, 0x7e,
	0xcb, 0x7f, 0x2b, 0xf9, 0xf3, 0xb7, 0xa3, 0x98, 0xca, 0xf4, 0x1e, 0xce, 0x0f, 0x3a, 0x26, 0xe7,
	0x65, 0x72, 0xe8, 0x23, 0x3b, 0xf9, 0x07, 0x3c, 0x23, 0xe8, 0x1d, 0x2c, 0x15, 0x7a, 0xc3, 0x16,
	0x19, 0xc9, 0xd7, 0xfb, 0xab, 0xbf, 0xc8, 0x9f, 0x3f, 0xf9, 0x23, 0x7a, 0x53, 0x0c, 0xc0, 0x26,
	0x83, 0x65, 0x9f, 0x28, 0x83, 0x55, 0x29, 0xd5, 0x51, 0xfb, 0x03, 0x23, 0x19, 0xc9, 0xcf, 0x8a,
	0x29, 0x3e, 0x5c, 0xbf, 0xe6, 0x95, 0x4b, 0xb6, 0x2b, 0xb9, 0xc2, 0x46, 0xf4, 0x62, 0xfb, 0x3e,
	0x1c, 0x21, 0x5b, 0x27, 0x66, 0xf3, 0x94, 0xa7, 0xc3, 0x2c, 0x37, 0x5f, 0x01, 0x00, 0x00, 0xff,
	0xff, 0x5b, 0xf8, 0x5a, 0x3a, 0x6b, 0x01, 0x00, 0x00,
}
