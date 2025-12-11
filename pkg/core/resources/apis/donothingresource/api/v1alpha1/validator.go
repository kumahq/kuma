package v1alpha1

import (
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/validators"
)

func validateResource(r *core_model.Res[*DoNothingResource]) error {
	var verr validators.ValidationError
	_ = validators.RootedAt("spec")
	return verr.OrNil()
}
