package api_server

import (
	"context"
	"fmt"

	"github.com/emicklei/go-restful/v3"

	api_server_types "github.com/kumahq/kuma/v3/pkg/api-server/types"
	kuma_cp "github.com/kumahq/kuma/v3/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/v3/pkg/core/resources/access"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice"
	meshservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/registry"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	"github.com/kumahq/kuma/v3/pkg/core/xds/inspect"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	"github.com/kumahq/kuma/v3/pkg/xds/envoy"
	"github.com/kumahq/kuma/v3/pkg/xds/sync"
)

// getMatchedPolicies returns information about sidecar dataplanes as well as
// the proxy and a potential error.
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
	for _, desc := range registry.Global().ObjectDescriptors(core_model.AllowedToInspect()) {
		ws.Route(
			ws.GET(fmt.Sprintf("/meshes/{mesh}/%s/{name}/dataplanes", desc.WsPath)).To(handle(inspectPolicies(desc.Name, builder, cfg))).
				Doc("inspect policies").
				Param(ws.PathParameter("mesh", "mesh name").DataType("string")).
				Param(ws.PathParameter("name", "resource name").DataType("string")).
				Returns(200, "OK", nil),
		)
	}

	ws.Route(
		ws.GET("/meshes/{mesh}/meshservices/{name}/_dataplanes").To(handle(inspectMeshServiceDataplanes(rm, resourceAccess))).
			Doc("inspect MeshService").
			Param(ws.PathParameter("mesh", "mesh name").DataType("string")).
			Param(ws.PathParameter("name", "resource name").DataType("string")).
			Returns(200, "OK", nil),
	)
}

// inspectMeshServiceDataplanes provides the standardized /_dataplanes endpoint.
// Uses exact tag matching via meshservice.MatchesDataplane() to fix multizone aggregation issues.
func inspectMeshServiceDataplanes(
	rm manager.ResourceManager,
	resourceAccess access.ResourceAccess,
) handlerFunc {
	return func(request *restful.Request) (any, error) {
		return matchingDataplanesForFilter(
			request,
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
) handlerFunc {
	return func(request *restful.Request) (any, error) {
		ctx := request.Request.Context()
		meshName := request.PathParameter("mesh")
		policyName := request.PathParameter("name")

		meshContext, err := builder.Build(ctx, meshName)
		if err != nil {
			return nil, withTitle(err, "Could not list Dataplanes")
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
				return nil, withTitle(err, fmt.Sprintf("Could not get MatchedPolicies for %v", dpKey))
			}
			for policy, attachments := range inspect.GroupByPolicy(&proxy.Policies) {
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

		return result, nil
	}
}
