package api_server

import (
	"context"
	"io"

	"github.com/emicklei/go-restful/v3"

	api_server_types "github.com/kumahq/kuma/v3/pkg/api-server/types"
	meshtrust_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshtrust/api/v1alpha1"
	resource_labels "github.com/kumahq/kuma/v3/pkg/core/resources/labels"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
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

	r.applyBeforeWriteHook(resRest, meshName, name)

	res := r.descriptor.NewObject()
	_ = res.SetSpec(resRest.GetSpec())
	res.SetMeta(resRest.GetMeta())

	labels, err := r.computeLabels(res.Descriptor(), res.GetSpec(), res.GetMeta(), meshName, name)
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

	currentLabels, err := r.computeLabels(currentRes.Descriptor(), currentRes.GetSpec(), currentRes.GetMeta(), meshName, currentRes.GetMeta().GetName())
	if err != nil {
		return nil, withTitle(err, "Could not compute current labels")
	}

	_ = currentRes.SetSpec(newResRest.GetSpec())

	labels, err := r.computeLabels(currentRes.Descriptor(), currentRes.GetSpec(), newResRest.GetMeta(), meshName, currentRes.GetMeta().GetName())
	if err != nil {
		return nil, withTitle(err, "Could not compute labels for a resource")
	}

	if validationErr := r.validateImmutableLabels(currentLabels, labels); validationErr.HasViolations() {
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
