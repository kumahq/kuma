package api_server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/emicklei/go-restful/v3"
	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	api_server_types "github.com/kumahq/kuma/v3/pkg/api-server/types"
	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	core_plugins "github.com/kumahq/kuma/v3/pkg/core/plugins"
	"github.com/kumahq/kuma/v3/pkg/core/resources/access"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	meshtrust_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshtrust/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/system"
	resource_labels "github.com/kumahq/kuma/v3/pkg/core/resources/labels"
	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	rest_errors "github.com/kumahq/kuma/v3/pkg/core/rest/errors"
	"github.com/kumahq/kuma/v3/pkg/core/rest/errors/types"
	"github.com/kumahq/kuma/v3/pkg/core/user"
	"github.com/kumahq/kuma/v3/pkg/core/validators"
	meshhttproute_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	meshtcproute_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtcproute/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/k8s"
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/v3/pkg/util/maps"
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

type resourceEndpoints struct {
	resourceEndpointsContext

	federatedZone   bool
	k8sMapper       k8s.ResourceMapperFunc
	filter          func(request *restful.Request) (store.ListFilterFunc, error)
	systemNamespace string
	isK8s           bool

	disableOriginLabelValidation bool

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

func (r *resourceEndpoints) findResource(withInsight bool) func(request *restful.Request, response *restful.Response) {
	return func(request *restful.Request, response *restful.Response) {
		name := request.PathParameter("name")
		meshName, err := r.meshFromRequest(request)
		if err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Failed to retrieve Mesh")
			return
		}

		if err := r.resourceAccess.ValidateGet(
			request.Request.Context(),
			core_model.ResourceKey{Mesh: meshName, Name: name},
			r.descriptor,
			user.FromCtx(request.Request.Context()),
		); err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Access Denied")
			return
		}

		resource := r.descriptor.NewObject()
		if err := r.resManager.Get(request.Request.Context(), resource, store.GetByKey(name, meshName)); err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Could not retrieve a resource")
			return
		}
		if withInsight {
			insight := r.descriptor.NewInsight()
			if err := r.resManager.Get(request.Request.Context(), insight, store.GetByKey(name, meshName)); err != nil && !store.IsNotFound(err) {
				rest_errors.HandleError(request.Request.Context(), response, err, "Could not retrieve insights")
				return
			}
			overview, ok := r.descriptor.NewOverview().(core_model.OverviewResource)
			if !ok {
				rest_errors.HandleError(request.Request.Context(), response, fmt.Errorf("type withInsight for '%s' doesn't implement core_model.OverviewResource this shouldn't happen", r.descriptor.Name), "Could not retrieve insights")
				return
			}
			if err := overview.SetOverviewSpec(resource, insight); err != nil {
				rest_errors.HandleError(request.Request.Context(), response, err, "Could not retrieve insights")
				return
			}
			resource = overview.(core_model.Resource)
		}
		var res any

		res, err = formatResource(resource, request.QueryParameter("format"), r.k8sMapper, request.QueryParameter("namespace"))
		if err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Could not retrieve a resource")
			return
		}
		if err := response.WriteAsJson(res); err != nil {
			log.Error(err, "Could not write the find response")
		}
	}
}

func formatResource(resource core_model.Resource, format string, k8sMapper k8s.ResourceMapperFunc, namespace string) (any, error) {
	switch format {
	case "k8s", "kubernetes":
		res, err := k8sMapper(resource, namespace)
		if err != nil {
			return nil, err
		}
		return res, nil
	case "universal", "":
		return rest.From.Resource(resource), nil
	default:
		err := validators.MakeFieldMustBeOneOfErr("format", "k8s", "kubernetes", "universal")
		return nil, err.OrNil()
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

func (r *resourceEndpoints) listResources(withInsight bool) func(request *restful.Request, response *restful.Response) {
	return func(request *restful.Request, response *restful.Response) {
		page, err := pagination(request)
		if err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Could not retrieve resources")
			return
		}
		filter, err := r.filter(request)
		if err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Could not retrieve resources")
			return
		}
		nameContains := request.QueryParameter("name")

		meshName, err := r.meshFromRequest(request)
		if err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Failed to retrieve Mesh")
			return
		}

		if err := r.resourceAccess.ValidateList(
			request.Request.Context(),
			meshName,
			r.descriptor,
			user.FromCtx(request.Request.Context()),
		); err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Access Denied")
			return
		}
		list := r.descriptor.NewList()
		if err := r.resManager.List(request.Request.Context(), list, store.ListByMesh(meshName), store.ListByNameContains(nameContains), store.ListByFilterFunc(filter), store.ListByPage(page.size, page.offset)); err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Could not retrieve resources")
			return
		}
		if withInsight {
			// we cannot paginate insights since there is no guarantee that the insights elements will be the same as regular entities
			// Extract ResourceKeys from filtered dataplanes
			resourceKeys := make([]core_model.ResourceKey, 0, len(list.GetItems()))
			for _, item := range list.GetItems() {
				resourceKeys = append(resourceKeys, core_model.MetaToResourceKey(item.GetMeta()))
			}

			// Fetch insights only for filtered dataplanes
			insights := r.descriptor.NewInsightList()
			if err := r.resManager.List(request.Request.Context(), insights, store.ListByMesh(meshName), store.ListByResourceKeys(resourceKeys)); err != nil {
				rest_errors.HandleError(request.Request.Context(), response, err, "Could not retrieve resources")
				return
			}
			list, err = r.MergeInOverview(list, insights)
			if err != nil {
				rest_errors.HandleError(request.Request.Context(), response, err, "Failed merging overview and insights")
				return
			}
		}
		restList := rest.From.ResourceList(list)
		restList.Next = nextLink(request, list.GetPagination().NextOffset)
		if err := response.WriteAsJson(restList); err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Could not list resources")
		}
	}
}

func (r *resourceEndpoints) MergeInOverview(resources core_model.ResourceList, insights core_model.ResourceList) (core_model.ResourceList, error) {
	insightsByKey := map[core_model.ResourceKey]core_model.Resource{}
	for _, insight := range insights.GetItems() {
		insightsByKey[core_model.MetaToResourceKey(insight.GetMeta())] = insight
	}

	items := r.descriptor.NewOverviewList()
	for _, resource := range resources.GetItems() {
		overview, ok := items.NewItem().(core_model.OverviewResource)
		if !ok {
			return nil, fmt.Errorf("type overview for '%s' doesn't implement core_model.OverviewResource this shouldn't happen", r.descriptor.Name)
		}
		if err := overview.SetOverviewSpec(resource, insightsByKey[core_model.MetaToResourceKey(resource.GetMeta())]); err != nil {
			return nil, err
		}

		if err := items.AddItem(overview.(core_model.Resource)); err != nil {
			return nil, err
		}
	}
	items.SetPagination(*resources.GetPagination())
	return items, nil
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

func (r *resourceEndpoints) createOrUpdateResource(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")
	meshName, err := r.meshFromRequest(request)
	if err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Failed to retrieve Mesh")
		return
	}

	bodyBytes, err := io.ReadAll(request.Request.Body)
	if err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Could not process a resource")
		return
	}

	resourceRest, err := rest.JSON.Unmarshal(bodyBytes, r.descriptor)
	if err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Could not process a resource")
		return
	}

	create := false
	resource := r.descriptor.NewObject()
	if err := r.resManager.Get(request.Request.Context(), resource, store.GetByKey(name, meshName)); err != nil && store.IsNotFound(err) {
		create = true
	} else if err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Failed to find a resource")
		return
	}

	if err := r.validateResourceRequest(name, meshName, resourceRest); err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Could not process a resource")
		return
	}

	if create {
		r.createResource(request.Request.Context(), name, meshName, resourceRest, response)
	} else {
		r.updateResource(request.Request.Context(), resource, resourceRest, response, meshName)
	}
}

func (r *resourceEndpoints) clearMeshTrustOrigin(resRest rest.Resource, meshName string, name string) {
	if r.descriptor.Name == meshtrust_api.MeshTrustType {
		if resRest.GetStatus() != nil {
			status, ok := resRest.GetStatus().(*meshtrust_api.MeshTrustStatus)
			if ok && status != nil && status.Origin != nil {
				log.Info("ignoring status.origin as it is read-only", "mesh", meshName, "name", name)
				status.Origin = nil
			}
		}
	}
}

// computeLabels derives the full label set for a resource from its descriptor,
// spec and meta, applying the control-plane mode, zone, k8s and namespace context
// shared by create and update.
func (r *resourceEndpoints) computeLabels(
	descriptor core_model.ResourceTypeDescriptor,
	spec core_model.ResourceSpec,
	meta core_model.ResourceMeta,
	meshName string,
	name string,
) (map[string]string, error) {
	return resource_labels.Compute(
		descriptor,
		spec,
		meta.GetLabels(),
		meshName,
		name,
		resource_labels.WithNamespace(resource_labels.GetNamespace(meta, r.systemNamespace)),
		resource_labels.WithMode(r.mode),
		resource_labels.WithK8s(r.isK8s),
		resource_labels.WithZone(r.zoneName),
	)
}

func (r *resourceEndpoints) createResource(
	ctx context.Context,
	name string,
	meshName string,
	resRest rest.Resource,
	response *restful.Response,
) {
	if err := r.resourceAccess.ValidateCreate(
		ctx,
		core_model.ResourceKey{Mesh: meshName, Name: name},
		resRest.GetSpec(),
		r.descriptor,
		user.FromCtx(ctx),
	); err != nil {
		rest_errors.HandleError(ctx, response, err, "Access Denied")
		return
	}

	r.clearMeshTrustOrigin(resRest, meshName, name)

	res := r.descriptor.NewObject()
	_ = res.SetSpec(resRest.GetSpec())
	res.SetMeta(resRest.GetMeta())

	// Validate workload label on Universal Zone dataplanes
	if r.descriptor.Name == core_mesh.DataplaneType && !r.isK8s {
		if resLabels := res.GetMeta().GetLabels(); resLabels != nil {
			if workloadName, ok := resLabels[metadata.KumaWorkload]; ok && workloadName != "" {
				if r.mode == config_core.Global {
					err := rest_errors.NewBadRequestError("labels[\"kuma.io/workload\"]: not allowed on Global control plane")
					rest_errors.HandleError(ctx, response, err, "Invalid workload label")
					return
				}
				if validationErrs := apimachineryvalidation.NameIsDNS1035Label(workloadName, false); len(validationErrs) != 0 {
					err := rest_errors.NewBadRequestError(fmt.Sprintf("labels[\"kuma.io/workload\"]: must be a valid DNS-1035 label (at most 63 characters, matching regex [a-z]([-a-z0-9]*[a-z0-9])?): %s", strings.Join(validationErrs, "; ")))
					rest_errors.HandleError(ctx, response, err, "Invalid workload label")
					return
				}
			}
		}
	}

	labels, err := r.computeLabels(res.Descriptor(), res.GetSpec(), res.GetMeta(), meshName, name)
	if err != nil {
		rest_errors.HandleError(ctx, response, err, "Could not compute labels for a resource")
		return
	}

	if err := r.resManager.Create(ctx, res, store.CreateByKey(name, meshName), store.CreateWithLabels(labels)); err != nil {
		rest_errors.HandleError(ctx, response, err, "Failed to create a resource")
		return
	}

	resp := api_server_types.CreateOrUpdateSuccessResponse{Warnings: core_model.Deprecations(res)}
	if err := response.WriteHeaderAndJson(http.StatusCreated, resp, "application/json"); err != nil {
		log.Error(err, "Could not write the create response")
	}
}

func (r *resourceEndpoints) updateResource(
	ctx context.Context,
	currentRes core_model.Resource,
	newResRest rest.Resource,
	response *restful.Response,
	meshName string,
) {
	if err := r.resourceAccess.ValidateUpdate(
		ctx,
		core_model.ResourceKey{Mesh: currentRes.GetMeta().GetMesh(), Name: currentRes.GetMeta().GetName()},
		currentRes.GetSpec(),
		newResRest.GetSpec(),
		r.descriptor,
		user.FromCtx(ctx),
	); err != nil {
		rest_errors.HandleError(ctx, response, err, "Access Denied")
		return
	}

	r.clearMeshTrustOrigin(newResRest, meshName, currentRes.GetMeta().GetName())

	// Compute labels for current state BEFORE modifying spec
	currentLabels, err := r.computeLabels(currentRes.Descriptor(), currentRes.GetSpec(), currentRes.GetMeta(), meshName, currentRes.GetMeta().GetName())
	if err != nil {
		rest_errors.HandleError(ctx, response, err, "Could not compute current labels")
		return
	}

	_ = currentRes.SetSpec(newResRest.GetSpec())

	// Compute labels for new request
	labels, err := r.computeLabels(currentRes.Descriptor(), currentRes.GetSpec(), newResRest.GetMeta(), meshName, currentRes.GetMeta().GetName())
	if err != nil {
		rest_errors.HandleError(ctx, response, err, "Could not compute labels for a resource")
		return
	}

	// Validate immutable labels by comparing computed results
	if validationErr := r.validateImmutableLabels(currentLabels, labels); validationErr.HasViolations() {
		var err validators.ValidationError
		err.AddError("labels", validationErr)
		rest_errors.HandleError(ctx, response, &err, "Could not update a resource")
		return
	}

	if err := r.resManager.Update(ctx, currentRes, store.UpdateWithLabels(labels)); err != nil {
		rest_errors.HandleError(ctx, response, err, "Failed to update a resource")
		return
	}

	resp := api_server_types.CreateOrUpdateSuccessResponse{Warnings: core_model.Deprecations(currentRes)}
	if err := response.WriteHeaderAndJson(http.StatusOK, resp, "application/json"); err != nil {
		log.Error(err, "Could not write the update response")
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

func (r *resourceEndpoints) deleteResource(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")
	meshName, err := r.meshFromRequest(request)
	if err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Failed to retrieve Mesh")
		return
	}

	resource := r.descriptor.NewObject()

	if err := r.resManager.Get(request.Request.Context(), resource, store.GetByKey(name, meshName)); err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Could not delete a resource")
		return
	}

	if verr := r.validateOriginForWrite(resource.GetMeta()); verr.HasViolations() {
		rest_errors.HandleError(request.Request.Context(), response, verr.OrNil(), "Could not delete a resource")
		return
	}

	if err := r.resourceAccess.ValidateDelete(
		request.Request.Context(),
		core_model.ResourceKey{Mesh: meshName, Name: name},
		resource.GetSpec(),
		resource.Descriptor(),
		user.FromCtx(request.Request.Context()),
	); err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Access Denied")
		return
	}

	if err := r.resManager.Delete(request.Request.Context(), resource, store.DeleteByKey(name, meshName)); err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Could not delete a resource")
		return
	}

	resp := api_server_types.DeleteSuccessResponse{}
	if err := response.WriteHeaderAndJson(http.StatusOK, resp, "application/json"); err != nil {
		log.Error(err, "Could not write the delete response")
	}
}

func (r *resourceEndpoints) validateResourceRequest(name string, meshName string, resource rest.Resource) error {
	var err validators.ValidationError
	if name != resource.GetMeta().Name {
		err.AddViolation("name", "name from the URL has to be the same as in body")
	}
	if r.federatedZone && !r.doesNameLengthFitsGlobal(name) {
		err.AddViolation("name", "the length of the name must be shorter")
	}
	if string(r.descriptor.Name) != resource.GetMeta().Type {
		err.AddViolation("type", "type from the URL has to be the same as in body")
	}
	if r.descriptor.Scope == core_model.ScopeMesh && meshName != resource.GetMeta().Mesh {
		err.AddViolation("mesh", "mesh from the URL has to be the same as in body")
	}

	err.AddError("labels", r.validateLabels(resource))
	err.AddError("", core_mesh.ValidateMeta(resource.GetMeta(), r.descriptor.Scope))

	return err.OrNil()
}

func (r *resourceEndpoints) validateOriginForWrite(meta core_model.ResourceMeta) validators.ValidationError {
	var err validators.ValidationError
	origin, ok := core_model.ResourceOrigin(meta)

	if !r.disableOriginLabelValidation && r.mode == config_core.Global {
		if ok && origin != mesh_proto.GlobalResourceOrigin {
			err.AddViolationAt(validators.Root().Key(mesh_proto.ResourceOriginLabel), fmt.Sprintf("the origin label must be set to '%s'", mesh_proto.GlobalResourceOrigin))
		}
	}

	if !r.disableOriginLabelValidation && r.federatedZone && r.descriptor.IsPluginOriginated {
		if !ok || origin != mesh_proto.ZoneResourceOrigin {
			err.AddViolationAt(validators.Root().Key(mesh_proto.ResourceOriginLabel), fmt.Sprintf("the origin label must be set to '%s'", mesh_proto.ZoneResourceOrigin))
		}
	}
	return err
}

func (r *resourceEndpoints) validateLabels(resource rest.Resource) validators.ValidationError {
	var err validators.ValidationError

	origin, ok := core_model.ResourceOrigin(resource.GetMeta())
	if ok {
		if oerr := origin.IsValid(); oerr != nil {
			err.AddViolationAt(validators.Root().Key(mesh_proto.ResourceOriginLabel), oerr.Error())
		}
	}

	err.AddError("", r.validateOriginForWrite(resource.GetMeta()))

	if r.mode != config_core.Global {
		if origin != mesh_proto.GlobalResourceOrigin {
			zoneTag, ok := resource.GetMeta().GetLabels()[mesh_proto.ZoneTag]
			if ok && zoneTag != r.zoneName {
				err.AddViolationAt(validators.Root().Key(mesh_proto.ZoneTag), fmt.Sprintf("%s label should have %s value", mesh_proto.ZoneTag, r.zoneName))
			}
			if meshLabelValue, ok := resource.GetMeta().GetLabels()[mesh_proto.MeshTag]; ok && meshLabelValue != resource.GetMeta().GetMesh() {
				err.AddViolationAt(validators.Root().Key(mesh_proto.MeshTag), fmt.Sprintf("%s label must not differ from mesh set on resource", mesh_proto.MeshTag))
			}
		}
	}

	if r.descriptor.IsPluginOriginated && r.descriptor.IsPolicy {
		err.AddError("", r.validatePolicyRole(resource))
	}

	for _, k := range maps.SortedKeys(resource.GetMeta().GetLabels()) {
		for _, msg := range validation.IsQualifiedName(k) {
			err.AddViolationAt(validators.Root().Key(k), msg)
		}
		for _, msg := range validation.IsValidLabelValue(resource.GetMeta().GetLabels()[k]) {
			err.AddViolationAt(validators.Root().Key(k), msg)
		}
	}
	return err
}

func (r *resourceEndpoints) validateImmutableLabels(currentComputedLabels, newComputedLabels map[string]string) validators.ValidationError {
	var err validators.ValidationError

	immutableLabels := []string{
		mesh_proto.ResourceOriginLabel,
		mesh_proto.ZoneTag,
	}

	for _, label := range immutableLabels {
		currentVal, currentExists := currentComputedLabels[label]
		newVal, newExists := newComputedLabels[label]

		if currentExists && !newExists {
			err.AddViolationAt(
				validators.Root().Key(label),
				fmt.Sprintf("is immutable, cannot be removed (was %q)", currentVal),
			)
		} else if currentExists && currentVal != newVal {
			err.AddViolationAt(
				validators.Root().Key(label),
				fmt.Sprintf("is immutable, cannot be changed from %q to %q", currentVal, newVal),
			)
		}
	}

	return err
}

func (r *resourceEndpoints) validatePolicyRole(resource rest.Resource) validators.ValidationError {
	var err validators.ValidationError
	policyRole := core_model.PolicyRole(resource.GetMeta())
	// at the moment on universal all policies have system policy role
	if policyRole != mesh_proto.SystemPolicyRole {
		err.AddViolationAt(validators.Root().Key(mesh_proto.PolicyRoleLabel), fmt.Sprintf("%s label should have %s value, got %s", mesh_proto.PolicyRoleLabel, mesh_proto.SystemPolicyRole, policyRole))
	}
	return err
}

// The resource is prefixed with the zone name when it is synchronized
// to global control-plane. It is important to notice that the zone is unaware
// of the type of the store used by the global control-plane, so we must prepare
// for the worst-case scenario. We don't have to check other plugabble policies
// because zone doesn't allow to create policies on the zone.
func (r *resourceEndpoints) doesNameLengthFitsGlobal(name string) bool {
	return len(fmt.Sprintf("%s.%s", r.zoneName, name)) < 253
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
