package webhooks

import (
	"context"
	"fmt"
	"net/http"

	kube_core "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/validators"
)

// ServiceValidator validates Kuma-specific annotations on Services.
type ServiceValidator struct {
	Decoder *admission.Decoder
}

// Handle admits a Service only if Kuma-specific annotations have proper values.
func (v *ServiceValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	svc := &kube_core.Service{}

	err := v.Decoder.Decode(req, svc)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if err := v.validate(svc); err != nil {
		if verr, ok := err.(*validators.ValidationError); ok {
			return convertValidationErrorOf(*verr, svc, svc)
		}
		return admission.Denied(err.Error())
	}

	return admission.Allowed("")
}

func (v *ServiceValidator) validate(svc *kube_core.Service) error {
	verr := &validators.ValidationError{}
	for _, svcPort := range svc.Spec.Ports {
		protocolAnnotation := fmt.Sprintf("%d.service.kuma.io/protocol", svcPort.Port)
		protocolAnnotationValue, exists := svc.Annotations[protocolAnnotation]
		if exists && core_mesh.ParseProtocol(protocolAnnotationValue) == core_mesh.ProtocolUnknown {
			verr.AddViolationAt(validators.RootedAt("metadata").Field("annotations").Key(protocolAnnotation),
				fmt.Sprintf("value %q is not valid. %s", protocolAnnotationValue, core_mesh.AllowedValuesHint(core_mesh.SupportedProtocols.Strings()...)))
		}
	}
	return verr.OrNil()
}
