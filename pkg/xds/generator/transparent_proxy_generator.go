package generator

import (
	"context"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	core_meta "github.com/kumahq/kuma/v2/pkg/core/metadata"
	"github.com/kumahq/kuma/v2/pkg/core/naming"
	unified_naming "github.com/kumahq/kuma/v2/pkg/core/naming/unified-naming"
	model "github.com/kumahq/kuma/v2/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/v2/pkg/core/xds/types"
	xds_context "github.com/kumahq/kuma/v2/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/v2/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/v2/pkg/xds/envoy/clusters"
	envoy_listeners "github.com/kumahq/kuma/v2/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/v2/pkg/xds/generator/metadata"
)

type TransparentProxyGenerator struct{}

func (tpg TransparentProxyGenerator) Generate(_ context.Context, _ *model.ResourceSet, xdsCtx xds_context.Context, proxy *model.Proxy) (*model.ResourceSet, error) {
	resources := model.NewResourceSet()

	tpCfg := proxy.GetTransparentProxy()
	if !tpCfg.Redirect.Outbound.Enabled || proxy.Metadata.HasFeature(xds_types.FeatureBindOutbounds) {
		return resources, nil
	}
	noDestCluster, err := CreateNoDestinationCluster(proxy.APIVersion, naming.ContextualNoDestinationClusterName("inbound"))
	if err != nil {
		return nil, err
	}
	resources.Add(&model.Resource{
		Name:     noDestCluster.GetName(),
		Origin:   metadata.OriginTransparent,
		Resource: noDestCluster,
	})

	resourcesIPv4, err := tpg.generateIPv4(xdsCtx, proxy)
	if err != nil {
		return nil, err
	}
	resources.Add(resourcesIPv4.List()...)

	if tpCfg.EnabledIPv6() {
		resourcesIPv6, err := tpg.generateIPv6(xdsCtx, proxy)
		if err != nil {
			return nil, err
		}
		resources.Add(resourcesIPv6.List()...)
	}

	return resources, nil
}

// CreateOutboundPassThroughCluster creates an outbound passthrough cluster
func CreateOutboundPassThroughCluster(apiVersion model.APIVersion, clusterName string) (envoy_common.NamedResource, error) {
	cluster, err := envoy_clusters.NewClusterBuilder(apiVersion, clusterName).
		Configure(envoy_clusters.PassThroughCluster()).
		Configure(envoy_clusters.DefaultTimeout()).
		Build()
	if err != nil {
		return nil, errors.Wrapf(err, "could not generate outbound cluster: %s", clusterName)
	}
	return cluster, nil
}

// CreateOutboundListener creates an outbound listener for transparent proxy
func CreateOutboundPassthroughListener(
	proxy *model.Proxy,
	ctx xds_context.Context,
	listenerName string,
	allIP string,
	outboundPort uint32,
	statPrefix string,
) (envoy_common.NamedResource, error) {
	sourceService := proxy.Dataplane.Spec.GetIdentifyingService()
	meshName := ctx.Mesh.Resource.GetMeta().GetName()

	listener, err := envoy_listeners.NewOutboundListenerBuilder(proxy.APIVersion, allIP, outboundPort, model.SocketAddressProtocolTCP).
		WithOverwriteName(listenerName).
		Configure(envoy_listeners.StatPrefix(statPrefix)).
		Configure(envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, envoy_common.AnonymousResource).
			Configure(envoy_listeners.TcpProxyDeprecated(listenerName, envoy_common.NewCluster(envoy_common.WithService(listenerName)))).
			Configure(envoy_listeners.NetworkAccessLog(
				meshName,
				envoy_common.TrafficDirectionUnspecified,
				sourceService,
				"external",
				ctx.Mesh.GetLoggingBackend(proxy.Policies.TrafficLogs[core_meta.PassThroughServiceName]),
				proxy,
			)))).
		Configure(envoy_listeners.OriginalDstForwarder()).
		Build()
	if err != nil {
		return nil, errors.Wrapf(err, "could not generate listener: %s", listenerName)
	}
	return listener, nil
}

// CreateInboundPassThroughCluster creates an inbound passthrough cluster
func CreateInboundPassThroughCluster(apiVersion model.APIVersion, clusterName string, inPassThroughIP string) (envoy_common.NamedResource, error) {
	cluster, err := envoy_clusters.NewClusterBuilder(apiVersion, clusterName).
		Configure(envoy_clusters.PassThroughCluster()).
		Configure(envoy_clusters.UpstreamBindConfig(inPassThroughIP, 0)).
		Configure(envoy_clusters.DefaultTimeout()).
		Build()
	if err != nil {
		return nil, errors.Wrapf(err, "could not generate cluster: %s", clusterName)
	}
	return cluster, nil
}

// CreateNoDestinationCluster creates a cluster which has no destination
func CreateNoDestinationCluster(apiVersion model.APIVersion, clusterName string) (envoy_common.NamedResource, error) {
	cluster, err := envoy_clusters.NewClusterBuilder(apiVersion, clusterName).
		Configure(envoy_clusters.StaticNoEndpointsCluster()).
		Configure(envoy_clusters.DefaultTimeout()).
		Build()
	if err != nil {
		return nil, errors.Wrapf(err, "could not generate cluster: %s", clusterName)
	}
	return cluster, nil
}

// CreateInboundListener creates an inbound listener for transparent proxy
func CreateInboundPassthroughListener(
	proxy *model.Proxy,
	listenerName string,
	allIP string,
	inboundPort uint32,
	useStrictInboundPorts bool,
	statPrefix string,
) (envoy_common.NamedResource, error) {
	inboundListenerBuilder := envoy_listeners.NewInboundListenerBuilder(proxy.APIVersion, allIP, inboundPort, model.SocketAddressProtocolTCP).
		WithOverwriteName(listenerName).
		Configure(envoy_listeners.StatPrefix(statPrefix)).
		Configure(envoy_listeners.OriginalDstForwarder())

	if useStrictInboundPorts && len(proxy.Dataplane.Spec.Networking.Inbound) > 0 {
		for _, inbound := range proxy.Dataplane.Spec.Networking.Inbound {
			// if service doesn't have any port we don't need to expose listener
			if inbound.Port == mesh_proto.TCPPortReserved {
				inboundListenerBuilder.Configure(envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, envoy_common.AnonymousResource).
					Configure(envoy_listeners.TcpProxyDeprecated(listenerName, envoy_common.NewCluster(envoy_common.WithService(naming.ContextualNoDestinationClusterName("inbound"))))).
					Configure(envoy_listeners.MatchDestiantionPort(inbound.Port))))
			} else {
				inboundListenerBuilder.Configure(envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, envoy_common.AnonymousResource).
					Configure(envoy_listeners.TcpProxyDeprecated(listenerName, envoy_common.NewCluster(envoy_common.WithService(listenerName)))).
					Configure(envoy_listeners.MatchDestiantionPort(inbound.Port))))
			}
		}
	} else {
		inboundListenerBuilder.Configure(envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, envoy_common.AnonymousResource).
			Configure(envoy_listeners.TcpProxyDeprecated(listenerName, envoy_common.NewCluster(envoy_common.WithService(listenerName))))))
	}

	listener, err := inboundListenerBuilder.Build()
	if err != nil {
		return nil, err
	}
	return listener, nil
}

func (TransparentProxyGenerator) generate(
	ctx xds_context.Context,
	proxy *model.Proxy,
	outboundName string,
	inboundName string,
	allIP string,
	inPassThroughIP string,
	unifiedNaming bool,
) (*model.ResourceSet, error) {
	resources := model.NewResourceSet()
	tpCfg := proxy.GetTransparentProxy()

	var outboundPassThroughCluster envoy_common.NamedResource
	var err error

	if ctx.Mesh.Resource.Spec.IsPassthrough() {
		outboundPassThroughCluster, err = CreateOutboundPassThroughCluster(proxy.APIVersion, outboundName)
		if err != nil {
			return nil, err
		}
	}

	outboundStatPrefix := ""
	inboundStatPrefix := ""
	if unifiedNaming {
		outboundStatPrefix = outboundName
		inboundStatPrefix = inboundName
	}

	outboundListener, err := CreateOutboundPassthroughListener(proxy, ctx, outboundName, allIP, tpCfg.Redirect.Outbound.Port.Uint32(), outboundStatPrefix)
	if err != nil {
		return nil, err
	}

	inboundPassThroughCluster, err := CreateInboundPassThroughCluster(proxy.APIVersion, inboundName, inPassThroughIP)
	if err != nil {
		return nil, err
	}

	useStrictInboundPorts := proxy.Dataplane != nil &&
		proxy.Metadata.HasFeature(xds_types.FeatureStrictInboundPorts) &&
		proxy.Dataplane.Spec.Networking != nil &&
		ctx.Mesh.Resource.MTLSEnabled() &&
		ctx.Mesh.Resource.GetEnabledCertificateAuthorityBackend().Mode == mesh_proto.CertificateAuthorityBackend_STRICT

	inboundListener, err := CreateInboundPassthroughListener(proxy, inboundName, allIP, tpCfg.Redirect.Inbound.Port.Uint32(), useStrictInboundPorts, inboundStatPrefix)
	if err != nil {
		return nil, err
	}

	resources.Add(&model.Resource{
		Name:     outboundListener.GetName(),
		Origin:   metadata.OriginTransparent,
		Resource: outboundListener,
	})

	if ctx.Mesh.Resource.Spec.IsPassthrough() {
		resources.Add(&model.Resource{
			Name:     outboundPassThroughCluster.GetName(),
			Origin:   metadata.OriginTransparent,
			Resource: outboundPassThroughCluster,
		})
	}
	resources.Add(&model.Resource{
		Name:     inboundListener.GetName(),
		Origin:   metadata.OriginTransparent,
		Resource: inboundListener,
	})
	resources.Add(&model.Resource{
		Name:     inboundPassThroughCluster.GetName(),
		Origin:   metadata.OriginTransparent,
		Resource: inboundPassThroughCluster,
	})
	return resources, nil
}

func (tpg TransparentProxyGenerator) generateIPv4(ctx xds_context.Context, proxy *model.Proxy) (*model.ResourceSet, error) {
	unifiedNaming := unified_naming.Enabled(proxy.Metadata, ctx.Mesh.Resource)
	nameOrDefault := naming.GetNameOrFallbackFunc(unifiedNaming)
	return tpg.generate(
		ctx,
		proxy,
		nameOrDefault(naming.ContextualTransparentProxyName("outbound", 4), metadata.TransparentOutboundNameIPv4),
		nameOrDefault(naming.ContextualTransparentProxyName("inbound", 4), metadata.TransparentInboundNameIPv4),
		metadata.TransparentAllIPv4,
		metadata.TransparentInPassThroughIPv4,
		unifiedNaming,
	)
}

func (tpg TransparentProxyGenerator) generateIPv6(ctx xds_context.Context, proxy *model.Proxy) (*model.ResourceSet, error) {
	unifiedNaming := unified_naming.Enabled(proxy.Metadata, ctx.Mesh.Resource)
	nameOrDefault := naming.GetNameOrFallbackFunc(unifiedNaming)
	return tpg.generate(
		ctx,
		proxy,
		nameOrDefault(naming.ContextualTransparentProxyName("outbound", 6), metadata.TransparentOutboundNameIPv6),
		nameOrDefault(naming.ContextualTransparentProxyName("inbound", 6), metadata.TransparentInboundNameIPv6),
		metadata.TransparentAllIPv6,
		metadata.TransparentInPassThroughIPv6,
		unifiedNaming,
	)
}
