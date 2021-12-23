package webhooks

import (
	"context"
	"net/http"

	"k8s.io/api/admission/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/kumahq/kuma/pkg/core/managers/apis/dataplane"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/validators"
	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
)

func NewDataplaneValidatorWebhook(validator dataplane.Validator, converter k8s_common.Converter, resourceManager manager.ResourceManager) k8s_common.AdmissionValidator {
	return &DataplaneValidator{
		validator:       validator,
		converter:       converter,
		resourceManager: resourceManager,
	}
}

type DataplaneValidator struct {
	validator       dataplane.Validator
	converter       k8s_common.Converter
	decoder         *admission.Decoder
	resourceManager manager.ResourceManager
}

func (h *DataplaneValidator) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

func (h *DataplaneValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	switch req.Operation {
	case v1.Create:
		return h.ValidateCreate(ctx, req)
	case v1.Update:
		return h.ValidateUpdate(ctx, req)
	}
	return admission.Allowed("")
}

func (h *DataplaneValidator) ValidateCreate(ctx context.Context, req admission.Request) admission.Response {
	coreRes := core_mesh.NewDataplaneResource()
	k8sRes := &mesh_k8s.Dataplane{}
	if err := h.decoder.Decode(req, k8sRes); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	if err := h.converter.ToCoreResource(k8sRes, coreRes); err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	mesh := core_mesh.NewMeshResource()
	if err := h.resourceManager.Get(ctx, mesh, core_store.GetByKey(coreRes.GetMeta().GetMesh(), core_model.NoMesh)); err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	if err := h.validator.ValidateCreate(ctx, core_model.MetaToResourceKey(coreRes.GetMeta()), coreRes, mesh); err != nil {
		if kumaErr, ok := err.(*validators.ValidationError); ok {
			return convertSpecValidationError(kumaErr, k8sRes)
		}
		return admission.Denied(err.Error())
	}
	return admission.Allowed("")
}

func (h *DataplaneValidator) ValidateUpdate(ctx context.Context, req admission.Request) admission.Response {
	coreRes := core_mesh.NewDataplaneResource()
	k8sRes := &mesh_k8s.Dataplane{}
	if err := h.decoder.DecodeRaw(req.Object, k8sRes); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	if err := h.converter.ToCoreResource(k8sRes, coreRes); err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	oldCoreRes := core_mesh.NewDataplaneResource()
	oldK8sRes := &mesh_k8s.Dataplane{}
	if err := h.decoder.DecodeRaw(req.OldObject, oldK8sRes); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	if err := h.converter.ToCoreResource(oldK8sRes, oldCoreRes); err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	mesh := core_mesh.NewMeshResource()
	if err := h.resourceManager.Get(ctx, mesh, core_store.GetByKey(coreRes.GetMeta().GetMesh(), core_model.NoMesh)); err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	if err := h.validator.ValidateUpdate(ctx, coreRes, mesh); err != nil {
		if kumaErr, ok := err.(*validators.ValidationError); ok {
			return convertSpecValidationError(kumaErr, k8sRes)
		}
		return admission.Denied(err.Error())
	}
	return admission.Allowed("")
}

func (h *DataplaneValidator) Supports(req admission.Request) bool {
	gvk := mesh_k8s.GroupVersion.WithKind("Dataplane")
	return req.Kind.Kind == gvk.Kind && req.Kind.Version == gvk.Version && req.Kind.Group == gvk.Group
}
