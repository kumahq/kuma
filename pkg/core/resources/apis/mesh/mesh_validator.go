package mesh

import (
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/validators"
)

func (m *MeshResource) Validate() error {
	var verr validators.ValidationError
	if meta := m.GetMeta(); meta != nil {
		verr.Add(validators.ValidateRFC1035Name(validators.RootedAt("name"), core_model.GetDisplayName(meta)))
	}
	// The deprecated meshServices field is still accepted for mixed-version
	// multizone upgrades, but only Exclusive is a truthful value: every other
	// mode would silently behave as Exclusive on 3.0
	// (https://github.com/kumahq/kuma/issues/18868).
	if m.Spec.MeshServices != nil && m.Spec.GetMeshServices().GetMode() != mesh_proto.Mesh_MeshServices_Exclusive { //nolint:staticcheck // deprecated on purpose
		verr.AddViolation(
			"meshServices.mode",
			"meshServices.mode was removed in 3.0 and every mesh behaves as Exclusive; remove the field or set it to Exclusive",
		)
	}
	return verr.OrNil()
}
