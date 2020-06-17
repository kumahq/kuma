package generator

import (
	"github.com/pkg/errors"

	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	model "github.com/Kong/kuma/pkg/core/xds"
	xds_context "github.com/Kong/kuma/pkg/xds/context"
	envoy_common "github.com/Kong/kuma/pkg/xds/envoy"
	envoy_clusters "github.com/Kong/kuma/pkg/xds/envoy/clusters"
	envoy_listeners "github.com/Kong/kuma/pkg/xds/envoy/listeners"
)

type TransparentProxyGenerator struct {
}

func (_ TransparentProxyGenerator) Generate(ctx xds_context.Context, proxy *model.Proxy) ([]*model.Resource, error) {
	redirectPort := proxy.Dataplane.Spec.Networking.GetTransparentProxying().GetRedirectPort()
	if redirectPort == 0 {
		return nil, nil
	}
	sourceService := proxy.Dataplane.Spec.GetIdentifyingService()
	meshName := ctx.Mesh.Resource.GetMeta().GetName()
	listener, err := envoy_listeners.NewListenerBuilder().
		Configure(envoy_listeners.OutboundListener("catch_all", "0.0.0.0", redirectPort)).
		Configure(envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder().
			Configure(envoy_listeners.TcpProxy("pass_through", envoy_common.ClusterSubset{ClusterName: "pass_through"})).
			Configure(envoy_listeners.NetworkAccessLog(meshName, envoy_listeners.TrafficDirectionUnspecified, sourceService, "external", proxy.Logs[mesh_core.PassThroughService], proxy)))).
		Configure(envoy_listeners.OriginalDstForwarder()).
		Build()
	if err != nil {
		return nil, errors.Wrapf(err, "could not generate listener: catch_all")
	}
	cluster, err := envoy_clusters.NewClusterBuilder().
		Configure(envoy_clusters.PassThroughCluster("pass_through")).
		Build()
	if err != nil {
		return nil, errors.Wrapf(err, "could not generate cluster: pass_through")
	}
	return []*model.Resource{
		&model.Resource{
			Name:     "catch_all",
			Version:  proxy.Dataplane.Meta.GetVersion(),
			Resource: listener,
		},
		&model.Resource{
			Name:     "pass_through",
			Version:  proxy.Dataplane.Meta.GetVersion(),
			Resource: cluster,
		},
	}, nil
}
