package api_server

import (
	"context"
	"fmt"
	"strconv"

	"github.com/emicklei/go-restful/v3"

	"github.com/kumahq/kuma/v3/pkg/api-server/filters"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	rest_errors "github.com/kumahq/kuma/v3/pkg/core/rest/errors"
	"github.com/kumahq/kuma/v3/pkg/core/user"
	"github.com/kumahq/kuma/v3/pkg/core/validators"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/k8s"
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
