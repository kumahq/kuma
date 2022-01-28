package api_server

import (
	"context"
	"fmt"
	"sort"

	"github.com/emicklei/go-restful"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	api_server_types "github.com/kumahq/kuma/pkg/api-server/types"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	rest_errors "github.com/kumahq/kuma/pkg/core/rest/errors"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/server/callbacks"
	"github.com/kumahq/kuma/pkg/xds/sync"
)

type fakeDataSourceLoader struct {
}

func (f fakeDataSourceLoader) Load(ctx context.Context, mesh string, source *system_proto.DataSource) ([]byte, error) {
	return []byte("secret"), nil
}

func getMatchedPolicies(cfg *kuma_cp.Config, meshContext xds_context.MeshContext, dataplaneKey core_model.ResourceKey) (*core_xds.MatchedPolicies, *core_mesh.DataplaneResource, error) {
	proxyBuilder := sync.DefaultDataplaneProxyBuilder(
		&fakeDataSourceLoader{},
		*cfg,
		callbacks.NewDataplaneMetadataTracker(),
		envoy.APIV3)
	if proxy, err := proxyBuilder.Build(dataplaneKey, meshContext); err != nil {
		return nil, nil, err
	} else {
		return &proxy.Policies, proxy.Dataplane, nil
	}
}

func addInspectEndpoints(
	ws *restful.WebService,
	cfg *kuma_cp.Config,
	builder xds_context.MeshContextBuilder,
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
}

func inspectDataplane(cfg *kuma_cp.Config, builder xds_context.MeshContextBuilder) restful.RouteFunction {
	return func(request *restful.Request, response *restful.Response) {
		meshName := request.PathParameter("mesh")
		dataplaneName := request.PathParameter("dataplane")

		meshContext, err := builder.Build(context.Background(), meshName)
		if err != nil {
			rest_errors.HandleError(response, err, "Could not build MeshContext")
			return
		}

		matchedPolicies, dp, err := getMatchedPolicies(cfg, meshContext, core_model.ResourceKey{Mesh: meshName, Name: dataplaneName})
		if err != nil {
			rest_errors.HandleError(response, err, "Could not get MatchedPolicies")
			return
		}

		entries := newDataplaneInspectResponse(matchedPolicies, dp)
		result := &api_server_types.DataplaneInspectEntryList{
			Items: entries,
			Total: uint32(len(entries)),
		}

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
		meshName := request.PathParameter("mesh")
		policyName := request.PathParameter("name")

		meshContext, err := builder.Build(context.Background(), meshName)
		if err != nil {
			rest_errors.HandleError(response, err, "Could not list Dataplanes")
			return
		}

		result := &api_server_types.PolicyInspectEntryList{}

		for _, dp := range meshContext.Resources.Dataplanes().Items {
			dpKey := core_model.MetaToResourceKey(dp.GetMeta())
			matchedPolicies, _, err := getMatchedPolicies(cfg, meshContext, dpKey)
			if err != nil {
				rest_errors.HandleError(response, err, fmt.Sprintf("Could not get MatchedPolicies for %v", dpKey))
				return
			}
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
					result.Items = append(result.Items, &api_server_types.PolicyInspectEntry{
						DataplaneKey: api_server_types.ResourceKeyEntry{
							Mesh: dpKey.Mesh,
							Name: dpKey.Name,
						},
						Attachments: attachmentList,
					})
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
