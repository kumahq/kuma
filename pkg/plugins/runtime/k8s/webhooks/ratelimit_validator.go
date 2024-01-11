package webhooks

import (
	"context"
	"net/http"

	v1 "k8s.io/api/admission/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	ratelimit_managers "github.com/kumahq/kuma/pkg/core/managers/apis/ratelimit"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/validators"
	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
)

func NewRateLimitValidatorWebhook(validator ratelimit_managers.RateLimitValidator, converter k8s_common.Converter) k8s_common.AdmissionValidator {
	return &RateLimitValidator{
		validator: validator,
		converter: converter,
	}
}

type RateLimitValidator struct {
	validator ratelimit_managers.RateLimitValidator
	converter k8s_common.Converter
	decoder   *admission.Decoder
}

func (h *RateLimitValidator) InjectDecoder(d *admission.Decoder) {
	h.decoder = d
}

func (h *RateLimitValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
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

func (h *RateLimitValidator) ValidateDelete(ctx context.Context, req admission.Request) admission.Response {
	if err := h.validator.ValidateDelete(ctx, req.Name); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	return admission.Allowed("")
}

func (h *RateLimitValidator) ValidateCreate(ctx context.Context, req admission.Request) admission.Response {
	coreRes := core_mesh.NewRateLimitResource()
	k8sRes := &mesh_k8s.RateLimit{}
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

func (h *RateLimitValidator) ValidateUpdate(ctx context.Context, req admission.Request) admission.Response {
	coreRes := core_mesh.NewRateLimitResource()
	k8sRes := &mesh_k8s.RateLimit{}
	if err := h.decoder.DecodeRaw(req.Object, k8sRes); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	if err := h.converter.ToCoreResource(k8sRes, coreRes); err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	oldCoreRes := core_mesh.NewRateLimitResource()
	oldK8sRes := &mesh_k8s.RateLimit{}
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

func (h *RateLimitValidator) Supports(req admission.Request) bool {
	gvk := mesh_k8s.GroupVersion.WithKind("RateLimit")
	return req.Kind.Kind == gvk.Kind && req.Kind.Version == gvk.Version && req.Kind.Group == gvk.Group
}
