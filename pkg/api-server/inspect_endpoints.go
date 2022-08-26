package api_server

import (
	"context"
	"fmt"
	"sort"

	"github.com/emicklei/go-restful/v3"

	api_server_types "github.com/kumahq/kuma/pkg/api-server/types"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/core/policy"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	rest_errors "github.com/kumahq/kuma/pkg/core/rest/errors"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/route"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/server/callbacks"
	"github.com/kumahq/kuma/pkg/xds/sync"
)

var CustomizeProxy func(meshContext xds_context.MeshContext, proxy *core_xds.Proxy) error

// getMatchedPolicies returns information about either sidecar dataplanes or
// builtin gateway dataplanes as well as the proxy and a potential error.
func getMatchedPolicies(
	ctx context.Context, cfg *kuma_cp.Config, meshContext xds_context.MeshContext, dataplaneKey core_model.ResourceKey,
) (
	*core_xds.MatchedPolicies, []gateway.GatewayListenerInfo, core_xds.Proxy, error,
) {
	proxyBuilder := sync.DefaultDataplaneProxyBuilder(
		*cfg,
		callbacks.NewDataplaneMetadataTracker(),
		envoy.APIV3)
	if proxy, err := proxyBuilder.Build(ctx, dataplaneKey, meshContext); err != nil {
		return nil, nil, core_xds.Proxy{}, err
	} else {
		if CustomizeProxy != nil {
			if err := CustomizeProxy(meshContext, proxy); err != nil {
				return nil, nil, core_xds.Proxy{}, err
			}
		}
		if proxy.Dataplane.Spec.IsBuiltinGateway() {
			entries, err := gateway.GatewayListenerInfoFromProxy(
				ctx, meshContext, proxy, proxyBuilder.Zone,
			)
			if err != nil {
				return nil, nil, core_xds.Proxy{}, err
			}

			return nil, entries, *proxy, nil
		}
		return &proxy.Policies, nil, *proxy, nil
	}
}

func addInspectEndpoints(
	ws *restful.WebService,
	cfg *kuma_cp.Config,
	builder xds_context.MeshContextBuilder,
	rm manager.ResourceManager,
) {
	ws.Route(
		ws.GET("/meshes/{mesh}/dataplanes/{dataplane}/policies").To(inspectDataplane(cfg, builder)).
			Doc("inspect dataplane matched policies").
			Param(ws.PathParameter("mesh", "mesh name").DataType("string")).
			Param(ws.PathParameter("dataplane", "dataplane name").DataType("string")).
			Returns(200, "OK", nil),
	)

	for _, desc := range registry.Global().ObjectDescriptors(core_model.AllowedToInspect()) {
		ws.Route(
			ws.GET(fmt.Sprintf("/meshes/{mesh}/%s/{name}/dataplanes", desc.WsPath)).To(inspectPolicies(desc.Name, builder, cfg)).
				Doc("inspect policies").
				Param(ws.PathParameter("mesh", "mesh name").DataType("string")).
				Param(ws.PathParameter("name", "resource name").DataType("string")).
				Returns(200, "OK", nil),
		)
	}

	ws.Route(
		ws.GET("/meshes/{mesh}/meshgateways/{name}/dataplanes").To(inspectGatewayDataplanes(builder, rm)).
			Doc("inspect MeshGateway").
			Param(ws.PathParameter("mesh", "mesh name").DataType("string")).
			Param(ws.PathParameter("name", "resource name").DataType("string")).
			Returns(200, "OK", nil),
	)
	ws.Route(
		ws.GET("/meshes/{mesh}/meshgatewayroutes/{name}/dataplanes").To(inspectGatewayRouteDataplanes(cfg, builder, rm)).
			Doc("inspect MeshGatewayRoute").
			Param(ws.PathParameter("mesh", "mesh name").DataType("string")).
			Param(ws.PathParameter("name", "resource name").DataType("string")).
			Returns(200, "OK", nil),
	)
}

func inspectDataplane(cfg *kuma_cp.Config, builder xds_context.MeshContextBuilder) restful.RouteFunction {
	return func(request *restful.Request, response *restful.Response) {
		ctx := request.Request.Context()
		meshName := request.PathParameter("mesh")
		dataplaneName := request.PathParameter("dataplane")

		meshContext, err := builder.Build(ctx, meshName)
		if err != nil {
			rest_errors.HandleError(response, err, "Could not build MeshContext")
			return
		}

		matchedPolicies, gatewayEntries, proxy, err := getMatchedPolicies(
			request.Request.Context(), cfg, meshContext, core_model.ResourceKey{Mesh: meshName, Name: dataplaneName},
		)
		if err != nil {
			rest_errors.HandleError(response, err, "Could not get MatchedPolicies")
			return
		}

		var result api_server_types.DataplaneInspectResponse
		if matchedPolicies != nil {
			inner := api_server_types.NewDataplaneInspectEntryList()
			inner.Items = append(inner.Items, newDataplaneInspectResponse(matchedPolicies, proxy.Dataplane)...)
			inner.Total = uint32(len(inner.Items))
			result = api_server_types.NewDataplaneInspectResponse(inner)
		} else {
			inner := newGatewayDataplaneInspectResponse(proxy, gatewayEntries)
			result = api_server_types.NewDataplaneInspectResponse(&inner)
		}
		if err := response.WriteAsJson(result); err != nil {
			rest_errors.HandleError(response, err, "Could not write response")
			return
		}
	}
}

func inspectGatewayDataplanes(
	builder xds_context.MeshContextBuilder,
	rm manager.ReadOnlyResourceManager,
) restful.RouteFunction {
	return func(request *restful.Request, response *restful.Response) {
		ctx := request.Request.Context()
		meshName := request.PathParameter("mesh")
		gatewayName := request.PathParameter("name")

		meshContext, err := builder.Build(ctx, meshName)
		if err != nil {
			rest_errors.HandleError(response, err, "Could not build mesh context")
			return
		}

		meshGateway := core_mesh.NewMeshGatewayResource()
		if err := rm.Get(ctx, meshGateway, store.GetByKey(gatewayName, meshName)); err != nil {
			rest_errors.HandleError(response, err, "Could not find MeshGateway")
			return
		}

		result := api_server_types.NewGatewayDataplanesInspectResult()
		for _, dp := range meshContext.Resources.Dataplanes().Items {
			if !dp.Spec.IsBuiltinGateway() {
				continue
			}
			if p := policy.SelectDataplanePolicyWithMatcher(dp.Spec.Matches, []policy.DataplanePolicy{meshGateway}); p != nil {
				result.Items = append(
					result.Items,
					api_server_types.GatewayDataplanesInspectEntry{
						DataplaneKey: api_server_types.ResourceKeyEntryFromModelKey(core_model.MetaToResourceKey(dp.GetMeta())),
					},
				)
			}
		}

		result.Total = uint32(len(result.Items))

		if err := response.WriteAsJson(result); err != nil {
			rest_errors.HandleError(response, err, "Could not write response")
			return
		}
	}
}

func inspectGatewayRouteDataplanes(
	cfg *kuma_cp.Config,
	builder xds_context.MeshContextBuilder,
	rm manager.ReadOnlyResourceManager,
) restful.RouteFunction {
	return func(request *restful.Request, response *restful.Response) {
		ctx := request.Request.Context()
		meshName := request.PathParameter("mesh")
		gatewayRouteName := request.PathParameter("name")

		meshContext, err := builder.Build(ctx, meshName)
		if err != nil {
			rest_errors.HandleError(response, err, "Could not build mesh context")
			return
		}

		gatewayRoute := core_mesh.NewMeshGatewayRouteResource()
		if err := rm.Get(ctx, gatewayRoute, store.GetByKey(gatewayRouteName, meshName)); err != nil {
			rest_errors.HandleError(response, err, "Could not find MeshGatewayRoute")
			return
		}

		dataplanes := map[core_model.ResourceKey]struct{}{}

		for _, dp := range meshContext.Resources.Dataplanes().Items {
			if !dp.Spec.IsBuiltinGateway() {
				continue
			}
			key := core_model.MetaToResourceKey(dp.GetMeta())
			_, listeners, _, err := getMatchedPolicies(request.Request.Context(), cfg, meshContext, key)
			if err != nil {
				rest_errors.HandleError(response, err, "Could not generate listener info")
				return
			}
			for _, listener := range listeners {
				for _, host := range listener.HostInfos {
					for _, entry := range host.Entries {
						if entry.Route != gatewayRoute.GetMeta().GetName() {
							continue
						}
						dataplanes[key] = struct{}{}
					}
				}
			}
		}

		result := api_server_types.NewGatewayDataplanesInspectResult()
		for key := range dataplanes {
			result.Items = append(
				result.Items,
				api_server_types.GatewayDataplanesInspectEntry{
					DataplaneKey: api_server_types.ResourceKeyEntryFromModelKey(key),
				},
			)
		}

		result.Total = uint32(len(result.Items))

		if err := response.WriteAsJson(result); err != nil {
			rest_errors.HandleError(response, err, "Could not write response")
			return
		}
	}
}

func inspectPolicies(
	resType core_model.ResourceType,
	builder xds_context.MeshContextBuilder,
	cfg *kuma_cp.Config,
) restful.RouteFunction {
	return func(request *restful.Request, response *restful.Response) {
		ctx := request.Request.Context()
		meshName := request.PathParameter("mesh")
		policyName := request.PathParameter("name")

		meshContext, err := builder.Build(ctx, meshName)
		if err != nil {
			rest_errors.HandleError(response, err, "Could not list Dataplanes")
			return
		}

		result := api_server_types.NewPolicyInspectEntryList()

		for _, dp := range meshContext.Resources.Dataplanes().Items {
			dpKey := core_model.MetaToResourceKey(dp.GetMeta())
			resourceKey := api_server_types.ResourceKeyEntry{
				Mesh: dpKey.Mesh,
				Name: dpKey.Name,
			}
			matchedPolicies, gatewayEntries, proxy, err := getMatchedPolicies(request.Request.Context(), cfg, meshContext, dpKey)
			if err != nil {
				rest_errors.HandleError(response, err, fmt.Sprintf("Could not get MatchedPolicies for %v", dpKey))
				return
			}
			if matchedPolicies != nil {
				for policy, attachments := range core_xds.GroupByPolicy(matchedPolicies, dp.Spec.Networking) {
					if policy.Type == resType && policy.Key.Name == policyName && policy.Key.Mesh == meshName {
						attachmentList := []api_server_types.AttachmentEntry{}
						for _, attachment := range attachments {
							attachmentList = append(attachmentList, api_server_types.AttachmentEntry{
								Type:    attachment.Type.String(),
								Name:    attachment.Name,
								Service: attachment.Service,
							})
						}
						entry := api_server_types.NewPolicyInspectSidecarEntry(resourceKey)
						entry.Attachments = attachmentList
						result.Items = append(result.Items, api_server_types.NewPolicyInspectEntry(&entry))
					}
				}
			} else {
				for policy, attachments := range gatewayEntriesByPolicy(proxy, gatewayEntries) {
					if policy.Type == resType && policy.Key.Name == policyName && policy.Key.Mesh == meshName {
						result.Items = append(result.Items, attachments...)
					}
				}
			}
		}

		result.Total = uint32(len(result.Items))

		if err := response.WriteAsJson(result); err != nil {
			rest_errors.HandleError(response, err, "Could not write response")
			return
		}
	}
}

func routeDestinationToAPIDestination(des route.Destination) api_server_types.Destination {
	policies := api_server_types.PolicyMap{}
	for kind, p := range des.Policies {
		policies[kind] = rest.From.Resource(p)
	}

	return api_server_types.Destination{
		Tags:     des.Destination,
		Policies: policies,
	}
}

func newDataplaneInspectResponse(matchedPolicies *core_xds.MatchedPolicies, dp *core_mesh.DataplaneResource) []*api_server_types.DataplaneInspectEntry {
	attachmentMap := core_xds.GroupByAttachment(matchedPolicies, dp.Spec.Networking)

	entries := make([]*api_server_types.DataplaneInspectEntry, 0, len(attachmentMap))
	attachments := []core_xds.Attachment{}
	for attachment := range attachmentMap {
		attachments = append(attachments, attachment)
	}

	sort.Stable(core_xds.AttachmentList(attachments))

	for _, attachment := range attachments {
		entry := &api_server_types.DataplaneInspectEntry{
			AttachmentEntry: api_server_types.AttachmentEntry{
				Type:    attachment.Type.String(),
				Name:    attachment.Name,
				Service: attachment.Service,
			},
			MatchedPolicies: map[core_model.ResourceType][]*rest.Resource{},
		}
		for typ, resList := range attachmentMap[attachment] {
			for _, res := range resList {
				entry.MatchedPolicies[typ] = append(entry.MatchedPolicies[typ], rest.From.Resource(res))
			}
		}

		entries = append(entries, entry)
	}

	return entries
}

func newGatewayDataplaneInspectResponse(
	proxy core_xds.Proxy,
	listenerInfos []gateway.GatewayListenerInfo,
) api_server_types.GatewayDataplaneInspectResult {
	var listeners []api_server_types.GatewayListenerInspectEntry

	for _, info := range listenerInfos {
		var hosts []api_server_types.HostInspectEntry
		for _, info := range info.HostInfos {
			var routes []api_server_types.RouteInspectEntry
			routeMap := map[string][]api_server_types.Destination{}
			for _, entry := range info.Entries {
				destinations := routeMap[entry.Route]

				if entry.Mirror != nil {
					destinations = append(destinations, routeDestinationToAPIDestination(entry.Mirror.Forward))
				}

				for _, forward := range entry.Action.Forward {
					routeMap[entry.Route] = append(destinations, routeDestinationToAPIDestination(forward))
				}
			}
			for name, destinations := range routeMap {
				if len(destinations) > 0 {
					sort.SliceStable(destinations, func(i, j int) bool {
						return destinations[i].Tags.String() < destinations[j].Tags.String()
					})
					routes = append(routes, api_server_types.RouteInspectEntry{
						Route:        name,
						Destinations: destinations,
					})
				}
			}
			sort.SliceStable(routes, func(i, j int) bool {
				return routes[i].Route < routes[j].Route
			})
			hosts = append(hosts, api_server_types.HostInspectEntry{
				HostName: info.Host.Hostname,
				Routes:   routes,
			})
		}
		sort.SliceStable(hosts, func(i, j int) bool {
			return hosts[i].HostName < hosts[j].HostName
		})
		listeners = append(listeners, api_server_types.GatewayListenerInspectEntry{
			Port:     info.Listener.Port,
			Protocol: info.Listener.Protocol.String(),
			Hosts:    hosts,
		})
	}

	result := api_server_types.NewGatewayDataplaneInspectResult()
	result.Listeners = listeners

	if len(listeners) > 0 {
		gatewayKey := core_model.MetaToResourceKey(listenerInfos[0].Gateway.GetMeta())
		result.Gateway = api_server_types.ResourceKeyEntry{
			Mesh: gatewayKey.Mesh,
			Name: gatewayKey.Name,
		}
	}

	gatewayPolicies := api_server_types.PolicyMap{}

	// TrafficLog and TrafficeTrace are applied to the entire MeshGateway
	// see pkg/plugins/runtime/gateway.newFilterChain
	if logging, ok := proxy.Policies.TrafficLogs[core_mesh.PassThroughService]; ok {
		gatewayPolicies[core_mesh.TrafficLogType] = rest.From.Resource(logging)
	}
	if trace := proxy.Policies.TrafficTrace; trace != nil {
		gatewayPolicies[core_mesh.TrafficTraceType] = rest.From.Resource(trace)
	}

	if len(gatewayPolicies) > 0 {
		result.Policies = gatewayPolicies
	}

	return result
}

func routeToPolicyInspect(
	policyMap map[core_xds.PolicyKey][]envoy.Tags,
	des route.Destination,
) map[core_xds.PolicyKey][]envoy.Tags {
	for kind, p := range des.Policies {
		policyKey := core_xds.PolicyKey{
			Type: kind,
			Key:  core_model.MetaToResourceKey(p.GetMeta()),
		}

		policies := policyMap[policyKey]
		policies = append(policies, des.Destination)
		policyMap[policyKey] = policies
	}

	return policyMap
}

func gatewayEntriesByPolicy(
	proxy core_xds.Proxy,
	listenerInfos []gateway.GatewayListenerInfo,
) map[core_xds.PolicyKey][]api_server_types.PolicyInspectEntry {
	policyMap := map[core_xds.PolicyKey][]api_server_types.PolicyInspectEntry{}

	dpKey := core_model.MetaToResourceKey(proxy.Dataplane.GetMeta())
	resourceKey := api_server_types.ResourceKeyEntry{
		Mesh: dpKey.Mesh,
		Name: dpKey.Name,
	}

	listenersMap := map[core_xds.PolicyKey][]api_server_types.PolicyInspectGatewayListenerEntry{}
	for _, info := range listenerInfos {
		hostMap := map[core_xds.PolicyKey][]api_server_types.PolicyInspectGatewayHostEntry{}
		for _, info := range info.HostInfos {
			routeMap := map[core_xds.PolicyKey][]api_server_types.PolicyInspectGatewayRouteEntry{}
			for _, entry := range info.Entries {
				entryMap := map[core_xds.PolicyKey][]envoy.Tags{}
				if entry.Mirror != nil {
					entryMap = routeToPolicyInspect(entryMap, entry.Mirror.Forward)
				}

				for _, forward := range entry.Action.Forward {
					entryMap = routeToPolicyInspect(entryMap, forward)
				}

				for policy, destinations := range entryMap {
					routeMap[policy] = append(
						routeMap[policy],
						api_server_types.PolicyInspectGatewayRouteEntry{
							Route:        entry.Route,
							Destinations: destinations,
						})
				}
			}

			for policy, routes := range routeMap {
				hostMap[policy] = append(
					hostMap[policy],
					api_server_types.PolicyInspectGatewayHostEntry{
						HostName: info.Host.Hostname,
						Routes:   routes,
					})
			}
		}

		for policy, hosts := range hostMap {
			listenersMap[policy] = append(
				listenersMap[policy],
				api_server_types.PolicyInspectGatewayListenerEntry{
					Port:     info.Listener.Port,
					Protocol: info.Listener.Protocol.String(),
					Hosts:    hosts,
				},
			)
		}
	}

	var gatewayKey api_server_types.ResourceKeyEntry
	if len(listenerInfos) > 0 {
		gatewayKey = api_server_types.ResourceKeyEntryFromModelKey(core_model.MetaToResourceKey(listenerInfos[0].Gateway.GetMeta()))
	}
	for policy, listeners := range listenersMap {
		result := api_server_types.NewPolicyInspectGatewayEntry(resourceKey, gatewayKey)

		result.Listeners = listeners

		result.Gateway = gatewayKey

		policyMap[policy] = append(
			policyMap[policy],
			api_server_types.NewPolicyInspectEntry(&result),
		)
	}

	if logging, ok := proxy.Policies.TrafficLogs[core_mesh.PassThroughService]; ok {
		wholeGateway := api_server_types.NewPolicyInspectGatewayEntry(resourceKey, gatewayKey)
		policyKey := core_xds.PolicyKey{
			Type: core_mesh.TrafficLogType,
			Key:  core_model.MetaToResourceKey(logging.GetMeta()),
		}
		policyMap[policyKey] = append(
			policyMap[policyKey],
			api_server_types.NewPolicyInspectEntry(&wholeGateway),
		)
	}
	if trace := proxy.Policies.TrafficTrace; trace != nil {
		wholeGateway := api_server_types.NewPolicyInspectGatewayEntry(resourceKey, gatewayKey)
		policyKey := core_xds.PolicyKey{
			Type: core_mesh.TrafficTraceType,
			Key:  core_model.MetaToResourceKey(trace.GetMeta()),
		}
		policyMap[policyKey] = append(
			policyMap[policyKey],
			api_server_types.NewPolicyInspectEntry(&wholeGateway),
		)
	}

	return policyMap
}
