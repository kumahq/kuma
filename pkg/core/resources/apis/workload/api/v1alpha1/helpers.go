package v1alpha1

import (
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
)

func (w *WorkloadResource) IsLocalWorkload() bool {
	origin, ok := w.GetMeta().GetLabels()[mesh_proto.ResourceOriginLabel]
	if !ok {
		return true // no origin label means local resource
	}
	return origin == string(mesh_proto.ZoneResourceOrigin)
}

func (w *WorkloadResource) Hash() []byte {
	// xDS currently depends on Workload identity/membership only. Future
	// xDS-relevant Workload fields must revisit this hash; re-adding version
	// or status would restore mesh-wide recomputes on status-only writes.
	return core_model.HashMetaIdentity(w)
}
