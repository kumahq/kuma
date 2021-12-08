package generator

import (
	"github.com/pkg/errors"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	model "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
)

// OriginTransparent is a marker to indicate by which ProxyGenerator resources were generated.
const (
	OriginTransparent = "transparent"
	outboundNameIPv4  = "outbound:passthrough:ipv4"
	outboundNameIPv6  = "outbound:passthrough:ipv6"
	inboundNameIPv4   = "inbound:passthrough:ipv4"
	inboundNameIPv6   = "inbound:passthrough:ipv6"
	allIPv4           = "0.0.0.0"
	allIPv6           = "::"
	inPassThroughIPv4 = "127.0.0.6"
	inPassThroughIPv6 = "::6"
)

type TransparentProxyGenerator struct {
}

func (tpg TransparentProxyGenerator) Generate(ctx xds_context.Context, proxy *model.Proxy) (*model.ResourceSet, error) {
	resources := model.NewResourceSet()
	redirectPortOutbound := proxy.Dataplane.Spec.GetNetworking().GetTransparentProxying().GetRedirectPortOutbound()
	if redirectPortOutbound == 0 {
		return resources, nil
	}

	redirectPortInbound := proxy.Dataplane.Spec.Networking.GetTransparentProxying().GetRedirectPortInbound()
	resourcesIPv4, err := tpg.generate(ctx, proxy, outboundNameIPv4, inboundNameIPv4, allIPv4, inPassThroughIPv4, redirectPortOutbound, redirectPortInbound)
	if err != nil {
		return nil, err
	}
	resources.Add(resourcesIPv4.List()...)

	redirectPortInboundV6 := proxy.Dataplane.Spec.Networking.GetTransparentProxying().GetRedirectPortInboundV6()
	if redirectPortInboundV6 != 0 {
		resourcesIPv6, err := tpg.generate(ctx, proxy, outboundNameIPv6, inboundNameIPv6, allIPv6, inPassThroughIPv6, redirectPortOutbound, redirectPortInboundV6)
		if err != nil {
			return nil, err
		}
		resources.Add(resourcesIPv6.List()...)
	}

	return resources, nil
}

func (_ TransparentProxyGenerator) generate(ctx xds_context.Context, proxy *model.Proxy,
	outboundName, inboundName, allIP, inPassThroughIP string,
	redirectPortOutbound, redirectPortInbound uint32) (*model.ResourceSet, error) {
	resources := model.NewResourceSet()

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
		Configure(envoy_listeners.OutboundListener(outboundName, allIP, redirectPortOutbound, model.SocketAddressProtocolTCP)).
		Configure(envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder(proxy.APIVersion).
			Configure(envoy_listeners.TcpProxy(outboundName, envoy_common.NewCluster(envoy_common.WithService(outboundName)))).
			Configure(envoy_listeners.NetworkAccessLog(
				meshName,
				envoy_common.TrafficDirectionUnspecified,
				sourceService,
				"external",
				ctx.Mesh.GetLoggingBackend(proxy.Policies.TrafficLogs[core_mesh.PassThroughService]),
				proxy,
			)))).
		Configure(envoy_listeners.OriginalDstForwarder()).
		Build()
	if err != nil {
		return nil, errors.Wrapf(err, "could not generate listener: %s", outboundName)
	}

	inboundPassThroughCluster, err := envoy_clusters.NewClusterBuilder(proxy.APIVersion).
		Configure(envoy_clusters.PassThroughCluster(inboundName)).
		Configure(envoy_clusters.UpstreamBindConfig(inPassThroughIP, 0)).
		Build()
	if err != nil {
		return nil, errors.Wrapf(err, "could not generate cluster: %s", inboundName)
	}

	inboundListener, err := envoy_listeners.NewListenerBuilder(proxy.APIVersion).
		Configure(envoy_listeners.InboundListener(inboundName, allIP, redirectPortInbound, model.SocketAddressProtocolTCP)).
		Configure(envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder(proxy.APIVersion).
			Configure(envoy_listeners.TcpProxy(inboundName, envoy_common.NewCluster(envoy_common.WithService(inboundName)))))).
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
