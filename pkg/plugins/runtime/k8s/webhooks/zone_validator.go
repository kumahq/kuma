package webhooks

import (
	"context"
	"net/http"

	v1 "k8s.io/api/admission/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/kumahq/kuma/pkg/core/managers/apis/zone"
	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
)

func NewZoneValidatorWebhook(validator zone.Validator, unsafeDelete bool) k8s_common.AdmissionValidator {
	return &ZoneValidator{
		validator:    validator,
		unsafeDelete: unsafeDelete,
	}
}

type ZoneValidator struct {
	validator    zone.Validator
	unsafeDelete bool
}

func (z *ZoneValidator) InjectDecoder(admission.Decoder) {
}

func (z *ZoneValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	if req.Operation == v1.Delete {
		return z.ValidateDelete(ctx, req)
	}
	return admission.Allowed("")
}

func (z *ZoneValidator) ValidateDelete(ctx context.Context, req admission.Request) admission.Response {
	if !z.unsafeDelete {
		if err := z.validator.ValidateDelete(ctx, req.Name); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
	}
	return admission.Allowed("")
}

func (z *ZoneValidator) Supports(req admission.Request) bool {
	gvk := mesh_k8s.GroupVersion.WithKind("Zone")
	return req.Kind.Kind == gvk.Kind && req.Kind.Version == gvk.Version && req.Kind.Group == gvk.Group
}
