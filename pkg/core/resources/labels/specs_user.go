package labels

import (
	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
)

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
