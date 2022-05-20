package service

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/kds/util"
)

type GlobalKDSServiceServer struct {
	xdsConfigStreams XDSConfigStreams
}

func NewGlobalKDSServiceServer(xdsConfigStreams XDSConfigStreams) *GlobalKDSServiceServer {
	return &GlobalKDSServiceServer{
		xdsConfigStreams: xdsConfigStreams,
	}
}

var _ mesh_proto.GlobalKDSServiceServer = &GlobalKDSServiceServer{}

func (g *GlobalKDSServiceServer) StreamXDSConfigs(stream mesh_proto.GlobalKDSService_StreamXDSConfigsServer) error {
	zone, err := util.ClientIDFromIncomingCtx(stream.Context())
	if err != nil {
		return err
	}
	core.Log.Info("XDSConfigs stream started", "zone", zone)
	g.xdsConfigStreams.ZoneConnected(zone, stream)
	defer g.xdsConfigStreams.ZoneDisconnected(zone)
	for {
		resp, err := stream.Recv()
		if err != nil {
			return err
		}
		core.Log.V(1).Info("XDSConfigs response received", "zone", zone, "requestId", resp.RequestId)
		if err := g.xdsConfigStreams.ResponseReceived(zone, resp); err != nil {
			return err
		}
	}
}
