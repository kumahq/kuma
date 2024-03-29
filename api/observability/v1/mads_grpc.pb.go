// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package v1

import (
	context "context"
	v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// MonitoringAssignmentDiscoveryServiceClient is the client API for MonitoringAssignmentDiscoveryService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MonitoringAssignmentDiscoveryServiceClient interface {
	// HTTP
	FetchMonitoringAssignments(ctx context.Context, in *v3.DiscoveryRequest, opts ...grpc.CallOption) (*v3.DiscoveryResponse, error)
}

type monitoringAssignmentDiscoveryServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewMonitoringAssignmentDiscoveryServiceClient(cc grpc.ClientConnInterface) MonitoringAssignmentDiscoveryServiceClient {
	return &monitoringAssignmentDiscoveryServiceClient{cc}
}

func (c *monitoringAssignmentDiscoveryServiceClient) FetchMonitoringAssignments(ctx context.Context, in *v3.DiscoveryRequest, opts ...grpc.CallOption) (*v3.DiscoveryResponse, error) {
	out := new(v3.DiscoveryResponse)
	err := c.cc.Invoke(ctx, "/kuma.observability.v1.MonitoringAssignmentDiscoveryService/FetchMonitoringAssignments", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MonitoringAssignmentDiscoveryServiceServer is the server API for MonitoringAssignmentDiscoveryService service.
// All implementations must embed UnimplementedMonitoringAssignmentDiscoveryServiceServer
// for forward compatibility
type MonitoringAssignmentDiscoveryServiceServer interface {
	// HTTP
	FetchMonitoringAssignments(context.Context, *v3.DiscoveryRequest) (*v3.DiscoveryResponse, error)
	mustEmbedUnimplementedMonitoringAssignmentDiscoveryServiceServer()
}

// UnimplementedMonitoringAssignmentDiscoveryServiceServer must be embedded to have forward compatible implementations.
type UnimplementedMonitoringAssignmentDiscoveryServiceServer struct {
}

func (UnimplementedMonitoringAssignmentDiscoveryServiceServer) FetchMonitoringAssignments(context.Context, *v3.DiscoveryRequest) (*v3.DiscoveryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FetchMonitoringAssignments not implemented")
}
func (UnimplementedMonitoringAssignmentDiscoveryServiceServer) mustEmbedUnimplementedMonitoringAssignmentDiscoveryServiceServer() {
}

// UnsafeMonitoringAssignmentDiscoveryServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MonitoringAssignmentDiscoveryServiceServer will
// result in compilation errors.
type UnsafeMonitoringAssignmentDiscoveryServiceServer interface {
	mustEmbedUnimplementedMonitoringAssignmentDiscoveryServiceServer()
}

func RegisterMonitoringAssignmentDiscoveryServiceServer(s grpc.ServiceRegistrar, srv MonitoringAssignmentDiscoveryServiceServer) {
	s.RegisterService(&MonitoringAssignmentDiscoveryService_ServiceDesc, srv)
}

func _MonitoringAssignmentDiscoveryService_FetchMonitoringAssignments_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v3.DiscoveryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MonitoringAssignmentDiscoveryServiceServer).FetchMonitoringAssignments(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kuma.observability.v1.MonitoringAssignmentDiscoveryService/FetchMonitoringAssignments",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MonitoringAssignmentDiscoveryServiceServer).FetchMonitoringAssignments(ctx, req.(*v3.DiscoveryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// MonitoringAssignmentDiscoveryService_ServiceDesc is the grpc.ServiceDesc for MonitoringAssignmentDiscoveryService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var MonitoringAssignmentDiscoveryService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "kuma.observability.v1.MonitoringAssignmentDiscoveryService",
	HandlerType: (*MonitoringAssignmentDiscoveryServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "FetchMonitoringAssignments",
			Handler:    _MonitoringAssignmentDiscoveryService_FetchMonitoringAssignments_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/observability/v1/mads.proto",
}
