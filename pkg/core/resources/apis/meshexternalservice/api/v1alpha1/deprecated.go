package v1alpha1

import (
	"fmt"
	"strings"

	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"

	"github.com/kumahq/kuma/v2/pkg/core/kri"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/resources/sni"
)

func (t *MeshExternalServiceResource) Deprecations() []string {
	var deprecations []string

	name := model.GetDisplayName(t.GetMeta())
	if allErrs := apimachineryvalidation.NameIsDNS1035Label(name, false); len(allErrs) != 0 {
		deprecations = append(deprecations, fmt.Sprintf(
			"Invalid %s resource name: '%s'. It does not conform to the DNS format (RFC 1035). This is deprecated. Errors: %s",
			MeshExternalServiceResourceTypeDescriptor.Name, name, strings.Join(allErrs, "; ")))
	}

	portName := t.Spec.Match.GetName()
	id := kri.WithSectionName(kri.From(t), portName)
	for _, err := range sni.ValidateKRI(id) {
		deprecations = append(deprecations, fmt.Sprintf(
			"Invalid %s SNI (port %q): %s. This is deprecated.",
			MeshExternalServiceResourceTypeDescriptor.Name, portName, err))
	}

	if t.Spec.Tls != nil && t.Spec.Tls.Verification != nil {
		v := t.Spec.Tls.Verification
		for _, ds := range []struct {
			field  string
			source *VerificationDataSource
		}{{"caCert", v.CaCert}, {"clientCert", v.ClientCert}, {"clientKey", v.ClientKey}} {
			if ds.source != nil && ds.source.IsLegacy() {
				deprecations = append(deprecations, fmt.Sprintf(
					"'spec.tls.verification.%s' uses 'secret', 'inline' or 'inlineString', which are deprecated and no longer read in 3.0. Use 'type: Secret' with 'secretRef', or 'type: InsecureInline' with 'insecureInline.value' (plain text, not base64).",
					ds.field))
			}
		}
	}

	return deprecations
}
