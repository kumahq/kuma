package catalog

import (
	"google.golang.org/grpc"

	"github.com/kumahq/kuma/api/system/v1alpha1"
)

type Client interface {
	Close() error
	v1alpha1.InterCpPingServiceClient
}

type grpcClient struct {
	v1alpha1.InterCpPingServiceClient
	*grpc.ClientConn
}

var _ Client = &grpcClient{}

func NewGRPCClient(conn *grpc.ClientConn) Client {
	return &grpcClient{
		InterCpPingServiceClient: v1alpha1.NewInterCpPingServiceClient(conn),
		ClientConn:               conn,
	}
}
