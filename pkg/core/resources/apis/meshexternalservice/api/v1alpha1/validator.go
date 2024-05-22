package v1alpha1

import (
	"github.com/kumahq/kuma/pkg/core/validators"
)

func (r *MeshExternalServiceResource) validate() error {
	var verr validators.ValidationError
	return verr.OrNil()
}
