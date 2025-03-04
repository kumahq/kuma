package v1alpha1

import (
	"fmt"
	"strings"

	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"
)

func (t *MeshServiceResource) Deprecations() []string {
	var deprecations []string

	allErrs := apimachineryvalidation.NameIsDNS1035Label(t.GetMeta().GetName(), false)
	if len(allErrs) != 0 {
		nameDeprecationMsg := fmt.Sprintf("Name that doesn't conform DNS (RFC 1035) format is deprecated: %s",
			strings.Join(allErrs, "; "))
		deprecations = append(deprecations, nameDeprecationMsg)
	}

	return deprecations
}
