package api_server

import (
	"context"
	"fmt"
	"net"
	"sort"

	"github.com/emicklei/go-restful"

	api_server_types "github.com/kumahq/kuma/pkg/api-server/types"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/core"
	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	"github.com/kumahq/kuma/pkg/core/datasource"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	rest_errors "github.com/kumahq/kuma/pkg/core/rest/errors"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/sync"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

type simpleMatchedPolicyGetter struct {
	cfg        *kuma_cp.Config
	resManager core_manager.ResourceManager
	cfgManager config_manager.ConfigManager
}

func NewSimpleMatchedPolicyGetter(cfg *kuma_cp.Config, resManager core_manager.ResourceManager, cfgManager config_manager.ConfigManager) core_xds.MatchedPoliciesGetter {
	return &simpleMatchedPolicyGetter{
		cfg:        cfg,
		resManager: resManager,
		cfgManager: cfgManager,
	}
}

type fakeMetadataTracker struct {
}

func (f fakeMetadataTracker) Metadata(dpKey core_model.ResourceKey) *core_xds.DataplaneMetadata {
	return nil
}

func (s *simpleMatchedPolicyGetter) Get(ctx context.Context, dataplaneKey core_model.ResourceKey) (*core_xds.MatchedPolicies, error) {
	dataplane := core_mesh.NewDataplaneResource()
	if err := s.resManager.Get(ctx, dataplane, store.GetBy(dataplaneKey)); err != nil {
		return nil, err
	}

	mesh := core_mesh.NewMeshResource()
	if err := s.resManager.Get(ctx, mesh, store.GetByKey(dataplaneKey.Mesh, core_model.NoMesh)); err != nil {
		return nil, err
	}

	dataplanes, err := xds_topology.GetDataplanes(core.Log, ctx, s.resManager, net.LookupIP, dataplaneKey.Mesh)
	if err != nil {
		return nil, err
	}

	// todo(lobkovilya): share DataplaneProxyBuilder with xDS code (instead of creating a new one)
	proxyBuilder := sync.DefaultDataplaneProxyBuilder(s.resManager, s.resManager, net.LookupIP,
		datasource.NewDataSourceLoader(s.resManager), *s.cfg, s.cfgManager, &fakeMetadataTracker{}, envoy.APIV3)
	proxy, err := proxyBuilder.Build(core_model.MetaToResourceKey(dataplane.GetMeta()), &xds_context.Context{Mesh: xds_context.MeshContext{
		Resource:   mesh,
		Dataplanes: dataplanes,
	}})
	if err != nil {
		return nil, err
	}
	return &proxy.Policies, nil
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
	resManager core_manager.ResourceManager,
	mpg core_xds.MatchedPoliciesGetter,
) {
	ws.Route(
		ws.GET("/meshes/{mesh}/dataplanes/{dataplane}/policies").To(inspectDataplane(mpg)).
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
			ws.GET(fmt.Sprintf("/meshes/{mesh}/%s/{name}/dataplanes", def.WsPath)).To(inspectPolicies(def.Name, mpg, resManager)).
				Doc("inspect policies").
				Param(ws.PathParameter("mesh", "mesh name").DataType("string")).
				Param(ws.PathParameter("name", "resource name").DataType("string")).
				Returns(200, "OK", nil),
		)
	}
}

func inspectDataplane(mpg core_xds.MatchedPoliciesGetter) restful.RouteFunction {
	return func(request *restful.Request, response *restful.Response) {
		meshName := request.PathParameter("mesh")
		dataplaneName := request.PathParameter("dataplane")

		matchedPolicies, err := mpg.Get(context.Background(), core_model.ResourceKey{Mesh: meshName, Name: dataplaneName})
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
	mpg core_xds.MatchedPoliciesGetter,
	rm core_manager.ResourceManager,
) restful.RouteFunction {
	return func(request *restful.Request, response *restful.Response) {
		meshName := request.PathParameter("mesh")
		policyName := request.PathParameter("name")

		dataplanes := &core_mesh.DataplaneResourceList{}
		if err := rm.List(context.Background(), dataplanes, store.ListByMesh(meshName)); err != nil {
			rest_errors.HandleError(response, err, "Could not list Dataplanes")
			return
		}

		result := []api_server_types.PolicyInspectEntry{}

		for _, dp := range dataplanes.Items {
			dpKey := core_model.MetaToResourceKey(dp.GetMeta())
			matchedPolicies, err := mpg.Get(context.Background(), dpKey)
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
					result = append(result, api_server_types.PolicyInspectEntry{
						DataplaneKey: api_server_types.ResourceKeyEntry{
							Mesh: dpKey.Mesh,
							Name: dpKey.Name,
						},
						Attachments: attachmentList,
					})
				}
			}
		}

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
