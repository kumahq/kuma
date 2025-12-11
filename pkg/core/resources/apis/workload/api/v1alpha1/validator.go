package v1alpha1

import (
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/validators"
)

const maxNameLength = 253

func validateResource(r *model.ResStatus[*Workload, *WorkloadStatus]) error {
	var verr validators.ValidationError
	name := model.GetDisplayName(r.GetMeta())
	verr.Add(validators.ValidateLength(validators.RootedAt("name"), maxNameLength, name))

	return verr.OrNil()
}
