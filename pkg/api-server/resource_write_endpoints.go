package api_server

import (
	"context"
	"fmt"
	"io"

	"github.com/emicklei/go-restful/v3"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	api_server_types "github.com/kumahq/kuma/v3/pkg/api-server/types"
	meshtrust_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshtrust/api/v1alpha1"
	resource_labels "github.com/kumahq/kuma/v3/pkg/core/resources/labels"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/core/resources/validator"
	"github.com/kumahq/kuma/v3/pkg/core/user"
	"github.com/kumahq/kuma/v3/pkg/core/validators"
)

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
	var previousLabels map[string]string
	resource := r.descriptor.NewObject()
	if err := r.resManager.Get(request.Request.Context(), resource, store.GetByKey(name, meshName)); err != nil && store.IsNotFound(err) {
		create = true
	} else if err != nil {
		return nil, withTitle(err, "Failed to find a resource")
	} else {
		previousLabels = resource.GetMeta().GetLabels()
	}

	if err := r.validateResourceRequest(name, meshName, resourceRest, previousLabels); err != nil {
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

func (r *resourceCrudHandler) applyBeforeWriteHook(resRest rest.Resource, meshName string, name string) {
	if hook, ok := beforeWriteHooks[r.descriptor.Name]; ok {
		hook(resRest, meshName, name)
	}
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

	r.applyBeforeWriteHook(resRest, meshName, name)

	res := r.descriptor.NewObject()
	_ = res.SetSpec(resRest.GetSpec())
	res.SetMeta(resRest.GetMeta())

	labels, err := resource_labels.Compute(resource_labels.Write{
		Descriptor:  r.descriptor,
		Spec:        res.GetSpec(),
		Namespace:   resource_labels.GetNamespace(res.GetMeta(), r.systemNamespace),
		Mesh:        meshName,
		DisplayName: name,
		Labels:      res.GetMeta().GetLabels(),
	}, r.cp)
	if err != nil {
		return nil, withTitle(err, "Could not compute labels for a resource")
	}

	if err := r.resManager.Create(ctx, res, store.CreateByKey(name, meshName), store.CreateWithLabels(labels)); err != nil {
		return nil, withTitle(err, "Failed to create a resource")
	}

	return created(api_server_types.CreateOrUpdateSuccessResponse{Warnings: core_model.Deprecations(res)}), nil
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

	r.applyBeforeWriteHook(newResRest, meshName, currentRes.GetMeta().GetName())

	newRes := r.descriptor.NewObject()
	_ = newRes.SetSpec(newResRest.GetSpec())
	newRes.SetMeta(currentRes.GetMeta())
	if err := validator.ValidateUpdate(currentRes, newRes); err != nil {
		return nil, withTitle(err, "Could not update a resource")
	}

	_ = currentRes.SetSpec(newResRest.GetSpec())

	labels, err := resource_labels.Compute(resource_labels.Write{
		Descriptor:  r.descriptor,
		Spec:        currentRes.GetSpec(),
		Namespace:   resource_labels.GetNamespace(newResRest.GetMeta(), r.systemNamespace),
		Mesh:        meshName,
		DisplayName: currentRes.GetMeta().GetName(),
		Labels:      newResRest.GetMeta().GetLabels(),
	}, r.cp)
	if err != nil {
		return nil, withTitle(err, "Could not compute labels for a resource")
	}

	if stored, ok := currentRes.GetMeta().GetLabels()[mesh_proto.ResourceOriginLabel]; ok && stored != labels[mesh_proto.ResourceOriginLabel] {
		var err validators.ValidationError
		err.AddViolationAt(
			validators.RootedAt("labels").Key(mesh_proto.ResourceOriginLabel),
			fmt.Sprintf("is immutable, cannot be changed from %q to %q", stored, labels[mesh_proto.ResourceOriginLabel]),
		)
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
