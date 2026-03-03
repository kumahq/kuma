package v1alpha1

import (
	"encoding/json"
	"fmt"
	"net"
	net_url "net/url"
	"strconv"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/kri"
	unified_naming "github.com/kumahq/kuma/v2/pkg/core/naming/unified-naming"
	core_plugins "github.com/kumahq/kuma/v2/pkg/core/plugins"
	"github.com/kumahq/kuma/v2/pkg/core/resources/apis/core/destinationname"
	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	motb_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshopentelemetrybackend/api/v1alpha1"
	workload_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/workload/api/v1alpha1"
	core_system_names "github.com/kumahq/kuma/v2/pkg/core/system_names"
	"github.com/kumahq/kuma/v2/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/v2/pkg/core/xds/types"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/matchers"
	core_rules "github.com/kumahq/kuma/v2/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/rules/resolve"
	policies_xds "github.com/kumahq/kuma/v2/pkg/plugins/policies/core/xds"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/xds/meshroute"
	api "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshtrace/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/meshtrace/dpapi"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/meshtrace/metadata"
	plugin_xds "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshtrace/plugin/xds"
	k8s_metadata "github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/v2/pkg/xds/context"
	"github.com/kumahq/kuma/v2/pkg/xds/dynconf"
	"github.com/kumahq/kuma/v2/pkg/xds/envoy/clusters"
	xds_topology "github.com/kumahq/kuma/v2/pkg/xds/topology"
)

var _ core_plugins.PolicyPlugin = &plugin{}

type plugin struct{}

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources, opts ...core_plugins.MatchedPoliciesOption) (xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshTraceType, dataplane, resources, opts...)
}

func (p plugin) Apply(rs *xds.ResourceSet, ctx xds_context.Context, proxy *xds.Proxy) error {
	policies, ok := proxy.Policies.Dynamic[api.MeshTraceType]
	if !ok || len(policies.SingleItemRules.Rules) == 0 {
		return nil
	}

	listeners := policies_xds.GatherListeners(rs)
	if err := applyToInbounds(ctx, policies.SingleItemRules, listeners.Inbound, proxy); err != nil {
		return err
	}
	if err := applyToOutbounds(ctx, policies.SingleItemRules, listeners.Outbound, proxy); err != nil {
		return err
	}
	if err := applyToClusters(ctx, policies.SingleItemRules, rs, proxy); err != nil {
		return err
	}
	if err := applyToGateway(ctx, policies.SingleItemRules, listeners.Gateway, ctx.Mesh.Resources.MeshLocalResources, proxy); err != nil {
		return err
	}
	if err := applyToRealResources(ctx, policies.SingleItemRules, rs, proxy); err != nil {
		return err
	}
	if proxy.Metadata.HasFeature(xds_types.FeatureOtelViaKumaDp) {
		if err := configureDynamicDPConfig(ctx, policies.SingleItemRules, rs, proxy); err != nil {
			return err
		}
	}

	return nil
}

func applyToGateway(ctx xds_context.Context, rules core_rules.SingleItemRules, gatewayListeners map[core_rules.InboundListener]*envoy_listener.Listener, resources xds_context.ResourceMap, proxy *xds.Proxy) error {
	var gateways *core_mesh.MeshGatewayResourceList
	if rawList := resources[core_mesh.MeshGatewayType]; rawList != nil {
		gateways = rawList.(*core_mesh.MeshGatewayResourceList)
	} else {
		return nil
	}

	dataplane := proxy.Dataplane
	gateway := xds_topology.SelectGateway(gateways.Items, dataplane.Spec.Matches)
	if gateway == nil {
		return nil
	}

	for _, listener := range gateway.Spec.GetConf().GetListeners() {
		address := dataplane.Spec.GetNetworking().Address
		port := listener.GetPort()
		inboundListener := core_rules.InboundListener{
			Address: address,
			Port:    port,
		}
		listener, ok := gatewayListeners[inboundListener]
		if !ok {
			continue
		}

		if err := configureListener(ctx, rules, proxy, listener, ""); err != nil {
			return err
		}
	}

	return nil
}

func applyToInbounds(ctx xds_context.Context, rules core_rules.SingleItemRules, inboundListeners map[core_rules.InboundListener]*envoy_listener.Listener, proxy *xds.Proxy) error {
	for _, inboundListener := range inboundListeners {
		if err := configureListener(ctx, rules, proxy, inboundListener, ""); err != nil {
			return err
		}
	}

	return nil
}

func applyToOutbounds(ctx xds_context.Context, rules core_rules.SingleItemRules, outboundListeners map[mesh_proto.OutboundInterface]*envoy_listener.Listener, proxy *xds.Proxy) error {
	outbounds := proxy.Outbounds
	dataplane := proxy.Dataplane
	for _, outbound := range outbounds.Filter(xds_types.NonBackendRefFilter) {
		oface := dataplane.Spec.Networking.ToOutboundInterface(outbound.LegacyOutbound)

		listener, ok := outboundListeners[oface]
		if !ok {
			continue
		}

		serviceName := outbound.LegacyOutbound.GetService()

		if err := configureListener(ctx, rules, proxy, listener, serviceName); err != nil {
			return err
		}
	}

	return nil
}

func applyToRealResources(ctx xds_context.Context, rules core_rules.SingleItemRules, rs *xds.ResourceSet, proxy *xds.Proxy) error {
	for uri, resType := range rs.IndexByOrigin(xds.NonMeshExternalService) {
		service, port, found := meshroute.DestinationPortFromRef(ctx.Mesh, &resolve.RealResourceBackendRef{
			Resource: uri,
		})
		if !found {
			continue
		}
		for typ, resources := range resType {
			if typ == envoy_resource.ListenerType {
				for _, listener := range resources {
					if err := configureListener(ctx, rules, proxy, listener.Resource.(*envoy_listener.Listener), destinationname.MustResolve(false, service, port)); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func configureListener(ctx xds_context.Context, rules core_rules.SingleItemRules, proxy *xds.Proxy, listener *envoy_listener.Listener, destination string) error {
	serviceName := proxy.Dataplane.IdentifyingName(ctx.ControlPlane != nil && ctx.ControlPlane.InboundTagsDisabled)
	rawConf := rules.Rules[0].Conf
	conf := rawConf.(api.Conf)

	var workloadKRI string
	if workloadName := proxy.Dataplane.GetMeta().GetLabels()[k8s_metadata.KumaWorkload]; workloadName != "" {
		id := kri.Identifier{
			ResourceType: workload_api.WorkloadType,
			Mesh:         proxy.Dataplane.GetMeta().GetMesh(),
			Zone:         proxy.Zone,
			Namespace:    proxy.Dataplane.GetMeta().GetLabels()[mesh_proto.KubeNamespaceTag],
			Name:         workloadName,
		}
		workloadKRI = id.String()
	}
	resolved := resolveOtelBackendInfo(conf, ctx.Mesh.Resources, proxy.Metadata.GetDynamicMetadata(xds.FieldDynamicHostIP))

	configurer := plugin_xds.Configurer{
		Conf:                  conf,
		Service:               serviceName,
		TrafficDirection:      listener.TrafficDirection,
		Destination:           destination,
		IsGateway:             proxy.Dataplane.Spec.IsBuiltinGateway(),
		UnifiedResourceNaming: unified_naming.Enabled(proxy.Metadata, ctx.Mesh.Resource),
		Mesh:                  proxy.Dataplane.GetMeta().GetMesh(),
		Zone:                  proxy.Zone,
		WorkloadKRI:           workloadKRI,
		SkipOpenTelemetry:     shouldSkipUnresolvedOpenTelemetryBackendRef(conf, resolved),
	}
	if resolved != nil {
		configurer.ResolvedOtelName = resolved.Name
		// When kuma-dp acts as intermediary for a backendRef backend, Envoy
		// always speaks gRPC to the pipe cluster. Only fall back to HTTP config
		// when using direct-to-collector mode (inline endpoint or no feature).
		usePipe := hasOtelBackendRef(conf) && proxy.Metadata.HasFeature(xds_types.FeatureOtelViaKumaDp)
		if !usePipe && resolved.Protocol == motb_api.ProtocolHTTP {
			configurer.ResolvedOtelUseHTTP = true
			host := net.JoinHostPort(resolved.Endpoint.Target, strconv.Itoa(int(resolved.Endpoint.Port)))
			configurer.ResolvedOtelURI = fmt.Sprintf("http://%s%s", host, resolved.FullPath(policies_xds.OtelTracesPathSuffix))
		}
	}

	for _, chain := range listener.FilterChains {
		if err := configurer.Configure(chain); err != nil {
			return err
		}
	}

	return nil
}

func hasOtelBackendRef(conf api.Conf) bool {
	backends := pointer.Deref(conf.Backends)
	if len(backends) == 0 {
		return false
	}
	otel := backends[0].OpenTelemetry
	return otel != nil && otel.BackendRef != nil
}

func shouldSkipUnresolvedOpenTelemetryBackendRef(
	conf api.Conf,
	resolved *policies_xds.ResolvedOtelBackend,
) bool {
	if resolved != nil {
		return false
	}
	return hasOtelBackendRef(conf)
}

func applyToClusters(ctx xds_context.Context, rules core_rules.SingleItemRules, rs *xds.ResourceSet, proxy *xds.Proxy) error {
	rawConf := rules.Rules[0].Conf

	conf := rawConf.(api.Conf)

	var backend api.Backend
	if backends := pointer.Deref(conf.Backends); len(backends) == 0 {
		return nil
	} else {
		backend = backends[0]
	}

	var endpoint *xds.Endpoint
	var provider string
	useHTTP2 := false

	getNameOrDefault := core_system_names.GetNameOrDefault(unified_naming.Enabled(proxy.Metadata, ctx.Mesh.Resource))
	name := ""
	switch {
	case backend.Zipkin != nil:
		endpoint = endpointForZipkin(backend.Zipkin)
		provider = plugin_xds.ZipkinProviderName
		name = getNameOrDefault(
			core_system_names.AsSystemName(core_system_names.JoinSections("meshtrace_zipkin", core_system_names.CleanName(backend.Zipkin.Url))),
			plugin_xds.GetTracingClusterName(provider),
		)
	case backend.Datadog != nil:
		endpoint = endpointForDatadog(backend.Datadog)
		provider = plugin_xds.DatadogProviderName
		name = getNameOrDefault(
			core_system_names.AsSystemName(core_system_names.JoinSections("meshtrace_datadog", core_system_names.CleanName(backend.Datadog.Url))),
			plugin_xds.GetTracingClusterName(provider),
		)
	case backend.OpenTelemetry != nil:
		resolved := policies_xds.ResolveOtelBackend(
			backend.OpenTelemetry.BackendRef,
			backend.OpenTelemetry.Endpoint, //nolint:staticcheck // inline endpoint still supported for backward compat
			policies_xds.ParseOtelEndpoint,
			func(ep string) string { return ep },
			ctx.Mesh.Resources,
			proxy.Metadata.GetDynamicMetadata(xds.FieldDynamicHostIP),
		)
		if resolved == nil {
			return nil
		}
		if backend.OpenTelemetry.BackendRef != nil && proxy.Metadata.HasFeature(xds_types.FeatureOtelViaKumaDp) {
			// Route through kuma-dp Unix socket; kuma-dp forwards to the real collector.
			socketPath := xds.OpenTelemetrySocketName(proxy.Metadata.WorkDir, resolved.Name)
			endpoint = &xds.Endpoint{UnixDomainPath: socketPath}
			useHTTP2 = true // Envoy→kuma-dp leg is always gRPC
		} else {
			endpoint = resolved.Endpoint
			useHTTP2 = resolved.Protocol != motb_api.ProtocolHTTP
		}
		provider = plugin_xds.OpenTelemetryProviderName
		name = getNameOrDefault(
			core_system_names.AsSystemName(core_system_names.JoinSections("meshtrace_otel", core_system_names.CleanName(resolved.Name))),
			plugin_xds.GetTracingClusterName(provider),
		)
	}
	builder := clusters.NewClusterBuilder(proxy.APIVersion, name)
	if backend.OpenTelemetry != nil && useHTTP2 {
		builder.Configure(clusters.Http2())
	}

	res, err := builder.Configure(clusters.ProvidedEndpointCluster(proxy.Dataplane.IsIPv6(), *endpoint)).
		Configure(clusters.ClientSideTLS([]xds.Endpoint{*endpoint})).
		Configure(clusters.DefaultTimeout()).Build()
	if err != nil {
		return err
	}

	rs.Add(&xds.Resource{
		Name:     getNameOrDefault(name, plugin_xds.GetTracingClusterName(provider)),
		Origin:   metadata.OriginMeshTrace,
		Resource: res,
	})

	return nil
}

func endpointForZipkin(cfg *api.ZipkinBackend) *xds.Endpoint {
	url, _ := net_url.ParseRequestURI(cfg.Url)
	port, _ := strconv.ParseInt(url.Port(), 10, 32)
	return &xds.Endpoint{
		Target: url.Hostname(),
		Port:   uint32(port),
		ExternalService: &xds.ExternalService{
			TLSEnabled:         url.Scheme == "https",
			AllowRenegotiation: true,
		},
	}
}

func resolveOtelBackendInfo(conf api.Conf, resources xds_context.Resources, nodeHostIP string) *policies_xds.ResolvedOtelBackend {
	backends := pointer.Deref(conf.Backends)
	if len(backends) == 0 {
		return nil
	}
	backend := backends[0]
	if backend.OpenTelemetry == nil {
		return nil
	}
	return policies_xds.ResolveOtelBackend(
		backend.OpenTelemetry.BackendRef,
		backend.OpenTelemetry.Endpoint, //nolint:staticcheck // inline endpoint still supported for backward compat
		policies_xds.ParseOtelEndpoint,
		func(ep string) string { return ep },
		resources,
		nodeHostIP,
	)
}

func endpointForDatadog(cfg *api.DatadogBackend) *xds.Endpoint {
	url, _ := net_url.ParseRequestURI(cfg.Url)
	port, _ := strconv.ParseInt(url.Port(), 10, 32)

	return &xds.Endpoint{
		Target: url.Hostname(),
		Port:   uint32(port),
	}
}

func configureDynamicDPConfig(ctx xds_context.Context, rules core_rules.SingleItemRules, rs *xds.ResourceSet, proxy *xds.Proxy) error {
	conf := rules.Rules[0].Conf.(api.Conf)
	if !hasOtelBackendRef(conf) {
		return nil
	}
	resolved := resolveOtelBackendInfo(conf, ctx.Mesh.Resources, proxy.Metadata.GetDynamicMetadata(xds.FieldDynamicHostIP))
	if resolved == nil {
		return nil
	}

	endpoint := policies_xds.CollectorEndpointString(resolved.Endpoint)

	dpConfig := dpapi.MeshTraceDpConfig{
		Backends: []dpapi.OtelBackendConfig{
			{
				SocketPath: xds.OpenTelemetrySocketName(proxy.Metadata.WorkDir, resolved.Name),
				Endpoint:   endpoint,
				UseHTTP:    resolved.Protocol == motb_api.ProtocolHTTP,
				Path:       pointer.Deref(resolved.Path),
			},
		},
	}

	marshal, err := json.Marshal(dpConfig)
	if err != nil {
		return err
	}
	unifiedNamingEnabled := unified_naming.Enabled(proxy.Metadata, ctx.Mesh.Resource)
	getNameOrDefault := core_system_names.GetNameOrDefault(unifiedNamingEnabled)
	return dynconf.AddConfigRoute(proxy, rs, unifiedNamingEnabled, getNameOrDefault("meshtrace", dpapi.PATH), dpapi.PATH, marshal)
}
