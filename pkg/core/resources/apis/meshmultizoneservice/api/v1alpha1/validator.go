package v1alpha1

import (
	"github.com/kumahq/kuma/pkg/core/validators"
)

func (r *MeshMultiZoneServiceResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")
	if len(r.Spec.Selector.MeshService.MatchLabels) == 0 {
		verr.AddViolationAt(path.Field("selector").Field("meshService").Field("matchLabels"), "cannot be empty")
	}
	if len(r.Spec.Ports) == 0 {
		verr.AddViolationAt(path.Field("ports"), "cannot be empty")
	}
	return verr.OrNil()
}
