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

	id := kri.WithSectionName(kri.From(t), t.Spec.Match.GetName())
	for _, err := range sni.ValidateKRI(id) {
		deprecations = append(deprecations, fmt.Sprintf(
			"Invalid %s SNI: %s. It does not conform to the DNS format (RFC 1123). This is deprecated.",
			MeshExternalServiceResourceTypeDescriptor.Name, err))
	}

	return deprecations
}
