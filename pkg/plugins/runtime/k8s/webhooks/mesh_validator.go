package webhooks

import (
	"context"
	"net/http"

	"k8s.io/api/admission/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	managers_mesh "github.com/kumahq/kuma/pkg/core/managers/apis/mesh"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/validators"
	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
)

func NewMeshValidatorWebhook(validator managers_mesh.MeshValidator, converter k8s_common.Converter) k8s_common.AdmissionValidator {
	return &MeshValidator{
		validator: validator,
		converter: converter,
	}
}

type MeshValidator struct {
	validator managers_mesh.MeshValidator
	converter k8s_common.Converter
	decoder   *admission.Decoder
}

func (h *MeshValidator) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

func (h *MeshValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	switch req.Operation {
	case v1.Delete:
		return h.ValidateDelete(ctx, req)
	case v1.Create:
		return h.ValidateCreate(ctx, req)
	case v1.Update:
		return h.ValidateUpdate(ctx, req)
	}
	return admission.Allowed("")
}

func (h *MeshValidator) ValidateDelete(ctx context.Context, req admission.Request) admission.Response {
	if err := h.validator.ValidateDelete(ctx, req.Name); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	return admission.Allowed("")
}

func (h *MeshValidator) ValidateCreate(ctx context.Context, req admission.Request) admission.Response {
	coreRes := core_mesh.NewMeshResource()
	k8sRes := &mesh_k8s.Mesh{}
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
	coreRes := core_mesh.NewMeshResource()
	k8sRes := &mesh_k8s.Mesh{}
	if err := h.decoder.DecodeRaw(req.Object, k8sRes); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	if err := h.converter.ToCoreResource(k8sRes, coreRes); err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	oldCoreRes := core_mesh.NewMeshResource()
	oldK8sRes := &mesh_k8s.Mesh{}
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
