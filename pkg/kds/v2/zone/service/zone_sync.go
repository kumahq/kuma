package v2

import (
	"time"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/kds/util"
	stream_v2 "github.com/kumahq/kuma/pkg/kds/v2/stream"
)

type Callbacks interface {
	OnGlobalToZoneSyncConnect(session stream_v2.Session) error
}

type OnGlobalToZoneSyncConnectFunc func(session stream_v2.Session) error

func (f OnGlobalToZoneSyncConnectFunc) OnGlobalToZoneSyncConnect(session stream_v2.Session) error {
	return f(session)
}

var clientLog = core.Log.WithName("kds-mux-client")

type KDSSyncServiceServer struct {
	timeout  time.Duration
	callback Callbacks
	filters  []stream_v2.Filter
	mesh_proto.UnimplementedKDSSyncServiceServer
}

func NewKDSSyncServiceServer(callback Callbacks, timeout time.Duration, filters []stream_v2.Filter) *KDSSyncServiceServer {
	return &KDSSyncServiceServer{
		callback: callback,
		timeout:  timeout,
		filters:  filters,
	}
}

var _ mesh_proto.KDSSyncServiceServer = &KDSSyncServiceServer{}

func (g *KDSSyncServiceServer) GlobalToZoneSync(stream mesh_proto.KDSSyncService_GlobalToZoneSyncServer) error {
	clientID, err := util.ClientIDFromIncomingCtx(stream.Context())
	if err != nil {
		return err
	}

	bufferSize := len(registry.Global().ObjectTypes())
	session := stream_v2.NewSession(clientID, stream, uint32(bufferSize), g.timeout)
	for _, filter := range g.filters {
		if err := filter.InterceptSession(session); err != nil {
			clientLog.Error(err, "closing KDS stream following a callback error")
			return err
		}
	}
	if err := g.callback.OnGlobalToZoneSyncConnect(session); err != nil {
		clientLog.Error(err, "closing KDS stream following a callback error")
		return err
	}
	err = <-session.Error()
	clientLog.Info("KDS stream is closed", "reason", err.Error())
	return nil
}