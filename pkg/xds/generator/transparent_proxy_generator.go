package generator

import (
	"context"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
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
	OutboundNameIPv4  = "outbound:passthrough:ipv4"
	OutboundNameIPv6  = "outbound:passthrough:ipv6"
	InboundNameIPv4   = "inbound:passthrough:ipv4"
	InboundNameIPv6   = "inbound:passthrough:ipv6"
	InPassThroughIPv4 = "127.0.0.6"
	InPassThroughIPv6 = "::6"
	allIPv4           = "0.0.0.0"
	allIPv6           = "::"
)

type TransparentProxyGenerator struct{}

func (tpg TransparentProxyGenerator) Generate(_ context.Context, _ *model.ResourceSet, xdsCtx xds_context.Context, proxy *model.Proxy) (*model.ResourceSet, error) {
	resources := model.NewResourceSet()
	redirectPortOutbound := proxy.Dataplane.Spec.GetNetworking().GetTransparentProxying().GetRedirectPortOutbound()
	if redirectPortOutbound == 0 {
		return resources, nil
	}

	redirectPortInbound := proxy.Dataplane.Spec.Networking.GetTransparentProxying().GetRedirectPortInbound()
	resourcesIPv4, err := tpg.generate(xdsCtx, proxy, OutboundNameIPv4, InboundNameIPv4, allIPv4, InPassThroughIPv4, redirectPortOutbound, redirectPortInbound)
	if err != nil {
		return nil, err
	}
	resources.Add(resourcesIPv4.List()...)

	ipFamilyMode := proxy.Dataplane.Spec.Networking.GetTransparentProxying().GetIpFamilyMode()
	if ipFamilyMode != mesh_proto.Dataplane_Networking_TransparentProxying_IPv4 {
		resourcesIPv6, err := tpg.generate(xdsCtx, proxy, OutboundNameIPv6, InboundNameIPv6, allIPv6, InPassThroughIPv6, redirectPortOutbound, redirectPortInbound)
		if err != nil {
			return nil, err
		}
		resources.Add(resourcesIPv6.List()...)
	}

	return resources, nil
}

func (TransparentProxyGenerator) generate(ctx xds_context.Context, proxy *model.Proxy,
	outboundName, inboundName, allIP, inPassThroughIP string,
	redirectPortOutbound, redirectPortInbound uint32,
) (*model.ResourceSet, error) {
	resources := model.NewResourceSet()

	sourceService := proxy.Dataplane.Spec.GetIdentifyingService()
	meshName := ctx.Mesh.Resource.GetMeta().GetName()

	var outboundPassThroughCluster envoy_common.NamedResource
	var outboundListener envoy_common.NamedResource
	var err error

	if ctx.Mesh.Resource.Spec.IsPassthrough() {
		outboundPassThroughCluster, err = envoy_clusters.NewClusterBuilder(proxy.APIVersion, outboundName).
			Configure(envoy_clusters.PassThroughCluster()).
			Configure(envoy_clusters.DefaultTimeout()).
			Build()
		if err != nil {
			return nil, errors.Wrapf(err, "could not generate outbound cluster: %s", outboundName)
		}
	}

	outboundListener, err = envoy_listeners.NewOutboundListenerBuilder(proxy.APIVersion, allIP, redirectPortOutbound, model.SocketAddressProtocolTCP).
		WithOverwriteName(outboundName).
		Configure(envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, envoy_common.AnonymousResource).
			Configure(envoy_listeners.TcpProxyDeprecated(outboundName, envoy_common.NewCluster(envoy_common.WithService(outboundName)))).
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

	inboundPassThroughCluster, err := envoy_clusters.NewClusterBuilder(proxy.APIVersion, inboundName).
		Configure(envoy_clusters.PassThroughCluster()).
		Configure(envoy_clusters.UpstreamBindConfig(inPassThroughIP, 0)).
		Configure(envoy_clusters.DefaultTimeout()).
		Build()
	if err != nil {
		return nil, errors.Wrapf(err, "could not generate cluster: %s", inboundName)
	}

	inboundListener, err := envoy_listeners.NewInboundListenerBuilder(proxy.APIVersion, allIP, redirectPortInbound, model.SocketAddressProtocolTCP).
		WithOverwriteName(inboundName).
		Configure(envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, envoy_common.AnonymousResource).
			Configure(envoy_listeners.TcpProxyDeprecated(inboundName, envoy_common.NewCluster(envoy_common.WithService(inboundName)))))).
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
