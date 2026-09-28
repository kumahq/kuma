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
	// Only Exclusive is a truthful value on 3.0; every other mode would
	// silently behave as Exclusive.
	if m.Spec.MeshServices != nil && m.Spec.GetMeshServices().GetMode() != mesh_proto.Mesh_MeshServices_Exclusive { //nolint:staticcheck // deprecated on purpose
		verr.AddViolation(
			"meshServices.mode",
			"removed in 3.0 and every mesh behaves as Exclusive; remove the field or set it to Exclusive",
		)
	}
	return verr.OrNil()
}
