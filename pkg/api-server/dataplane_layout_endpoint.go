package api_server

import (
	"net/http"
	"slices"
	"strings"

	"github.com/emicklei/go-restful/v3"

	"github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/api/openapi/types"
	api_common "github.com/kumahq/kuma/v3/api/openapi/types/common"
	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	"github.com/kumahq/kuma/v3/pkg/core/naming"
	"github.com/kumahq/kuma/v3/pkg/core/resources/access"
	core_resources "github.com/kumahq/kuma/v3/pkg/core/resources/apis/core"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	meshidentity_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshidentity/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	rest_errors "github.com/kumahq/kuma/v3/pkg/core/rest/errors"
	"github.com/kumahq/kuma/v3/pkg/core/user"
	util_slices "github.com/kumahq/kuma/v3/pkg/util/slices"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	"github.com/kumahq/kuma/v3/pkg/xds/generator/zoneproxy"
)

type dataplaneLayoutEndpoint struct {
	resManager         manager.ReadOnlyResourceManager
	meshContextBuilder xds_context.MeshContextBuilder
	resourceAccess     access.ResourceAccess
	zone               string
	environment        config_core.EnvironmentType
}

func newDataplaneLayoutEndpoint(
	resManager manager.ReadOnlyResourceManager,
	meshContextBuilder xds_context.MeshContextBuilder,
	resourceAccess access.ResourceAccess,
	zone string,
	environment config_core.EnvironmentType,
) *dataplaneLayoutEndpoint {
	return &dataplaneLayoutEndpoint{
		resManager:         resManager,
		meshContextBuilder: meshContextBuilder,
		resourceAccess:     resourceAccess,
		zone:               zone,
		environment:        environment,
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
		return
	}

	dataplane := core_mesh.NewDataplaneResource()
	err = dle.resManager.Get(request.Request.Context(), dataplane, store.GetByKey(dataplaneName, meshName))
	if err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Failed to retrieve Dataplane")
		return
	}

	inbounds := util_slices.Map(dataplane.Spec.GetNetworking().GetInbound(), func(inbound *v1alpha1.Dataplane_Networking_Inbound) api_common.DataplaneInbound {
		return api_common.DataplaneInbound{
			Kri:               kri.WithSectionName(kri.From(dataplane), inbound.GetSectionName()).String(),
			Port:              int32(inbound.GetPort()),
			Protocol:          inbound.GetProtocol(),
			ProxyResourceName: naming.MustContextualInboundName(dataplane, inbound.GetSectionName()),
		}
	})

	meshResources := baseMeshContext.Resources()
	listeners := []api_common.DataplaneListener{}
	for _, listener := range dataplane.Spec.GetNetworking().GetListeners() {
		sectionName := listener.GetSectionName()
		var listenerType api_common.DataplaneListenerType
		var proxyResourceName string
		var clusters []api_common.DataplaneListenerCluster
		switch listener.GetType() {
		case v1alpha1.Dataplane_Networking_Listener_ZoneIngress:
			listenerType = api_common.ZoneIngress
			proxyResourceName = naming.ContextualZoneIngressListenerName(sectionName)
			clusters = clustersForDestinations(zoneproxy.IngressDestinations(meshResources), allPorts)
		case v1alpha1.Dataplane_Networking_Listener_ZoneEgress:
			listenerType = api_common.ZoneEgress
			proxyResourceName = naming.ContextualZoneEgressListenerName(sectionName)
			clusters = clustersForDestinations(zoneproxy.EgressDestinations(meshResources), firstPort)
		default:
			continue
		}
		listeners = append(listeners, api_common.DataplaneListener{
			Kri:               kri.WithSectionName(kri.From(dataplane), sectionName).String(),
			Type:              listenerType,
			Port:              int32(listener.GetPort()),
			ProxyResourceName: proxyResourceName,
			Clusters:          clusters,
		})
	}

	outbounds := []api_common.DataplaneOutbound{}
	reachableOutbounds, _ := baseMeshContext.DestinationIndex.GetReachableBackends(dataplane)
	for outboundKri, port := range reachableOutbounds {
		outbounds = append(outbounds, api_common.DataplaneOutbound{
			Kri:               outboundKri.String(),
			Port:              port.GetValue(),
			Protocol:          string(port.GetProtocol()),
			ProxyResourceName: outboundKri.String(),
		})
	}

	slices.SortStableFunc(outbounds, func(a, b api_common.DataplaneOutbound) int {
		return strings.Compare(a.Kri, b.Kri)
	})

	networkingLayout := types.DataplaneNetworkingLayout{
		Inbounds:  inbounds,
		Kri:       kri.From(dataplane).String(),
		Labels:    dataplane.GetMeta().GetLabels(),
		Listeners: listeners,
		Outbounds: outbounds,
		SpiffeId:  dle.computeSpiffeID(request, meshName, dataplane),
	}

	err = response.WriteHeaderAndJson(http.StatusOK, networkingLayout, "application/json")
	if err != nil {
		log.Error(err, "Could not write response")
	}
}

func allPorts(destination core_resources.Destination) []core_resources.Port {
	return destination.GetPorts()
}

// firstPort matches the egress generator, which builds a single filter chain
// per MeshExternalService off its first port.
func firstPort(destination core_resources.Destination) []core_resources.Port {
	if ports := destination.GetPorts(); len(ports) > 0 {
		return ports[:1]
	}
	return nil
}

// clustersForDestinations reports the clusters a zone proxy listener serves.
// Like the outbounds above it describes what is configured, so a destination
// with no endpoints yet is still listed.
func clustersForDestinations(
	destinations []core_resources.Destination,
	ports func(core_resources.Destination) []core_resources.Port,
) []api_common.DataplaneListenerCluster {
	clusters := []api_common.DataplaneListenerCluster{}
	for _, destination := range destinations {
		origin := kri.From(destination)
		for _, port := range ports(destination) {
			id := kri.WithSectionName(origin, port.GetName()).String()
			clusters = append(clusters, api_common.DataplaneListenerCluster{
				Kri:               id,
				Port:              port.GetValue(),
				Protocol:          string(port.GetProtocol()),
				ProxyResourceName: id,
			})
		}
	}
	slices.SortStableFunc(clusters, func(a, b api_common.DataplaneListenerCluster) int {
		return strings.Compare(a.Kri, b.Kri)
	})
	return clusters
}

func (dle *dataplaneLayoutEndpoint) computeSpiffeID(request *restful.Request, meshName string, dataplane *core_mesh.DataplaneResource) *string {
	ctx := request.Request.Context()
	identities := &meshidentity_api.MeshIdentityResourceList{}
	if err := dle.resManager.List(ctx, identities, store.ListByMesh(meshName)); err != nil {
		log.V(1).Info("could not list MeshIdentity resources", "mesh", meshName, "err", err)
		return nil
	}
	identity, ok := meshidentity_api.BestMatched(dataplane.GetMeta().GetLabels(), identities.Items)
	if !ok || !identity.Status.IsInitialized() {
		return nil
	}
	trustDomain, err := identity.Spec.GetTrustDomain(identity.GetMeta(), dle.zone)
	if err != nil {
		log.V(1).Info("could not get trust domain", "mesh", meshName, "dataplane", dataplane.GetMeta().GetName(), "err", err)
		return nil
	}
	spiffeID, err := identity.Spec.GetSpiffeID(trustDomain, dataplane.GetMeta(), dle.environment)
	if err != nil {
		log.V(1).Info("could not compute SPIFFE ID", "mesh", meshName, "dataplane", dataplane.GetMeta().GetName(), "err", err)
		return nil
	}
	return &spiffeID
}
