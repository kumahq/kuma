package v1alpha1

import (
	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/validators"
)

func (r *MeshServiceResource) validate() error {
	var verr validators.ValidationError

	if meta := r.GetMeta(); meta != nil {
		verr.Add(validators.ValidateRFC1035Name(validators.RootedAt("name"), model.GetDisplayName(meta)))
	}

	portsPath := validators.RootedAt("spec").Field("ports")
	for i, port := range r.Spec.Ports {
		if port.AppProtocol != "" && !core_meta.SupportedProtocols.Contains(core_meta.ParseProtocol(port.AppProtocol)) {
			verr.AddViolationAt(portsPath.Index(i).Field("appProtocol"), validators.MustBeOneOf("appProtocol", core_meta.SupportedProtocols.Strings()...))
		}
	}

	// Validate selector mutual exclusivity
	var setSelectors []bool
	if r.Spec.Selector.DataplaneRef != nil {
		setSelectors = append(setSelectors, true)
	}
	if r.Spec.Selector.DataplaneLabels != nil {
		setSelectors = append(setSelectors, true)
	}

	if len(setSelectors) > 1 {
		verr.AddViolationAt(validators.RootedAt("spec").Field("selector"), "must specify only one of: dataplaneRef or dataplaneLabels")
	}

	return verr.OrNil()
}
