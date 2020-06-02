package webhooks

import (
	"context"
	"fmt"
	"net/http"

	"k8s.io/api/admission/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	managers_mesh "github.com/Kong/kuma/pkg/core/managers/apis/mesh"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/core/validators"
	"github.com/Kong/kuma/pkg/plugins/resources/k8s"
	"github.com/Kong/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	mesh_k8s "github.com/Kong/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
)

func NewMeshValidatorWebhook(validator managers_mesh.MeshValidator, converter k8s.Converter, resourceManager manager.ResourceManager) AdmissionValidator {
	return &MeshValidator{
		validator:       validator,
		converter:       converter,
		resourceManager: resourceManager,
	}
}

type MeshValidator struct {
	validator       managers_mesh.MeshValidator
	converter       k8s.Converter
	decoder         *admission.Decoder
	resourceManager manager.ResourceManager
}

func (h *MeshValidator) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

func (h *MeshValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	switch req.Operation {
	case v1beta1.Delete:
		return h.ValidateDelete(ctx, req)
	case v1beta1.Create:
		return h.ValidateCreate(ctx, req)
	case v1beta1.Update:
		return h.ValidateUpdate(ctx, req)
	}
	return admission.Allowed("")
}

func (h *MeshValidator) ValidateDelete(ctx context.Context, req admission.Request) admission.Response {
	dps := mesh_core.DataplaneResourceList{}
	if err := h.resourceManager.List(ctx, &dps, store.ListByMesh(req.Name)); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	if len(dps.Items) != 0 {
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("unable to delete mesh, there are still some dataplanes attached"))
	}
	return admission.Allowed("")
}

func (h *MeshValidator) ValidateCreate(ctx context.Context, req admission.Request) admission.Response {
	coreRes := &mesh_core.MeshResource{}
	k8sRes := &v1alpha1.Mesh{}
	if err := h.decoder.Decode(req, k8sRes); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	if err := h.converter.ToCoreResource(k8sRes, coreRes); err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	if err := h.validator.ValidateCreate(ctx, req.Name, coreRes); err != nil {
		if kumaErr, ok := err.(*validators.ValidationError); ok {
			return convertSpecValidationError(kumaErr, k8sRes)
		}
		return admission.Denied(err.Error())
	}
	return admission.Allowed("")
}

func (h *MeshValidator) ValidateUpdate(ctx context.Context, req admission.Request) admission.Response {
	coreRes := &mesh_core.MeshResource{}
	k8sRes := &v1alpha1.Mesh{}
	if err := h.decoder.DecodeRaw(req.Object, k8sRes); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	if err := h.converter.ToCoreResource(k8sRes, coreRes); err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	oldCoreRes := &mesh_core.MeshResource{}
	oldK8sRes := &v1alpha1.Mesh{}
	if err := h.decoder.DecodeRaw(req.OldObject, oldK8sRes); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	if err := h.converter.ToCoreResource(oldK8sRes, oldCoreRes); err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	if err := h.validator.ValidateUpdate(ctx, oldCoreRes, coreRes); err != nil {
		if kumaErr, ok := err.(*validators.ValidationError); ok {
			return convertSpecValidationError(kumaErr, k8sRes)
		}
		return admission.Denied(err.Error())
	}
	return admission.Allowed("")
}

func (h *MeshValidator) Supports(req admission.Request) bool {
	gvk := mesh_k8s.GroupVersion.WithKind("Mesh")
	return req.Kind.Kind == gvk.Kind && req.Kind.Version == gvk.Version && req.Kind.Group == gvk.Group
}
