package service

import (
<<<<<<< HEAD
=======
	"context"
	"fmt"
	"io"
	"time"

	"github.com/sethvargo/go-retry"
>>>>>>> 3ab585c92 (feat(kds): better error handling (#7868))
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/kds/util"
	util_grpc "github.com/kumahq/kuma/pkg/util/grpc"
)

var log = core.Log.WithName("kds-service")

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
		return status.Error(codes.InvalidArgument, err.Error())
	}
<<<<<<< HEAD
	core.Log.Info("Envoy Admin RPC stream started", "rpc", rpcName, "zone", zone)
	rpc.ClientConnected(zone, stream)
	defer rpc.ClientDisconnected(zone)
=======
	clientID := ClientID(stream.Context(), zone)
	logger := log.WithValues("rpc", rpcName, "clientID", clientID)
	logger.Info("Envoy Admin RPC stream started")
	rpc.ClientConnected(clientID, stream)
	if err := g.storeStreamConnection(stream.Context(), zone, rpcName, g.instanceID); err != nil {
		logger.Error(err, "could not store stream connection")
		return status.Error(codes.Internal, "could not store stream connection")
	}
	defer func() {
		rpc.ClientDisconnected(clientID)
		// stream.Context() is cancelled here, we need to use another ctx
		ctx := multitenant.CopyIntoCtx(stream.Context(), context.Background())
		if err := g.storeStreamConnection(ctx, zone, rpcName, ""); err != nil {
			logger.Error(err, "could not clear stream connection information in ZoneInsight")
		}
	}()
>>>>>>> 3ab585c92 (feat(kds): better error handling (#7868))
	for {
		resp, err := recv()
		if err == io.EOF {
			return nil
		}
<<<<<<< HEAD
		core.Log.V(1).Info("Envoy Admin RPC response received", "rpc", rpc, "zone", zone, "requestId", resp.GetRequestId())
		if err := rpc.ResponseReceived(zone, resp); err != nil {
			return err
=======
		if err != nil {
			logger.Error(err, "could not receive a message")
			return status.Error(codes.Internal, "could not receive a message")
		}
		logger.V(1).Info("Envoy Admin RPC response received", "requestId", resp.GetRequestId())
		if err := rpc.ResponseReceived(clientID, resp); err != nil {
			logger.Error(err, "could not mark the response as received")
			return status.Error(codes.InvalidArgument, "could not mark the response as received")
>>>>>>> 3ab585c92 (feat(kds): better error handling (#7868))
		}
	}
}
