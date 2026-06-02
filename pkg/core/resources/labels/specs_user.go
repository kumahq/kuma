package labels

import (
	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
)

// User-controlled labels. These are reserved by prefix but are user-settable
// flags. AllowedValues constrains the set of accepted values; an empty
// AllowedValues means any value is accepted.

func init() {
	register(LabelSpec{
		Key:           mesh_proto.EffectLabel,
		Owner:         OwnerUser,
		AllowedValues: []string{"", "shadow"},
	})

	register(LabelSpec{
		Key:           mesh_proto.KDSSyncLabel,
		Owner:         OwnerUser,
		AllowedValues: []string{"", "enabled", "disabled"},
	})
}
