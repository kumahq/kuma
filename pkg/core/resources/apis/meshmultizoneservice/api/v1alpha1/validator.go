package v1alpha1

import (
	"slices"

	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
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
	for i, port := range r.Spec.Ports {
		if port.Name != nil && *port.Name == "" {
			verr.AddViolationAt(path.Field("ports").Index(i).Field("name"), validators.MustNotBeEmpty)
		}
		if !slices.Contains(mesh.SupportedProtocols, port.AppProtocol) {
			verr.AddViolationAt(path.Field("ports").Index(i).Field("appProtocol"), validators.MustBeOneOf("appProtocol", mesh.SupportedProtocols.Strings()...))
		}
	}
	return verr.OrNil()
}
