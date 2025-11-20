package v1alpha1

import (
	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
)

func (w *WorkloadResource) IsLocalWorkload() bool {
	if len(w.GetMeta().GetLabels()) == 0 {
		return true // no labels mean that it's a local resource
	}
	origin, ok := w.GetMeta().GetLabels()[mesh_proto.ResourceOriginLabel]
	if !ok {
		return true // no origin label means that it's a local resource
	}
	return origin == string(mesh_proto.ZoneResourceOrigin)
}
