package hds

import (
	"context"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_service_health "github.com/envoyproxy/go-control-plane/envoy/service/health/v3"
	"google.golang.org/grpc"

	"github.com/kumahq/kuma/pkg/core"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/hds/cache"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
)

var (
	hdsServerLog = core.Log.WithName("hds-server")
)

func RegisterHDS(rt core_runtime.Runtime, grpcSrv *grpc.Server) error {
	hasher := hasher{}
	snapshotCache := cache.NewSnapshotCache(false, hasher, util_xds.NewLogger(hdsServerLog))

	callbacks, err := DefaultCallbacks(rt, snapshotCache)
	if err != nil {
		return err
	}

	srv := NewServer(context.Background(), snapshotCache, callbacks)

	hdsServerLog.Info("registering Health Discovery Service in Dataplane Server")
	envoy_service_health.RegisterHealthDiscoveryServiceServer(grpcSrv, srv)
	return nil
}

func DefaultCallbacks(rt core_runtime.Runtime, cache cache.SnapshotCache) (Callbacks, error) {
	return NewTracker(rt.ResourceManager(), cache, rt.Config().DpServer.Hds), nil
}

type hasher struct {
}

func (_ hasher) ID(node *envoy_core.Node) string {
	return node.Id
}
