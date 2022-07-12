package service

import (
	"google.golang.org/grpc"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/kds/util"
	util_grpc "github.com/kumahq/kuma/pkg/util/grpc"
)

type GlobalKDSServiceServer struct {
	envoyAdminRPCs EnvoyAdminRPCs
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
	core.Log.Info("Envoy Admin RPC stream started", "rpc", rpcName, "zone", zone)
	rpc.ClientConnected(zone, stream)
	defer rpc.ClientDisconnected(zone)
	for {
		resp, err := recv()
		if err != nil {
			return err
		}
		core.Log.V(1).Info("Envoy Admin RPC response received", "rpc", rpc, "zone", zone, "requestId", resp.GetRequestId())
		if err := rpc.ResponseReceived(zone, resp); err != nil {
			return err
		}
	}
}
