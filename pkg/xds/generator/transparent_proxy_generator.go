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
const (
	OriginTransparent = "transparent"
	outboundName      = "outbound:passthrough"
	inboundName       = "inbound:passthrough"
)

type TransparentProxyGenerator struct {
}

func (_ TransparentProxyGenerator) Generate(ctx xds_context.Context, proxy *model.Proxy) (*model.ResourceSet, error) {
	redirectPortOutbound := proxy.Dataplane.Spec.GetNetworking().GetTransparentProxying().GetRedirectPortOutbound()
	resources := model.NewResourceSet()
	if redirectPortOutbound == 0 {
		return resources, nil
	}
	sourceService := proxy.Dataplane.Spec.GetIdentifyingService()
	meshName := ctx.Mesh.Resource.GetMeta().GetName()

	var outboundPassThroughCluster envoy_common.NamedResource
	var outboundListener envoy_common.NamedResource
	var err error

	if ctx.Mesh.Resource.Spec.IsPassthrough() {
		outboundPassThroughCluster, err = envoy_clusters.NewClusterBuilder(proxy.APIVersion).
			Configure(envoy_clusters.PassThroughCluster(outboundName)).
			Build()
		if err != nil {
			return nil, errors.Wrapf(err, "could not generate outbound cluster: %s", outboundName)
		}
	}

	outboundListener, err = envoy_listeners.NewListenerBuilder(proxy.APIVersion).
		Configure(envoy_listeners.OutboundListener(outboundName, "0.0.0.0", redirectPortOutbound)).
		Configure(envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder(proxy.APIVersion).
			Configure(envoy_listeners.TcpProxy(outboundName, envoy_common.ClusterSubset{ClusterName: outboundName})).
			Configure(envoy_listeners.NetworkAccessLog(meshName, envoy_common.TrafficDirectionUnspecified, sourceService, "external", proxy.Policies.Logs[mesh_core.PassThroughService], proxy)))).
		Configure(envoy_listeners.OriginalDstForwarder()).
		Build()
	if err != nil {
		return nil, errors.Wrapf(err, "could not generate listener: %s", outboundName)
	}

	redirectPortInbound := proxy.Dataplane.Spec.Networking.GetTransparentProxying().GetRedirectPortInbound()

	inboundPassThroughCluster, err := envoy_clusters.NewClusterBuilder(proxy.APIVersion).
		Configure(envoy_clusters.PassThroughCluster(inboundName)).
		Configure(envoy_clusters.UpstreamBindConfig("127.0.0.6", 0)).
		Build()
	if err != nil {
		return nil, errors.Wrapf(err, "could not generate cluster: %s", inboundName)
	}

	inboundListener, err := envoy_listeners.NewListenerBuilder(proxy.APIVersion).
		Configure(envoy_listeners.InboundListener(inboundName, "0.0.0.0", redirectPortInbound)).
		Configure(envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder(proxy.APIVersion).
			Configure(envoy_listeners.TcpProxy(inboundName, envoy_common.ClusterSubset{ClusterName: inboundName})))).
		Configure(envoy_listeners.OriginalDstForwarder()).
		Build()
	if err != nil {
		return nil, errors.Wrapf(err, "could not generate listener: %s", inboundName)
	}

	resources.Add(&model.Resource{
		Name:     outboundListener.GetName(),
		Origin:   OriginTransparent,
		Resource: outboundListener,
	})

	if ctx.Mesh.Resource.Spec.IsPassthrough() {
		resources.Add(&model.Resource{
			Name:     outboundPassThroughCluster.GetName(),
			Origin:   OriginTransparent,
			Resource: outboundPassThroughCluster,
		})
	}
	resources.Add(&model.Resource{
		Name:     inboundListener.GetName(),
		Origin:   OriginTransparent,
		Resource: inboundListener,
	})
	resources.Add(&model.Resource{
		Name:     inboundPassThroughCluster.GetName(),
		Origin:   OriginTransparent,
		Resource: inboundPassThroughCluster,
	})
	return resources, nil
}
