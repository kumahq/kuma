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
		Description:   "Per-policy effect modifier. 'shadow' makes the policy observable without changing DPP configs.",
		Owner:         OwnerUser,
		AllowedValues: []string{"", "shadow"},
	})

	register(LabelSpec{
		Key:           mesh_proto.KDSSyncLabel,
		Description:   "Controls KDS sync behavior for the resource.",
		Owner:         OwnerUser,
		AllowedValues: []string{"", "enabled", "disabled"},
	})
}
