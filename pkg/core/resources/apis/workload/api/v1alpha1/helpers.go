package v1alpha1

import (
	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
)

func (w *WorkloadResource) IsLocalWorkload() bool {
	origin, ok := w.GetMeta().GetLabels()[mesh_proto.ResourceOriginLabel]
	if !ok {
		return true // no origin label means local resource
	}
	return origin == string(mesh_proto.ZoneResourceOrigin)
}
