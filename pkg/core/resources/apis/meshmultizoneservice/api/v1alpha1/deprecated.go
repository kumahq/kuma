package v1alpha1

import (
	"fmt"
	"strings"

	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"

	"github.com/kumahq/kuma/v3/pkg/core/kri"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/sni"
)

func (t *MeshMultiZoneServiceResource) Deprecations() []string {
	var deprecations []string

	name := model.GetDisplayName(t.GetMeta())
	if allErrs := apimachineryvalidation.NameIsDNS1035Label(name, false); len(allErrs) != 0 {
		deprecations = append(deprecations, fmt.Sprintf(
			"Invalid %s resource name: '%s'. It does not conform to the DNS format (RFC 1035). This is deprecated. Errors: %s",
			MeshMultiZoneServiceResourceTypeDescriptor.Name, name, strings.Join(allErrs, "; ")))
	}

	base := kri.From(t)
	for _, port := range t.Spec.Ports {
		for _, err := range sni.ValidateKRI(kri.WithSectionName(base, port.GetName())) {
			deprecations = append(deprecations, fmt.Sprintf(
				"Invalid %s SNI (port %q): %s. This is deprecated.",
				MeshMultiZoneServiceResourceTypeDescriptor.Name, port.GetName(), err))
		}
	}

	return deprecations
}
