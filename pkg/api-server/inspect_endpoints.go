package api_server

import (
	"context"
	"fmt"
	"sort"

	"github.com/emicklei/go-restful/v3"

	api_server_types "github.com/kumahq/kuma/v3/pkg/api-server/types"
	kuma_cp "github.com/kumahq/kuma/v3/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/v3/pkg/core/resources/access"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice"
	meshservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model/rest/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/registry"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	rest_errors "github.com/kumahq/kuma/v3/pkg/core/rest/errors"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	"github.com/kumahq/kuma/v3/pkg/core/xds/inspect"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	"github.com/kumahq/kuma/v3/pkg/xds/envoy"
	"github.com/kumahq/kuma/v3/pkg/xds/sync"
)

// getMatchedPolicies returns information about either sidecar dataplanes or
// builtin gateway dataplanes as well as the proxy and a potential error.
func getMatchedPolicies(
	ctx context.Context, cfg *kuma_cp.Config, meshContext xds_context.MeshContext, dataplaneKey core_model.ResourceKey,
) (
	*core_xds.Proxy, error,
) {
	proxyBuilder := sync.DefaultDataplaneProxyBuilder(*cfg, envoy.APIV3)
	proxy, err := proxyBuilder.Build(ctx, dataplaneKey, &core_xds.DataplaneMetadata{}, meshContext)
	if err != nil {
		return nil, err
	}
	return proxy, nil
}

func addInspectEndpoints(
	ws *restful.WebService,
	cfg *kuma_cp.Config,
	builder xds_context.MeshContextBuilder,
	rm manager.ResourceManager,
	resourceAccess access.ResourceAccess,
) {
	ws.Route(
		ws.GET("/meshes/{mesh}/dataplanes/{dataplane}/policies").To(inspectDataplane(cfg, builder)).
			Doc("inspect dataplane matched policies").
			Param(ws.PathParameter("mesh", "mesh name").DataType("string")).
			Param(ws.PathParameter("dataplane", "dataplane name").DataType("string")).
			Returns(200, "OK", nil),
	)

	ws.Route(
		ws.GET("/meshes/{mesh}/dataplanes/{dataplane}/rules").To(inspectRulesAttachment(cfg, builder)).
			Doc("inspect dataplane matched rules").
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
		ws.GET("/meshes/{mesh}/meshservices/{name}/_dataplanes").To(inspectMeshServiceDataplanes(rm, resourceAccess)).
			Doc("inspect MeshService").
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
			rest_errors.HandleError(request.Request.Context(), response, err, "Could not build MeshContext")
			return
		}

		proxy, err := getMatchedPolicies(
			request.Request.Context(), cfg, meshContext, core_model.ResourceKey{Mesh: meshName, Name: dataplaneName},
		)
		if err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Could not get MatchedPolicies")
			return
		}

		inner := api_server_types.NewDataplaneInspectEntryList()
		inner.Items = append(inner.Items, newDataplaneInspectResponse(&proxy.Policies, proxy.Dataplane)...)
		inner.Total = uint32(len(inner.Items))
		result := api_server_types.NewDataplaneInspectResponse(inner)
		if err := response.WriteAsJson(result); err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Could not write response")
			return
		}
	}
}

// inspectMeshServiceDataplanes provides standardized /_dataplanes endpoint following MeshGateway/MeshGatewayRoute pattern.
// Uses exact tag matching via meshservice.MatchesDataplane() to fix multizone aggregation issues.
// Legacy endpoint /meshes/{mesh}/meshservices/{name}/_resources/dataplanes (inspect_mesh_service.go:38) remains for backward compatibility.
func inspectMeshServiceDataplanes(
	rm manager.ResourceManager,
	resourceAccess access.ResourceAccess,
) restful.RouteFunction {
	return func(request *restful.Request, response *restful.Response) {
		matchingDataplanesForFilter(
			request,
			response,
			meshservice_api.MeshServiceResourceTypeDescriptor,
			rm,
			resourceAccess,
			func(resource core_model.Resource) store.ListFilterFunc {
				meshService := resource.(*meshservice_api.MeshServiceResource)
				return func(rs core_model.Resource) bool {
					return meshservice.MatchesDataplane(meshService.Spec, rs.(*core_mesh.DataplaneResource))
				}
			},
		)
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
			proxy, err := getMatchedPolicies(request.Request.Context(), cfg, meshContext, dpKey)
			if err != nil {
				rest_errors.HandleError(request.Request.Context(), response, err, fmt.Sprintf("Could not get MatchedPolicies for %v", dpKey))
				return
			}
			for policy, attachments := range inspect.GroupByPolicy(&proxy.Policies, dp.Spec.Networking) {
				if policy.Type != resType || policy.Key.Name != policyName || policy.Key.Mesh != meshName {
					continue
				}
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

		result.Total = uint32(len(result.Items))

		if err := response.WriteAsJson(result); err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Could not write response")
			return
		}
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

func inspectRulesAttachment(cfg *kuma_cp.Config, builder xds_context.MeshContextBuilder) restful.RouteFunction {
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
			ctx, cfg, meshContext, core_model.ResourceKey{Mesh: meshName, Name: dataplaneName},
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
