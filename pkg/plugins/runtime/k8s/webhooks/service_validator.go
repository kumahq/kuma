package webhooks

import (
	"context"
	"fmt"
	"net/http"

	kube_core "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/validators"
)

// ServiceValidator validates Kuma-specific annotations on Services.
type ServiceValidator struct {
	decoder *admission.Decoder
}

// Handle admits a Service only if Kuma-specific annotations have proper values.
func (v *ServiceValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	svc := &kube_core.Service{}

	err := v.decoder.Decode(req, svc)
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
		if exists && mesh_core.ParseProtocol(protocolAnnotationValue) == mesh_core.ProtocolUnknown {
			verr.AddViolationAt(validators.RootedAt("metadata").Field("annotations").Key(protocolAnnotation),
				fmt.Sprintf("value %q is not valid. %s", protocolAnnotationValue, mesh_core.AllowedValuesHint(mesh_core.SupportedProtocols.Strings()...)))
		}
	}
	return verr.OrNil()
}

func (v *ServiceValidator) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
}
