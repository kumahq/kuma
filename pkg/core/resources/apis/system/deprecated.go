package system

import (
	"fmt"
	"strings"

	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"

	"github.com/kumahq/kuma/pkg/core/resources/model"
)

func (t *ZoneResource) Deprecations() []string {
	var deprecations []string

	name := model.GetDisplayName(t.GetMeta())
	allErrs := apimachineryvalidation.NameIsDNS1035Label(name, false)
	if len(allErrs) != 0 {
		nameDeprecationMsg := fmt.Sprintf("Name that doesn't conform DNS (RFC 1035) format is deprecated, name %s, error: %s",
			name, strings.Join(allErrs, "; "))
		deprecations = append(deprecations, nameDeprecationMsg)
	}

	return deprecations
}
