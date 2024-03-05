package webhooks

import (
	"context"
	"net/http"

	v1 "k8s.io/api/admission/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	externalservice_managers "github.com/kumahq/kuma/pkg/core/managers/apis/external_service"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/validators"
	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
)

func NewExternalServiceValidatorWebhook(validator externalservice_managers.ExternalServiceValidator, converter k8s_common.Converter) k8s_common.AdmissionValidator {
	return &ExternalServiceValidator{
		validator: validator,
		converter: converter,
	}
}

type ExternalServiceValidator struct {
	validator externalservice_managers.ExternalServiceValidator
	converter k8s_common.Converter
	decoder   *admission.Decoder
}

func (h *ExternalServiceValidator) InjectDecoder(d *admission.Decoder) {
	h.decoder = d
}

func (h *ExternalServiceValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
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

func (h *ExternalServiceValidator) ValidateDelete(ctx context.Context, req admission.Request) admission.Response {
	if err := h.validator.ValidateDelete(ctx, req.Name); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	return admission.Allowed("")
}

func (h *ExternalServiceValidator) ValidateCreate(ctx context.Context, req admission.Request) admission.Response {
	coreRes := core_mesh.NewExternalServiceResource()
	k8sRes := &mesh_k8s.ExternalService{}
	if err := h.decoder.Decode(req, k8sRes); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	if err := h.converter.ToCoreResource(k8sRes, coreRes); err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	if err := h.validator.ValidateCreate(ctx, k8sRes.Mesh, coreRes); err != nil {
		if kumaErr, ok := err.(*validators.ValidationError); ok {
			return convertSpecValidationError(kumaErr, false, k8sRes)
		}
		return admission.Denied(err.Error())
	}
	return admission.Allowed("")
}

func (h *ExternalServiceValidator) ValidateUpdate(ctx context.Context, req admission.Request) admission.Response {
	coreRes := core_mesh.NewExternalServiceResource()
	k8sRes := &mesh_k8s.ExternalService{}
	if err := h.decoder.DecodeRaw(req.Object, k8sRes); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	if err := h.converter.ToCoreResource(k8sRes, coreRes); err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	oldCoreRes := core_mesh.NewExternalServiceResource()
	oldK8sRes := &mesh_k8s.ExternalService{}
	if err := h.decoder.DecodeRaw(req.OldObject, oldK8sRes); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	if err := h.converter.ToCoreResource(oldK8sRes, oldCoreRes); err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	if err := h.validator.ValidateUpdate(ctx, oldCoreRes, coreRes); err != nil {
		if kumaErr, ok := err.(*validators.ValidationError); ok {
			return convertSpecValidationError(kumaErr, false, k8sRes)
		}
		return admission.Denied(err.Error())
	}
	return admission.Allowed("")
}

func (h *ExternalServiceValidator) Supports(req admission.Request) bool {
	gvk := mesh_k8s.GroupVersion.WithKind("ExternalService")
	return req.Kind.Kind == gvk.Kind && req.Kind.Version == gvk.Version && req.Kind.Group == gvk.Group
}
