package api_server

import (
	"context"
	"net/http"
	"slices"
	"strings"

	"github.com/emicklei/go-restful/v3"

	"github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/api/openapi/types"
	api_common "github.com/kumahq/kuma/api/openapi/types/common"
	"github.com/kumahq/kuma/pkg/core/kri"
	"github.com/kumahq/kuma/pkg/core/resources/access"
	"github.com/kumahq/kuma/pkg/core/resources/apis/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	rest_errors "github.com/kumahq/kuma/pkg/core/rest/errors"
	"github.com/kumahq/kuma/pkg/core/user"
	util_slices "github.com/kumahq/kuma/pkg/util/slices"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

type dataplaneLayoutEndpoint struct {
	resManager         manager.ResourceManager
	meshContextBuilder xds_context.MeshContextBuilder
	resourceAccess     access.ResourceAccess
}

func newDataplaneLayoutEndpoint(resManager manager.ResourceManager, meshContextBuilder xds_context.MeshContextBuilder, resourceAccess access.ResourceAccess) *dataplaneLayoutEndpoint {
	return &dataplaneLayoutEndpoint{
		resManager:         resManager,
		meshContextBuilder: meshContextBuilder,
		resourceAccess:     resourceAccess,
	}
}

func (dle *dataplaneLayoutEndpoint) addEndpoint(ws *restful.WebService) {
	ws.Route(
		ws.GET("/meshes/{mesh}/dataplanes/{name}/_layout").
			To(dle.getLayout).
			Doc("Get Dataplane Layout").
			Returns(http.StatusOK, "OK", nil),
	)
}

func (dle *dataplaneLayoutEndpoint) getLayout(request *restful.Request, response *restful.Response) {
	meshName := request.PathParameter("mesh")
	dataplaneName := request.PathParameter("name")

	if err := dle.resourceAccess.ValidateGet(
		request.Request.Context(),
		core_model.ResourceKey{Mesh: meshName, Name: dataplaneName},
		core_mesh.DataplaneResourceTypeDescriptor,
		user.FromCtx(request.Request.Context()),
	); err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Access Denied")
		return
	}

	baseMeshContext, err := dle.meshContextBuilder.BuildBaseMeshContextIfChanged(request.Request.Context(), meshName, nil)
	if err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Failed to build MeshContext")
	}

	if baseMeshContext.Mesh.Spec.GetMeshServices().GetMode() == v1alpha1.Mesh_MeshServices_Disabled {
		rest_errors.HandleError(request.Request.Context(), response, rest_errors.NewBadRequestError("can't use _layout endpoint without meshService enabled"), "Bad Request")
		return
	}

	dataplane, err := dle.getDataplaneResource(request.Request.Context(), meshName, dataplaneName)
	if err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Failed to retrieve Dataplane")
		return
	}

	inbounds := util_slices.Map(dataplane.Spec.GetNetworking().GetInbound(), func(inbound *v1alpha1.Dataplane_Networking_Inbound) api_common.DataplaneInbound {
		sectionName := inbound.GetSectionName()

		return api_common.DataplaneInbound{
			Kri:      kri.From(dataplane, sectionName).String(),
			Port:     int32(inbound.GetPort()),
			Protocol: inbound.GetProtocol(),
		}
	})

	var outbounds []api_common.DataplaneOutbound
	if dataplane.Spec.GetNetworking().GetOutbound() != nil {
		// TODO handle user defined outbounds on universal without transparent proxy
	} else {
		// handle reachable backend and outbounds when using transparent proxy
		reachableBackends := baseMeshContext.DestinationIndex.GetReachableBackends(baseMeshContext.Mesh, dataplane)
		outbounds = getOutbounds(baseMeshContext.DestinationIndex, reachableBackends)
	}

	slices.SortStableFunc(outbounds, func(a, b api_common.DataplaneOutbound) int {
		return strings.Compare(a.Kri, b.Kri)
	})

	networkingLayout := types.DataplaneNetworkingLayout{
		Inbounds:  inbounds,
		Kri:       kri.From(dataplane, "").String(),
		Labels:    dataplane.GetMeta().GetLabels(),
		Outbounds: outbounds,
	}

	err = response.WriteHeaderAndJson(http.StatusOK, networkingLayout, "application/json")
	if err != nil {
		log.Error(err, "Could not write response")
	}
}

func (dle *dataplaneLayoutEndpoint) getDataplaneResource(ctx context.Context, meshName string, dataplaneName string) (*core_mesh.DataplaneResource, error) {
	dataplaneResource := core_mesh.NewDataplaneResource()
	err := dle.resManager.Get(ctx, dataplaneResource, store.GetByKey(dataplaneName, meshName))
	if err != nil {
		return nil, err
	}
	return dataplaneResource, nil
}

func getOutbounds(destinationContext *xds_context.DestinationIndex, reachableBackends *xds_context.ReachableBackends) []api_common.DataplaneOutbound {
	var outbounds []api_common.DataplaneOutbound
	if reachableBackends == nil {
		for destinationKri, destination := range destinationContext.GetAllDestinations() {
			outbounds = append(outbounds, expandDestination(destinationKri, destination)...)
		}
		return outbounds
	}

	for destinationKri := range *reachableBackends {
		dest := destinationContext.GetDestinationByKri(destinationKri)
		if dest == nil {
			continue
		}
		if destinationKri.HasSectionName() {
			port, ok := dest.FindPortByName(destinationKri.SectionName)
			if !ok {
				continue
			}
			outbounds = append(outbounds, api_common.DataplaneOutbound{
				Kri:      destinationKri.String(),
				Port:     port.GetValue(),
				Protocol: string(port.GetProtocol()),
			})
		} else {
			outbounds = append(outbounds, expandDestination(destinationKri, dest)...)
		}
	}

	return outbounds
}

func expandDestination(destinationKri kri.Identifier, dest core.Destination) []api_common.DataplaneOutbound {
	var outbounds []api_common.DataplaneOutbound
	for _, port := range dest.GetPorts() {
		outbounds = append(outbounds, api_common.DataplaneOutbound{
			Kri:      kri.WithSectionName(destinationKri, port.GetName()).String(),
			Port:     port.GetValue(),
			Protocol: string(port.GetProtocol()),
		})
	}
	return outbounds
}
