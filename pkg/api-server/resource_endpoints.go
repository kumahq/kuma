package api_server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/util/validation"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	api_types "github.com/kumahq/kuma/api/openapi/types"
	api_common "github.com/kumahq/kuma/api/openapi/types/common"
	oapi_helpers "github.com/kumahq/kuma/pkg/api-server/oapi-helpers"
	api_server_types "github.com/kumahq/kuma/pkg/api-server/types"
	config_core "github.com/kumahq/kuma/pkg/config/core"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	"github.com/kumahq/kuma/pkg/core/policy"
	"github.com/kumahq/kuma/pkg/core/resources/access"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	rest_errors "github.com/kumahq/kuma/pkg/core/rest/errors"
	"github.com/kumahq/kuma/pkg/core/rest/errors/types"
	"github.com/kumahq/kuma/pkg/core/user"
	"github.com/kumahq/kuma/pkg/core/validators"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/core/xds/inspect"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/matchers"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/ordered"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s"
	"github.com/kumahq/kuma/pkg/util/maps"
	"github.com/kumahq/kuma/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	xds_hooks "github.com/kumahq/kuma/pkg/xds/hooks"
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

type resourceEndpoints struct {
	mode               config_core.CpMode
	federatedZone      bool
	zoneName           string
	resManager         manager.ResourceManager
	descriptor         model.ResourceTypeDescriptor
	resourceAccess     access.ResourceAccess
	k8sMapper          k8s.ResourceMapperFunc
	filter             func(request *restful.Request) (store.ListFilterFunc, error)
	meshContextBuilder xds_context.MeshContextBuilder
	xdsHooks           []xds_hooks.ResourceSetHook
	systemNamespace    string
	isK8s              bool

	disableOriginLabelValidation bool
}

func typeToLegacyOverviewPath(resourceType model.ResourceType) string {
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
		ws.Route(ws.GET(pathPrefix+"/{name}/_resources/dataplanes").To(r.matchingDataplanesForPolicy()).
			Doc(fmt.Sprintf("Get matching dataplanes of a %s", r.descriptor.Name)).
			Param(ws.PathParameter("name", fmt.Sprintf("Name of a %s", r.descriptor.Name)).DataType("string")).
			Returns(200, "OK", nil).
			Returns(404, "Not found", nil))
	}
	if r.descriptor.Name == core_mesh.DataplaneType || r.descriptor.Name == core_mesh.MeshGatewayType {
		ws.Route(ws.GET(pathPrefix+"/{name}/_rules").To(r.rulesForResource()).
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
			ws.Route(ws.GET(pathPrefix+"/{name}/_config").To(r.configForProxy()).
				Doc(fmt.Sprintf("Get proxy config%s", r.descriptor.Name)).
				Param(ws.PathParameter("name", fmt.Sprintf("Name of a %s", r.descriptor.Name)).DataType("string")).
				Returns(200, "OK", nil).
				Returns(404, "Not found", nil))
		}
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
			model.ResourceKey{Mesh: meshName, Name: name},
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
			if err := r.resManager.Get(request.Request.Context(), insight, store.GetByKey(name, meshName)); err != nil && !store.IsResourceNotFound(err) {
				rest_errors.HandleError(request.Request.Context(), response, err, "Could not retrieve insights")
				return
			}
			overview, ok := r.descriptor.NewOverview().(model.OverviewResource)
			if !ok {
				rest_errors.HandleError(request.Request.Context(), response, fmt.Errorf("type withInsight for '%s' doesn't implement model.OverviewResource this shouldn't happen", r.descriptor.Name), "Could not retrieve insights")
				return
			}
			if err := overview.SetOverviewSpec(resource, insight); err != nil {
				rest_errors.HandleError(request.Request.Context(), response, err, "Could not retrieve insights")
				return
			}
			resource = overview.(model.Resource)
		}
		var res interface{}
		switch request.QueryParameter("format") {
		case "k8s", "kubernetes":
			var err error
			res, err = r.k8sMapper(resource, request.QueryParameter("namespace"))
			if err != nil {
				rest_errors.HandleError(request.Request.Context(), response, err, "k8s mapping failed")
				return
			}
		case "universal", "":
			res = rest.From.Resource(resource)
		default:
			err := validators.MakeFieldMustBeOneOfErr("format", "k8s", "kubernetes", "universal")
			rest_errors.HandleError(request.Request.Context(), response, err.OrNil(), "invalid format")
		}
		if err := response.WriteAsJson(res); err != nil {
			log.Error(err, "Could not write the response")
		}
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
			insights := r.descriptor.NewInsightList()
			if err := r.resManager.List(request.Request.Context(), insights, store.ListByMesh(meshName), store.ListByNameContains(nameContains), store.ListByFilterFunc(filter)); err != nil {
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

func (r *resourceEndpoints) MergeInOverview(resources model.ResourceList, insights model.ResourceList) (model.ResourceList, error) {
	insightsByKey := map[model.ResourceKey]model.Resource{}
	for _, insight := range insights.GetItems() {
		insightsByKey[model.MetaToResourceKey(insight.GetMeta())] = insight
	}

	items := r.descriptor.NewOverviewList()
	for _, resource := range resources.GetItems() {
		overview, ok := items.NewItem().(model.OverviewResource)
		if !ok {
			return nil, fmt.Errorf("type overview for '%s' doesn't implement model.OverviewResource this shouldn't happen", r.descriptor.Name)
		}
		if err := overview.SetOverviewSpec(resource, insightsByKey[model.MetaToResourceKey(resource.GetMeta())]); err != nil {
			return nil, err
		}

		if err := items.AddItem(overview.(model.Resource)); err != nil {
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
	if err := r.resManager.Get(request.Request.Context(), resource, store.GetByKey(name, meshName)); err != nil && store.IsResourceNotFound(err) {
		create = true
	} else if err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Could not find a resource")
	}

	if err := r.validateResourceRequest(name, meshName, resourceRest, create); err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Could not process a resource")
		return
	}

	if create {
		r.createResource(request.Request.Context(), name, meshName, resourceRest, response)
	} else {
		r.updateResource(request.Request.Context(), resource, resourceRest, response)
	}
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
		model.ResourceKey{Mesh: meshName, Name: name},
		resRest.GetSpec(),
		r.descriptor,
		user.FromCtx(ctx),
	); err != nil {
		rest_errors.HandleError(ctx, response, err, "Access Denied")
		return
	}

	res := r.descriptor.NewObject()
	_ = res.SetSpec(resRest.GetSpec())
	res.SetMeta(resRest.GetMeta())
	if r.descriptor.HasStatus {
		_ = res.SetStatus(resRest.GetStatus())
	}

	labels, err := model.ComputeLabels(
		res.Descriptor(),
		res.GetSpec(),
		res.GetMeta().GetLabels(),
		model.GetNamespace(res.GetMeta(), r.systemNamespace),
		meshName,
		r.mode,
		r.isK8s,
		r.zoneName,
	)
	if err != nil {
		rest_errors.HandleError(ctx, response, err, "Could not compute labels for a resource")
		return
	}

	if err := r.resManager.Create(ctx, res, store.CreateByKey(name, meshName), store.CreateWithLabels(labels)); err != nil {
		rest_errors.HandleError(ctx, response, err, "Could not create a resource")
		return
	}

	if warnings := model.Deprecations(res); len(warnings) > 0 {
		if err := response.WriteHeaderAndJson(201, api_server_types.CreateOrUpdateSuccessResponse{Warnings: warnings}, "application/json"); err != nil {
			log.Error(err, "Could not write the response")
		}
	} else {
		response.WriteHeader(201)
	}
}

func (r *resourceEndpoints) updateResource(
	ctx context.Context,
	currentRes model.Resource,
	newResRest rest.Resource,
	response *restful.Response,
) {
	if err := r.resourceAccess.ValidateUpdate(
		ctx,
		model.ResourceKey{Mesh: currentRes.GetMeta().GetMesh(), Name: currentRes.GetMeta().GetName()},
		currentRes.GetSpec(),
		newResRest.GetSpec(),
		r.descriptor,
		user.FromCtx(ctx),
	); err != nil {
		rest_errors.HandleError(ctx, response, err, "Access Denied")
		return
	}

	_ = currentRes.SetSpec(newResRest.GetSpec())
	if r.descriptor.HasStatus { // todo(jakubdyszkiewicz) should we always override this?
		_ = currentRes.SetStatus(newResRest.GetStatus())
	}

	if err := r.resManager.Update(ctx, currentRes, store.UpdateWithLabels(newResRest.GetMeta().GetLabels())); err != nil {
		rest_errors.HandleError(ctx, response, err, "Could not update a resource")
		return
	}

	if warnings := model.Deprecations(currentRes); len(warnings) > 0 {
		if err := response.WriteHeaderAndJson(200, api_server_types.CreateOrUpdateSuccessResponse{Warnings: warnings}, "application/json"); err != nil {
			log.Error(err, "Could not write the response")
		}
	} else {
		response.WriteHeader(200)
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

	if err := r.resourceAccess.ValidateDelete(
		request.Request.Context(),
		model.ResourceKey{Mesh: meshName, Name: name},
		resource.GetSpec(),
		resource.Descriptor(),
		user.FromCtx(request.Request.Context()),
	); err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Access Denied")
		return
	}

	if err := r.resManager.Delete(request.Request.Context(), resource, store.DeleteByKey(name, meshName)); err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Could not delete a resource")
	}
}

func (r *resourceEndpoints) validateResourceRequest(name string, meshName string, resource rest.Resource, create bool) error {
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
	if r.descriptor.Scope == model.ScopeMesh && meshName != resource.GetMeta().Mesh {
		err.AddViolation("mesh", "mesh from the URL has to be the same as in body")
	}

	err.AddError("labels", r.validateLabels(resource))

	if create {
		err.AddError("", core_mesh.ValidateMeta(resource.GetMeta(), r.descriptor.Scope))
	} else {
		if verr, msg := core_mesh.ValidateMetaBackwardsCompatible(resource.GetMeta(), r.descriptor.Scope); verr.HasViolations() {
			err.AddError("", verr)
		} else if msg != "" {
			log.Info(msg, "type", r.descriptor.Name, "mesh", resource.GetMeta().Mesh, "name", resource.GetMeta().Name)
		}
	}
	return err.OrNil()
}

func (r *resourceEndpoints) validateLabels(resource rest.Resource) validators.ValidationError {
	var err validators.ValidationError

	origin, ok := model.ResourceOrigin(resource.GetMeta())
	if ok {
		if oerr := origin.IsValid(); oerr != nil {
			err.AddViolationAt(validators.Root().Key(mesh_proto.ResourceOriginLabel), oerr.Error())
		}
	}

	if !r.disableOriginLabelValidation && r.federatedZone && r.descriptor.IsPluginOriginated {
		if !ok || origin != mesh_proto.ZoneResourceOrigin {
			err.AddViolationAt(validators.Root().Key(mesh_proto.ResourceOriginLabel), fmt.Sprintf("the origin label must be set to '%s'", mesh_proto.ZoneResourceOrigin))
		}
	}

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
		r.validatePolicyRole(resource)
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

func (r *resourceEndpoints) meshFromRequest(request *restful.Request) (string, error) {
	if r.descriptor.Scope == model.ScopeMesh {
		meshName := request.PathParameter("mesh")
		if meshName == "" { // Handle lists across all meshes
			return "", nil
		}
		mRes := core_mesh.MeshResourceTypeDescriptor.NewObject()
		if err := r.resManager.Get(request.Request.Context(), mRes, store.GetByKey(meshName, model.NoMesh)); err != nil {
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

func (r *resourceEndpoints) matchingDataplanesForPolicy() restful.RouteFunction {
	return func(request *restful.Request, response *restful.Response) {
		policyName := request.PathParameter("name")
		page, err := pagination(request)
		if err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Could not retrieve policy")
			return
		}
		nameContains := request.QueryParameter("name")
		meshName, err := r.meshFromRequest(request)
		if err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Failed to retrieve Mesh")
			return
		}

		if err := r.resourceAccess.ValidateGet(
			request.Request.Context(),
			model.ResourceKey{Mesh: meshName, Name: policyName},
			r.descriptor,
			user.FromCtx(request.Request.Context()),
		); err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Access Denied")
			return
		}
		policyResource := r.descriptor.NewObject()
		if err := r.resManager.Get(request.Request.Context(), policyResource, store.GetByKey(policyName, meshName)); err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Could not retrieve policy")
			return
		}

		var dependentTypes []model.ResourceType
		if r.descriptor.IsTargetRefBased {
			dependentTypes = []model.ResourceType{meshhttproute_api.MeshHTTPRouteType, core_mesh.MeshGatewayType}
		} else if r.descriptor.Name == core_mesh.MeshGatewayRouteType {
			dependentTypes = []model.ResourceType{core_mesh.MeshGatewayType}
		}
		dependentResources := xds_context.NewResources()
		for _, dependentType := range dependentTypes {
			hl, err := registry.Global().NewList(dependentType)
			if err != nil {
				rest_errors.HandleError(request.Request.Context(), response, err, "failed inspect")
				return
			}
			if err := r.resManager.List(request.Request.Context(), hl, store.ListByMesh(meshName)); err != nil {
				rest_errors.HandleError(request.Request.Context(), response, err, "failed inspect")
				return
			}
			dependentResources.MeshLocalResources[dependentType] = hl
		}
		filter := func(rs model.Resource) bool {
			dpp := rs.(*core_mesh.DataplaneResource)
			if r.descriptor.IsTargetRefBased {
				res, _ := matchers.PolicyMatches(policyResource, dpp, dependentResources)
				return res
			} else if dpPolicy, ok := policyResource.(policy.DataplanePolicy); ok {
				for _, s := range dpPolicy.Selectors() {
					if dpp.Spec.Matches(s.GetMatch()) {
						return true
					}
				}
			} else if connPolicy, ok := policyResource.(policy.ConnectionPolicy); ok {
				for _, s := range connPolicy.Sources() {
					if dpp.Spec.Matches(s.GetMatch()) {
						return true
					}
				}
				for _, s := range connPolicy.Destinations() {
					if dpp.Spec.Matches(s.GetMatch()) {
						return true
					}
				}
			}
			return false
		}
		dppList := registry.Global().MustNewList(core_mesh.DataplaneType)
		err = r.resManager.List(request.Request.Context(), dppList,
			store.ListByMesh(meshName),
			store.ListByNameContains(nameContains),
			store.ListByFilterFunc(filter),
			store.ListByPage(page.size, page.offset),
		)
		if err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "failed inspect")
			return
		}
		items := make([]api_common.Meta, len(dppList.GetItems()))
		for i, elt := range dppList.GetItems() {
			items[i] = oapi_helpers.ResourceToMeta(elt)
		}
		out := api_types.InspectDataplanesForPolicyResponse{
			Total: int(dppList.GetPagination().Total),
			Items: items,
			Next:  nextLink(request, dppList.GetPagination().NextOffset),
		}
		if err := response.WriteAsJson(out); err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Failed writing response")
		}
	}
}

func (r *resourceEndpoints) configForProxy() restful.RouteFunction {
	return func(request *restful.Request, response *restful.Response) {
		ctx := request.Request.Context()

		name := request.PathParameter("name")
		mesh, err := r.meshFromRequest(request)
		if err != nil {
			rest_errors.HandleError(ctx, response, err, "Failed to retrieve Mesh")
			return
		}
		qparams, err := r.configForProxyParams(request)
		if err != nil {
			rest_errors.HandleError(ctx, response, err, "Failed to parse query parameters")
			return
		}

		mc, err := r.meshContextBuilder.Build(ctx, mesh)
		if err != nil {
			rest_errors.HandleError(ctx, response, err, "Failed to build mesh context")
			return
		}

		inspector, err := inspect.NewProxyConfigInspector(mc, r.zoneName, r.xdsHooks...)
		if err != nil {
			rest_errors.HandleError(ctx, response, err, "Failed to create proxy config inspector")
			return
		}

		config, err := inspector.Get(ctx, name, *qparams.Shadow)
		if err != nil {
			rest_errors.HandleError(ctx, response, err, "Failed to inspect proxy config")
			return
		}

		out := &api_types.InspectDataplanesConfig{
			Xds: config,
		}

		if slices.Contains(*qparams.Include, api_types.Diff) {
			currentConfig, err := inspector.Get(ctx, name, false)
			if err != nil {
				rest_errors.HandleError(ctx, response, err, "Failed to inspect current proxy config")
				return
			}
			diff, err := inspect.Diff(currentConfig, config)
			if err != nil {
				rest_errors.HandleError(ctx, response, err, "Failed to compute diff")
				return
			}
			out.Diff = &diff
		}

		if err := response.WriteAsJson(out); err != nil {
			rest_errors.HandleError(ctx, response, err, "Failed writing response")
		}
	}
}

func (r *resourceEndpoints) configForProxyParams(request *restful.Request) (*api_types.GetMeshesMeshDataplanesNameConfigParams, error) {
	params := &api_types.GetMeshesMeshDataplanesNameConfigParams{
		Shadow:  pointer.To(false),
		Include: &[]api_types.GetMeshesMeshDataplanesNameConfigParamsInclude{},
	}

	if shadow := request.QueryParameter("shadow"); shadow != "" {
		if b, err := strconv.ParseBool(shadow); err != nil {
			return nil, rest_errors.NewBadRequestError("unsupported value for query parameter 'shadow'")
		} else {
			params.Shadow = &b
		}
	}

	if include := request.QueryParameter("include"); include != "" {
		for _, v := range strings.Split(include, ",") {
			switch api_types.GetMeshesMeshDataplanesNameConfigParamsInclude(v) {
			case api_types.Diff:
			default:
				return nil, rest_errors.NewBadRequestError("unsupported value for query parameter 'include'")
			}
			*params.Include = append(*params.Include, api_types.GetMeshesMeshDataplanesNameConfigParamsInclude(v))
		}
	}

	return params, nil
}

func (r *resourceEndpoints) rulesForResource() restful.RouteFunction {
	return func(request *restful.Request, response *restful.Response) {
		resourceName := request.PathParameter("name")
		meshName, err := r.meshFromRequest(request)
		if err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Failed to retrieve Mesh")
			return
		}

		if err := r.resourceAccess.ValidateGet(
			request.Request.Context(),
			model.ResourceKey{Mesh: meshName, Name: resourceName},
			r.descriptor,
			user.FromCtx(request.Request.Context()),
		); err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Access Denied")
			return
		}

		resource := r.descriptor.NewObject()
		if err := r.resManager.Get(request.Request.Context(), resource, store.GetByKey(resourceName, meshName)); err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, fmt.Sprintf("Could not retrieve %s", r.descriptor.Name))
			return
		}
		var dp *core_mesh.DataplaneResource
		switch {
		case r.descriptor.Name == core_mesh.DataplaneType:
			dp = resource.(*core_mesh.DataplaneResource)
		case r.descriptor.Name == core_mesh.MeshGatewayType:
			// Create a dataplane that would match this gateway.
			// It might not show all policies but most of the ones matching this specific gateway and its routes
			gw := resource.(*core_mesh.MeshGatewayResource)
			if len(gw.Spec.Selectors) == 0 {
				rest_errors.HandleError(request.Request.Context(), response, errors.New("no selectors on MeshGateway this is not supported"), "Invalid MeshGateway")
				return
			}
			dp = &core_mesh.DataplaneResource{
				Meta: gw.Meta,
				Spec: &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Gateway: &mesh_proto.Dataplane_Networking_Gateway{
							Type: mesh_proto.Dataplane_Networking_Gateway_BUILTIN,
							Tags: gw.Spec.Selectors[0].Match,
						},
					},
				},
			}
		// In the future we will probably add externalService
		default:
			rest_errors.HandleError(request.Request.Context(), response, fmt.Errorf("rules not supported for type %s", r.descriptor.Name), "Unsupported resource type")
			return
		}
		baseMeshContext, err := r.meshContextBuilder.BuildBaseMeshContextIfChanged(request.Request.Context(), meshName, nil)
		if err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Failed to build Mesh context")
		}

		resources := xds_context.Resources{
			CrossMeshResources: map[core_xds.MeshName]xds_context.ResourceMap{},
			MeshLocalResources: baseMeshContext.ResourceMap,
		}
		matchesByHash := map[common_api.MatchesHash][]meshhttproute_api.Match{}
		// Get all the matching policies
		allPlugins := core_plugins.Plugins().PolicyPlugins(ordered.Policies)
		rules := []api_common.InspectRule{}
		for _, policyPlugin := range allPlugins {
			res, err := policyPlugin.Plugin.MatchedPolicies(dp, resources)
			if res.Type == meshhttproute_api.MeshHTTPRouteType {
				for _, pol := range res.ToRules.Rules {
					for _, r := range pol.Conf.(meshhttproute_api.PolicyDefault).Rules {
						matchesByHash[meshhttproute_api.HashMatches(r.Matches)] = r.Matches
					}
				}
			}
			if err != nil {
				rest_errors.HandleError(request.Request.Context(), response, err, fmt.Sprintf("could not apply policy plugin %s", policyPlugin.Name))
			}
			if res.Type == "" {
				rest_errors.HandleError(request.Request.Context(), response, err, fmt.Sprintf("matched policy didn't set type for policy plugin %s", policyPlugin.Name))
			}

			if len(res.ToRules.Rules) == 0 && len(res.FromRules.Rules) == 0 && len(res.SingleItemRules.Rules) == 0 {
				continue
			}
			toRules := []api_common.Rule{}
			if baseMeshContext.Mesh.Spec.MeshServicesMode() != mesh_proto.Mesh_MeshServices_Exclusive {
				// Old 'ToRules' don't affect outbounds that were produced by real resources.
				// That's why we don't have to set them when the mode is Exclusive
				for _, ruleItem := range res.ToRules.Rules {
					toRules = append(toRules, api_common.Rule{
						Conf:     ruleItem.Conf,
						Matchers: oapi_helpers.SubsetToRuleMatcher(ruleItem.Subset),
						Origin:   oapi_helpers.ResourceMetaListToMetaList(res.Type, ruleItem.Origin),
					})
				}
			}
			var proxyRule *api_common.ProxyRule
			if len(res.SingleItemRules.Rules) > 0 {
				proxyRule = &api_common.ProxyRule{
					Conf:   res.SingleItemRules.Rules[0].Conf,
					Origin: oapi_helpers.ResourceMetaListToMetaList(res.Type, res.SingleItemRules.Rules[0].Origin),
				}
			}

			fromRules := []api_common.FromRule{}
			if len(res.FromRules.Rules) > 0 {
				for inbound, rulesForInbound := range res.FromRules.Rules {
					if len(rulesForInbound) == 0 {
						continue
					}
					fromRulesForInbound := make([]api_common.Rule, len(rulesForInbound))
					for i := range rulesForInbound {
						fromRulesForInbound[i] = api_common.Rule{
							Conf:     rulesForInbound[i].Conf,
							Matchers: oapi_helpers.SubsetToRuleMatcher(rulesForInbound[i].Subset),
							Origin:   oapi_helpers.ResourceMetaListToMetaList(res.Type, rulesForInbound[i].Origin),
						}
					}
					var tags map[string]string
					if dp.Spec.IsBuiltinGateway() || dp.Spec.IsDelegatedGateway() {
						tags = dp.Spec.Networking.Gateway.Tags
					} else {
						tags = dp.Spec.GetNetworking().GetInboundForPort(inbound.Port).Tags
					}
					fromRules = append(fromRules, api_common.FromRule{
						Inbound: api_common.Inbound{
							Tags: tags,
							Port: int(inbound.Port),
						},
						Rules: fromRulesForInbound,
					})
				}
			}
			toResourceRules := []api_common.ResourceRule{}
			for itemIdentifier, resourceRuleItem := range res.ToRules.ResourceRules {
				toResourceRules = append(toResourceRules, api_common.ResourceRule{
					Conf:                resourceRuleItem.Conf,
					Origin:              oapi_helpers.OriginListToResourceRuleOrigin(res.Type, resourceRuleItem.Origin),
					ResourceMeta:        oapi_helpers.ResourceMetaToMeta(itemIdentifier.ResourceType, resourceRuleItem.Resource),
					ResourceSectionName: &resourceRuleItem.ResourceSectionName,
				})
			}
			sort.Slice(toResourceRules, func(i, j int) bool {
				return toResourceRules[i].ResourceMeta.Name < toResourceRules[j].ResourceMeta.Name
			})

			if proxyRule == nil && len(fromRules) == 0 && len(toRules) == 0 && len(toResourceRules) == 0 {
				// No matches for this policy, keep going...
				continue
			}
			warnings := res.Warnings
			if warnings == nil {
				warnings = []string{}
			}
			rules = append(rules, api_common.InspectRule{
				Type:            string(res.Type),
				ToRules:         &toRules,
				ToResourceRules: &toResourceRules,
				FromRules:       &fromRules,
				ProxyRule:       proxyRule,
				Warnings:        &warnings,
			})
		}
		httpMatches := []api_common.HttpMatch{}
		for k, v := range matchesByHash {
			httpMatches = append(httpMatches, api_common.HttpMatch{
				Match: v,
				Hash:  string(k),
			})
		}
		sort.Slice(httpMatches, func(i, j int) bool {
			return httpMatches[i].Hash < httpMatches[j].Hash
		})
		out := api_types.InspectRulesResponse{
			HttpMatches: httpMatches,
			Resource:    oapi_helpers.ResourceToMeta(resource),
			Rules:       rules,
		}
		if err := response.WriteAsJson(out); err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Failed writing response")
		}
	}
}
