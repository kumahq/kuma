package webhooks

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/kumahq/kuma/pkg/dns"
	"github.com/kumahq/kuma/pkg/dns/resolver"

	kube_core "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/validators"
)

// ServiceValidator validates Kuma-specific annotations on Services.
type ServiceValidator struct {
	decoder  *admission.Decoder
	Resolver resolver.DNSResolver
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

	// validate <port>.service.kuma.io/protocol annotation
	for _, svcPort := range svc.Spec.Ports {
		protocolAnnotation := fmt.Sprintf("%d.service.kuma.io/protocol", svcPort.Port)
		protocolAnnotationValue, exists := svc.Annotations[protocolAnnotation]
		if exists && mesh_core.ParseProtocol(protocolAnnotationValue) == mesh_core.ProtocolUnknown {
			verr.AddViolationAt(validators.RootedAt("metadata").Field("annotations").Key(protocolAnnotation),
				fmt.Sprintf("value %q is not valid. %s", protocolAnnotationValue, mesh_core.AllowedValuesHint(mesh_core.SupportedProtocols.Strings()...)))
		}
	}

	// validate ExternalName services
	if svc.Spec.Type == "ExternalName" && svc.Spec.ExternalName != "" {
		verr.Add(v.validateExternalName(svc))
	}

	return verr.OrNil()
}

func (v *ServiceValidator) validateExternalName(svc *kube_core.Service) validators.ValidationError {
	verr := validators.ValidationError{}
	externalName := svc.Spec.ExternalName
	if !strings.HasSuffix(externalName, v.Resolver.GetDomain()) {
		return validators.ValidationError{}
	}

	if strings.Contains(externalName, "_") {
		dnsName, err := dns.KumaCompliantToDnsName(externalName)
		if err != nil {
			verr.AddViolationAt(validators.RootedAt("service").Field("Spec").Field("ExternalName"), "unable to convert "+externalName)
		} else {
			verr.AddViolationAt(validators.RootedAt("service").Field("Spec").Field("ExternalName"), externalName+" is not a canonical DNS name, please consider using "+dnsName)

			_, err := v.Resolver.ForwardLookupFQDN(dnsName)
			if err != nil {
				verr.AddViolationAt(validators.RootedAt("service").Field("Spec").Field("ExternalName"),
					"unable to resolve "+dnsName+". Please check the format is <service-name>.<namespace>.svc.<port>.mesh")
			}
		}
	} else {
		kumaName, err := dns.DnsNameToKumaCompliant(externalName)
		if err != nil {
			verr.AddViolationAt(validators.RootedAt("service").Field("Spec").Field("ExternalName"), "unable to convert "+externalName)
		} else {
			_, err := v.Resolver.ForwardLookupFQDN(kumaName)
			if err != nil {
				verr.AddViolationAt(validators.RootedAt("service").Field("Spec").Key("ExternalName"),
					"unable to resolve "+externalName+". Please check the format is <service-name>.<namespace>.svc.<port>.mesh")
			}
		}
	}

	return verr
}

func (v *ServiceValidator) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
}
