package v1alpha1

import (
	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/validators"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

func (r *MeshMultiZoneServiceResource) validate() error {
	var verr validators.ValidationError
	if meta := r.GetMeta(); meta != nil {
		verr.Add(validators.ValidateRFC1035Name(validators.RootedAt("name"), core_model.GetDisplayName(meta)))
	}

	path := validators.RootedAt("spec")
	if len(pointer.Deref(r.Spec.Selector.MeshService.MatchLabels)) == 0 {
		verr.AddViolationAt(path.Field("selector").Field("meshService").Field("matchLabels"), "cannot be empty")
	}
	if len(r.Spec.Ports) == 0 {
		verr.AddViolationAt(path.Field("ports"), "cannot be empty")
	}
	for i, port := range r.Spec.Ports {
		if port.Name != nil && *port.Name == "" {
			verr.AddViolationAt(path.Field("ports").Index(i).Field("name"), validators.MustNotBeEmpty)
		}
		if !core_meta.SupportedProtocols.Contains(port.AppProtocol) {
			verr.AddViolationAt(path.Field("ports").Index(i).Field("appProtocol"), validators.MustBeOneOf("appProtocol", core_meta.SupportedProtocols.Strings()...))
		}
	}
	return verr.OrNil()
}
