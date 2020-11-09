// Code generated by protoc-gen-go. DO NOT EDIT.
// source: system/v1alpha1/zone.proto

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

// Zone defines the Zone configuration used at the Global Control Plane
// within a distributed deployment
type Zone struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Zone) Reset()         { *m = Zone{} }
func (m *Zone) String() string { return proto.CompactTextString(m) }
func (*Zone) ProtoMessage()    {}
func (*Zone) Descriptor() ([]byte, []int) {
	return fileDescriptor_0b77e158d4963c0f, []int{0}
}

func (m *Zone) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Zone.Unmarshal(m, b)
}
func (m *Zone) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Zone.Marshal(b, m, deterministic)
}
func (m *Zone) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Zone.Merge(m, src)
}
func (m *Zone) XXX_Size() int {
	return xxx_messageInfo_Zone.Size(m)
}
func (m *Zone) XXX_DiscardUnknown() {
	xxx_messageInfo_Zone.DiscardUnknown(m)
}

var xxx_messageInfo_Zone proto.InternalMessageInfo

func init() {
	proto.RegisterType((*Zone)(nil), "kuma.system.v1alpha1.Zone")
}

func init() { proto.RegisterFile("system/v1alpha1/zone.proto", fileDescriptor_0b77e158d4963c0f) }

var fileDescriptor_0b77e158d4963c0f = []byte{
	// 83 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0x2a, 0xae, 0x2c, 0x2e,
	0x49, 0xcd, 0xd5, 0x2f, 0x33, 0x4c, 0xcc, 0x29, 0xc8, 0x48, 0x34, 0xd4, 0xaf, 0xca, 0xcf, 0x4b,
	0xd5, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x12, 0xc9, 0x2e, 0xcd, 0x4d, 0xd4, 0x83, 0x28, 0xd0,
	0x83, 0x29, 0x50, 0x62, 0xe3, 0x62, 0x89, 0xca, 0xcf, 0x4b, 0x75, 0xe2, 0x8a, 0xe2, 0x80, 0x89,
	0x25, 0xb1, 0x81, 0x35, 0x18, 0x03, 0x02, 0x00, 0x00, 0xff, 0xff, 0xa4, 0xd6, 0x08, 0x45, 0x4e,
	0x00, 0x00, 0x00,
}
