package service

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/kds/util"
	"github.com/kumahq/kuma/pkg/multitenant"
	util_grpc "github.com/kumahq/kuma/pkg/util/grpc"
)

type GlobalKDSServiceServer struct {
	envoyAdminRPCs EnvoyAdminRPCs
	mesh_proto.UnimplementedGlobalKDSServiceServer
}

func NewGlobalKDSServiceServer(envoyAdminRPCs EnvoyAdminRPCs) *GlobalKDSServiceServer {
	return &GlobalKDSServiceServer{
		envoyAdminRPCs: envoyAdminRPCs,
	}
}

var _ mesh_proto.GlobalKDSServiceServer = &GlobalKDSServiceServer{}

func (g *GlobalKDSServiceServer) StreamXDSConfigs(stream mesh_proto.GlobalKDSService_StreamXDSConfigsServer) error {
	return g.streamEnvoyAdminRPC("XDS Config Dump", g.envoyAdminRPCs.XDSConfigDump, stream, func() (util_grpc.ReverseUnaryMessage, error) {
		return stream.Recv()
	})
}

func (g *GlobalKDSServiceServer) StreamStats(stream mesh_proto.GlobalKDSService_StreamStatsServer) error {
	return g.streamEnvoyAdminRPC("Stats", g.envoyAdminRPCs.Stats, stream, func() (util_grpc.ReverseUnaryMessage, error) {
		return stream.Recv()
	})
}

func (g *GlobalKDSServiceServer) StreamClusters(stream mesh_proto.GlobalKDSService_StreamClustersServer) error {
	return g.streamEnvoyAdminRPC("Clusters", g.envoyAdminRPCs.Clusters, stream, func() (util_grpc.ReverseUnaryMessage, error) {
		return stream.Recv()
	})
}

func (g *GlobalKDSServiceServer) streamEnvoyAdminRPC(
	rpcName string,
	rpc util_grpc.ReverseUnaryRPCs,
	stream grpc.ServerStream,
	recv func() (util_grpc.ReverseUnaryMessage, error),
) error {
	zone, err := util.ClientIDFromIncomingCtx(stream.Context())
	if err != nil {
		return err
	}
	clientID := ClientID(stream.Context(), zone)
	core.Log.Info("Envoy Admin RPC stream started", "rpc", rpcName, "clientID", clientID)
	rpc.ClientConnected(clientID, stream)
	defer rpc.ClientDisconnected(clientID)
	for {
		resp, err := recv()
		if err != nil {
			return err
		}
		core.Log.V(1).Info("Envoy Admin RPC response received", "rpc", rpc, "clientID", clientID, "requestId", resp.GetRequestId())
		if err := rpc.ResponseReceived(clientID, resp); err != nil {
			return err
		}
	}
}

func ClientID(ctx context.Context, zone string) string {
	tenantID, _ := multitenant.TenantFromCtx(ctx)
	return fmt.Sprintf("%s:%s", zone, tenantID)
}
