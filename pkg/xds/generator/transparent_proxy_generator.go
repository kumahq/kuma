package generator

import (
	"github.com/pkg/errors"

	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	model "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
)

// OriginTransparent is a marker to indicate by which ProxyGenerator resources were generated.
const OriginTransparent = "transparent"

type TransparentProxyGenerator struct {
}

func (_ TransparentProxyGenerator) Generate(ctx xds_context.Context, proxy *model.Proxy) (*model.ResourceSet, error) {
	redirectPortOutbound := proxy.Dataplane.Spec.Networking.GetTransparentProxying().GetRedirectPortOutbound()
	resources := model.NewResourceSet()
	if redirectPortOutbound == 0 {
		return resources, nil
	}
	sourceService := proxy.Dataplane.Spec.GetIdentifyingService()
	meshName := ctx.Mesh.Resource.GetMeta().GetName()
	const outboundName = "outbound:passthrough"
	listener, err := envoy_listeners.NewListenerBuilder().
		Configure(envoy_listeners.OutboundListener(outboundName, "0.0.0.0", redirectPortOutbound)).
		Configure(envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder().
			Configure(envoy_listeners.TcpProxy(outboundName, envoy_common.ClusterSubset{ClusterName: outboundName})).
			Configure(envoy_listeners.NetworkAccessLog(meshName, envoy_listeners.TrafficDirectionUnspecified, sourceService, "external", proxy.Logs[mesh_core.PassThroughService], proxy)))).
		Configure(envoy_listeners.OriginalDstForwarder()).
		Build()
	if err != nil {
		return nil, errors.Wrapf(err, "could not generate listener: %s", outboundName)
	}
	if err := listener.Validate(); err != nil {
		return nil, errors.Wrapf(err, "could not generate listener: %s", outboundName)
	}

	cluster, err := envoy_clusters.NewClusterBuilder().
		Configure(envoy_clusters.PassThroughCluster(outboundName)).
		Build()
	if err != nil {
		return nil, errors.Wrapf(err, "could not generate cluster: %s", outboundName)
	}

	redirectPortInbound := proxy.Dataplane.Spec.Networking.GetTransparentProxying().GetRedirectPortInbound()
	const inboundName = "inbound:passthrough"
	inboundPassThroughCluster, err := envoy_clusters.NewClusterBuilder().
		Configure(envoy_clusters.PassThroughCluster(inboundName)).
		Configure(envoy_clusters.UpstreamBindConfig("127.0.0.6", 0)).
		Build()
	if err != nil {
		return nil, errors.Wrapf(err, "could not generate cluster: %s", inboundName)
	}

	inboundListener, err := envoy_listeners.NewListenerBuilder().
		Configure(envoy_listeners.InboundListener(inboundName, "0.0.0.0", redirectPortInbound)).
		Configure(envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder().
			Configure(envoy_listeners.TcpProxy(inboundName, envoy_common.ClusterSubset{ClusterName: inboundName})))).
		Configure(envoy_listeners.OriginalDstForwarder()).
		Build()
	if err != nil {
		return nil, errors.Wrapf(err, "could not generate listener: %s", inboundName)
	}
	if err := inboundListener.Validate(); err != nil {
		return nil, errors.Wrapf(err, "could not generate listener: %s", inboundName)
	}

	resources.Add(&model.Resource{
		Name:     listener.Name,
		Origin:   OriginTransparent,
		Resource: listener,
	})
	resources.Add(&model.Resource{
		Name:     cluster.Name,
		Origin:   OriginTransparent,
		Resource: cluster,
	})
	resources.Add(&model.Resource{
		Name:     inboundListener.Name,
		Origin:   OriginTransparent,
		Resource: inboundListener,
	})
	resources.Add(&model.Resource{
		Name:     inboundPassThroughCluster.Name,
		Origin:   OriginTransparent,
		Resource: inboundPassThroughCluster,
	})
	return resources, nil
}
