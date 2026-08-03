package labels

import (
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
)

// expectedOrigin returns the kuma.io/origin value the control plane assigns
// based on the mode it runs in. Standalone is normalized to zone at CP startup.
func expectedOrigin(ctx ValidationContext) string {
	if ctx.Mode == config_core.Global {
		return string(mesh_proto.GlobalResourceOrigin)
	}
	return string(mesh_proto.ZoneResourceOrigin)
}
