package api_server

import (
	"context"
	"fmt"
	"sort"

	"github.com/emicklei/go-restful/v3"

	api_server_types "github.com/kumahq/kuma/pkg/api-server/types"
	"github.com/kumahq/kuma/pkg/core/policy"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	rest_errors "github.com/kumahq/kuma/pkg/core/rest/errors"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/core/xds/inspect"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/route"
	"github.com/kumahq/kuma/pkg/util/maps"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
	"github.com/kumahq/kuma/pkg/xds/sync"
)

// getMatchedPolicies returns information about either sidecar dataplanes or
// builtin gateway dataplanes as well as the proxy and a potential error.
func getMatchedPolicies(
	ctx context.Context, zoneName string, meshContext xds_context.MeshContext, dataplaneKey core_model.ResourceKey,
) (
	*core_xds.Proxy, error,
) {
	proxyBuilder := sync.DefaultDataplaneProxyBuilder(zoneName, envoy.APIV3)
	proxy, err := proxyBuilder.Build(ctx, dataplaneKey, meshContext)
	if err != nil {
		return nil, err
	}
	return proxy, nil
}

func addInspectEndpoints(
	ws *restful.WebService,
	zoneName string,
	builder xds_context.MeshContextBuilder,
	rm manager.ResourceManager,
) {
	ws.Route(
		ws.GET("/meshes/{mesh}/dataplanes/{dataplane}/policies").To(inspectDataplane(zoneName, builder)).
			Doc("inspect dataplane matched policies").
			Param(ws.PathParameter("mesh", "mesh name").DataType("string")).
			Param(ws.PathParameter("dataplane", "dataplane name").DataType("string")).
			Returns(200, "OK", nil),
	)

	ws.Route(
		ws.GET("/meshes/{mesh}/dataplanes/{dataplane}/rules").To(inspectRulesAttachment(zoneName, builder)).
			Doc("inspect dataplane matched rules").
			Param(ws.PathParameter("mesh", "mesh name").DataType("string")).
			Param(ws.PathParameter("dataplane", "dataplane name").DataType("string")).
			Returns(200, "OK", nil),
	)

	for _, desc := range registry.Global().ObjectDescriptors(core_model.AllowedToInspect()) {
		ws.Route(
			ws.GET(fmt.Sprintf("/meshes/{mesh}/%s/{name}/dataplanes", desc.WsPath)).To(inspectPolicies(desc.Name, builder, zoneName)).
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
		ws.GET("/meshes/{mesh}/meshgatewayroutes/{name}/dataplanes").To(inspectGatewayRouteDataplanes(zoneName, builder, rm)).
			Doc("inspect MeshGatewayRoute").
			Param(ws.PathParameter("mesh", "mesh name").DataType("string")).
			Param(ws.PathParameter("name", "resource name").DataType("string")).
			Returns(200, "OK", nil),
	)
}

func inspectDataplane(zoneName string, builder xds_context.MeshContextBuilder) restful.RouteFunction {
	return func(request *restful.Request, response *restful.Response) {
		ctx := request.Request.Context()
		meshName := request.PathParameter("mesh")
		dataplaneName := request.PathParameter("dataplane")

		meshContext, err := builder.Build(ctx, meshName)
		if err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Could not build MeshContext")
			return
		}

		proxy, err := getMatchedPolicies(
			request.Request.Context(), zoneName, meshContext, core_model.ResourceKey{Mesh: meshName, Name: dataplaneName},
		)
		if err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Could not get MatchedPolicies")
			return
		}

		var result api_server_types.DataplaneInspectResponse
		if !proxy.Dataplane.Spec.IsBuiltinGateway() {
			inner := api_server_types.NewDataplaneInspectEntryList()
			inner.Items = append(inner.Items, newDataplaneInspectResponse(&proxy.Policies, proxy.Dataplane)...)
			inner.Total = uint32(len(inner.Items))
			result = api_server_types.NewDataplaneInspectResponse(inner)
		} else {
			inner := newGatewayDataplaneInspectResponse(proxy)
			result = api_server_types.NewDataplaneInspectResponse(&inner)
		}
		if err := response.WriteAsJson(result); err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Could not write response")
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
			rest_errors.HandleError(request.Request.Context(), response, err, "Could not build mesh context")
			return
		}

		meshGateway := core_mesh.NewMeshGatewayResource()
		if err := rm.Get(ctx, meshGateway, store.GetByKey(gatewayName, meshName)); err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Could not find MeshGateway")
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
			rest_errors.HandleError(request.Request.Context(), response, err, "Could not write response")
			return
		}
	}
}

func inspectGatewayRouteDataplanes(
	zoneName string,
	builder xds_context.MeshContextBuilder,
	rm manager.ReadOnlyResourceManager,
) restful.RouteFunction {
	return func(request *restful.Request, response *restful.Response) {
		ctx := request.Request.Context()
		meshName := request.PathParameter("mesh")
		gatewayRouteName := request.PathParameter("name")

		meshContext, err := builder.Build(ctx, meshName)
		if err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Could not build mesh context")
			return
		}

		gatewayRoute := core_mesh.NewMeshGatewayRouteResource()
		if err := rm.Get(ctx, gatewayRoute, store.GetByKey(gatewayRouteName, meshName)); err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Could not find MeshGatewayRoute")
			return
		}

		dataplanes := map[core_model.ResourceKey]struct{}{}

		for _, dp := range meshContext.Resources.Dataplanes().Items {
			if !dp.Spec.IsBuiltinGateway() {
				continue
			}
			key := core_model.MetaToResourceKey(dp.GetMeta())
			proxy, err := getMatchedPolicies(request.Request.Context(), zoneName, meshContext, key)
			if err != nil {
				rest_errors.HandleError(request.Request.Context(), response, err, "Could not generate listener info")
				return
			}
			for _, listener := range gateway.ExtractGatewayListeners(proxy) {
				for _, listenerHostname := range listener.ListenerHostnames {
					for _, host := range listenerHostname.HostInfos {
						for _, entry := range host.Entries() {
							if entry.Route != gatewayRoute.GetMeta().GetName() {
								continue
							}
							dataplanes[key] = struct{}{}
						}
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
			rest_errors.HandleError(request.Request.Context(), response, err, "Could not write response")
			return
		}
	}
}

func inspectPolicies(
	resType core_model.ResourceType,
	builder xds_context.MeshContextBuilder,
	zoneName string,
) restful.RouteFunction {
	return func(request *restful.Request, response *restful.Response) {
		ctx := request.Request.Context()
		meshName := request.PathParameter("mesh")
		policyName := request.PathParameter("name")

		meshContext, err := builder.Build(ctx, meshName)
		if err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Could not list Dataplanes")
			return
		}

		result := api_server_types.NewPolicyInspectEntryList()

		for _, dp := range meshContext.Resources.Dataplanes().Items {
			dpKey := core_model.MetaToResourceKey(dp.GetMeta())
			resourceKey := api_server_types.ResourceKeyEntry{
				Mesh: dpKey.Mesh,
				Name: dpKey.Name,
			}
			proxy, err := getMatchedPolicies(request.Request.Context(), zoneName, meshContext, dpKey)
			if err != nil {
				rest_errors.HandleError(request.Request.Context(), response, err, fmt.Sprintf("Could not get MatchedPolicies for %v", dpKey))
				return
			}
			if proxy.Dataplane.Spec.IsBuiltinGateway() {
				for policy, attachments := range gatewayEntriesByPolicy(proxy) {
					if policy.Type == resType && policy.Key.Name == policyName && policy.Key.Mesh == meshName {
						result.Items = append(result.Items, attachments...)
					}
				}
			} else {
				for policy, attachments := range inspect.GroupByPolicy(&proxy.Policies, dp.Spec.Networking) {
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
			}
		}

		result.Total = uint32(len(result.Items))

		if err := response.WriteAsJson(result); err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Could not write response")
			return
		}
	}
}

func routeDestinationToAPIDestination(des route.Destination) api_server_types.Destination {
	policies := api_server_types.PolicyMap{}
	for kind, p := range des.Policies {
		policies[kind] = rest.From.Meta(p)
	}

	return api_server_types.Destination{
		Tags:     des.Destination,
		Policies: policies,
	}
}

func newDataplaneInspectResponse(matchedPolicies *core_xds.MatchedPolicies, dp *core_mesh.DataplaneResource) []*api_server_types.DataplaneInspectEntry {
	attachmentMap := inspect.GroupByAttachment(matchedPolicies, dp.Spec.Networking)

	entries := make([]*api_server_types.DataplaneInspectEntry, 0, len(attachmentMap))
	attachments := []inspect.Attachment{}
	for attachment := range attachmentMap {
		attachments = append(attachments, attachment)
	}

	sort.Stable(inspect.AttachmentList(attachments))

	for _, attachment := range attachments {
		entry := &api_server_types.DataplaneInspectEntry{
			AttachmentEntry: api_server_types.AttachmentEntry{
				Type:    attachment.Type.String(),
				Name:    attachment.Name,
				Service: attachment.Service,
			},
			MatchedPolicies: map[core_model.ResourceType][]v1alpha1.ResourceMeta{},
		}
		for typ, resList := range attachmentMap[attachment] {
			for _, res := range resList {
				entry.MatchedPolicies[typ] = append(entry.MatchedPolicies[typ], rest.From.Meta(res))
			}
		}

		entries = append(entries, entry)
	}

	return entries
}

func newGatewayDataplaneInspectResponse(
	proxy *core_xds.Proxy,
) api_server_types.GatewayDataplaneInspectResult {
	result := api_server_types.NewGatewayDataplaneInspectResult()
	gwListeners := gateway.ExtractGatewayListeners(proxy)
	for _, port := range maps.SortedKeys(gwListeners) {
		meta := gwListeners[port].Gateway.GetMeta()
		result.Gateway = api_server_types.ResourceKeyEntry{
			Mesh: meta.GetMesh(),
			Name: meta.GetName(),
		}
		break
	}

	for _, info := range gwListeners {
		listener := api_server_types.GatewayListenerInspectEntry{
			Port:     info.Listener.Port,
			Protocol: info.Listener.Protocol.String(),
			Hosts:    []api_server_types.HostInspectEntry{},
		}
		for _, listenerHostname := range info.ListenerHostnames {
			for _, info := range listenerHostname.HostInfos {
				hostInspectEntry := api_server_types.HostInspectEntry{
					HostName: info.Host.Hostname,
					Routes:   []api_server_types.RouteInspectEntry{},
				}
				routeMap := map[string][]api_server_types.Destination{}
				for _, entry := range info.Entries() {
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
						hostInspectEntry.Routes = append(hostInspectEntry.Routes, api_server_types.RouteInspectEntry{
							Route:        name,
							Destinations: destinations,
						})
					}
				}
				sort.SliceStable(hostInspectEntry.Routes, func(i, j int) bool {
					return hostInspectEntry.Routes[i].Route < hostInspectEntry.Routes[j].Route
				})
				listener.Hosts = append(listener.Hosts, hostInspectEntry)
			}
		}
		sort.SliceStable(listener.Hosts, func(i, j int) bool {
			return listener.Hosts[i].HostName < listener.Hosts[j].HostName
		})
		result.Listeners = append(result.Listeners, listener)
	}

	gatewayPolicies := api_server_types.PolicyMap{}

	// TrafficLog and TrafficeTrace are applied to the entire MeshGateway
	// see pkg/plugins/runtime/gateway.newFilterChain
	if logging, ok := proxy.Policies.TrafficLogs[core_mesh.PassThroughService]; ok {
		gatewayPolicies[core_mesh.TrafficLogType] = rest.From.Meta(logging)
	}
	if trace := proxy.Policies.TrafficTrace; trace != nil {
		gatewayPolicies[core_mesh.TrafficTraceType] = rest.From.Meta(trace)
	}

	if len(gatewayPolicies) > 0 {
		result.Policies = gatewayPolicies
	}

	return result
}

func routeToPolicyInspect(
	policyMap map[inspect.PolicyKey][]tags.Tags,
	des route.Destination,
) map[inspect.PolicyKey][]tags.Tags {
	for kind, p := range des.Policies {
		policyKey := inspect.PolicyKey{
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
	proxy *core_xds.Proxy,
) map[inspect.PolicyKey][]api_server_types.PolicyInspectEntry {
	policyMap := map[inspect.PolicyKey][]api_server_types.PolicyInspectEntry{}
	gwListeners := gateway.ExtractGatewayListeners(proxy)

	dpKey := core_model.MetaToResourceKey(proxy.Dataplane.GetMeta())
	resourceKey := api_server_types.ResourceKeyEntry{
		Mesh: dpKey.Mesh,
		Name: dpKey.Name,
	}

	var gatewayKey api_server_types.ResourceKeyEntry
	listenersMap := map[inspect.PolicyKey][]api_server_types.PolicyInspectGatewayListenerEntry{}
	for _, port := range maps.SortedKeys(gwListeners) {
		info := gwListeners[port]
		hostMap := map[inspect.PolicyKey][]api_server_types.PolicyInspectGatewayHostEntry{}
		for _, listenerHostname := range info.ListenerHostnames {
			for _, info := range listenerHostname.HostInfos {
				routeMap := map[inspect.PolicyKey][]api_server_types.PolicyInspectGatewayRouteEntry{}
				for _, entry := range info.Entries() {
					entryMap := map[inspect.PolicyKey][]tags.Tags{}
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
		// This should be identical between all listeners
		gatewayKey = api_server_types.ResourceKeyEntryFromModelKey(core_model.MetaToResourceKey(info.Gateway.GetMeta()))
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
		policyKey := inspect.PolicyKey{
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
		policyKey := inspect.PolicyKey{
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

func inspectRulesAttachment(zoneName string, builder xds_context.MeshContextBuilder) restful.RouteFunction {
	return func(request *restful.Request, response *restful.Response) {
		ctx := request.Request.Context()
		meshName := request.PathParameter("mesh")
		dataplaneName := request.PathParameter("dataplane")

		meshContext, err := builder.Build(ctx, meshName)
		if err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Could not build MeshContext")
			return
		}

		proxy, err := getMatchedPolicies(
			ctx, zoneName, meshContext, core_model.ResourceKey{Mesh: meshName, Name: dataplaneName},
		)
		if err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Could not get MatchedPolicies")
			return
		}
		rulesAttachments := inspect.BuildRulesAttachments(proxy.Policies.Dynamic, proxy.Dataplane.Spec.Networking, meshContext.VIPDomains)
		resp := api_server_types.RuleInspectResponse{
			Items: []api_server_types.RuleInspectEntry{},
		}
		for _, ruleAttachment := range rulesAttachments {
			subset := map[string]string{}
			for _, tag := range ruleAttachment.Rule.Subset {
				value := tag.Value
				if tag.Not {
					value = "!" + value
				}
				subset[tag.Key] = value
			}

			var origins []api_server_types.ResourceKeyEntry
			for _, origin := range ruleAttachment.Rule.Origin {
				origins = append(origins, api_server_types.ResourceKeyEntry{
					Mesh: origin.GetMesh(),
					Name: origin.GetName(),
				})
			}

			resp.Items = append(resp.Items, api_server_types.RuleInspectEntry{
				Type:       ruleAttachment.Type,
				Name:       ruleAttachment.Name,
				Addresses:  ruleAttachment.Addresses,
				Service:    ruleAttachment.Service,
				Tags:       ruleAttachment.Tags,
				PolicyType: string(ruleAttachment.PolicyType),
				Subset:     subset,
				Conf:       ruleAttachment.Rule.Conf,
				Origins:    origins,
			})
		}
		resp.Total = uint32(len(resp.Items))
		if err := response.WriteAsJson(resp); err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Could not write response")
			return
		}
	}
}
