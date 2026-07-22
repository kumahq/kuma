package v1alpha1

import (
	"fmt"
	"net"
	net_url "net/url"
	"strconv"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core"
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	"github.com/kumahq/kuma/v3/pkg/core/naming"
	unified_naming "github.com/kumahq/kuma/v3/pkg/core/naming/unified-naming"
	core_plugins "github.com/kumahq/kuma/v3/pkg/core/plugins"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/core/destinationname"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	motb_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshopentelemetrybackend/api/v1alpha1"
	workload_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/workload/api/v1alpha1"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	core_system_names "github.com/kumahq/kuma/v3/pkg/core/system_names"
	"github.com/kumahq/kuma/v3/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/v3/pkg/core/xds/types"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/matchers"
	core_rules "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/resolve"
	policies_xds "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/xds"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/xds/meshroute"
	api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtrace/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtrace/metadata"
	plugin_xds "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtrace/plugin/xds"
	k8s_metadata "github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	"github.com/kumahq/kuma/v3/pkg/xds/envoy/clusters"
)

var (
	_   core_plugins.PolicyPlugin = &plugin{}
	log                           = core.Log.WithName("MeshTrace")
)

type plugin struct{}

func (p plugin) Order() int { return api.MeshTraceResourceTypeDescriptor.Order }

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
	if err := applyToInbounds(ctx, policies, listeners.Inbound, proxy); err != nil {
		return err
	}
	if err := applyToOutbounds(ctx, policies.SingleItemRules, listeners.Outbound, proxy); err != nil {
		return err
	}
	if err := applyToZoneProxyListeners(ctx, policies, rs, proxy); err != nil {
		return err
	}
	if err := applyToClusters(ctx, policies.SingleItemRules, rs, proxy); err != nil {
		return err
	}
	if err := applyToRealResources(ctx, policies.SingleItemRules, rs, proxy); err != nil {
		return err
	}
	if proxy.Metadata.HasFeature(xds_types.FeatureOtelViaKumaDp) && proxy.OtelPipeBackends != nil {
		addToOtelPipeBackends(ctx, policies.SingleItemRules, proxy)
	}

	return nil
}

func applyToInbounds(ctx xds_context.Context, policies xds.TypedMatchingPolicies, inboundListeners map[core_rules.InboundListener]*envoy_listener.Listener, proxy *xds.Proxy) error {
	sectionNames := inboundSectionNames(proxy.Dataplane.Spec.GetNetworking())
	for key, inboundListener := range inboundListeners {
		listenerRules, err := buildListenerScopedSingleItemRules(policies, sectionNames[key])
		if err != nil {
			return err
		}
		if len(listenerRules.Rules) == 0 {
			continue
		}
		if err := configureListener(ctx, listenerRules, proxy, inboundListener, "", inboundListener.TrafficDirection); err != nil {
			return err
		}
	}

	return nil
}

func inboundSectionNames(n *mesh_proto.Dataplane_Networking) map[core_rules.InboundListener]string {
	result := map[core_rules.InboundListener]string{}
	for _, inb := range n.GetInbound() {
		iface := n.ToInboundInterface(inb)
		key := core_rules.InboundListener{Address: iface.DataplaneIP, Port: iface.DataplanePort}
		result[key] = inb.GetSectionName()
	}
	return result
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

		if err := configureListener(ctx, rules, proxy, listener, serviceName, listener.TrafficDirection); err != nil {
			return err
		}
	}

	return nil
}

func applyToZoneProxyListeners(ctx xds_context.Context, policies xds.TypedMatchingPolicies, rs *xds.ResourceSet, proxy *xds.Proxy) error {
	if !proxy.Dataplane.Spec.GetNetworking().HasZoneProxyListeners() {
		return nil
	}
	listenerRes := rs.Resources(envoy_resource.ListenerType)
	for _, l := range proxy.Dataplane.Spec.GetNetworking().GetListeners() {
		var name string
		switch l.GetType() {
		case mesh_proto.Dataplane_Networking_Listener_ZoneIngress:
			name = naming.ContextualZoneIngressListenerName(l.GetSectionName())
		case mesh_proto.Dataplane_Networking_Listener_ZoneEgress:
			name = naming.ContextualZoneEgressListenerName(l.GetSectionName())
		default:
			continue
		}
		res, ok := listenerRes[name]
		if !ok {
			log.V(1).Info("zone-proxy listener declared on Dataplane but missing from xDS; skipping tracing for this chain",
				"dataplane", proxy.Dataplane.GetMeta().GetName(),
				"listener", name)
			continue
		}
		listener, ok := res.Resource.(*envoy_listener.Listener)
		if !ok {
			continue
		}
		listenerRules, err := buildListenerScopedSingleItemRules(policies, l.GetSectionName())
		if err != nil {
			return err
		}
		if len(listenerRules.Rules) == 0 {
			continue
		}
		// Zone-proxy listener TrafficDirection is INBOUND on the wire (Envoy receives
		// from local sidecars), but a zone-egress span represents an outbound hop. Pass
		// UNSPECIFIED so Datadog SplitService skips the misleading "_INBOUND" suffix.
		if err := configureListener(ctx, listenerRules, proxy, listener, "", envoy_core.TrafficDirection_UNSPECIFIED); err != nil {
			return err
		}
	}
	return nil
}

// buildListenerScopedSingleItemRules returns rules scoped to the given embedded listener.
// Policies with `kind: Dataplane` and a non-empty `sectionName` that does not match the
// listener are excluded; everything else (including the matcher's proxy-wide rule when no
// per-policy info is available) is preserved.
func buildListenerScopedSingleItemRules(policies xds.TypedMatchingPolicies, sectionName string) (core_rules.SingleItemRules, error) {
	if len(policies.DataplanePolicies) == 0 {
		return policies.SingleItemRules, nil
	}
	filtered := make([]core_model.Resource, 0, len(policies.DataplanePolicies))
	for _, p := range policies.DataplanePolicies {
		policy, ok := p.GetSpec().(core_model.Policy)
		if !ok {
			continue
		}
		targetRef := policy.GetTargetRef()
		if targetRef.Kind == common_api.Dataplane {
			if sn := pointer.Deref(targetRef.SectionName); sn != "" && sn != sectionName {
				continue
			}
		}
		filtered = append(filtered, p)
	}
	return core_rules.BuildSingleItemRules(filtered)
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
					l := listener.Resource.(*envoy_listener.Listener)
					if err := configureListener(ctx, rules, proxy, l, destinationname.MustResolve(false, service, port), l.TrafficDirection); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func configureListener(ctx xds_context.Context, rules core_rules.SingleItemRules, proxy *xds.Proxy, listener *envoy_listener.Listener, destination string, direction envoy_core.TrafficDirection) error {
	serviceName := proxy.Dataplane.IdentifyingName(ctx.ControlPlane != nil && ctx.ControlPlane.InboundTagsDisabled)
	// IdentifyingName falls back to "unknown" on a zone-proxy-only Dataplane (no service tag).
	// Prefer the workload label (stable across pod restarts on K8s) and fall back to the
	// Dataplane name (= pod name on K8s) so span service names remain meaningful.
	if serviceName == mesh_proto.ServiceUnknown && proxy.Dataplane.Spec.GetNetworking().IsZoneProxyOnly() {
		if workloadName := proxy.Dataplane.GetMeta().GetLabels()[k8s_metadata.KumaWorkload]; workloadName != "" {
			serviceName = workloadName
		} else {
			serviceName = proxy.Dataplane.GetMeta().GetName()
		}
	}
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
	resolved := resolveOtelBackendInfo(conf, ctx.Mesh.Resources)

	configurer := plugin_xds.Configurer{
		Conf:                  conf,
		Service:               serviceName,
		TrafficDirection:      direction,
		Destination:           destination,
		IsGateway:             proxy.Dataplane.Spec.IsBuiltinGateway(),
		UnifiedResourceNaming: unified_naming.Enabled(proxy.Metadata, ctx.Mesh.Resource),
		Mesh:                  proxy.Dataplane.GetMeta().GetMesh(),
		Zone:                  proxy.Zone,
		WorkloadKRI:           workloadKRI,
		SkipOpenTelemetry:     shouldSkipUnresolvedOpenTelemetryBackend(conf, resolved),
	}
	if resolved != nil {
		configurer.ResolvedOtelName = resolved.Name
		// When kuma-dp acts as intermediary for the resolved backend, Envoy
		// always speaks gRPC to the pipe cluster. Only fall back to HTTP config
		// when using direct-to-collector mode (FeatureOtelViaKumaDp not enabled).
		usePipe := hasOtelBackendRef(conf) && proxy.Metadata.HasFeature(xds_types.FeatureOtelViaKumaDp)
		if !usePipe && resolved.Protocol == motb_api.ProtocolHTTP {
			configurer.ResolvedOtelUseHTTP = true
			host := net.JoinHostPort(resolved.Endpoint.Target, strconv.Itoa(int(resolved.Endpoint.Port)))
			scheme := "http"
			if resolved.UseHTTPS {
				scheme = "https"
			}
			configurer.ResolvedOtelURI = fmt.Sprintf("%s://%s%s", scheme, host, resolved.FullPath(policies_xds.OtelTracesPathSuffix))
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
	otel := openTelemetryBackend(conf)
	return otel != nil && otel.BackendRef != nil
}

func hasOpenTelemetryBackend(conf api.Conf) bool {
	return openTelemetryBackend(conf) != nil
}

func openTelemetryBackend(conf api.Conf) *api.OpenTelemetryBackend {
	backends := pointer.Deref(conf.Backends)
	if len(backends) == 0 {
		return nil
	}
	return backends[0].OpenTelemetry
}

func shouldSkipUnresolvedOpenTelemetryBackend(
	conf api.Conf,
	resolved *policies_xds.ResolvedOtelBackend,
) bool {
	if resolved != nil {
		return false
	}
	return hasOpenTelemetryBackend(conf)
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
			"",
			policies_xds.ParseOtelEndpoint,
			func(ep string) string { return ep },
			ctx.Mesh.Resources,
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
			endpoint = policies_xds.EndpointForDirectOtelExport(resolved, proxy.Metadata.GetDynamicMetadata(xds.FieldDynamicHostIP))
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

func resolveOtelBackendInfo(conf api.Conf, resources xds_context.Resources) *policies_xds.ResolvedOtelBackend {
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
		"",
		policies_xds.ParseOtelEndpoint,
		func(ep string) string { return ep },
		resources,
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

func addToOtelPipeBackends(ctx xds_context.Context, rules core_rules.SingleItemRules, proxy *xds.Proxy) {
	conf := rules.Rules[0].Conf.(api.Conf)
	if !hasOtelBackendRef(conf) {
		return
	}

	resolved := resolveOtelBackendInfo(conf, ctx.Mesh.Resources)
	if resolved == nil {
		return
	}

	base := policies_xds.BuildResolvedPipeBackend(proxy.Metadata.WorkDir, resolved)
	plan := policies_xds.BuildSignalRuntimePlan(
		proxy.Metadata.GetOtelEnvInventory(),
		base.EnvPolicy,
		xds.OtelSignalTraces,
		policies_xds.AddResolvedBackendOptions{},
	)
	proxy.OtelPipeBackends.AddSignal(resolved.Name, base, xds.OtelSignalTraces, plan)
}
