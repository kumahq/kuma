package v1alpha1

import (
	"fmt"
	"strings"

	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"

	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
)

func (t *MeshMultiZoneServiceResource) Deprecations() []string {
	var deprecations []string

	name := model.GetDisplayName(t.GetMeta())
	allErrs := apimachineryvalidation.NameIsDNS1035Label(name, false)
	if len(allErrs) != 0 {
		nameDeprecationMsg := fmt.Sprintf("Invalid %s resource name: '%s'. It does not conform to the DNS format (RFC 1035). This is deprecated. Errors: %s",
			MeshMultiZoneServiceResourceTypeDescriptor.Name, name, strings.Join(allErrs, "; "))
		deprecations = append(deprecations, nameDeprecationMsg)
	}

	if len(name) > 63 {
		lengthDeprecationMsg := fmt.Sprintf("Invalid %s resource name: '%s'. Name must be no more than 63 characters long. This is deprecated and will be rejected in a future release.",
			MeshMultiZoneServiceResourceTypeDescriptor.Name, name)
		deprecations = append(deprecations, lengthDeprecationMsg)
	}

	return deprecations
}
