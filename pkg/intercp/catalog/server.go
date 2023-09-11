package catalog

import (
	"context"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
)

var serverLog = core.Log.WithName("intercp").WithName("catalog").WithName("server")

type server struct {
	heartbeats *Heartbeats
	leaderInfo component.LeaderInfo

	system_proto.UnimplementedInterCpPingServiceServer
}

var _ system_proto.InterCpPingServiceServer = &server{}

func NewServer(heartbeats *Heartbeats, leaderInfo component.LeaderInfo) system_proto.InterCpPingServiceServer {
	return &server{
		heartbeats: heartbeats,
		leaderInfo: leaderInfo,
	}
}

func (s *server) Ping(_ context.Context, request *system_proto.PingRequest) (*system_proto.PingResponse, error) {
	serverLog.V(1).Info("received ping", "instanceID", request.InstanceId, "address", request.Address, "ready", request.Ready)
	instance := Instance{
		Id:          request.InstanceId,
		Address:     request.Address,
		InterCpPort: uint16(request.InterCpPort),
		Leader:      false,
	}
	if request.Ready {
		s.heartbeats.Add(instance)
	} else {
		s.heartbeats.Remove(instance)
	}
	return &system_proto.PingResponse{
		Leader: s.leaderInfo.IsLeader(),
	}, nil
}
