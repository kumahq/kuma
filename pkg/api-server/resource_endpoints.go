package api_server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/emicklei/go-restful/v3"

	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	core_plugins "github.com/kumahq/kuma/v3/pkg/core/plugins"
	"github.com/kumahq/kuma/v3/pkg/core/resources/access"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	rest_errors "github.com/kumahq/kuma/v3/pkg/core/rest/errors"
	"github.com/kumahq/kuma/v3/pkg/core/rest/errors/types"
	meshhttproute_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	meshtcproute_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtcproute/api/v1alpha1"
	util_slices "github.com/kumahq/kuma/v3/pkg/util/slices"
)

const (
	k8sReadOnlyMessage = "On Kubernetes you cannot change the state of Kuma resources with 'kumactl apply' or via the HTTP API." +
		" As a best practice, you should always be using 'kubectl apply' instead." +
		" You can still use 'kumactl' or the HTTP API to make read-only operations. On Universal this limitation does not apply.\n"
	globalReadOnlyMessage = "On global control plane you can not modify dataplane resources with 'kumactl apply' or via the HTTP API." +
		" You can still use 'kumactl' or the HTTP API to modify them on the zone control plane.\n"
	zoneReadOnlyMessage = "On zone control plane you can only modify zone resources with 'kumactl apply' or via the HTTP API." +
		" You can still use 'kumactl' or the HTTP API to modify the rest of the resource on the global control plane.\n"
)

// resourceEndpointsContext holds the dependencies shared by the resource handlers.
type resourceEndpointsContext struct {
	mode           config_core.CpMode
	zoneName       string
	resManager     manager.ResourceManager
	descriptor     core_model.ResourceTypeDescriptor
	resourceAccess access.ResourceAccess
}

// resourceEndpoints registers the resource routes and delegates them to the CRUD
// and inspect handlers.
type resourceEndpoints struct {
	*resourceCrudHandler

	inspect *resourceInspectHandler
}

func typeToLegacyOverviewPath(resourceType core_model.ResourceType) string {
	switch resourceType {
	case core_mesh.ZoneEgressType:
		return "zoneegressoverviews"
	case core_mesh.ZoneIngressType:
		return "zoneingresses+insights"
	case core_mesh.DataplaneType:
		return "dataplanes+insights"
	case system.ZoneType:
		return "zones+insights"
	default:
		return ""
	}
}

func (r *resourceEndpoints) addFindEndpoint(ws *restful.WebService, pathPrefix string) {
	ws.Route(ws.GET(pathPrefix+"/{name}").To(r.findResource(false)).
		Doc(fmt.Sprintf("Get a %s", r.descriptor.WsPath)).
		Param(ws.PathParameter("name", fmt.Sprintf("Name of a %s", r.descriptor.Name)).DataType("string")).
		Returns(200, "OK", nil).
		Returns(404, "Not found", nil))
	if r.descriptor.HasInsights() {
		route := r.findResource(true)
		ws.Route(ws.GET(pathPrefix+"/{name}/_overview").To(route).
			Doc(fmt.Sprintf("Get overview of a %s", r.descriptor.Name)).
			Param(ws.PathParameter("name", fmt.Sprintf("Name of a %s", r.descriptor.Name)).DataType("string")).
			Returns(200, "OK", nil).
			Returns(404, "Not found", nil))
		// Backward compatibility with previous path for overviews
		if legacyPath := typeToLegacyOverviewPath(r.descriptor.Name); legacyPath != "" {
			ws.Route(ws.GET(strings.Replace(pathPrefix, r.descriptor.WsPath, legacyPath, 1)+"/{name}").To(route).
				Doc(fmt.Sprintf("Get overview of a %s", r.descriptor.Name)).
				Param(ws.PathParameter("name", fmt.Sprintf("Name of a %s", r.descriptor.Name)).DataType("string")).
				Param(ws.QueryParameter("name", fmt.Sprintf("Name of a %s", r.descriptor.Name)).DataType("string")).
				Returns(200, "OK", nil).
				Returns(404, "Not found", nil))
		}
	}
	if r.descriptor.IsPolicy {
		ws.Route(ws.GET(pathPrefix+"/{name}/_resources/dataplanes").To(r.inspect.matchingDataplanesForPolicy()).
			Doc(fmt.Sprintf("Get matching dataplanes of a %s", r.descriptor.Name)).
			Param(ws.PathParameter("name", fmt.Sprintf("Name of a %s", r.descriptor.Name)).DataType("string")).
			Returns(200, "OK", nil).
			Returns(404, "Not found", nil))
	}
	if r.descriptor.Name == core_mesh.DataplaneType || r.descriptor.Name == core_mesh.MeshGatewayType {
		ws.Route(ws.GET(pathPrefix+"/{name}/_rules").To(r.inspect.rulesForResource()).
			Doc(fmt.Sprintf("Get matching rules %s", r.descriptor.Name)).
			Param(ws.PathParameter("name", fmt.Sprintf("Name of a %s", r.descriptor.Name)).DataType("string")).
			Returns(200, "OK", nil).
			Returns(404, "Not found", nil))
		if r.mode == config_core.Global {
			msg := "Not allowed on global CP"
			ws.Route(ws.GET(pathPrefix+"/{name}/_config").To(r.methodNotAllowed(msg)).
				Doc(msg).
				Returns(http.StatusMethodNotAllowed, msg, restful.ServiceError{}))
		} else {
			ws.Route(ws.GET(pathPrefix+"/{name}/_config").To(r.inspect.configForProxy()).
				Doc(fmt.Sprintf("Get proxy config%s", r.descriptor.Name)).
				Param(ws.PathParameter("name", fmt.Sprintf("Name of a %s", r.descriptor.Name)).DataType("string")).
				Returns(200, "OK", nil).
				Returns(404, "Not found", nil))
		}
	}
	if r.descriptor.Name == core_mesh.DataplaneType {
		ws.Route(ws.GET(pathPrefix+"/{name}/_policies").To(r.inspect.getPoliciesConf(core_plugins.Plugins().PolicyPlugins(), matchedPoliciesToProxyPolicy)).
			Doc(fmt.Sprintf("Get policy config %s", r.descriptor.Name)).
			Param(ws.PathParameter("name", fmt.Sprintf("Name of a %s", r.descriptor.Name)).DataType("string")).
			Returns(200, "OK", nil).
			Returns(404, "Not found", nil))
		ws.Route(ws.GET(pathPrefix+"/{name}/_inbounds/{inbound_kri}/_policies").To(r.inspect.getPoliciesConf(core_plugins.Plugins().PolicyPlugins(), matchedPoliciesToInboundConfig)).
			Doc("Get policy config for inbound").
			Param(ws.PathParameter("name", fmt.Sprintf("Name of a %s", r.descriptor.Name)).DataType("string")).
			Param(ws.PathParameter("inbound_kri", "KRI of a inbound").DataType("string")).
			Returns(200, "OK", nil).
			Returns(404, "Not found", nil))
		ws.Route(ws.GET(pathPrefix+"/{name}/_outbounds/{outbound_kri}/_policies").To(r.inspect.getPoliciesConf(core_plugins.Plugins().PolicyPlugins(), matchedPoliciesToOutboundPolicy)).
			Doc("Get policy config for outbound").
			Param(ws.PathParameter("name", fmt.Sprintf("Name of a %s", r.descriptor.Name)).DataType("string")).
			Param(ws.PathParameter("outbound_kri", "KRI of a outbound").DataType("string")).
			Returns(200, "OK", nil).
			Returns(404, "Not found", nil))
		ws.Route(ws.GET(pathPrefix+"/{name}/_outbounds/{outbound_kri}/_routes").To(r.inspect.getPoliciesConf(
			util_slices.Filter(core_plugins.Plugins().PolicyPlugins(), func(p core_plugins.RegisteredPolicyPlugin) bool {
				return p.Name == core_plugins.PluginName(meshhttproute_api.MeshHTTPRouteResourceTypeDescriptor.KumactlArg) ||
					p.Name == core_plugins.PluginName(meshtcproute_api.MeshTCPRouteResourceTypeDescriptor.KumactlArg)
			}),
			matchedPoliciesToRoutes,
		)).
			Doc("Get policy config for outbound").
			Param(ws.PathParameter("name", fmt.Sprintf("Name of a %s", r.descriptor.Name)).DataType("string")).
			Param(ws.PathParameter("outbound_kri", "KRI of a outbound").DataType("string")).
			Returns(200, "OK", nil).
			Returns(404, "Not found", nil))
		ws.Route(ws.GET(pathPrefix+"/{name}/_outbounds/{outbound_kri}/_routes/{route_kri}/_policies").To(r.inspect.getPoliciesConf(
			util_slices.Filter(core_plugins.Plugins().PolicyPlugins(), func(p core_plugins.RegisteredPolicyPlugin) bool {
				return p.Name != core_plugins.PluginName(meshhttproute_api.MeshHTTPRouteResourceTypeDescriptor.KumactlArg) &&
					p.Name != core_plugins.PluginName(meshtcproute_api.MeshTCPRouteResourceTypeDescriptor.KumactlArg)
			}), matchedPoliciesToRouteConfig)).
			Doc("Get policy config for route").
			Param(ws.PathParameter("name", fmt.Sprintf("Name of a %s", r.descriptor.Name)).DataType("string")).
			Param(ws.PathParameter("outbound_kri", "KRI of a outbound").DataType("string")).
			Param(ws.PathParameter("route_kri", "KRI of a route").DataType("string")).
			Returns(200, "OK", nil).
			Returns(404, "Not found", nil))
	}
}

func (r *resourceEndpoints) methodNotAllowed(detail string) func(request *restful.Request, response *restful.Response) {
	return func(request *restful.Request, response *restful.Response) {
		err := &types.Error{
			Status: 405,
			Title:  "Method not allowed",
			Detail: detail,
		}
		rest_errors.HandleError(request.Request.Context(), response, err, "")
	}
}

func (r *resourceEndpoints) addListEndpoint(ws *restful.WebService, pathPrefix string) {
	ws.Route(ws.GET(pathPrefix).To(r.listResources(false)).
		Doc(fmt.Sprintf("List of %s", r.descriptor.Name)).
		Param(ws.QueryParameter("size", "size of page").DataType("int")).
		Param(ws.QueryParameter("offset", "offset of page to list").DataType("string")).
		Param(ws.QueryParameter("name", "a pattern to select only resources that contain these characters").DataType("string")).
		Returns(200, "OK", nil))
	if r.descriptor.HasInsights() {
		route := r.listResources(true)
		ws.Route(ws.GET(pathPrefix+"/_overview").To(route).
			Doc(fmt.Sprintf("Get a %s", r.descriptor.WsPath)).
			Param(ws.QueryParameter("size", "size of page").DataType("int")).
			Param(ws.QueryParameter("offset", "offset of page to list").DataType("string")).
			Param(ws.PathParameter("name", "a pattern to select only resources that contain these characters").DataType("string")).
			Returns(200, "OK", nil).
			Returns(404, "Not found", nil))
		// Backward compatibility with previous path for overviews
		if legacyPath := typeToLegacyOverviewPath(r.descriptor.Name); legacyPath != "" {
			ws.Route(ws.GET(strings.Replace(pathPrefix, r.descriptor.WsPath, legacyPath, 1)).To(route).
				Doc(fmt.Sprintf("Get a %s", r.descriptor.WsPath)).
				Param(ws.QueryParameter("name", "a pattern to select only resources that contain these characters").DataType("string")).
				Returns(200, "OK", nil).
				Returns(404, "Not found", nil))
		}
	}
}

func (r *resourceEndpoints) addCreateOrUpdateEndpoint(ws *restful.WebService, pathPrefix string) {
	if r.descriptor.ReadOnly {
		ws.Route(ws.PUT(pathPrefix+"/{name}").To(r.methodNotAllowed(r.readOnlyMessage())).
			Doc("Not allowed in read-only mode.").
			Returns(http.StatusMethodNotAllowed, "Not allowed in read-only mode.", restful.ServiceError{}))
	} else {
		ws.Route(ws.PUT(pathPrefix+"/{name}").To(r.createOrUpdateResource).
			Doc(fmt.Sprintf("Updates a %s", r.descriptor.WsPath)).
			Param(ws.PathParameter("name", fmt.Sprintf("Name of the %s", r.descriptor.WsPath)).DataType("string")).
			Returns(200, "OK", nil).
			Returns(201, "Created", nil))
	}
}

func (r *resourceEndpoints) addDeleteEndpoint(ws *restful.WebService, pathPrefix string) {
	if r.descriptor.ReadOnly {
		ws.Route(ws.DELETE(pathPrefix+"/{name}").To(r.methodNotAllowed(r.readOnlyMessage())).
			Doc("Not allowed in read-only mode.").
			Returns(http.StatusMethodNotAllowed, "Not allowed in read-only mode.", restful.ServiceError{}))
	} else {
		ws.Route(ws.DELETE(pathPrefix+"/{name}").To(r.deleteResource).
			Doc(fmt.Sprintf("Deletes a %s", r.descriptor.Name)).
			Param(ws.PathParameter("name", fmt.Sprintf("Name of a %s", r.descriptor.Name)).DataType("string")).
			Returns(200, "OK", nil))
	}
}

func (r *resourceEndpointsContext) meshFromRequest(request *restful.Request) (string, error) {
	if r.descriptor.Scope == core_model.ScopeMesh {
		meshName := request.PathParameter("mesh")
		if meshName == "" { // Handle lists across all meshes
			return "", nil
		}
		mRes := core_mesh.MeshResourceTypeDescriptor.NewObject()
		if err := r.resManager.Get(request.Request.Context(), mRes, store.GetByKey(meshName, core_model.NoMesh)); err != nil {
			return "", err
		}
		return meshName, nil
	}
	return "", nil
}

func (r *resourceEndpoints) readOnlyMessage() string {
	switch r.mode {
	case config_core.Global:
		return globalReadOnlyMessage
	case config_core.Zone:
		return zoneReadOnlyMessage
	default:
		return k8sReadOnlyMessage
	}
}
