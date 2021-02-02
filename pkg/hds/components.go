package hds

import (
	"context"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_service_health "github.com/envoyproxy/go-control-plane/envoy/service/health/v3"

	hds_metrics "github.com/kumahq/kuma/pkg/hds/metrics"

	"github.com/kumahq/kuma/pkg/hds/authn"
	hds_callbacks "github.com/kumahq/kuma/pkg/hds/callbacks"
	hds_server "github.com/kumahq/kuma/pkg/hds/server"
	"github.com/kumahq/kuma/pkg/hds/tracker"
	"github.com/kumahq/kuma/pkg/xds/auth/components"

	"github.com/kumahq/kuma/pkg/core"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
)

var (
	hdsServerLog = core.Log.WithName("hds-server")
)

func Setup(rt core_runtime.Runtime) error {
	if rt.Config().DpServer.Hds.Enabled {
		return nil
	}

	snapshotCache := util_xds_v3.NewSnapshotCache(false, hasher{}, util_xds.NewLogger(hdsServerLog))

	callbacks, err := DefaultCallbacks(rt, snapshotCache)
	if err != nil {
		return err
	}

	srv := hds_server.New(context.Background(), snapshotCache, callbacks)

	hdsServerLog.Info("registering Health Discovery Service in Dataplane Server")
	envoy_service_health.RegisterHealthDiscoveryServiceServer(rt.DpServer().GrpcServer(), srv)
	return nil
}

func DefaultCallbacks(rt core_runtime.Runtime, cache util_xds_v3.SnapshotCache) (hds_callbacks.Callbacks, error) {
	authenticator, err := components.DefaultAuthenticator(rt)
	if err != nil {
		return nil, err
	}

	metrics, err := hds_metrics.NewMetrics(rt.Metrics())
	if err != nil {
		return nil, err
	}

	return hds_callbacks.Chain{
		authn.NewCallbacks(rt.ResourceManager(), authenticator),
		tracker.NewCallbacks(hdsServerLog, rt.ResourceManager(), rt.ReadOnlyResourceManager(),
			cache, rt.Config().DpServer.Hds, hasher{}, metrics),
	}, nil
}

type hasher struct {
}

func (_ hasher) ID(node *envoy_core.Node) string {
	return node.Id
}
