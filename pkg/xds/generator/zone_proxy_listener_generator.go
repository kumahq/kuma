package generator

import (
	"context"

	envoy_tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core"
	"github.com/kumahq/kuma/v2/pkg/core/kri"
	core_meta "github.com/kumahq/kuma/v2/pkg/core/metadata"
	"github.com/kumahq/kuma/v2/pkg/core/naming"
	core_resources "github.com/kumahq/kuma/v2/pkg/core/resources/apis/core"
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
// embedded in a regular Dataplane resource.
// It is a no-op when the dataplane has no zone proxy listeners.
// Only MeshExternalService and MeshIdentity are supported; legacy ExternalService
// and mesh.mtls are not handled here.
type ZoneProxyListenerGenerator struct{}

func (g ZoneProxyListenerGenerator) Generate(
	_ context.Context,
	_ *core_xds.ResourceSet,
	xdsCtx xds_context.Context,
	proxy *core_xds.Proxy,
) (*core_xds.ResourceSet, error) {
	if !proxy.Dataplane.Spec.GetNetworking().HasZoneProxyListeners() {
		return nil, nil
	}

	if xdsCtx.Mesh.Resource.Spec.MeshServicesMode() != mesh_proto.Mesh_MeshServices_Exclusive {
		zoneProxyLog.Info("skipping zone proxy listeners: MeshServices must be in Exclusive mode",
			"mesh", xdsCtx.Mesh.Resource.GetMeta().GetName(),
		)
		return nil, nil
	}

	rs := core_xds.NewResourceSet()

	localResources := xds_context.Resources{MeshLocalResources: xdsCtx.Mesh.Resources.MeshLocalResources}

	for _, l := range proxy.Dataplane.Spec.GetNetworking().GetListeners() {
		switch l.Type {
		case mesh_proto.Dataplane_Networking_Listener_ZoneIngress:
			generated, err := g.generateIngressListener(proxy, xdsCtx, l)
			if err != nil {
				return nil, err
			}
			rs.AddSet(generated)
		case mesh_proto.Dataplane_Networking_Listener_ZoneEgress:
			if proxy.WorkloadIdentity == nil {
				zoneProxyLog.Info("skipping zone egress listener: WorkloadIdentity is required for egress mTLS",
					"mesh", xdsCtx.Mesh.Resource.GetMeta().GetName(),
				)
				continue
			}
			generated, err := g.generateEgressListener(
				proxy,
				l,
				localResources.MeshExternalServices().GetDestinations(),
				xdsCtx.Mesh.DataplaneZoneEgressEndpointMap,
			)
			if err != nil {
				return nil, err
			}
			rs.AddSet(generated)
		}
	}

	return rs, nil
}

func (g ZoneProxyListenerGenerator) generateIngressListener(
	proxy *core_xds.Proxy,
	xdsCtx xds_context.Context,
	listener *mesh_proto.Dataplane_Networking_Listener,
) (*core_xds.ResourceSet, error) {
	rs := core_xds.NewResourceSet()
	cp := xdsCtx.ControlPlane
	meshName := xdsCtx.Mesh.Resource.GetMeta().GetName()

	address := listener.Address
	port := listener.Port

	listenerName := naming.ContextualZoneIngressListenerName(listener.Name)

	listenerBuilder := envoy_listeners.NewListenerBuilder(proxy.APIVersion, listenerName).
		Configure(envoy_listeners.InboundListener(address, port, core_xds.SocketAddressProtocolTCP)).
		Configure(envoy_listeners.StatPrefix(listenerName)).
		Configure(envoy_listeners.TLSInspector())

	meshResources := xds_context.Resources{MeshLocalResources: xdsCtx.Mesh.Resources.MeshLocalResources}

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

	services := zoneproxy.GetServices(dest, xdsCtx.Mesh.DataplaneZoneIngressEndpointMap, nil, true)
	clusters := services.Clusters()

	cds, err := zoneproxy.GenerateCDS(proxy, dest, services, meshName, metadata.OriginIngress, true)
	if err != nil {
		return nil, err
	}
	rs.AddSet(cds)

	eds, err := zoneproxy.GenerateEDS(proxy, xdsCtx.Mesh.DataplaneZoneIngressEndpointMap, services, meshName, metadata.OriginIngress, true)
	if err != nil {
		return nil, err
	}
	rs.AddSet(eds)

	for _, cluster := range clusters {
		listenerBuilder.Configure(envoy_listeners.FilterChain(zoneproxy.CreateFilterChain(proxy, cluster)))
	}

	if len(clusters) == 0 {
		return nil, nil
	}

	resource, err := listenerBuilder.Build()
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
	listener *mesh_proto.Dataplane_Networking_Listener,
	destinations []core_resources.Destination,
	endpointMap core_xds.EgressEndpointMap,
) (*core_xds.ResourceSet, error) {
	if len(destinations) == 0 {
		return nil, nil
	}

	rs := core_xds.NewResourceSet()

	address := listener.Address
	port := listener.Port

	zoneEgressListenerName := naming.ContextualZoneEgressListenerName(listener.Name)

	listenerBuilder := envoy_listeners.NewListenerBuilder(proxy.APIVersion, zoneEgressListenerName).
		Configure(envoy_listeners.InboundListener(address, port, core_xds.SocketAddressProtocolTCP)).
		Configure(envoy_listeners.StatPrefix(zoneEgressListenerName)).
		Configure(envoy_listeners.TLSInspector())

	downstreamTLS, err := meshIdentityDownstreamTLS(proxy)
	if err != nil {
		return nil, err
	}

	addedFilterChains := 0
	for _, dst := range destinations {
		esPort := dst.GetPorts()[0]
		id := kri.WithSectionName(kri.From(dst), esPort.GetName())
		sni := tls.SNIForResource(id.Name, id.Mesh, id.ResourceType, esPort.GetValue(), nil)
		clusterName := id.String()
		group := endpointMap[clusterName]

		if len(group.Endpoints) == 0 {
			continue
		}

		split := plugins_xds.NewSplitBuilder().
			WithClusterName(clusterName).
			WithExternalService(true).
			WithWeight(1).
			Build()

		listenerBuilder.Configure(envoy_listeners.FilterChain(g.buildEgressFilterChain(proxy, group.Protocol, downstreamTLS, split, sni)))
		cds, err := g.genClusterCDS(proxy, group, clusterName)
		if err != nil {
			return nil, err
		}
		rs.Add(cds)
		addedFilterChains++
	}

	if addedFilterChains == 0 {
		return nil, nil
	}

	resource, err := listenerBuilder.Build()
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
	group core_xds.EgressEndpointGroup,
	clusterName string,
) (*core_xds.Resource, error) {
	ipv6 := proxy.Dataplane.IsIPv6()
	systemCAPath := proxy.Metadata.GetSystemCaPath()

	resource, err := envoy_clusters.NewClusterBuilder(proxy.APIVersion, clusterName).
		Configure(envoy_clusters.DefaultTimeout()).
		ConfigureIf(core_meta.IsHTTP(group.Protocol), envoy_clusters.Http()).
		ConfigureIf(core_meta.IsHTTP2Based(group.Protocol), envoy_clusters.Http2()).
		Configure(envoy_clusters.ProvidedCustomEndpointCluster(ipv6, true, group.Endpoints...)).
		Configure(envoy_clusters.MeshExternalServiceClientSideTLS(group.Endpoints, systemCAPath, true)).
		Build()
	if err != nil {
		return nil, err
	}
	return &core_xds.Resource{
		Name:           resource.GetName(),
		Origin:         metadata.OriginEgress,
		Resource:       resource,
		Protocol:       group.Protocol,
		ResourceOrigin: group.OwnerResource,
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
