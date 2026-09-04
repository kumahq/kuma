package api_server

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/emicklei/go-restful/v3"
	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/api-server/filters"
	api_server_types "github.com/kumahq/kuma/v3/pkg/api-server/types"
	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	meshtrust_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshtrust/api/v1alpha1"
	resource_labels "github.com/kumahq/kuma/v3/pkg/core/resources/labels"
	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	rest_errors "github.com/kumahq/kuma/v3/pkg/core/rest/errors"
	"github.com/kumahq/kuma/v3/pkg/core/user"
	"github.com/kumahq/kuma/v3/pkg/core/validators"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/k8s"
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/v3/pkg/util/maps"
)

// resourceCrudHandler serves the resource CRUD endpoints and their validation.
type resourceCrudHandler struct {
	resourceEndpointsContext

	federatedZone   bool
	k8sMapper       k8s.ResourceMapperFunc
	filter          func(request *restful.Request) (store.ListFilterFunc, error)
	systemNamespace string
	isK8s           bool

	disableOriginLabelValidation bool
}

// overviewForResource merges a resource with its insight. A missing insight is
// not an error: the overview is returned with an empty insight, like it is for
// a proxy that never connected.
func overviewForResource(
	ctx context.Context,
	resManager manager.ResourceManager,
	descriptor core_model.ResourceTypeDescriptor,
	resource core_model.Resource,
	name string,
	meshName string,
) (core_model.Resource, error) {
	insight := descriptor.NewInsight()
	if err := resManager.Get(ctx, insight, store.GetByKey(name, meshName)); err != nil && !store.IsNotFound(err) {
		return nil, err
	}
	overview, ok := descriptor.NewOverview().(core_model.OverviewResource)
	if !ok {
		return nil, fmt.Errorf("type withInsight for '%s' doesn't implement core_model.OverviewResource this shouldn't happen", descriptor.Name)
	}
	if err := overview.SetOverviewSpec(resource, insight); err != nil {
		return nil, err
	}
	return overview.(core_model.Resource), nil
}

func (r *resourceCrudHandler) findResource(withInsight bool) handlerFunc {
	return func(request *restful.Request) (any, error) {
		name := request.PathParameter("name")
		meshName, err := r.meshFromRequest(request)
		if err != nil {
			return nil, withTitle(err, "Failed to retrieve Mesh")
		}

		if err := r.resourceAccess.ValidateGet(
			request.Request.Context(),
			core_model.ResourceKey{Mesh: meshName, Name: name},
			r.descriptor,
			user.FromCtx(request.Request.Context()),
		); err != nil {
			return nil, withTitle(err, "Access Denied")
		}

		resource := r.descriptor.NewObject()
		if err := r.resManager.Get(request.Request.Context(), resource, store.GetByKey(name, meshName)); err != nil {
			return nil, withTitle(err, "Could not retrieve a resource")
		}
		if withInsight {
			resource, err = overviewForResource(request.Request.Context(), r.resManager, r.descriptor, resource, name, meshName)
			if err != nil {
				return nil, withTitle(err, "Could not retrieve insights")
			}
		}

		res, err := formatResource(resource, request.QueryParameter("format"), r.k8sMapper, request.QueryParameter("namespace"))
		if err != nil {
			return nil, withTitle(err, "Could not retrieve a resource")
		}
		return res, nil
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

func (r *resourceCrudHandler) listResources(withInsight bool) handlerFunc {
	return func(request *restful.Request) (any, error) {
		page, err := pagination(request)
		if err != nil {
			return nil, withTitle(err, "Could not retrieve resources")
		}
		filter, err := r.filter(request)
		if err != nil {
			return nil, withTitle(err, "Could not retrieve resources")
		}
		nameContains := request.QueryParameter("name")
		status, err := filters.Status(request)
		if err != nil {
			return nil, withTitle(err, "Could not retrieve resources")
		}
		if status != "" {
			if err := r.validateStatusFilter(withInsight); err != nil {
				return nil, withTitle(err, "Could not retrieve resources")
			}
		}

		meshName, err := r.meshFromRequest(request)
		if err != nil {
			return nil, withTitle(err, "Failed to retrieve Mesh")
		}

		if err := r.resourceAccess.ValidateList(
			request.Request.Context(),
			meshName,
			r.descriptor,
			user.FromCtx(request.Request.Context()),
		); err != nil {
			return nil, withTitle(err, "Access Denied")
		}
		list := r.descriptor.NewList()
		listOpts := []store.ListOptionsFunc{store.ListByMesh(meshName), store.ListByNameContains(nameContains), store.ListByFilterFunc(filter)}
		if status == "" {
			listOpts = append(listOpts, store.ListByPage(page.size, page.offset))
		}
		if err := r.resManager.List(request.Request.Context(), list, listOpts...); err != nil {
			return nil, withTitle(err, "Could not retrieve resources")
		}
		if withInsight {
			list, err = r.mergeInsights(request.Request.Context(), list, meshName, status, page)
			if err != nil {
				return nil, err
			}
		}
		restList := rest.From.ResourceList(list)
		restList.Next = nextLink(request, list.GetPagination().NextOffset)
		return restList, nil
	}
}

// mergeInsights merges the list with the insights of its items into overviews.
// Status lives on the insight, so when a status filter is set the page can only
// be cut once the overviews are merged: everything is listed and paginated here.
func (r *resourceCrudHandler) mergeInsights(ctx context.Context, list core_model.ResourceList, meshName string, status core_mesh.Status, page page) (core_model.ResourceList, error) {
	resourceKeys := make([]core_model.ResourceKey, 0, len(list.GetItems()))
	for _, item := range list.GetItems() {
		resourceKeys = append(resourceKeys, core_model.MetaToResourceKey(item.GetMeta()))
	}

	insights := r.descriptor.NewInsightList()
	if err := r.resManager.List(ctx, insights, store.ListByMesh(meshName), store.ListByResourceKeys(resourceKeys)); err != nil {
		return nil, withTitle(err, "Could not retrieve resources")
	}
	merged, err := r.MergeInOverview(list, insights)
	if err != nil {
		return nil, withTitle(err, "Failed merging overview and insights")
	}
	if status != "" {
		merged, err = r.pageByStatus(merged, status, page)
		if err != nil {
			return nil, withTitle(err, "Could not retrieve resources")
		}
	}
	return merged, nil
}

// statusAware is implemented by overviews exposing a status computed from their
// insight, which is what StatusFilterParam matches against.
type statusAware interface {
	Status() (core_mesh.Status, []string)
}

func (r *resourceCrudHandler) validateStatusFilter(withInsight bool) error {
	if !withInsight {
		return rest_errors.NewBadRequestError("filtering by status is only supported on the _overview endpoint")
	}
	if _, ok := r.descriptor.NewOverview().(statusAware); !ok {
		return rest_errors.NewBadRequestError(fmt.Sprintf("filtering by status is not supported for %s", r.descriptor.Name))
	}
	return nil
}

// pageByStatus keeps the overviews matching status and cuts the requested page
// out of them, mirroring what the store does for the other filters.
func (r *resourceCrudHandler) pageByStatus(list core_model.ResourceList, status core_mesh.Status, page page) (core_model.ResourceList, error) {
	var matching []core_model.Resource
	for _, item := range list.GetItems() {
		if itemStatus, _ := item.(statusAware).Status(); itemStatus == status {
			matching = append(matching, item)
		}
	}

	offset := 0
	if page.offset != "" {
		o, err := strconv.Atoi(page.offset)
		if err != nil {
			return nil, store.ErrorInvalid(fmt.Sprintf("invalid offset: %s", err.Error()))
		}
		if o < 0 {
			return nil, store.ErrorInvalid("invalid offset: must be non-negative")
		}
		offset = o
	}

	paged := r.descriptor.NewOverviewList()
	for i := offset; i < offset+page.size && i < len(matching); i++ {
		if err := paged.AddItem(matching[i]); err != nil {
			return nil, err
		}
	}
	nextOffset := ""
	if offset+page.size < len(matching) {
		nextOffset = strconv.Itoa(offset + page.size)
	}
	paged.GetPagination().SetNextOffset(nextOffset)
	paged.GetPagination().SetTotal(uint32(len(matching)))
	return paged, nil
}

func (r *resourceCrudHandler) MergeInOverview(resources core_model.ResourceList, insights core_model.ResourceList) (core_model.ResourceList, error) {
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

func (r *resourceCrudHandler) createOrUpdateResource(request *restful.Request) (any, error) {
	name := request.PathParameter("name")
	meshName, err := r.meshFromRequest(request)
	if err != nil {
		return nil, withTitle(err, "Failed to retrieve Mesh")
	}

	bodyBytes, err := io.ReadAll(request.Request.Body)
	if err != nil {
		return nil, withTitle(err, "Could not process a resource")
	}

	resourceRest, err := rest.JSON.Unmarshal(bodyBytes, r.descriptor)
	if err != nil {
		return nil, withTitle(err, "Could not process a resource")
	}

	create := false
	resource := r.descriptor.NewObject()
	if err := r.resManager.Get(request.Request.Context(), resource, store.GetByKey(name, meshName)); err != nil && store.IsNotFound(err) {
		create = true
	} else if err != nil {
		return nil, withTitle(err, "Failed to find a resource")
	}

	if err := r.validateResourceRequest(name, meshName, resourceRest); err != nil {
		return nil, withTitle(err, "Could not process a resource")
	}

	if create {
		return r.createResource(request.Request.Context(), name, meshName, resourceRest)
	}
	return r.updateResource(request.Request.Context(), resource, resourceRest, meshName)
}

// beforeWriteHooks mutate the REST representation of a resource of the given
// type before it is persisted on create or update.
var beforeWriteHooks = map[core_model.ResourceType]func(resRest rest.Resource, meshName string, name string){
	meshtrust_api.MeshTrustType: clearMeshTrustOrigin,
}

// validateCreateHooks run extra validation on create for the given resource
// type.
var validateCreateHooks = map[core_model.ResourceType]func(r *resourceCrudHandler, res core_model.Resource) error{
	core_mesh.DataplaneType: (*resourceCrudHandler).validateUniversalDataplaneWorkloadLabel,
}

func clearMeshTrustOrigin(resRest rest.Resource, meshName string, name string) {
	if resRest.GetStatus() != nil {
		status, ok := resRest.GetStatus().(*meshtrust_api.MeshTrustStatus)
		if ok && status != nil && status.Origin != nil {
			log.Info("ignoring status.origin as it is read-only", "mesh", meshName, "name", name)
			status.Origin = nil
		}
	}
}

// computeLabels derives the full label set for a resource from its descriptor,
// spec and meta, applying the control-plane mode, zone, k8s and namespace context
// shared by create and update.
func (r *resourceCrudHandler) computeLabels(
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

func (r *resourceCrudHandler) createResource(
	ctx context.Context,
	name string,
	meshName string,
	resRest rest.Resource,
) (any, error) {
	if err := r.resourceAccess.ValidateCreate(
		ctx,
		core_model.ResourceKey{Mesh: meshName, Name: name},
		resRest.GetSpec(),
		r.descriptor,
		user.FromCtx(ctx),
	); err != nil {
		return nil, withTitle(err, "Access Denied")
	}

	if hook, ok := beforeWriteHooks[r.descriptor.Name]; ok {
		hook(resRest, meshName, name)
	}

	res := r.descriptor.NewObject()
	_ = res.SetSpec(resRest.GetSpec())
	res.SetMeta(resRest.GetMeta())

	if hook, ok := validateCreateHooks[r.descriptor.Name]; ok {
		if err := hook(r, res); err != nil {
			return nil, err
		}
	}

	labels, err := r.computeLabels(res.Descriptor(), res.GetSpec(), res.GetMeta(), meshName, name)
	if err != nil {
		return nil, withTitle(err, "Could not compute labels for a resource")
	}

	if err := r.resManager.Create(ctx, res, store.CreateByKey(name, meshName), store.CreateWithLabels(labels)); err != nil {
		return nil, withTitle(err, "Failed to create a resource")
	}

	return created(api_server_types.CreateOrUpdateSuccessResponse{Warnings: core_model.Deprecations(res)}), nil
}

// validateUniversalDataplaneWorkloadLabel validates the workload label of
// Dataplanes created on Universal zones.
func (r *resourceCrudHandler) validateUniversalDataplaneWorkloadLabel(res core_model.Resource) error {
	if r.isK8s {
		return nil
	}
	workloadName, ok := res.GetMeta().GetLabels()[metadata.KumaWorkload]
	if !ok || workloadName == "" {
		return nil
	}
	if r.mode == config_core.Global {
		return withTitle(
			rest_errors.NewBadRequestError("labels[\"kuma.io/workload\"]: not allowed on Global control plane"),
			"Invalid workload label",
		)
	}
	if validationErrs := apimachineryvalidation.NameIsDNS1035Label(workloadName, false); len(validationErrs) != 0 {
		return withTitle(
			rest_errors.NewBadRequestError(fmt.Sprintf("labels[\"kuma.io/workload\"]: must be a valid DNS-1035 label (at most 63 characters, matching regex [a-z]([-a-z0-9]*[a-z0-9])?): %s", strings.Join(validationErrs, "; "))),
			"Invalid workload label",
		)
	}
	return nil
}

func (r *resourceCrudHandler) updateResource(
	ctx context.Context,
	currentRes core_model.Resource,
	newResRest rest.Resource,
	meshName string,
) (any, error) {
	if err := r.resourceAccess.ValidateUpdate(
		ctx,
		core_model.ResourceKey{Mesh: currentRes.GetMeta().GetMesh(), Name: currentRes.GetMeta().GetName()},
		currentRes.GetSpec(),
		newResRest.GetSpec(),
		r.descriptor,
		user.FromCtx(ctx),
	); err != nil {
		return nil, withTitle(err, "Access Denied")
	}

	if hook, ok := beforeWriteHooks[r.descriptor.Name]; ok {
		hook(newResRest, meshName, currentRes.GetMeta().GetName())
	}

	_ = currentRes.SetSpec(newResRest.GetSpec())

	labels, err := r.computeLabels(currentRes.Descriptor(), currentRes.GetSpec(), newResRest.GetMeta(), meshName, currentRes.GetMeta().GetName())
	if err != nil {
		return nil, withTitle(err, "Could not compute labels for a resource")
	}

	if validationErr := r.validateImmutableLabels(currentRes.GetMeta().GetLabels(), labels); validationErr.HasViolations() {
		var err validators.ValidationError
		err.AddError("labels", validationErr)
		return nil, withTitle(&err, "Could not update a resource")
	}

	if err := r.resManager.Update(ctx, currentRes, store.UpdateWithLabels(labels)); err != nil {
		return nil, withTitle(err, "Failed to update a resource")
	}

	return api_server_types.CreateOrUpdateSuccessResponse{Warnings: core_model.Deprecations(currentRes)}, nil
}

func (r *resourceCrudHandler) deleteResource(request *restful.Request) (any, error) {
	name := request.PathParameter("name")
	meshName, err := r.meshFromRequest(request)
	if err != nil {
		return nil, withTitle(err, "Failed to retrieve Mesh")
	}

	resource := r.descriptor.NewObject()

	if err := r.resManager.Get(request.Request.Context(), resource, store.GetByKey(name, meshName)); err != nil {
		return nil, withTitle(err, "Could not delete a resource")
	}

	if !core_model.IsLocallyOriginated(r.mode, resource.GetMeta().GetLabels()) {
		origin, _ := core_model.ResourceOrigin(resource.GetMeta())
		return nil, withTitle(rest_errors.NewBadRequestError(fmt.Sprintf("resource with %s=%s cannot be deleted on this control plane", mesh_proto.ResourceOriginLabel, origin)), "Could not delete a resource")
	}

	if verr := r.validateOriginForWrite(resource.GetMeta()); verr.HasViolations() {
		return nil, withTitle(verr.OrNil(), "Could not delete a resource")
	}

	if err := r.resourceAccess.ValidateDelete(
		request.Request.Context(),
		core_model.ResourceKey{Mesh: meshName, Name: name},
		resource.GetSpec(),
		resource.Descriptor(),
		user.FromCtx(request.Request.Context()),
	); err != nil {
		return nil, withTitle(err, "Access Denied")
	}

	if err := r.resManager.Delete(request.Request.Context(), resource, store.DeleteByKey(name, meshName)); err != nil {
		return nil, withTitle(err, "Could not delete a resource")
	}

	return api_server_types.DeleteSuccessResponse{}, nil
}

func (r *resourceCrudHandler) validateResourceRequest(name string, meshName string, resource rest.Resource) error {
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

func (r *resourceCrudHandler) validateOriginForWrite(meta core_model.ResourceMeta) validators.ValidationError {
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

func (r *resourceCrudHandler) validateLabels(resource rest.Resource) validators.ValidationError {
	var err validators.ValidationError

	origin, ok := core_model.ResourceOrigin(resource.GetMeta())
	if ok {
		if oerr := origin.IsValid(); oerr != nil {
			err.AddViolationAt(validators.Root().Key(mesh_proto.ResourceOriginLabel), oerr.Error())
		}
	}

	err.AddError("", r.validateOriginForWrite(resource.GetMeta()))

	zoneTag, hasZoneTag := resource.GetMeta().GetLabels()[mesh_proto.ZoneTag]
	if r.mode == config_core.Global {
		if hasZoneTag {
			err.AddViolationAt(validators.Root().Key(mesh_proto.ZoneTag), fmt.Sprintf("%s is not allowed on a global control plane", mesh_proto.ZoneTag))
		}
	} else if hasZoneTag && zoneTag != r.zoneName {
		err.AddViolationAt(validators.Root().Key(mesh_proto.ZoneTag), fmt.Sprintf("%s label should have %s value", mesh_proto.ZoneTag, r.zoneName))
	}
	if meshLabelValue, ok := resource.GetMeta().GetLabels()[mesh_proto.MeshTag]; ok && meshLabelValue != resource.GetMeta().GetMesh() {
		err.AddViolationAt(validators.Root().Key(mesh_proto.MeshTag), fmt.Sprintf("%s label must not differ from mesh set on resource", mesh_proto.MeshTag))
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

func (r *resourceCrudHandler) validateImmutableLabels(storedLabels, newComputedLabels map[string]string) validators.ValidationError {
	var err validators.ValidationError

	immutableLabels := []string{
		mesh_proto.ZoneTag,
		mesh_proto.ResourceOriginLabel,
	}

	for _, label := range immutableLabels {
		currentVal, currentExists := storedLabels[label]
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

func (r *resourceCrudHandler) validatePolicyRole(resource rest.Resource) validators.ValidationError {
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
func (r *resourceCrudHandler) doesNameLengthFitsGlobal(name string) bool {
	return len(fmt.Sprintf("%s.%s", r.zoneName, name)) < 253
}
