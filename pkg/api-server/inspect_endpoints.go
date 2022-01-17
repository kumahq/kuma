package api_server

import (
	"context"
	"net"
	"sort"

	"github.com/emicklei/go-restful"

	api_server_types "github.com/kumahq/kuma/pkg/api-server/types"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	"github.com/kumahq/kuma/pkg/core/datasource"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	rest_errors "github.com/kumahq/kuma/pkg/core/rest/errors"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/dns/vips"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/server"
	"github.com/kumahq/kuma/pkg/xds/sync"
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

	meshContextBuilder := xds_context.NewMeshContextBuilder(
		s.resManager,
		server.MeshResourceTypes(server.HashMeshExcludedResources),
		net.LookupIP,
		s.cfg.Multizone.Zone.Name,
		vips.NewPersistence(s.resManager, s.cfgManager),
		s.cfg.DNSServer.Domain,
	)

	meshContext, err := meshContextBuilder.Build(ctx, dataplaneKey.Mesh)
	if err != nil {
		return nil, err
	}

	// todo(lobkovilya): share DataplaneProxyBuilder with xDS code (instead of creating a new one)
	proxyBuilder := sync.DefaultDataplaneProxyBuilder(datasource.NewDataSourceLoader(s.resManager),
		*s.cfg, &fakeMetadataTracker{}, envoy.APIV3)
	proxy, err := proxyBuilder.Build(core_model.MetaToResourceKey(dataplane.GetMeta()), &xds_context.Context{
		Mesh: meshContext,
	})
	if err != nil {
		return nil, err
	}
	return &proxy.Policies, nil
}

func addInspectEndpoints(ws *restful.WebService, mpg core_xds.MatchedPoliciesGetter) {
	ws.Route(
		ws.GET("/meshes/{mesh}/dataplanes/{dataplane}/policies").To(inspectDataplane(mpg)).
			Doc("inspect dataplane matched policies").
			Param(ws.PathParameter("mesh", "mesh name").DataType("string")).
			Param(ws.PathParameter("dataplane", "dataplane name").DataType("string")).
			Returns(200, "OK", nil),
	)
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
func newDataplaneInspectResponse(matchedPolicies *core_xds.MatchedPolicies) []api_server_types.InspectEntry {
	attachmentMap := core_xds.GroupByAttachment(matchedPolicies)

	entries := make([]api_server_types.InspectEntry, 0, len(attachmentMap))
	attachments := []core_xds.Attachment{}
	for attachment := range attachmentMap {
		attachments = append(attachments, attachment)
	}

	sort.Stable(core_xds.AttachmentList(attachments))

	for _, attachment := range attachments {
		entry := api_server_types.InspectEntry{
			Type:            attachment.Type.String(),
			Name:            attachment.Name,
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
