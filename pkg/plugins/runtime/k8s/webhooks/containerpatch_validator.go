package webhooks

import (
	"context"
	"net/http"

	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	v1 "k8s.io/api/admission/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type ContainerPatchValidator struct {
	decoder *admission.Decoder
}

func NewContainerPatchValidatorWebhook() k8s_common.AdmissionValidator {
	return &ContainerPatchValidator{}
}

func (h *ContainerPatchValidator) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

func (h *ContainerPatchValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
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

func (h *ContainerPatchValidator) ValidateDelete(ctx context.Context, req admission.Request) admission.Response {
	return admission.Allowed("")
}

func validateSpec(containerPatch *mesh_k8s.ContainerPatch) error {
	for _, patch := range containerPatch.Spec.SidecarPatch {
		_, err := mesh_k8s.JsonPatchBlockToPatch(patch)
		if err != nil {
			return err
		}
	}

	for _, patch := range containerPatch.Spec.InitPatch {
		_, err := mesh_k8s.JsonPatchBlockToPatch(patch)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *ContainerPatchValidator) ValidateCreate(ctx context.Context, req admission.Request) admission.Response {
	k8sRes := &mesh_k8s.ContainerPatch{}
	if err := h.decoder.Decode(req, k8sRes); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if err := validateSpec(k8sRes); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	return admission.Allowed("")
}

func (h *ContainerPatchValidator) ValidateUpdate(ctx context.Context, req admission.Request) admission.Response {
	k8sRes := &mesh_k8s.ContainerPatch{}
	if err := h.decoder.DecodeRaw(req.Object, k8sRes); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	oldK8sRes := &mesh_k8s.ContainerPatch{}
	if err := h.decoder.DecodeRaw(req.OldObject, oldK8sRes); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if err := validateSpec(k8sRes); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	return admission.Allowed("")
}

func (h *ContainerPatchValidator) Supports(req admission.Request) bool {
	gvk := mesh_k8s.GroupVersion.WithKind("ContainerPatch")
	return req.Kind.Kind == gvk.Kind && req.Kind.Version == gvk.Version && req.Kind.Group == gvk.Group
}
