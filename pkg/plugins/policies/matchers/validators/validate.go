package validators

import (
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/validators"
)

// Validate does basic validation of the standard matchers using targetRef (madr-005)
func Validate(err validators.ValidationError, resource core_model.Resource) {
	// TODO @lobkovilya Implement standard matching strategy (also quite likely this resource type can be more specific)
}
