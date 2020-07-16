package generator

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/xds/envoy/names"

	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	model "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
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

	inboundPassThroughClusterName := "inbound:pass_through"
	inboundPassThroughCluster, err := envoy_clusters.NewClusterBuilder().
		Configure(envoy_clusters.PassThroughCluster(inboundPassThroughClusterName)).
		Configure(envoy_clusters.UpstreamBindConfig("127.0.0.6", 0)).
		Build()
	if err != nil {
		return nil, errors.Wrapf(err, "could not generate cluster: %s", inboundPassThroughClusterName)
	}

	inboundListenerName := names.GetInboundListenerName("0.0.0.0", 15006)
	inboundListener, err := envoy_listeners.NewListenerBuilder().
		Configure(envoy_listeners.InboundListener(inboundListenerName, "0.0.0.0", 15006)).
		Configure(envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder().
			Configure(envoy_listeners.TcpProxy(inboundPassThroughClusterName, envoy_common.ClusterSubset{ClusterName: inboundPassThroughClusterName})))).
		Configure(envoy_listeners.OriginalDstForwarder()).
		Build()
	if err != nil {
		return nil, errors.Wrapf(err, "could not generate listener: %s", inboundListenerName)
	}

	return []*model.Resource{
		{
			Name:     "catch_all",
			Version:  proxy.Dataplane.Meta.GetVersion(),
			Resource: listener,
		},
		{
			Name:     "pass_through",
			Version:  proxy.Dataplane.Meta.GetVersion(),
			Resource: cluster,
		},
		{
			Name:     inboundListenerName,
			Version:  proxy.Dataplane.Meta.GetVersion(),
			Resource: inboundListener,
		},
		{
			Name:     inboundPassThroughClusterName,
			Version:  proxy.Dataplane.Meta.GetVersion(),
			Resource: inboundPassThroughCluster,
		},
	}, nil
}
