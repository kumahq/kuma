package webhooks

import (
	"context"
	"net/http"
	"reflect"

	v1 "k8s.io/api/admission/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/validators"
	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
)

func NewGatewayInstanceValidatorWebhook(converter k8s_common.Converter, resourceManager manager.ResourceManager) k8s_common.AdmissionValidator {
	return &GatewayInstanceValidator{
		converter:       converter,
		resourceManager: resourceManager,
	}
}

type GatewayInstanceValidator struct {
	converter       k8s_common.Converter
	decoder         *admission.Decoder
	resourceManager manager.ResourceManager
}

func (h *GatewayInstanceValidator) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

func (h *GatewayInstanceValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
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

func (h *GatewayInstanceValidator) ValidateDelete(ctx context.Context, req admission.Request) admission.Response {
	return admission.Allowed("")
}

func (h *GatewayInstanceValidator) ValidateCreate(ctx context.Context, req admission.Request) admission.Response {
	gatewayInstance := &mesh_k8s.MeshGatewayInstance{}
	if err := h.decoder.Decode(req, gatewayInstance); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if resp := h.validateTags(gatewayInstance); !resp.Allowed {
		return resp
	}

	return admission.Allowed("")
}

func (h *GatewayInstanceValidator) ValidateUpdate(ctx context.Context, req admission.Request) admission.Response {
	gatewayInstance := &mesh_k8s.MeshGatewayInstance{}
	if err := h.decoder.Decode(req, gatewayInstance); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if resp := h.validateTags(gatewayInstance); !resp.Allowed {
		return resp
	}

	return admission.Allowed("")
}

func (h *GatewayInstanceValidator) validateTags(gatewayInstance *mesh_k8s.MeshGatewayInstance) admission.Response {
	tags := gatewayInstance.Spec.Tags

	err := core_mesh.ValidateTags(validators.RootedAt("tags"), tags, core_mesh.ValidateTagsOpts{
		RequireService: true,
	})

	if err.HasViolations() {
		return convertValidationErrorOf(err, gatewayInstance, gatewayInstance.GetObjectMeta())
	}

	return admission.Allowed("")
}

func (h *GatewayInstanceValidator) Supports(req admission.Request) bool {
	gvk := mesh_k8s.GroupVersion.WithKind(reflect.TypeOf(mesh_k8s.MeshGatewayInstance{}).Name())
	return req.Kind.Kind == gvk.Kind && req.Kind.Version == gvk.Version && req.Kind.Group == gvk.Group
}
