package v1alpha1

import (
	"bytes"
	"maps"
	"slices"
	"text/template"

	envoy_accesslog "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/kri"
	core_meta "github.com/kumahq/kuma/pkg/core/metadata"
	unified_naming "github.com/kumahq/kuma/pkg/core/naming/unified-naming"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	bldrs_accesslog "github.com/kumahq/kuma/pkg/envoy/builders/accesslog"
	. "github.com/kumahq/kuma/pkg/envoy/builders/common"
	"github.com/kumahq/kuma/pkg/envoy/builders/filter/network/hcm"
	bldrs_listener "github.com/kumahq/kuma/pkg/envoy/builders/listener"
	bldrs_matcher "github.com/kumahq/kuma/pkg/envoy/builders/matcher"
	bldrs_route "github.com/kumahq/kuma/pkg/envoy/builders/route"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/matchers"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	rules_inbound "github.com/kumahq/kuma/pkg/plugins/policies/core/rules/inbound"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/outbound"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/subsetutils"
	policies_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	. "github.com/kumahq/kuma/pkg/plugins/policies/meshaccesslog/plugin/xds"
	gateway_plugin "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	util_maps "github.com/kumahq/kuma/pkg/util/maps"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_slices "github.com/kumahq/kuma/pkg/util/slices"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	listeners_v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
	"github.com/kumahq/kuma/pkg/xds/envoy/names"
	"github.com/kumahq/kuma/pkg/xds/generator"
	"github.com/kumahq/kuma/pkg/xds/generator/model"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

var _ core_plugins.PolicyPlugin = &plugin{}

type plugin struct{}

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources, opts ...core_plugins.MatchedPoliciesOption) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshAccessLogType, dataplane, resources, opts...)
}

func (p plugin) Apply(rs *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error {
	policies, ok := proxy.Policies.Dynamic[api.MeshAccessLogType]
	if !ok {
		return nil
	}

	endpoints := &EndpointAccumulator{
		UnifiedResourceNaming: unified_naming.Enabled(proxy.Metadata, ctx.Mesh.Resource),
	}

	listeners := policies_xds.GatherListeners(rs)

	accessLogSocketPath := core_xds.AccessLogSocketName(proxy.Metadata.WorkDir, proxy.Id.ToResourceKey().Name, proxy.Id.ToResourceKey().Mesh)

	if err := applyToInbounds(policies.FromRules, listeners.Inbound, proxy.Dataplane, endpoints, accessLogSocketPath); err != nil {
		return err
	}
	if err := applyToOutbounds(policies.ToRules, listeners.Outbound, proxy.Outbounds, proxy.Dataplane, endpoints, accessLogSocketPath, ctx.Mesh); err != nil {
		return err
	}
	if err := applyToTransparentProxyListeners(policies, listeners.Ipv4Passthrough, listeners.Ipv6Passthrough, proxy.Dataplane, endpoints, accessLogSocketPath); err != nil {
		return err
	}
	if err := applyToDirectAccess(policies.ToRules, listeners.DirectAccess, proxy.Dataplane, endpoints, accessLogSocketPath); err != nil {
		return err
	}
	if err := applyToGateway(policies.GatewayRules, listeners.Gateway, ctx.Mesh.Resources.MeshLocalResources, proxy, endpoints, accessLogSocketPath); err != nil {
		return err
	}

	rctx := outbound.RootContext[api.Conf](ctx.Mesh.Resource, policies.ToRules.ResourceRules)
	for _, r := range util_slices.Filter(rs.List(), core_xds.HasAssociatedServiceResource) {
		svcCtx := rctx.
			WithID(kri.NoSectionName(r.ResourceOrigin)).
			WithID(r.ResourceOrigin)
		if err := applyToRealResource(svcCtx, r, proxy, endpoints, accessLogSocketPath); err != nil {
			return err
		}
	}

	if err := AddLogBackendConf(*endpoints, rs, proxy); err != nil {
		return errors.Wrap(err, "unable to add configuration for MeshAccessLog backends")
	}

	return nil
}

func applyToInbounds(
	rules core_rules.FromRules,
	inboundListeners map[core_rules.InboundListener]*envoy_listener.Listener,
	dataplane *core_mesh.DataplaneResource,
	backends *EndpointAccumulator,
	accessLogSocketPath string,
) error {
	for _, inbound := range dataplane.Spec.GetNetworking().GetInbound() {
		iface := dataplane.Spec.Networking.ToInboundInterface(inbound)

		listenerKey := core_rules.InboundListener{
			Address: iface.DataplaneIP,
			Port:    iface.DataplanePort,
		}
		listener, ok := inboundListeners[listenerKey]
		if !ok {
			continue
		}
		protocol := core_meta.ParseProtocol(inbound.GetProtocol())
		conf := rules_inbound.MatchesAllIncomingTraffic[api.Conf](rules.InboundRules[listenerKey])
		kumaValues := listeners_v3.KumaValues{
			SourceService:      mesh_proto.ServiceUnknown,
			SourceIP:           dataplane.GetIP(), // todo(lobkovilya): why do we set SourceIP always to DPP's address? see https://github.com/kumahq/kuma/issues/13635
			DestinationService: dataplane.Spec.GetIdentifyingService(),
			Mesh:               dataplane.GetMeta().GetMesh(),
			TrafficDirection:   envoy.TrafficDirectionInbound,
		}
		if err := configureListener(conf, listener, backends, DefaultFormat(protocol), kumaValues, accessLogSocketPath); err != nil {
			return err
		}
	}
	return nil
}

func applyToOutbounds(
	rules core_rules.ToRules,
	outboundListeners map[mesh_proto.OutboundInterface]*envoy_listener.Listener,
	outbounds xds_types.Outbounds,
	dataplane *core_mesh.DataplaneResource,
	backendsAcc *EndpointAccumulator,
	accessLogSocketPath string,
	meshCtx xds_context.MeshContext,
) error {
	for _, outbound := range outbounds.Filter(xds_types.NonBackendRefFilter) {
		oface := dataplane.Spec.Networking.ToOutboundInterface(outbound.LegacyOutbound)

		listener, ok := outboundListeners[oface]
		if !ok {
			continue
		}

		serviceName := outbound.LegacyOutbound.GetService()

		kumaValues := listeners_v3.KumaValues{
			SourceService:      dataplane.Spec.GetIdentifyingService(),
			SourceIP:           dataplane.GetIP(),
			DestinationService: outbound.LegacyOutbound.GetService(),
			Mesh:               dataplane.GetMeta().GetMesh(),
			TrafficDirection:   envoy.TrafficDirectionOutbound,
		}

		conf := core_rules.ComputeConf[api.Conf](rules.Rules, subsetutils.KumaServiceTagElement(serviceName))
		if conf == nil {
			continue
		}

		protocol := meshCtx.GetServiceProtocol(serviceName)
		if err := configureListener(*conf, listener, backendsAcc, DefaultFormat(protocol), kumaValues, accessLogSocketPath); err != nil {
			return err
		}
	}

	return nil
}

func applyToTransparentProxyListeners(
	policies core_xds.TypedMatchingPolicies, ipv4 *envoy_listener.Listener, ipv6 *envoy_listener.Listener, dataplane *core_mesh.DataplaneResource,
	backends *EndpointAccumulator, path string,
) error {
	conf := core_rules.ComputeConf[api.Conf](policies.ToRules.Rules, subsetutils.KumaServiceTagElement(core_meta.PassThroughServiceName))
	if conf == nil {
		return nil
	}

	kumaValues := listeners_v3.KumaValues{
		SourceService:      dataplane.Spec.GetIdentifyingService(),
		SourceIP:           dataplane.GetIP(),
		DestinationService: "external",
		Mesh:               dataplane.GetMeta().GetMesh(),
		TrafficDirection:   envoy.TrafficDirectionOutbound,
	}

	if ipv4 != nil {
		if err := configureListener(*conf, ipv4, backends, core_meta.ProtocolTCP, kumaValues, path); err != nil {
			return err
		}
	}

	if ipv6 != nil {
		return configureListener(*conf, ipv6, backends, core_meta.ProtocolTCP, kumaValues, path)
	}

	return nil
}

func applyToDirectAccess(
	rules core_rules.ToRules, directAccess map[model.Endpoint]*envoy_listener.Listener, dataplane *core_mesh.DataplaneResource,
	backends *EndpointAccumulator, path string,
) error {
	conf := core_rules.ComputeConf[api.Conf](rules.Rules, subsetutils.KumaServiceTagElement(core_meta.PassThroughServiceName))
	if conf == nil {
		return nil
	}

	for endpoint, listener := range directAccess {
		kumaValues := listeners_v3.KumaValues{
			SourceService:      dataplane.Spec.GetIdentifyingService(),
			SourceIP:           dataplane.GetIP(),
			DestinationService: generator.DirectAccessEndpointName(endpoint),
			Mesh:               dataplane.GetMeta().GetMesh(),
			TrafficDirection:   envoy.TrafficDirectionOutbound,
		}
		return configureListener(*conf, listener, backends, core_meta.ProtocolTCP, kumaValues, path)
	}

	return nil
}

func applyToGateway(
	rules core_rules.GatewayRules,
	gatewayListeners map[core_rules.InboundListener]*envoy_listener.Listener,
	resources xds_context.ResourceMap,
	proxy *core_xds.Proxy,
	backends *EndpointAccumulator,
	path string,
) error {
	var gateways *core_mesh.MeshGatewayResourceList
	if rawList := resources[core_mesh.MeshGatewayType]; rawList != nil {
		gateways = rawList.(*core_mesh.MeshGatewayResourceList)
	} else {
		return nil
	}

	gateway := xds_topology.SelectGateway(gateways.Items, proxy.Dataplane.Spec.Matches)
	if gateway == nil {
		return nil
	}

	for _, listenerInfo := range gateway_plugin.ExtractGatewayListeners(proxy) {
		address := proxy.Dataplane.Spec.GetNetworking().Address
		port := listenerInfo.Listener.Port
		listenerKey := core_rules.InboundListener{
			Address: address,
			Port:    port,
		}
		listener, ok := gatewayListeners[listenerKey]
		if !ok {
			continue
		}
		var protocol core_meta.Protocol
		if _, p, _, err := names.ParseGatewayListenerName(listener.GetName()); err != nil {
			return err
		} else {
			protocol = core_meta.ParseProtocol(p)
		}

		if toListenerRules, ok := rules.ToRules.ByListener[listenerKey]; ok {
			if conf := core_rules.ComputeConf[api.Conf](toListenerRules.Rules, subsetutils.MeshElement()); conf != nil {
				kumaValues := listeners_v3.KumaValues{
					SourceService:      proxy.Dataplane.Spec.GetIdentifyingService(),
					SourceIP:           proxy.Dataplane.GetIP(),
					DestinationService: mesh_proto.MatchAllTag,
					Mesh:               proxy.Dataplane.GetMeta().GetMesh(),
					TrafficDirection:   envoy.TrafficDirectionOutbound,
				}

				if err := configureListener(*conf, listener, backends, DefaultFormat(protocol), kumaValues, path); err != nil {
					return err
				}
			}
		}

		if fromListenerRules, ok := rules.InboundRules[listenerKey]; ok {
			conf := rules_inbound.MatchesAllIncomingTraffic[api.Conf](fromListenerRules)
			kumaValues := listeners_v3.KumaValues{
				SourceService:      mesh_proto.ServiceUnknown,
				SourceIP:           proxy.Dataplane.GetIP(), // todo(lobkovilya): why do we set SourceIP always to DPP's address? see https://github.com/kumahq/kuma/issues/13635
				DestinationService: proxy.Dataplane.Spec.GetIdentifyingService(),
				Mesh:               proxy.Dataplane.GetMeta().GetMesh(),
				TrafficDirection:   envoy.TrafficDirectionInbound,
			}
			if err := configureListener(conf, listener, backends, DefaultFormat(protocol), kumaValues, path); err != nil {
				return err
			}
		}
	}

	return nil
}

func configureListener[T ~string](
	conf api.Conf,
	listener *envoy_listener.Listener,
	backendsAcc *EndpointAccumulator,
	defaultFormat T,
	values listeners_v3.KumaValues,
	accessLogSocketPath string,
) error {
	return NewModifier(listener).
		Configure(bldrs_listener.AccessLogs(
			util_slices.Map(
				pointer.Deref(conf.Backends),
				func(b api.Backend) *Builder[envoy_accesslog.AccessLog] {
					return BaseAccessLogBuilder(b, string(defaultFormat), backendsAcc, values, accessLogSocketPath)
				}))).
		Modify()
}

func applyToRealResource(
	rctx *outbound.ResourceContext[api.Conf],
	r *core_xds.Resource,
	proxy *core_xds.Proxy,
	backendsAcc *EndpointAccumulator,
	accessLogSocketPath string,
) error {
	listener, ok := r.Resource.(*envoy_listener.Listener)
	if !ok {
		return nil
	}

	defaultFormat := DefaultFormat(r.Protocol)

	kumaValues := listeners_v3.KumaValues{
		SourceService:      proxy.Dataplane.Spec.GetIdentifyingService(),
		SourceIP:           proxy.Dataplane.GetIP(),
		DestinationService: r.ResourceOrigin.Name,
		Mesh:               proxy.Dataplane.GetMeta().GetMesh(),
		TrafficDirection:   envoy.TrafficDirectionOutbound,
	}

	backends := pointer.Deref(rctx.Conf().Backends)

	routesConfs, err := buildRoutesMap(listener, rctx)
	if err != nil {
		return err
	}

	routesIds := slices.SortedStableFunc(maps.Keys(routesConfs), kri.Compare)

	hasAtLeastOneBackend := len(util_slices.Filter(util_maps.AllValues(routesConfs), func(conf api.Conf) bool {
		return len(pointer.Deref(conf.Backends)) > 0
	})) > 0

	builderForSharedBackend := func(b api.Backend) *Builder[envoy_accesslog.AccessLog] {
		return BaseAccessLogBuilder(b, defaultFormat, backendsAcc, kumaValues, accessLogSocketPath).
			Configure(If(core_meta.IsHTTPBased(r.Protocol), bldrs_accesslog.MetadataFilter(true, bldrs_matcher.NewMetadataBuilder().
				Configure(bldrs_matcher.Key(namespace, routeMetadataKey)).
				Configure(bldrs_matcher.NullValue())),
			))
	}

	builderForRouteBackend := func(routeID kri.Identifier) func(b api.Backend) *Builder[envoy_accesslog.AccessLog] {
		return func(b api.Backend) *Builder[envoy_accesslog.AccessLog] {
			return BaseAccessLogBuilder(b, defaultFormat, backendsAcc, kumaValues, accessLogSocketPath).
				Configure(bldrs_accesslog.MetadataFilter(false, bldrs_matcher.NewMetadataBuilder().
					Configure(bldrs_matcher.Key(namespace, routeMetadataKey)).
					Configure(bldrs_matcher.ExactValue(routeID.String()))))
		}
	}

	return NewModifier(listener).
		Configure(bldrs_listener.AccessLogs(util_slices.Map(backends, builderForSharedBackend))).
		Configure(bldrs_listener.Routes(
			util_maps.MapValues(
				routesConfs,
				func(id kri.Identifier, _ api.Conf) Configurer[routev3.Route] {
					return bldrs_route.Metadata(routeMetadataKey, id.String())
				}))).
		Configure(bldrs_listener.AccessLogs(
			util_slices.FlatMap(
				routesIds,
				func(id kri.Identifier) []*Builder[envoy_accesslog.AccessLog] {
					return util_slices.Map(
						pointer.Deref(routesConfs[id].Backends),
						builderForRouteBackend(id),
					)
				}))).
		Configure(If(hasAtLeastOneBackend, bldrs_listener.HCM(hcm.LuaFilterAddFirst(setFilterMetadataAsDynamicLuaFilter(namespace, routeMetadataKey))))).
		Modify()
}

func buildRoutesMap(l *envoy_listener.Listener, svcCtx *outbound.ResourceContext[api.Conf]) (map[kri.Identifier]api.Conf, error) {
	routes := map[kri.Identifier]api.Conf{}
	if err := bldrs_listener.TraverseRoutes(l, func(route *routev3.Route) {
		if !kri.IsValid(route.Name) {
			return
		}

		id, _ := kri.FromString(route.Name)

		routeCtx := svcCtx.
			WithID(kri.NoSectionName(id)).
			WithID(id)

		if conf, isDirect := routeCtx.DirectConf(); isDirect {
			routes[id] = conf
		}
	}); err != nil {
		return nil, err
	}
	return routes, nil
}

const (
	routeMetadataKey = "route_kri"
	namespace        = "kuma.routes"
)

var luaTemplate = template.Must(template.New("luaFilter").Parse(`function envoy_on_request(handle)
  local meta = handle:metadata():get("{{ .Key }}")
  if meta ~= nil then
    handle:streamInfo():dynamicMetadata():set("{{ .Namespace }}", "{{ .Key }}", meta)
  end
end
`))

// setFilterMetadataAsDynamicLuaFilter returns a Lua filter that takes filter's metadata (set by the route)
// and set's this as a dynamic metadata (used by the AccessLog filter)
func setFilterMetadataAsDynamicLuaFilter(namespace, key string) string {
	var buf bytes.Buffer
	_ = luaTemplate.Execute(&buf, struct {
		Namespace string
		Key       string
	}{
		Namespace: namespace,
		Key:       key,
	})
	return buf.String()
}
