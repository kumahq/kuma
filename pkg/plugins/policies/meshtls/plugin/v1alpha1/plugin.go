package v1alpha1

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/spiffe/go-spiffe/v2/spiffeid"

	common_tls "github.com/kumahq/kuma/v3/api/common/v1alpha1/tls"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core"
	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	"github.com/kumahq/kuma/v3/pkg/core/naming"
	core_plugins "github.com/kumahq/kuma/v3/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/v3/pkg/core/xds/types"
	bldrs_common "github.com/kumahq/kuma/v3/pkg/envoy/builders/common"
	bldrs_core "github.com/kumahq/kuma/v3/pkg/envoy/builders/core"
	bldrs_matcher "github.com/kumahq/kuma/v3/pkg/envoy/builders/matcher"
	bldrs_tls "github.com/kumahq/kuma/v3/pkg/envoy/builders/tls"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/matchers"
	core_rules "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules"
	rules_inbound "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/inbound"
	policies_xds "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/xds"
	api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtls/api/v1alpha1"
	util_maps "github.com/kumahq/kuma/v3/pkg/util/maps"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
	"github.com/kumahq/kuma/v3/pkg/util/proto"
	util_slices "github.com/kumahq/kuma/v3/pkg/util/slices"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
	envoy_listeners "github.com/kumahq/kuma/v3/pkg/xds/envoy/listeners"
	xds_tls "github.com/kumahq/kuma/v3/pkg/xds/envoy/tls"
	"github.com/kumahq/kuma/v3/pkg/xds/generator"
	"github.com/kumahq/kuma/v3/pkg/xds/generator/metadata"
	"github.com/kumahq/kuma/v3/pkg/xds/generator/system_names"
)

var logger = core.Log.WithName("MeshTLS")

var _ core_plugins.PolicyPlugin = &plugin{}

type plugin struct{}

func (p plugin) Order() int { return api.MeshTLSResourceTypeDescriptor.Order }

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(
	dataplane *core_mesh.DataplaneResource,
	resources xds_context.Resources,
	opts ...core_plugins.MatchedPoliciesOption,
) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshTLSType, dataplane, resources, opts...)
}

func (p plugin) Apply(rs *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error {
	if proxy.Dataplane == nil {
		return nil
	}

	log := logger.WithValues(
		"proxyName", proxy.Dataplane.GetMeta().GetName(),
		"mesh", ctx.Mesh.Resource.GetMeta().GetName(),
	)

	switch {
	case ctx.Mesh.Resource.MTLSEnabled():
	case proxy.WorkloadIdentity != nil:
	default:
		log.V(1).Info("skip applying MeshTLS, mTLS is disabled")
		return nil
	}

	log.V(1).Info("applying")

	policies := proxy.Policies.Dynamic[api.MeshTLSType]

	listeners := policies_xds.GatherListeners(rs)
	clusters := policies_xds.GatherClusters(rs)

	if err := applyToInbounds(rs, policies.FromRules, listeners.Inbound, proxy, ctx); err != nil {
		return err
	}
	if err := applyToGateways(policies.GatewayRules, clusters.Gateway, ctx); err != nil {
		return err
	}
	if err := applyToRealResources(policies.FromRules, rs); err != nil {
		return err
	}

	return nil
}

func applyToInbounds(
	rs *core_xds.ResourceSet,
	fromRules core_rules.FromRules,
	inboundListeners map[core_rules.InboundListener]*envoy_listener.Listener,
	proxy *core_xds.Proxy,
	ctx xds_context.Context,
) error {
	for _, inbound := range proxy.Dataplane.Spec.GetNetworking().GetInbound() {
		iface := proxy.Dataplane.Spec.Networking.ToInboundInterface(inbound)

		listenerKey := core_rules.InboundListener{
			Address: iface.DataplaneIP,
			Port:    iface.DataplanePort,
		}

		listener, ok := inboundListeners[listenerKey]
		if !ok {
			continue
		}

		conf := rules_inbound.MatchesAllIncomingTraffic[api.Conf](fromRules.InboundRules[listenerKey])

		if resource, err := configureListener(proxy, ctx, iface, inbound, conf); err != nil {
			return err
		} else if resource != nil {
			rs.Remove(envoy_resource.ListenerType, listener.GetName())
			rs.Add(&core_xds.Resource{
				Name:     resource.GetName(),
				Origin:   metadata.OriginInbound,
				Resource: resource,
			})
		}

		if resource, err := configureInboundPassthroughListener(proxy, ctx, conf, false); err != nil {
			return err
		} else if resource != nil {
			rs.Remove(envoy_resource.ListenerType, resource.GetName())
			rs.Add(&core_xds.Resource{
				Name:     resource.GetName(),
				Origin:   metadata.OriginTransparent,
				Resource: resource,
			})
		}

		if resource, err := configureInboundPassthroughListener(proxy, ctx, conf, true); err != nil {
			return err
		} else if resource != nil {
			rs.Remove(envoy_resource.ListenerType, resource.GetName())
			rs.Add(&core_xds.Resource{
				Name:     resource.GetName(),
				Origin:   metadata.OriginTransparent,
				Resource: resource,
			})
		}
	}

	return nil
}

func applyToGateways(
	gatewayRules core_rules.GatewayRules,
	gatewayClusters map[string]*envoy_cluster.Cluster,
	ctx xds_context.Context,
) error {
	for serviceName, cluster := range gatewayClusters {
		// we shouldn't modify ExternalService
		// MeshExternalService has different origin
		if ctx.Mesh.IsExternalService(serviceName) {
			continue
		}
		// there is only one rule always because we're in `Mesh/Mesh`
		var conf api.Conf
		for _, r := range gatewayRules.InboundRules {
			conf = rules_inbound.MatchesAllIncomingTraffic[api.Conf](r)
			break
		}

		if err := configureTLSParams(conf, cluster); err != nil {
			return err
		}
	}
	return nil
}

func applyToRealResources(
	fromRules core_rules.FromRules,
	rs *core_xds.ResourceSet,
) error {
	for _, resType := range rs.IndexByOrigin(core_xds.NonMeshExternalService, core_xds.NonGatewayResources) {
		// there is only one rule always because we're in `Mesh/Mesh`
		var conf api.Conf
		for _, r := range fromRules.InboundRules {
			conf = rules_inbound.MatchesAllIncomingTraffic[api.Conf](r)
			break
		}

		for _, cluster := range resType[envoy_resource.ClusterType] {
			if err := configureTLSParams(conf, cluster.Resource.(*envoy_cluster.Cluster)); err != nil {
				return err
			}
		}
	}

	return nil
}

func configureTLSParams(conf api.Conf, cluster *envoy_cluster.Cluster) error {
	if cluster.TransportSocket.GetName() != wellknown.TransportSocketTLS {
		// we only want to configure TLS Version on listeners protected by Kuma's TLS
		return nil
	}

	var dst envoy_tls.UpstreamTlsContext
	if err := proto.UnmarshalAnyTo(cluster.TransportSocket.GetTypedConfig(), &dst); err != nil {
		return err
	}

	version := pointer.Deref(conf.TlsVersion)
	ciphers := pointer.Deref(conf.TlsCiphers)

	if len(ciphers) > 0 || version.Min != nil || version.Max != nil {
		dst.CommonTlsContext.TlsParams = &envoy_tls.TlsParameters{}
	}

	if len(ciphers) > 0 {
		dst.CommonTlsContext.TlsParams.CipherSuites = util_slices.Map(ciphers, common_tls.TlsCipher.String)
	}

	if version.Min != nil {
		dst.CommonTlsContext.TlsParams.TlsMinimumProtocolVersion = common_tls.ToTlsVersion(version.Min)
	}

	if version.Max != nil {
		dst.CommonTlsContext.TlsParams.TlsMaximumProtocolVersion = common_tls.ToTlsVersion(version.Max)
	}

	logger.V(1).Info("computed outbound tls params", "params", dst.CommonTlsContext.TlsParams)

	pbst, err := proto.MarshalAnyDeterministic(&dst)
	if err != nil {
		return err
	}

	cluster.TransportSocket = &envoy_core.TransportSocket{
		Name: "envoy.transport_sockets.tls",
		ConfigType: &envoy_core.TransportSocket_TypedConfig{
			TypedConfig: pbst,
		},
	}

	return nil
}

func configureInboundPassthroughListener(
	proxy *core_xds.Proxy,
	xdsCtx xds_context.Context,
	conf api.Conf,
	ipv6 bool,
) (envoy_common.NamedResource, error) {
	// mirror the conditions under which TransparentProxyGenerator emits the
	// passthrough listener, otherwise we replace a listener that doesn't exist
	// and point it at a cluster that was never generated
	tpCfg := proxy.GetTransparentProxy()
	if tpCfg == nil || !tpCfg.Redirect.Outbound.Enabled || proxy.Metadata.HasFeature(xds_types.FeatureBindOutbounds) {
		return nil, nil
	}
	if ipv6 && !tpCfg.EnabledIPv6() {
		return nil, nil
	}
	if xdsCtx.Mesh.Resource.GetEnabledCertificateAuthorityBackend() == nil &&
		proxy.WorkloadIdentity == nil &&
		!proxy.Metadata.HasFeature(xds_types.FeatureStrictInboundPorts) {
		return nil, nil
	}
	tlsMode := getMeshTLSMode(conf.Mode)
	address := metadata.TransparentAllIPv4
	inboundName := naming.ContextualTransparentProxyName("inbound", 4)
	if ipv6 {
		inboundName = naming.ContextualTransparentProxyName("inbound", 6)
		address = metadata.TransparentAllIPv6
	}
	statPrefix := inboundName
	switch tlsMode {
	case api.ModeStrict, api.ModePermissive:
		// only a sidecar that opted into strict inbound ports may restrict the
		// passthrough to the declared inbounds, and only in strict mode
		useStrictInboundPorts := tlsMode == api.ModeStrict &&
			proxy.Metadata.HasFeature(xds_types.FeatureStrictInboundPorts)
		return generator.CreateInboundPassthroughListener(
			proxy,
			inboundName,
			address,
			tpCfg.Redirect.Inbound.Port.Uint32(),
			useStrictInboundPorts,
			statPrefix,
		)
	}
	return nil, nil
}

func configureListener(
	proxy *core_xds.Proxy,
	xdsCtx xds_context.Context,
	iface mesh_proto.InboundInterface,
	inbound *mesh_proto.Dataplane_Networking_Inbound,
	conf api.Conf,
) (envoy_common.NamedResource, error) {
	inboundContextualID := naming.MustContextualInboundName(proxy.Dataplane, iface.InboundName)

	listener := envoy_listeners.NewListenerBuilder(proxy.APIVersion, inboundContextualID).
		Configure(envoy_listeners.InboundListener(iface.DataplaneIP, iface.DataplanePort, core_xds.SocketAddressProtocolTCP, proxy.Metadata.HasFeature(xds_types.FeatureReusePort))).
		Configure(envoy_listeners.StatPrefix(inboundContextualID)).
		Configure(envoy_listeners.TransparentProxying(proxy)).
		Configure(envoy_listeners.TagsMetadata(generator.InboundListenerTags(proxy.Dataplane, inboundContextualID)))

	downstreamCtx, err := downstreamTLSContext(xdsCtx, proxy, conf)
	if err != nil {
		return nil, err
	}

	protocol := core_meta.ParseProtocol(inbound.GetProtocol())

	filterChainKumaTLS := generator.FilterChainBuilder(protocol, proxy, iface).
		Configure(envoy_listeners.ServerSideMTLS(
			xdsCtx.Mesh.Resource,
			proxy.SecretsTracker,
			conf.TlsVersion,
			pointer.Deref(conf.TlsCiphers),
			len(xdsCtx.Mesh.CAsByTrustDomain) > 0,
		)).
		Configure(envoy_listeners.DownstreamTlsContext(downstreamCtx))

	if getMeshTLSMode(conf.Mode) == api.ModeStrict {
		return listener.Configure(envoy_listeners.FilterChain(filterChainKumaTLS)).Build()
	}

	filterChainRawBuffer := generator.FilterChainBuilder(protocol, proxy, iface).
		Configure(envoy_listeners.MatchTransportProtocol(core_meta.ProtocolRawBuffer))

	// we need to differentiate between just TLS and Kuma's TLS, because with permissive mode the app
	// itself might be protected by TLS
	filterChainTLS := generator.FilterChainBuilder(protocol, proxy, iface).
		Configure(envoy_listeners.MatchTransportProtocol(core_meta.ProtocolTLS))

	filterChainKumaTLS.
		Configure(envoy_listeners.MatchTransportProtocol(core_meta.ProtocolTLS)).
		Configure(envoy_listeners.MatchApplicationProtocols(xds_tls.KumaALPNProtocols...))

	return listener.
		Configure(envoy_listeners.TLSInspector()).
		Configure(envoy_listeners.FilterChain(filterChainRawBuffer)).
		Configure(envoy_listeners.FilterChain(filterChainTLS)).
		Configure(envoy_listeners.FilterChain(filterChainKumaTLS)).
		Build()
}

func downstreamTLSContext(xdsCtx xds_context.Context, proxy *core_xds.Proxy, conf api.Conf) (*envoy_tls.DownstreamTlsContext, error) {
	if proxy.WorkloadIdentity == nil {
		return nil, nil
	}

	var sanMatchers []*bldrs_common.Builder[envoy_tls.SubjectAltNameMatcher]
	// Spire delivers SANs validator and we don't support MeshTrust with spire
	// TODO: do we need this validator since we have a better validator of CA matched with TrustDomain
	// check: pkg/core/resources/apis/meshtrust/generator/v1alpha1/secrets.go
	if proxy.WorkloadIdentity.ManagementMode == core_xds.KumaManagementMode {
		for _, trustDomain := range util_maps.SortedKeys(xdsCtx.Mesh.CAsByTrustDomain) {
			id, err := spiffeid.TrustDomainFromString(trustDomain)
			if err != nil {
				return nil, err
			}
			conf := bldrs_tls.NewSubjectAltNameMatcher().Configure(bldrs_tls.URI(bldrs_matcher.NewStringMatcher().Configure(bldrs_matcher.PrefixMatcher(id.IDString() + "/"))))
			sanMatchers = append(sanMatchers, conf)
		}
	}

	validationCtx := func() bldrs_common.Configurer[envoy_tls.SdsSecretConfig] {
		return bldrs_tls.SdsSecretConfigSource(
			system_names.SystemResourceNameCABundle,
			bldrs_core.NewConfigSource().Configure(bldrs_core.Sds()),
		)
	}
	if proxy.WorkloadIdentity.ExternalValidationSourceConfigurer != nil {
		validationCtx = func() bldrs_common.Configurer[envoy_tls.SdsSecretConfig] {
			return proxy.WorkloadIdentity.ExternalValidationSourceConfigurer()
		}
	}

	tlsVersion := pointer.Deref(conf.TlsVersion)

	return bldrs_tls.NewDownstreamTLSContext().
		Configure(
			bldrs_tls.DownstreamCommonTlsContext(
				bldrs_tls.NewCommonTlsContext().
					Configure(bldrs_common.IfNotNil(conf.TlsCiphers, bldrs_tls.CipherSuites)).
					Configure(bldrs_common.If(tlsVersion.Min != nil, bldrs_tls.TlsMinVersion(tlsVersion.Min))).
					Configure(bldrs_common.If(tlsVersion.Max != nil, bldrs_tls.TlsMaxVersion(tlsVersion.Max))).
					Configure(
						bldrs_tls.CombinedCertificateValidationContext(
							bldrs_tls.NewCombinedCertificateValidationContext().
								Configure(
									bldrs_tls.DefaultValidationContext(
										bldrs_tls.NewDefaultValidationContext().
											Configure(bldrs_tls.SANs(sanMatchers)),
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

// getMeshTLSMode resolves the TLS mode of an inbound. Only a MeshTLS policy can
// select Permissive, everything else defaults to Strict.
func getMeshTLSMode(confMode *api.Mode) api.Mode {
	if confMode != nil {
		return *confMode
	}
	return api.ModeStrict
}
