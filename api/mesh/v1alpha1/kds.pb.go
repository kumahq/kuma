// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.14.0
// source: mesh/v1alpha1/kds.proto

package v1alpha1

import (
	context "context"
	v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	anypb "google.golang.org/protobuf/types/known/anypb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type KumaResource struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Meta *KumaResource_Meta `protobuf:"bytes,1,opt,name=meta,proto3" json:"meta,omitempty"`
	Spec *anypb.Any         `protobuf:"bytes,2,opt,name=spec,proto3" json:"spec,omitempty"`
}

func (x *KumaResource) Reset() {
	*x = KumaResource{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mesh_v1alpha1_kds_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KumaResource) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KumaResource) ProtoMessage() {}

func (x *KumaResource) ProtoReflect() protoreflect.Message {
	mi := &file_mesh_v1alpha1_kds_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KumaResource.ProtoReflect.Descriptor instead.
func (*KumaResource) Descriptor() ([]byte, []int) {
	return file_mesh_v1alpha1_kds_proto_rawDescGZIP(), []int{0}
}

func (x *KumaResource) GetMeta() *KumaResource_Meta {
	if x != nil {
		return x.Meta
	}
	return nil
}

func (x *KumaResource) GetSpec() *anypb.Any {
	if x != nil {
		return x.Spec
	}
	return nil
}

type KumaResource_Meta struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name             string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Mesh             string                 `protobuf:"bytes,2,opt,name=mesh,proto3" json:"mesh,omitempty"`
	CreationTime     *timestamppb.Timestamp `protobuf:"bytes,3,opt,name=creation_time,json=creationTime,proto3" json:"creation_time,omitempty"`
	ModificationTime *timestamppb.Timestamp `protobuf:"bytes,4,opt,name=modification_time,json=modificationTime,proto3" json:"modification_time,omitempty"`
	Version          string                 `protobuf:"bytes,5,opt,name=version,proto3" json:"version,omitempty"`
}

func (x *KumaResource_Meta) Reset() {
	*x = KumaResource_Meta{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mesh_v1alpha1_kds_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KumaResource_Meta) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KumaResource_Meta) ProtoMessage() {}

func (x *KumaResource_Meta) ProtoReflect() protoreflect.Message {
	mi := &file_mesh_v1alpha1_kds_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KumaResource_Meta.ProtoReflect.Descriptor instead.
func (*KumaResource_Meta) Descriptor() ([]byte, []int) {
	return file_mesh_v1alpha1_kds_proto_rawDescGZIP(), []int{0, 0}
}

func (x *KumaResource_Meta) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *KumaResource_Meta) GetMesh() string {
	if x != nil {
		return x.Mesh
	}
	return ""
}

func (x *KumaResource_Meta) GetCreationTime() *timestamppb.Timestamp {
	if x != nil {
		return x.CreationTime
	}
	return nil
}

func (x *KumaResource_Meta) GetModificationTime() *timestamppb.Timestamp {
	if x != nil {
		return x.ModificationTime
	}
	return nil
}

func (x *KumaResource_Meta) GetVersion() string {
	if x != nil {
		return x.Version
	}
	return ""
}

var File_mesh_v1alpha1_kds_proto protoreflect.FileDescriptor

var file_mesh_v1alpha1_kds_proto_rawDesc = []byte{
	0x0a, 0x17, 0x6d, 0x65, 0x73, 0x68, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f,
	0x6b, 0x64, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x12, 0x6b, 0x75, 0x6d, 0x61, 0x2e,
	0x6d, 0x65, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x1a, 0x1f, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74,
	0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x2a,
	0x65, 0x6e, 0x76, 0x6f, 0x79, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2f, 0x64, 0x69,
	0x73, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x79, 0x2f, 0x76, 0x33, 0x2f, 0x64, 0x69, 0x73, 0x63, 0x6f,
	0x76, 0x65, 0x72, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x19, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x61, 0x6e, 0x79, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xc8, 0x02, 0x0a, 0x0c, 0x4b, 0x75, 0x6d, 0x61, 0x52, 0x65,
	0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x39, 0x0a, 0x04, 0x6d, 0x65, 0x74, 0x61, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x25, 0x2e, 0x6b, 0x75, 0x6d, 0x61, 0x2e, 0x6d, 0x65, 0x73, 0x68,
	0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x4b, 0x75, 0x6d, 0x61, 0x52, 0x65,
	0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e, 0x4d, 0x65, 0x74, 0x61, 0x52, 0x04, 0x6d, 0x65, 0x74,
	0x61, 0x12, 0x28, 0x0a, 0x04, 0x73, 0x70, 0x65, 0x63, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x14, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2e, 0x41, 0x6e, 0x79, 0x52, 0x04, 0x73, 0x70, 0x65, 0x63, 0x1a, 0xd2, 0x01, 0x0a, 0x04,
	0x4d, 0x65, 0x74, 0x61, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6d, 0x65, 0x73, 0x68,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6d, 0x65, 0x73, 0x68, 0x12, 0x3f, 0x0a, 0x0d,
	0x63, 0x72, 0x65, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52,
	0x0c, 0x63, 0x72, 0x65, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x47, 0x0a,
	0x11, 0x6d, 0x6f, 0x64, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x74, 0x69,
	0x6d, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73,
	0x74, 0x61, 0x6d, 0x70, 0x52, 0x10, 0x6d, 0x6f, 0x64, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f,
	0x6e, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e,
	0x32, 0x8e, 0x01, 0x0a, 0x14, 0x4b, 0x75, 0x6d, 0x61, 0x44, 0x69, 0x73, 0x63, 0x6f, 0x76, 0x65,
	0x72, 0x79, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x76, 0x0a, 0x13, 0x53, 0x74, 0x72,
	0x65, 0x61, 0x6d, 0x4b, 0x75, 0x6d, 0x61, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73,
	0x12, 0x2c, 0x2e, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x2e, 0x64, 0x69, 0x73, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x79, 0x2e, 0x76, 0x33, 0x2e, 0x44, 0x69,
	0x73, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2d,
	0x2e, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x64,
	0x69, 0x73, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x79, 0x2e, 0x76, 0x33, 0x2e, 0x44, 0x69, 0x73, 0x63,
	0x6f, 0x76, 0x65, 0x72, 0x79, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x28, 0x01, 0x30,
	0x01, 0x42, 0x2a, 0x5a, 0x28, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x6b, 0x75, 0x6d, 0x61, 0x68, 0x71, 0x2f, 0x6b, 0x75, 0x6d, 0x61, 0x2f, 0x61, 0x70, 0x69, 0x2f,
	0x6d, 0x65, 0x73, 0x68, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_mesh_v1alpha1_kds_proto_rawDescOnce sync.Once
	file_mesh_v1alpha1_kds_proto_rawDescData = file_mesh_v1alpha1_kds_proto_rawDesc
)

func file_mesh_v1alpha1_kds_proto_rawDescGZIP() []byte {
	file_mesh_v1alpha1_kds_proto_rawDescOnce.Do(func() {
		file_mesh_v1alpha1_kds_proto_rawDescData = protoimpl.X.CompressGZIP(file_mesh_v1alpha1_kds_proto_rawDescData)
	})
	return file_mesh_v1alpha1_kds_proto_rawDescData
}

var file_mesh_v1alpha1_kds_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_mesh_v1alpha1_kds_proto_goTypes = []interface{}{
	(*KumaResource)(nil),          // 0: kuma.mesh.v1alpha1.KumaResource
	(*KumaResource_Meta)(nil),     // 1: kuma.mesh.v1alpha1.KumaResource.Meta
	(*anypb.Any)(nil),             // 2: google.protobuf.Any
	(*timestamppb.Timestamp)(nil), // 3: google.protobuf.Timestamp
	(*v3.DiscoveryRequest)(nil),   // 4: envoy.service.discovery.v3.DiscoveryRequest
	(*v3.DiscoveryResponse)(nil),  // 5: envoy.service.discovery.v3.DiscoveryResponse
}
var file_mesh_v1alpha1_kds_proto_depIdxs = []int32{
	1, // 0: kuma.mesh.v1alpha1.KumaResource.meta:type_name -> kuma.mesh.v1alpha1.KumaResource.Meta
	2, // 1: kuma.mesh.v1alpha1.KumaResource.spec:type_name -> google.protobuf.Any
	3, // 2: kuma.mesh.v1alpha1.KumaResource.Meta.creation_time:type_name -> google.protobuf.Timestamp
	3, // 3: kuma.mesh.v1alpha1.KumaResource.Meta.modification_time:type_name -> google.protobuf.Timestamp
	4, // 4: kuma.mesh.v1alpha1.KumaDiscoveryService.StreamKumaResources:input_type -> envoy.service.discovery.v3.DiscoveryRequest
	5, // 5: kuma.mesh.v1alpha1.KumaDiscoveryService.StreamKumaResources:output_type -> envoy.service.discovery.v3.DiscoveryResponse
	5, // [5:6] is the sub-list for method output_type
	4, // [4:5] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_mesh_v1alpha1_kds_proto_init() }
func file_mesh_v1alpha1_kds_proto_init() {
	if File_mesh_v1alpha1_kds_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_mesh_v1alpha1_kds_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*KumaResource); i {
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
		file_mesh_v1alpha1_kds_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*KumaResource_Meta); i {
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
			RawDescriptor: file_mesh_v1alpha1_kds_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_mesh_v1alpha1_kds_proto_goTypes,
		DependencyIndexes: file_mesh_v1alpha1_kds_proto_depIdxs,
		MessageInfos:      file_mesh_v1alpha1_kds_proto_msgTypes,
	}.Build()
	File_mesh_v1alpha1_kds_proto = out.File
	file_mesh_v1alpha1_kds_proto_rawDesc = nil
	file_mesh_v1alpha1_kds_proto_goTypes = nil
	file_mesh_v1alpha1_kds_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// KumaDiscoveryServiceClient is the client API for KumaDiscoveryService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type KumaDiscoveryServiceClient interface {
	StreamKumaResources(ctx context.Context, opts ...grpc.CallOption) (KumaDiscoveryService_StreamKumaResourcesClient, error)
}

type kumaDiscoveryServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewKumaDiscoveryServiceClient(cc grpc.ClientConnInterface) KumaDiscoveryServiceClient {
	return &kumaDiscoveryServiceClient{cc}
}

func (c *kumaDiscoveryServiceClient) StreamKumaResources(ctx context.Context, opts ...grpc.CallOption) (KumaDiscoveryService_StreamKumaResourcesClient, error) {
	stream, err := c.cc.NewStream(ctx, &_KumaDiscoveryService_serviceDesc.Streams[0], "/kuma.mesh.v1alpha1.KumaDiscoveryService/StreamKumaResources", opts...)
	if err != nil {
		return nil, err
	}
	x := &kumaDiscoveryServiceStreamKumaResourcesClient{stream}
	return x, nil
}

type KumaDiscoveryService_StreamKumaResourcesClient interface {
	Send(*v3.DiscoveryRequest) error
	Recv() (*v3.DiscoveryResponse, error)
	grpc.ClientStream
}

type kumaDiscoveryServiceStreamKumaResourcesClient struct {
	grpc.ClientStream
}

func (x *kumaDiscoveryServiceStreamKumaResourcesClient) Send(m *v3.DiscoveryRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *kumaDiscoveryServiceStreamKumaResourcesClient) Recv() (*v3.DiscoveryResponse, error) {
	m := new(v3.DiscoveryResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// KumaDiscoveryServiceServer is the server API for KumaDiscoveryService service.
type KumaDiscoveryServiceServer interface {
	StreamKumaResources(KumaDiscoveryService_StreamKumaResourcesServer) error
}

// UnimplementedKumaDiscoveryServiceServer can be embedded to have forward compatible implementations.
type UnimplementedKumaDiscoveryServiceServer struct {
}

func (*UnimplementedKumaDiscoveryServiceServer) StreamKumaResources(KumaDiscoveryService_StreamKumaResourcesServer) error {
	return status.Errorf(codes.Unimplemented, "method StreamKumaResources not implemented")
}

func RegisterKumaDiscoveryServiceServer(s *grpc.Server, srv KumaDiscoveryServiceServer) {
	s.RegisterService(&_KumaDiscoveryService_serviceDesc, srv)
}

func _KumaDiscoveryService_StreamKumaResources_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(KumaDiscoveryServiceServer).StreamKumaResources(&kumaDiscoveryServiceStreamKumaResourcesServer{stream})
}

type KumaDiscoveryService_StreamKumaResourcesServer interface {
	Send(*v3.DiscoveryResponse) error
	Recv() (*v3.DiscoveryRequest, error)
	grpc.ServerStream
}

type kumaDiscoveryServiceStreamKumaResourcesServer struct {
	grpc.ServerStream
}

func (x *kumaDiscoveryServiceStreamKumaResourcesServer) Send(m *v3.DiscoveryResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *kumaDiscoveryServiceStreamKumaResourcesServer) Recv() (*v3.DiscoveryRequest, error) {
	m := new(v3.DiscoveryRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _KumaDiscoveryService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "kuma.mesh.v1alpha1.KumaDiscoveryService",
	HandlerType: (*KumaDiscoveryServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StreamKumaResources",
			Handler:       _KumaDiscoveryService_StreamKumaResources_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "mesh/v1alpha1/kds.proto",
}
