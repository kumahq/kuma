package server

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core"
	core_discovery "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/discovery"
	core_store "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	core_runtime "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/runtime"
	core_xds "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/xds"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/template"
	envoy_xds "github.com/envoyproxy/go-control-plane/pkg/server"
)

var (
	xdsServerLog = core.Log.WithName("xds-server")
)

func SetupServer(rt core_runtime.Runtime) error {
	r := newReconciler(rt.XDS(), rt.ResourceStore())
	for _, ds := range rt.DiscoverySources() {
		ds.AddConsumer(r)
	}

	srv := envoy_xds.NewServer(rt.XDS().Cache(), DefaultDataplaneStatusTracker(rt))
	return core_runtime.Add(
		rt,
		// xDS gRPC API
		&grpcServer{srv, rt.Config().XdsServer.GrpcPort},
		// xDS HTTP API
		&httpGateway{srv, rt.Config().XdsServer.HttpPort},
		// diagnostics server
		&diagnosticsServer{rt.Config().XdsServer.DiagnosticsPort})
}

func newReconciler(xds core_xds.XdsContext, rs core_store.ResourceStore) core_discovery.DiscoveryConsumer {
	return &core_discovery.DiscoverySink{
		DataplaneConsumer: &reconciler{
			&templateSnapshotGenerator{
				ProxyTemplateResolver: &simpleProxyTemplateResolver{
					ResourceStore:        rs,
					DefaultProxyTemplate: template.TransparentProxyTemplate,
				},
			},
			&simpleSnapshotCacher{xds.Hasher(), xds.Cache()},
		},
	}
}
