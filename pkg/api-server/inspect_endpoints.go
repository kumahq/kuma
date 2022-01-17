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
	rest_errors "github.com/kumahq/kuma/pkg/core/rest/errors"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/sync"
)

type fakeMetadataTracker struct {
}

func (f fakeMetadataTracker) Metadata(dpKey core_model.ResourceKey) *core_xds.DataplaneMetadata {
	return nil
}

type fakeDataSourceLoader struct {
}

func (f fakeDataSourceLoader) Load(ctx context.Context, mesh string, source *system_proto.DataSource) ([]byte, error) {
	return []byte("secret"), nil
}

func getMatchedPolicies(cfg *kuma_cp.Config, meshContext xds_context.MeshContext, dataplaneKey core_model.ResourceKey) (*core_xds.MatchedPolicies, error) {
	proxyBuilder := sync.DefaultDataplaneProxyBuilder(
		&fakeDataSourceLoader{},
		*cfg,
		&fakeMetadataTracker{},
		envoy.APIV3)
	if proxy, err := proxyBuilder.Build(dataplaneKey, &xds_context.Context{
		Mesh: meshContext,
	}); err != nil {
		return nil, err
	} else {
		return &proxy.Policies, nil
	}
}

var policies = map[core_model.ResourceType]bool{
	core_mesh.TrafficPermissionType: true,
	core_mesh.FaultInjectionType:    true,
	core_mesh.RateLimitType:         true,
	core_mesh.TrafficLogType:        true,
	core_mesh.HealthCheckType:       true,
	core_mesh.CircuitBreakerType:    true,
	core_mesh.RetryType:             true,
	core_mesh.TimeoutType:           true,
	core_mesh.TrafficTraceType:      true,
}

func addInspectEndpoints(
	ws *restful.WebService,
	defs []core_model.ResourceTypeDescriptor,
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

	for _, def := range defs {
		if !policies[def.Name] {
			continue
		}
		ws.Route(
			ws.GET(fmt.Sprintf("/meshes/{mesh}/%s/{name}/dataplanes", def.WsPath)).To(inspectPolicies(def.Name, builder, cfg)).
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

		matchedPolicies, err := getMatchedPolicies(cfg, meshContext, core_model.ResourceKey{Mesh: meshName, Name: dataplaneName})
		if err != nil {
			rest_errors.HandleError(response, err, "Could not get MatchedPolicies")
			return
		}

		entries := newDataplaneInspectResponse(matchedPolicies)

		if err := response.WriteAsJson(entries); err != nil {
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
			matchedPolicies, err := getMatchedPolicies(cfg, meshContext, dpKey)
			if err != nil {
				rest_errors.HandleError(response, err, fmt.Sprintf("Could not get MatchedPolicies for %v", dpKey))
				return
			}
			for policy, attachments := range core_xds.GroupByPolicy(matchedPolicies) {
				if policy.Type == resType && policy.Key.Name == policyName && policy.Key.Mesh == meshName {
					attachmentList := []api_server_types.AttachmentEntry{}
					for _, attachment := range attachments {
						attachmentList = append(attachmentList, api_server_types.AttachmentEntry{
							Type: attachment.Type.String(),
							Name: attachment.Name,
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

func newDataplaneInspectResponse(matchedPolicies *core_xds.MatchedPolicies) []api_server_types.DataplaneInspectEntry {
	attachmentMap := core_xds.GroupByAttachment(matchedPolicies)

	entries := make([]api_server_types.DataplaneInspectEntry, 0, len(attachmentMap))
	attachments := []core_xds.Attachment{}
	for attachment := range attachmentMap {
		attachments = append(attachments, attachment)
	}

	sort.Stable(core_xds.AttachmentList(attachments))

	for _, attachment := range attachments {
		entry := api_server_types.DataplaneInspectEntry{
			AttachmentEntry: api_server_types.AttachmentEntry{
				Type: attachment.Type.String(),
				Name: attachment.Name,
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
