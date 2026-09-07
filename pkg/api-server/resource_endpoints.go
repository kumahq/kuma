package api_server

import (
	"fmt"
	"net/http"

	"github.com/emicklei/go-restful/v3"

	"github.com/kumahq/kuma/v3/pkg/api-server/authn"
	"github.com/kumahq/kuma/v3/pkg/api-server/filters"
	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	core_plugins "github.com/kumahq/kuma/v3/pkg/core/plugins"
	"github.com/kumahq/kuma/v3/pkg/core/resources/access"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/core/rest/errors/types"
	"github.com/kumahq/kuma/v3/pkg/core/runtime"
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
	mode                  config_core.CpMode
	zoneName              string
	resManager            manager.ResourceManager
	descriptor            core_model.ResourceTypeDescriptor
	resourceAccess        access.ResourceAccess
	routeMetadataProvider runtime.RouteMetadataProvider
}

// resourceEndpoints registers the resource routes and delegates them to the CRUD
// and inspect handlers.
type resourceEndpoints struct {
	*resourceCrudHandler

	inspect *resourceInspectHandler
}

// reservedRouteMetadataKeys are route metadata keys Kuma interprets itself, so a
// RouteMetadataProvider must not set them. If it does, route drops the value to
// stop an embedder from changing Kuma's own routing behavior (e.g. disabling
// authn via authn.MetadataAuthKey).
var reservedRouteMetadataKeys = map[string]struct{}{
	authn.MetadataAuthKey: {},
}

// route builds the RouteBuilder for a CRUD method/path and applies any provider
// metadata, naming the verb once so ws.<VERB> and the metadata can't desync.
func (r *resourceEndpoints) route(ws *restful.WebService, method, path string) *restful.RouteBuilder {
	var rb *restful.RouteBuilder
	switch method {
	case http.MethodGet:
		rb = ws.GET(path)
	case http.MethodPut:
		rb = ws.PUT(path)
	case http.MethodDelete:
		rb = ws.DELETE(path)
	default:
		panic(fmt.Sprintf("resource endpoints: unsupported method %q", method))
	}
	if r.routeMetadataProvider != nil {
		for k, v := range r.routeMetadataProvider(r.descriptor, method) {
			if _, reserved := reservedRouteMetadataKeys[k]; reserved {
				log.Info("ignoring reserved route metadata key from provider", "key", k, "type", r.descriptor.Name, "method", method)
				continue
			}
			rb = rb.Metadata(k, v)
		}
	}
	return rb
}

func (r *resourceEndpoints) addFindEndpoint(ws *restful.WebService, pathPrefix string) {
	ws.Route(r.route(ws, http.MethodGet, pathPrefix+"/{name}").To(handle(r.findResource(false))).
		Doc(fmt.Sprintf("Get a %s", r.descriptor.WsPath)).
		Param(ws.PathParameter("name", fmt.Sprintf("Name of a %s", r.descriptor.Name)).DataType("string")).
		Returns(200, "OK", nil).
		Returns(404, "Not found", nil))
	if r.descriptor.HasInsights() {
		route := r.findResource(true)
		ws.Route(ws.GET(pathPrefix+"/{name}/_overview").To(handle(route)).
			Doc(fmt.Sprintf("Get overview of a %s", r.descriptor.Name)).
			Param(ws.PathParameter("name", fmt.Sprintf("Name of a %s", r.descriptor.Name)).DataType("string")).
			Returns(200, "OK", nil).
			Returns(404, "Not found", nil))
	}
	if r.descriptor.IsPolicy {
		ws.Route(ws.GET(pathPrefix+"/{name}/_resources/dataplanes").To(handle(r.inspect.matchingDataplanesForPolicy())).
			Doc(fmt.Sprintf("Get matching dataplanes of a %s", r.descriptor.Name)).
			Param(ws.PathParameter("name", fmt.Sprintf("Name of a %s", r.descriptor.Name)).DataType("string")).
			Returns(200, "OK", nil).
			Returns(404, "Not found", nil))
	}
	if extra, ok := extraInspectRoutes[r.descriptor.Name]; ok {
		extra(r, ws, pathPrefix)
	}
}

// extraInspectRoutes registers type-specific GET routes under
// pathPrefix+"/{name}" for resource types that expose proxy inspect
// endpoints, keeping addFindEndpoint free of per-type branches.
var extraInspectRoutes = map[core_model.ResourceType]func(*resourceEndpoints, *restful.WebService, string){
	core_mesh.DataplaneType: (*resourceEndpoints).addDataplaneInspectRoutes,
}

func (r *resourceEndpoints) addDataplaneInspectRoutes(ws *restful.WebService, pathPrefix string) {
	ws.Route(ws.GET(pathPrefix+"/{name}/_rules").To(handle(r.inspect.rulesForResource())).
		Doc(fmt.Sprintf("Get matching rules %s", r.descriptor.Name)).
		Param(ws.PathParameter("name", fmt.Sprintf("Name of a %s", r.descriptor.Name)).DataType("string")).
		Returns(200, "OK", nil).
		Returns(404, "Not found", nil))
	if r.mode == config_core.Global {
		msg := "Not allowed on global CP"
		ws.Route(ws.GET(pathPrefix+"/{name}/_config").To(handle(r.methodNotAllowed(msg))).
			Doc(msg).
			Returns(http.StatusMethodNotAllowed, msg, restful.ServiceError{}))
	} else {
		ws.Route(ws.GET(pathPrefix+"/{name}/_config").To(handle(r.inspect.configForProxy())).
			Doc(fmt.Sprintf("Get proxy config%s", r.descriptor.Name)).
			Param(ws.PathParameter("name", fmt.Sprintf("Name of a %s", r.descriptor.Name)).DataType("string")).
			Returns(200, "OK", nil).
			Returns(404, "Not found", nil))
	}
	ws.Route(ws.GET(pathPrefix+"/{name}/_policies").To(handle(r.inspect.getPoliciesConf(core_plugins.Plugins().PolicyPlugins(), matchedPoliciesToProxyPolicy))).
		Doc(fmt.Sprintf("Get policy config %s", r.descriptor.Name)).
		Param(ws.PathParameter("name", fmt.Sprintf("Name of a %s", r.descriptor.Name)).DataType("string")).
		Returns(200, "OK", nil).
		Returns(404, "Not found", nil))
	ws.Route(ws.GET(pathPrefix+"/{name}/_inbounds/{inbound_kri}/_policies").To(handle(r.inspect.getPoliciesConf(core_plugins.Plugins().PolicyPlugins(), matchedPoliciesToInboundConfig))).
		Doc("Get policy config for inbound").
		Param(ws.PathParameter("name", fmt.Sprintf("Name of a %s", r.descriptor.Name)).DataType("string")).
		Param(ws.PathParameter("inbound_kri", "KRI of a inbound").DataType("string")).
		Returns(200, "OK", nil).
		Returns(404, "Not found", nil))
	ws.Route(ws.GET(pathPrefix+"/{name}/_outbounds/{outbound_kri}/_policies").To(handle(r.inspect.getPoliciesConf(core_plugins.Plugins().PolicyPlugins(), matchedPoliciesToOutboundPolicy))).
		Doc("Get policy config for outbound").
		Param(ws.PathParameter("name", fmt.Sprintf("Name of a %s", r.descriptor.Name)).DataType("string")).
		Param(ws.PathParameter("outbound_kri", "KRI of a outbound").DataType("string")).
		Returns(200, "OK", nil).
		Returns(404, "Not found", nil))
	ws.Route(ws.GET(pathPrefix+"/{name}/_outbounds/{outbound_kri}/_routes").To(handle(r.inspect.getPoliciesConf(
		util_slices.Filter(core_plugins.Plugins().PolicyPlugins(), func(p core_plugins.RegisteredPolicyPlugin) bool {
			return p.Name == core_plugins.PluginName(meshhttproute_api.MeshHTTPRouteResourceTypeDescriptor.KumactlArg) ||
				p.Name == core_plugins.PluginName(meshtcproute_api.MeshTCPRouteResourceTypeDescriptor.KumactlArg)
		}),
		matchedPoliciesToRoutes,
	))).
		Doc("Get policy config for outbound").
		Param(ws.PathParameter("name", fmt.Sprintf("Name of a %s", r.descriptor.Name)).DataType("string")).
		Param(ws.PathParameter("outbound_kri", "KRI of a outbound").DataType("string")).
		Returns(200, "OK", nil).
		Returns(404, "Not found", nil))
	ws.Route(ws.GET(pathPrefix+"/{name}/_outbounds/{outbound_kri}/_routes/{route_kri}/_policies").To(handle(r.inspect.getPoliciesConf(
		util_slices.Filter(core_plugins.Plugins().PolicyPlugins(), func(p core_plugins.RegisteredPolicyPlugin) bool {
			return p.Name != core_plugins.PluginName(meshhttproute_api.MeshHTTPRouteResourceTypeDescriptor.KumactlArg) &&
				p.Name != core_plugins.PluginName(meshtcproute_api.MeshTCPRouteResourceTypeDescriptor.KumactlArg)
		}), matchedPoliciesToRouteConfig))).
		Doc("Get policy config for route").
		Param(ws.PathParameter("name", fmt.Sprintf("Name of a %s", r.descriptor.Name)).DataType("string")).
		Param(ws.PathParameter("outbound_kri", "KRI of a outbound").DataType("string")).
		Param(ws.PathParameter("route_kri", "KRI of a route").DataType("string")).
		Returns(200, "OK", nil).
		Returns(404, "Not found", nil))
}

func (r *resourceEndpoints) methodNotAllowed(detail string) handlerFunc {
	return func(request *restful.Request) (any, error) {
		return nil, &types.Error{
			Status: 405,
			Title:  "Method not allowed",
			Detail: detail,
		}
	}
}

func (r *resourceEndpoints) addListEndpoint(ws *restful.WebService, pathPrefix string) {
	ws.Route(r.route(ws, http.MethodGet, pathPrefix).To(handle(r.listResources(false))).
		Doc(fmt.Sprintf("List of %s", r.descriptor.Name)).
		Param(ws.QueryParameter("size", "size of page").DataType("int")).
		Param(ws.QueryParameter("offset", "offset of page to list").DataType("string")).
		Param(ws.QueryParameter("name", "a pattern to select only resources that contain these characters").DataType("string")).
		Returns(200, "OK", nil))
	if r.descriptor.HasInsights() {
		route := r.listResources(true)
		ws.Route(ws.GET(pathPrefix+"/_overview").To(handle(route)).
			Doc(fmt.Sprintf("Get a %s", r.descriptor.WsPath)).
			Param(ws.QueryParameter("size", "size of page").DataType("int")).
			Param(ws.QueryParameter("offset", "offset of page to list").DataType("string")).
			Param(ws.QueryParameter(filters.StatusFilterParam, "select only resources with this status").DataType("string")).
			Param(ws.PathParameter("name", "a pattern to select only resources that contain these characters").DataType("string")).
			Returns(200, "OK", nil).
			Returns(404, "Not found", nil))
	}
}

func (r *resourceEndpoints) addCreateOrUpdateEndpoint(ws *restful.WebService, pathPrefix string) {
	if r.descriptor.ReadOnly {
		ws.Route(r.route(ws, http.MethodPut, pathPrefix+"/{name}").To(handle(r.methodNotAllowed(r.readOnlyMessage()))).
			Doc("Not allowed in read-only mode.").
			Returns(http.StatusMethodNotAllowed, "Not allowed in read-only mode.", restful.ServiceError{}))
	} else {
		ws.Route(r.route(ws, http.MethodPut, pathPrefix+"/{name}").To(handle(r.createOrUpdateResource)).
			Doc(fmt.Sprintf("Updates a %s", r.descriptor.WsPath)).
			Param(ws.PathParameter("name", fmt.Sprintf("Name of the %s", r.descriptor.WsPath)).DataType("string")).
			Returns(200, "OK", nil).
			Returns(201, "Created", nil))
	}
}

func (r *resourceEndpoints) addDeleteEndpoint(ws *restful.WebService, pathPrefix string) {
	if r.descriptor.ReadOnly {
		ws.Route(r.route(ws, http.MethodDelete, pathPrefix+"/{name}").To(handle(r.methodNotAllowed(r.readOnlyMessage()))).
			Doc("Not allowed in read-only mode.").
			Returns(http.StatusMethodNotAllowed, "Not allowed in read-only mode.", restful.ServiceError{}))
	} else {
		ws.Route(r.route(ws, http.MethodDelete, pathPrefix+"/{name}").To(handle(r.deleteResource)).
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
