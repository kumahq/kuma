package webhooks

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
)

type ContainerPatchValidator struct {
	SystemNamespace string
}

func NewContainerPatchValidatorWebhook() k8s_common.AdmissionValidator {
	return &ContainerPatchValidator{}
}

func (h *ContainerPatchValidator) InjectDecoder(d *admission.Decoder) {
}

func (h *ContainerPatchValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	if req.Namespace != h.SystemNamespace {
		return admission.Denied("ContainerPatch can only be placed in " + h.SystemNamespace + " namespace. It can be however referenced by pods in all namespaces")
	}
	return admission.Allowed("")
}

func (h *ContainerPatchValidator) Supports(req admission.Request) bool {
	gvk := mesh_k8s.GroupVersion.WithKind("ContainerPatch")
	return req.Kind.Kind == gvk.Kind && req.Kind.Version == gvk.Version && req.Kind.Group == gvk.Group
}
