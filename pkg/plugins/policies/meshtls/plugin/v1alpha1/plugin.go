package v1alpha1

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"

	common_tls "github.com/kumahq/kuma/api/common/v1alpha1/tls"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/matchers"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	policies_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshtls/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
	"github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
	xds_tls "github.com/kumahq/kuma/pkg/xds/envoy/tls"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

var (
	_   core_plugins.PolicyPlugin = &plugin{}
	log                           = core.Log.WithName("MeshTLS")
)

type plugin struct{}

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources, opts ...core_plugins.MatchedPoliciesOption) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshTLSType, dataplane, resources)
}

func (p plugin) Apply(rs *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error {
	if proxy.Dataplane == nil {
		return nil
	}
	log.V(1).Info("applying", "proxy-name", proxy.Dataplane.GetMeta().GetName())
	policies, ok := proxy.Policies.Dynamic[api.MeshTLSType]
	if !ok {
		return nil
	}
	listeners := policies_xds.GatherListeners(rs)
	clusters := policies_xds.GatherClusters(rs)

	if err := applyToInbounds(rs, policies.FromRules, listeners.Inbound, proxy, ctx); err != nil {
		return err
	}
	if err := applyToOutbounds(policies.FromRules, clusters.Outbound, clusters.OutboundSplit, proxy.Outbounds, ctx); err != nil {
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
		conf := core_rules.ComputeConf[api.Conf](fromRules.Rules[listenerKey], core_rules.MeshSubset())
		if conf == nil {
			continue
		}
		l, err := configure(proxy, listener, iface, inbound, conf, ctx)
		if err != nil {
			return err
		}
		if l != nil {
			rs.Remove(resource.ListenerType, listener.GetName())
			rs.Add(&core_xds.Resource{
				Name:     listener.Name,
				Origin:   generator.OriginInbound,
				Resource: l,
			})
		}
	}
	return nil
}

func applyToOutbounds(
	fromRules core_rules.FromRules,
	outboundClusters map[string]*envoy_cluster.Cluster,
	outboundSplitClusters map[string][]*envoy_cluster.Cluster,
	outbounds xds_types.Outbounds,
	ctx xds_context.Context,
) error {
	targetedClusters := policies_xds.GatherTargetedClusters(
		outbounds.Filter(xds_types.NonBackendRefFilter),
		outboundSplitClusters,
		outboundClusters,
	)
	for cluster, serviceName := range targetedClusters {
		// we shouldn't modify ExternalService
		// MeshExternalService has different origin
		if ctx.Mesh.IsExternalService(serviceName) {
			continue
		}
		// there is only one rule always because we're in `Mesh/Mesh`
		var conf *api.Conf
		for _, r := range fromRules.Rules {
			conf = core_rules.ComputeConf[api.Conf](r, core_rules.MeshSubset())
			break
		}
		if conf == nil {
			continue
		}
		if err := configureParams(conf, cluster); err != nil {
			return err
		}
	}

	return nil
}

func configureParams(conf *api.Conf, cluster *envoy_cluster.Cluster) error {
	if cluster.TransportSocket.GetName() != wellknown.TransportSocketTLS {
		// we only want to configure TLS Version on listeners protected by Kuma's TLS
		return nil
	}

	tlsContext := &envoy_tls.UpstreamTlsContext{CommonTlsContext: &envoy_tls.CommonTlsContext{}}

	version := conf.TlsVersion
	if version != nil {
		if version.Min != nil {
			tlsContext.CommonTlsContext.TlsParams = &envoy_tls.TlsParameters{
				TlsMinimumProtocolVersion: common_tls.ToTlsVersion(version.Min),
			}
		}
		if version.Max != nil {
			if tlsContext.CommonTlsContext.TlsParams == nil {
				tlsContext.CommonTlsContext.TlsParams = &envoy_tls.TlsParameters{
					TlsMaximumProtocolVersion: common_tls.ToTlsVersion(version.Max),
				}
			} else {
				tlsContext.CommonTlsContext.TlsParams.TlsMaximumProtocolVersion = common_tls.ToTlsVersion(version.Max)
			}
		}
	}

	ciphers := conf.TlsCiphers
	if ciphers != nil {
		if tlsContext.CommonTlsContext.TlsParams != nil {
			var cipherSuites []string
			for _, c := range ciphers {
				cipherSuites = append(cipherSuites, string(c))
			}
			tlsContext.CommonTlsContext.TlsParams.CipherSuites = cipherSuites
		}
	}

	log.V(1).Info("computed outbound tlsContext", "tlsContext", tlsContext)

	dst := envoy_tls.UpstreamTlsContext{}
	err := proto.UnmarshalAnyTo(cluster.TransportSocket.GetTypedConfig(), &dst)
	if err != nil {
		return err
	}

	// this relies on nothing before it modifying TlsParams
	dst.CommonTlsContext.TlsParams = tlsContext.CommonTlsContext.TlsParams
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

func configure(
	proxy *core_xds.Proxy,
	listener *envoy_listener.Listener,
	iface mesh_proto.InboundInterface,
	inbound *mesh_proto.Dataplane_Networking_Inbound,
	conf *api.Conf,
	xdsCtx xds_context.Context,
) (envoy_common.NamedResource, error) {
	mesh := xdsCtx.Mesh.Resource
	// Default Strict
	mode := pointer.DerefOr(conf.Mode, api.ModeStrict)
	protocol := core_mesh.ParseProtocol(inbound.GetProtocol())
	localClusterName := envoy_names.GetLocalClusterName(iface.WorkloadPort)
	cluster := envoy_common.NewCluster(envoy_common.WithService(localClusterName))
	service := inbound.GetService()
	routes := generator.GenerateRoutes(proxy, iface, cluster)
	listenerBuilder := envoy_listeners.NewInboundListenerBuilder(proxy.APIVersion, iface.DataplaneIP, iface.DataplanePort, core_xds.SocketAddressProtocolTCP).
		Configure(envoy_listeners.TransparentProxying(proxy.Dataplane.Spec.Networking.GetTransparentProxying())).
		Configure(envoy_listeners.TagsMetadata(inbound.GetTags()))

	switch mode {
	case api.ModeStrict:
		listenerBuilder.
			Configure(envoy_listeners.FilterChain(generator.FilterChainBuilder(true, protocol, proxy, localClusterName, xdsCtx, iface, service, &routes, conf.TlsVersion, conf.TlsCiphers).Configure(
				envoy_listeners.NetworkRBAC(listener.GetName(), mesh.MTLSEnabled(), proxy.Policies.TrafficPermissions[iface]),
			)))
	case api.ModePermissive:
		listenerBuilder.
			Configure(envoy_listeners.TLSInspector()).
			Configure(envoy_listeners.FilterChain(
				generator.FilterChainBuilder(false, protocol, proxy, localClusterName, xdsCtx, iface, service, &routes, conf.TlsVersion, conf.TlsCiphers).Configure(
					envoy_listeners.MatchTransportProtocol("raw_buffer"))),
			).
			Configure(envoy_listeners.FilterChain(
				// we need to differentiate between just TLS and Kuma's TLS, because with permissive mode
				// the app itself might be protected by TLS.
				generator.FilterChainBuilder(false, protocol, proxy, localClusterName, xdsCtx, iface, service, &routes, conf.TlsVersion, conf.TlsCiphers).Configure(
					envoy_listeners.MatchTransportProtocol("tls"))),
			).
			Configure(envoy_listeners.FilterChain(
				generator.FilterChainBuilder(true, protocol, proxy, localClusterName, xdsCtx, iface, service, &routes, conf.TlsVersion, conf.TlsCiphers).Configure(
					envoy_listeners.MatchTransportProtocol("tls"),
					envoy_listeners.MatchApplicationProtocols(xds_tls.KumaALPNProtocols...),
					envoy_listeners.NetworkRBAC(listener.GetName(), xdsCtx.Mesh.Resource.MTLSEnabled(), proxy.Policies.TrafficPermissions[iface]),
				)),
			)
	}
	return listenerBuilder.Build()
}
