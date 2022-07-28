package hds

import (
	"context"
	"time"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_service_health "github.com/envoyproxy/go-control-plane/envoy/service/health/v3"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/hds/authn"
	hds_callbacks "github.com/kumahq/kuma/pkg/hds/callbacks"
	hds_metrics "github.com/kumahq/kuma/pkg/hds/metrics"
	hds_server "github.com/kumahq/kuma/pkg/hds/server"
	"github.com/kumahq/kuma/pkg/hds/tracker"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
	"github.com/kumahq/kuma/pkg/xds/auth/components"
)

var (
	hdsServerLog = core.Log.WithName("hds-server")
)

func Setup(rt core_runtime.Runtime) error {
	if rt.Config().Mode == config_core.Global {
		return nil
	}
	if !rt.Config().DpServer.Hds.Enabled {
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
		authn.NewCallbacks(rt.ResourceManager(), authenticator, authn.DPNotFoundRetry{
			// Usually the difference between DP is created from ADS and HDS is initiated is less than 1 second, but just in case we set this higher.
			Backoff:  1 * time.Second,
			MaxTimes: 30,
		}),
		tracker.NewCallbacks(
			hdsServerLog,
			rt.ResourceManager(),
			rt.ReadOnlyResourceManager(),
			cache,
			rt.Config().DpServer.Hds,
			hasher{},
			metrics,
			rt.Config().GetEnvoyAdminPort(),
		),
	}, nil
}

type hasher struct {
}

func (_ hasher) ID(node *envoy_core.Node) string {
	return node.Id
}
