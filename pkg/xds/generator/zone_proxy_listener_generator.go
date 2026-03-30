package generator

import (
	"context"

	envoy_tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core"
	"github.com/kumahq/kuma/v2/pkg/core/kri"
	core_meta "github.com/kumahq/kuma/v2/pkg/core/metadata"
	"github.com/kumahq/kuma/v2/pkg/core/naming"
	meshservice_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
	bldrs_common "github.com/kumahq/kuma/v2/pkg/envoy/builders/common"
	bldrs_core "github.com/kumahq/kuma/v2/pkg/envoy/builders/core"
	bldrs_tls "github.com/kumahq/kuma/v2/pkg/envoy/builders/tls"
	plugins_xds "github.com/kumahq/kuma/v2/pkg/plugins/policies/core/xds"
	meshhttproute_api "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/meshhttproute/xds"
	xds_context "github.com/kumahq/kuma/v2/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/v2/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/v2/pkg/xds/envoy/clusters"
	envoy_listeners "github.com/kumahq/kuma/v2/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/v2/pkg/xds/envoy/tls"
	"github.com/kumahq/kuma/v2/pkg/xds/generator/metadata"
	system_names "github.com/kumahq/kuma/v2/pkg/xds/generator/system_names"
	"github.com/kumahq/kuma/v2/pkg/xds/generator/zoneproxy"
)

var zoneProxyLog = core.Log.WithName("xds").WithName("zone-proxy-listener-generator")

// ZoneProxyListenerGenerator generates Envoy listeners for zone proxy listeners
// embedded in a regular Dataplane resource (DataplaneZoneListeners).
// It is a no-op when proxy.DataplaneZoneListeners is nil.
// Only MeshExternalService and MeshIdentity are supported; legacy ExternalService
// and mesh.mtls are not handled here.
type ZoneProxyListenerGenerator struct{}

func (g ZoneProxyListenerGenerator) Generate(
	_ context.Context,
	_ *core_xds.ResourceSet,
	xdsCtx xds_context.Context,
	proxy *core_xds.Proxy,
) (*core_xds.ResourceSet, error) {
	if proxy.DataplaneZoneListeners == nil {
		return nil, nil
	}

	if xdsCtx.Mesh.Resource.Spec.MeshServicesMode() != mesh_proto.Mesh_MeshServices_Exclusive {
		zoneProxyLog.Info("skipping zone proxy listeners: MeshServices must be in Exclusive mode",
			"mesh", xdsCtx.Mesh.Resource.GetMeta().GetName(),
		)
		return nil, nil
	}

	rs := core_xds.NewResourceSet()

	for _, il := range proxy.DataplaneZoneListeners.IngressListeners {
		generated, err := g.generateIngressListener(proxy, xdsCtx, il)
		if err != nil {
			return nil, err
		}
		rs.AddSet(generated)
	}

	if proxy.WorkloadIdentity == nil {
		zoneProxyLog.Info("skipping zone egress listeners: WorkloadIdentity is required for egress mTLS",
			"mesh", xdsCtx.Mesh.Resource.GetMeta().GetName(),
		)
		return rs, nil
	}

	for _, el := range proxy.DataplaneZoneListeners.EgressListeners {
		generated, err := g.generateEgressListener(proxy, el)
		if err != nil {
			return nil, err
		}
		rs.AddSet(generated)
	}

	return rs, nil
}

func (g ZoneProxyListenerGenerator) generateIngressListener(
	proxy *core_xds.Proxy,
	xdsCtx xds_context.Context,
	il *core_xds.DataplaneListener,
) (*core_xds.ResourceSet, error) {
	rs := core_xds.NewResourceSet()
	cp := xdsCtx.ControlPlane
	mr := il.MeshResources
	meshName := mr.Mesh.GetMeta().GetName()

	address := il.Listener.Address
	port := il.Listener.Port

	listenerName := naming.ContextualZoneIngressListenerName(il.Listener.Name)

	listener := envoy_listeners.NewListenerBuilder(proxy.APIVersion, listenerName).
		Configure(envoy_listeners.InboundListener(address, port, core_xds.SocketAddressProtocolTCP)).
		Configure(envoy_listeners.StatPrefix(listenerName)).
		Configure(envoy_listeners.TLSInspector())

	meshResources := xds_context.Resources{MeshLocalResources: mr.Resources}

	// Only expose services local to this zone.
	localMS := &meshservice_api.MeshServiceResourceList{}
	for _, ms := range meshResources.MeshServices().GetItems() {
		if lbls := ms.GetMeta().GetLabels(); lbls == nil || lbls[mesh_proto.ZoneTag] == "" || lbls[mesh_proto.ZoneTag] == cp.Zone {
			_ = localMS.AddItem(ms)
		}
	}

	dest := zoneproxy.BuildMeshDestinations(
		nil, // no legacy available services
		cp.SystemNamespace,
		meshResources,
		localMS,
		meshResources.MeshMultiZoneServices(),
	)

	services := zoneproxy.GetServices(dest, mr.EndpointMap, nil, true)
	clusters := services.Clusters()

	cds, err := zoneproxy.GenerateCDS(proxy, dest, services, meshName, metadata.OriginIngress, true)
	if err != nil {
		return nil, err
	}
	rs.AddSet(cds)

	eds, err := zoneproxy.GenerateEDS(proxy, mr.EndpointMap, services, meshName, metadata.OriginIngress, true)
	if err != nil {
		return nil, err
	}
	rs.AddSet(eds)

	for _, cluster := range clusters {
		listener.Configure(envoy_listeners.FilterChain(zoneproxy.CreateFilterChain(proxy, cluster)))
	}

	if len(clusters) == 0 {
		return nil, nil
	}

	resource, err := listener.Build()
	if err != nil {
		return nil, err
	}
	rs.Add(&core_xds.Resource{
		Name:     resource.GetName(),
		Origin:   metadata.OriginIngress,
		Resource: resource,
	})
	return rs, nil
}

func (g ZoneProxyListenerGenerator) generateEgressListener(
	proxy *core_xds.Proxy,
	el *core_xds.DataplaneListener,
) (*core_xds.ResourceSet, error) {
	localResources := xds_context.Resources{MeshLocalResources: el.MeshResources.Resources}

	esDestinations := localResources.MeshExternalServices().GetDestinations()
	if len(esDestinations) == 0 {
		return nil, nil
	}

	rs := core_xds.NewResourceSet()

	address := el.Listener.Address
	port := el.Listener.Port

	zoneEgressListenerName := naming.ContextualZoneEgressListenerName(el.Listener.Name)

	listener := envoy_listeners.NewListenerBuilder(proxy.APIVersion, zoneEgressListenerName).
		Configure(envoy_listeners.InboundListener(address, port, core_xds.SocketAddressProtocolTCP)).
		Configure(envoy_listeners.StatPrefix(zoneEgressListenerName)).
		Configure(envoy_listeners.TLSInspector())

	downstreamTLS, err := meshIdentityDownstreamTLS(proxy)
	if err != nil {
		return nil, err
	}

	for _, dst := range esDestinations {
		esPort := dst.GetPorts()[0]
		id := kri.WithSectionName(kri.From(dst), esPort.GetName())
		sni := tls.SNIForResource(id.Name, id.Mesh, id.ResourceType, esPort.GetValue(), nil)
		clusterName := id.String()
		endpoints := el.MeshResources.EndpointMap[clusterName]

		split := plugins_xds.NewSplitBuilder().
			WithClusterName(clusterName).
			WithExternalService(true).
			WithWeight(1).
			Build()

		listener.Configure(envoy_listeners.FilterChain(g.buildEgressFilterChain(proxy, endpoints[0].Protocol(), downstreamTLS, split, sni)))
		cds, err := g.genClusterCDS(proxy, endpoints, clusterName)
		if err != nil {
			return nil, err
		}
		rs.Add(cds)
	}

	resource, err := listener.Build()
	if err != nil {
		return nil, err
	}
	rs.Add(&core_xds.Resource{
		Name:     resource.GetName(),
		Origin:   metadata.OriginEgress,
		Resource: resource,
	})

	return rs, nil
}

func (g ZoneProxyListenerGenerator) genClusterCDS(
	proxy *core_xds.Proxy,
	endpoints []core_xds.Endpoint,
	clusterName string,
) (*core_xds.Resource, error) {
	// This should never happen as callers should always provide at least one endpoint.
	// If it does happen, it indicates a bug in the calling code.
	if len(endpoints) == 0 {
		return nil, errors.New("endpoints cannot be empty")
	}

	ipv6 := proxy.Dataplane.IsIPv6()
	systemCAPath := proxy.Metadata.GetSystemCaPath()
	protocol := endpoints[0].Protocol()

	resource, err := envoy_clusters.NewClusterBuilder(proxy.APIVersion, clusterName).
		Configure(envoy_clusters.DefaultTimeout()).
		ConfigureIf(core_meta.IsHTTP(protocol), envoy_clusters.Http()).
		ConfigureIf(core_meta.IsHTTP2Based(protocol), envoy_clusters.Http2()).
		Configure(envoy_clusters.ProvidedCustomEndpointCluster(ipv6, true, endpoints...)).
		Configure(envoy_clusters.MeshExternalServiceClientSideTLS(endpoints, systemCAPath, true)).
		Build()
	if err != nil {
		return nil, err
	}
	return &core_xds.Resource{
		Name:           resource.GetName(),
		Origin:         metadata.OriginEgress,
		Resource:       resource,
		Protocol:       endpoints[0].ExternalService.Protocol,
		ResourceOrigin: endpoints[0].ExternalService.OwnerResource,
	}, nil
}

func (g ZoneProxyListenerGenerator) buildEgressFilterChain(
	proxy *core_xds.Proxy,
	protocol core_meta.Protocol,
	downstreamTLS *envoy_tls.DownstreamTlsContext,
	split envoy_common.Split,
	sni string,
) *envoy_listeners.FilterChainBuilder {
	esName := split.ClusterName()

	filterChain := envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, esName).
		Configure(envoy_listeners.MatchTransportProtocol(core_meta.ProtocolTLS)).
		Configure(envoy_listeners.MatchServerNames(sni)).
		Configure(envoy_listeners.DownstreamTlsContext(downstreamTLS))

	if !core_meta.IsHTTPBased(protocol) {
		return filterChain.Configure(envoy_listeners.TCPProxy(esName, split))
	}

	return filterChain.
		Configure(envoy_listeners.HttpConnectionManager(esName, false, proxy.InternalAddresses, proxy.Metadata.GetIPv6Enabled())).
		Configure(envoy_listeners.AddFilterChainConfigurer(&xds.HttpOutboundRouteConfigurer{
			RouteConfigName: esName,
			VirtualHostName: esName,
			Routes: []xds.OutboundRoute{
				{
					Match: meshhttproute_api.Match{
						Path: &meshhttproute_api.PathMatch{
							Type:  meshhttproute_api.PathPrefix,
							Value: "/",
						},
					},
					Split: []envoy_common.Split{split},
				},
			},
		}))
}

// meshIdentityDownstreamTLS builds a DownstreamTlsContext from the proxy's WorkloadIdentity.
// Returns nil when WorkloadIdentity is nil.
func meshIdentityDownstreamTLS(proxy *core_xds.Proxy) (*envoy_tls.DownstreamTlsContext, error) {
	if proxy.WorkloadIdentity == nil {
		return nil, nil
	}

	validationCtx := func() bldrs_common.Configurer[envoy_tls.SdsSecretConfig] {
		return bldrs_tls.SdsSecretConfigSource(
			system_names.SystemResourceNameCABundle,
			bldrs_core.NewConfigSource().Configure(bldrs_core.Sds()),
		)
	}
	if proxy.WorkloadIdentity.ExternalValidationSourceConfigurer != nil {
		ext := proxy.WorkloadIdentity.ExternalValidationSourceConfigurer
		validationCtx = func() bldrs_common.Configurer[envoy_tls.SdsSecretConfig] {
			return ext()
		}
	}

	return bldrs_tls.NewDownstreamTLSContext().
		Configure(
			bldrs_tls.DownstreamCommonTlsContext(
				bldrs_tls.NewCommonTlsContext().
					Configure(
						bldrs_tls.CombinedCertificateValidationContext(
							bldrs_tls.NewCombinedCertificateValidationContext().
								Configure(
									bldrs_tls.DefaultValidationContext(
										bldrs_tls.NewDefaultValidationContext(),
									),
								).
								Configure(
									bldrs_tls.ValidationContextSdsSecretConfig(
										bldrs_tls.NewTlsCertificateSdsSecretConfigs().Configure(validationCtx()),
									),
								),
						),
					).
					Configure(
						bldrs_tls.TlsCertificateSdsSecretConfigs([]*bldrs_common.Builder[envoy_tls.SdsSecretConfig]{
							bldrs_tls.NewTlsCertificateSdsSecretConfigs().Configure(
								proxy.WorkloadIdentity.IdentitySourceConfigurer(),
							),
						}),
					),
			),
		).
		Configure(bldrs_tls.RequireClientCertificate(true)).
		Build()
}
