package v1alpha1

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_wellknown "github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/structpb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/kri"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	bldrs_al "github.com/kumahq/kuma/pkg/envoy/builders/accesslog"
	. "github.com/kumahq/kuma/pkg/envoy/builders/common"
	bldrs_hcm "github.com/kumahq/kuma/pkg/envoy/builders/filter/network/hcm"
	bldrs_listener "github.com/kumahq/kuma/pkg/envoy/builders/listener"
	bldrs_matcher "github.com/kumahq/kuma/pkg/envoy/builders/matcher"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/matchers"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	rules_inbound "github.com/kumahq/kuma/pkg/plugins/policies/core/rules/inbound"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/outbound"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/subsetutils"
	policies_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	. "github.com/kumahq/kuma/pkg/plugins/policies/meshaccesslog/plugin/xds"
	gateway_plugin "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_slices "github.com/kumahq/kuma/pkg/util/slices"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	listeners_v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
	"github.com/kumahq/kuma/pkg/xds/envoy/names"
	"github.com/kumahq/kuma/pkg/xds/generator"
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

	endpoints := &EndpointAccumulator{}

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
		defaultFormat := DefaultFormat(r.Protocol)

		svcCtx := rctx.
			WithID(kri.NoSectionName(*r.ResourceOrigin)).
			WithID(*r.ResourceOrigin)

		kumaValues := listeners_v3.KumaValues{
			SourceService:      proxy.Dataplane.Spec.GetIdentifyingService(),
			SourceIP:           proxy.Dataplane.GetIP(),
			DestinationService: r.ResourceOrigin.Name,
			Mesh:               proxy.Dataplane.GetMeta().GetMesh(),
			TrafficDirection:   envoy.TrafficDirectionOutbound,
		}

		switch envoyResource := r.Resource.(type) {
		case *envoy_listener.Listener:
			routeSet := map[kri.Identifier]struct{}{}

			listenerModifer := NewModifier(envoyResource)
			for _, backend := range pointer.Deref(svcCtx.Conf().Backends) {
				listenerModifer.Configure(bldrs_listener.AccessLog(
					BaseAccessLogBuilder(backend, defaultFormat, endpoints, kumaValues, accessLogSocketPath).
						Configure(bldrs_al.MetadataFilter(true, bldrs_matcher.NewMetadataBuilder().
							Configure(bldrs_matcher.Key(envoy_wellknown.FileAccessLog, routeMetadataKey)).
							Configure(bldrs_matcher.NullValue())))))
			}

			listenerModifer.Configure(bldrs_listener.ForEachRoute(func(hcm *envoy_hcm.HttpConnectionManager, route *routev3.Route, id kri.Identifier) error {
				routeConf, isDirectConf := svcCtx.WithID(id).DirectConf()
				if !isDirectConf {
					return nil
				}

				setRouteMetadata(route, routeMetadataKey, route.Name)

				if _, routeProcessed := routeSet[id]; routeProcessed {
					return nil
				}

				hcmModifier := NewModifier(hcm)
				for _, backend := range pointer.Deref(routeConf.Backends) {
					hcmModifier.Configure(bldrs_hcm.AccessLog(
						BaseAccessLogBuilder(backend, defaultFormat, endpoints, kumaValues, accessLogSocketPath).
							Configure(bldrs_al.MetadataFilter(false, bldrs_matcher.NewMetadataBuilder().
								Configure(bldrs_matcher.Key(envoy_wellknown.FileAccessLog, routeMetadataKey)).
								Configure(bldrs_matcher.ExactValue(id.String())))))).
						ConfigureIf(len(routeSet) == 0, func() Configurer[envoy_hcm.HttpConnectionManager] {
							return bldrs_hcm.LuaFilterAddFirst(luaFilter)
						})
				}
				if err := hcmModifier.Modify(); err != nil {
					return err
				}
				routeSet[id] = struct{}{}
				return nil
			}))

			if err := listenerModifer.Modify(); err != nil {
				return err
			}
		}
	}

	if err := AddLogBackendConf(*endpoints, rs, proxy); err != nil {
		return errors.Wrap(err, "unable to add configuration for MeshAccessLog backends")
	}

	return nil
}

const routeMetadataKey = "route_kri"

var luaFilter = `function envoy_on_request(handle)
  local meta = handle:metadata():get("route_kri")
  if meta ~= nil then
    handle:streamInfo():dynamicMetadata():set("envoy.access_loggers.file", "route_kri", meta)
  end
end`

func setRouteMetadata(r *routev3.Route, key, value string) {
	if r.Metadata == nil {
		r.Metadata = &envoy_core.Metadata{}
	}
	if r.Metadata.FilterMetadata == nil {
		r.Metadata.FilterMetadata = map[string]*structpb.Struct{}
	}
	r.Metadata.FilterMetadata[envoy_wellknown.Lua] = &structpb.Struct{
		Fields: map[string]*structpb.Value{
			key: {Kind: &structpb.Value_StringValue{StringValue: value}},
		},
	}
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
		protocol := core_mesh.ParseProtocol(inbound.GetProtocol())
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

		conf := core_rules.ComputeConf[api.Conf](rules.Rules, subsetutils.MeshServiceElement(serviceName))
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
	conf := core_rules.ComputeConf[api.Conf](policies.ToRules.Rules, subsetutils.MeshServiceElement(core_mesh.PassThroughService))
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
		if err := configureListener(*conf, ipv4, backends, core_mesh.ProtocolTCP, kumaValues, path); err != nil {
			return err
		}
	}

	if ipv6 != nil {
		return configureListener(*conf, ipv6, backends, core_mesh.ProtocolTCP, kumaValues, path)
	}

	return nil
}

func applyToDirectAccess(
	rules core_rules.ToRules, directAccess map[generator.Endpoint]*envoy_listener.Listener, dataplane *core_mesh.DataplaneResource,
	backends *EndpointAccumulator, path string,
) error {
	conf := core_rules.ComputeConf[api.Conf](rules.Rules, subsetutils.MeshServiceElement(core_mesh.PassThroughService))
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
		return configureListener(*conf, listener, backends, core_mesh.ProtocolTCP, kumaValues, path)
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
		var protocol core_mesh.Protocol
		if _, p, _, err := names.ParseGatewayListenerName(listener.GetName()); err != nil {
			return err
		} else {
			protocol = core_mesh.ParseProtocol(p)
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

func configureListener(
	conf api.Conf,
	listener *envoy_listener.Listener,
	backendsAcc *EndpointAccumulator,
	defaultFormat string,
	values listeners_v3.KumaValues,
	accessLogSocketPath string,
) error {
	listenerModifier := NewModifier(listener)
	for _, backend := range pointer.Deref(conf.Backends) {
		listenerModifier.Configure(bldrs_listener.AccessLog(
			BaseAccessLogBuilder(backend, defaultFormat, backendsAcc, values, accessLogSocketPath)))
	}
	return listenerModifier.Modify()
}
